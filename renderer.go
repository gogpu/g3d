package g3d

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/gogpu/g3d/internal/gpu"
	"github.com/gogpu/g3d/internal/render"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

// Renderer orchestrates the g3d forward rendering pipeline. It manages the
// scene traversal, frustum culling, render list sorting, GPU resource creation,
// and draw call submission via wgpu.
//
// Two construction modes:
//   - NewRenderer(provider) — shared device from gogpu app
//   - NewRendererFromDevice(device, queue, format) — standalone (tests, headless)
//
// The renderer is NOT thread-safe. Call Render from a single goroutine.
type Renderer struct {
	gpuState *gpu.GPUState

	renderList *render.RenderList

	depthTexture *wgpu.Texture
	depthView    *wgpu.TextureView
	depthFormat  gputypes.TextureFormat

	width, height uint32
	surfaceFormat gputypes.TextureFormat

	// meshTable stores meshes referenced by DrawCall.MeshIndex during a frame.
	// Reused across frames to avoid allocation.
	meshTable []*Mesh
}

// NewRenderer creates a Renderer using a shared GPU device from a DeviceProvider
// (e.g., gogpu.App). Returns an error if the provider has a software adapter
// or the device type is incompatible.
func NewRenderer(provider gpucontext.DeviceProvider) (*Renderer, error) {
	r := &Renderer{
		gpuState:    gpu.NewGPUState(),
		renderList:  render.NewRenderList(),
		depthFormat: gputypes.TextureFormatDepth24Plus,
	}

	if err := r.gpuState.SetDeviceProvider(provider); err != nil {
		return nil, fmt.Errorf("g3d: init renderer: %w", err)
	}

	r.surfaceFormat = provider.SurfaceFormat()
	if r.surfaceFormat == gputypes.TextureFormatUndefined {
		r.surfaceFormat = gputypes.TextureFormatBGRA8Unorm
	}

	return r, nil
}

// NewRendererFromDevice creates a Renderer from an explicit wgpu device and queue.
// Use this for standalone rendering, tests, or headless scenarios where there is
// no gogpu application framework.
func NewRendererFromDevice(
	device *wgpu.Device,
	queue *wgpu.Queue,
	surfaceFormat gputypes.TextureFormat,
) (*Renderer, error) {
	r := &Renderer{
		gpuState:      gpu.NewGPUState(),
		renderList:    render.NewRenderList(),
		depthFormat:   gputypes.TextureFormatDepth24Plus,
		surfaceFormat: surfaceFormat,
	}

	if err := r.gpuState.SetDeviceDirect(device, queue); err != nil {
		return nil, fmt.Errorf("g3d: init renderer: %w", err)
	}

	return r, nil
}

// SetSize sets the render target dimensions and (re)creates the depth texture.
// Must be called before the first Render call and whenever the viewport resizes.
func (r *Renderer) SetSize(width, height uint32) {
	if width == r.width && height == r.height {
		return
	}
	r.width = width
	r.height = height
	r.recreateDepthTexture()
}

// Render draws the scene from the camera's perspective into the given target view.
//
// The render flow follows the validated TASK-G3D-007 design:
//  1. Update world transforms (propagate dirty matrices).
//  2. Extract frustum from camera, build sorted render list.
//  3. Upload frame uniforms (ViewProjection, camera, lights).
//  4. Record draw calls and submit command buffer.
//
// Per-frame GPU buffers are created with MappedAtCreation, filled, unmapped,
// used, and released after submission. Pipeline creation is lazy and cached.
func (r *Renderer) Render(scene *Scene, camera Camera, targetView *wgpu.TextureView) error {
	if err := r.validateRenderState(); err != nil {
		return err
	}

	// Step 1-2: Update transforms, build render list.
	scene.UpdateWorldTransforms()
	cameraWorldPos := camera.CameraNode().WorldPosition()
	r.buildRenderList(scene, camera, cameraWorldPos)

	// Step 3: Collect lights, build frame uniform data.
	frameData := r.buildFrameUniforms(scene, camera, cameraWorldPos)

	// Step 4: Record draw calls and submit.
	return r.submitFrame(scene, targetView, &frameData)
}

// validateRenderState checks that the renderer is ready for a frame.
func (r *Renderer) validateRenderState() error {
	if !r.gpuState.IsReady() {
		return fmt.Errorf("g3d: GPU not initialized, call SetSize or ensure device is ready")
	}
	if r.width == 0 || r.height == 0 {
		return fmt.Errorf("g3d: render target size is zero, call SetSize first")
	}
	if r.depthView == nil {
		return fmt.Errorf("g3d: depth texture not created, call SetSize first")
	}
	return nil
}

// buildRenderList traverses the scene, frustum-culls meshes, and populates
// the sorted render list.
func (r *Renderer) buildRenderList(scene *Scene, camera Camera, cameraPos Vec3) {
	frustum := camera.Frustum()

	r.renderList.Clear()
	r.meshTable = r.meshTable[:0]

	scene.TraverseVisible(func(node *Node) {
		mesh := meshFromNode(node)
		if mesh == nil || mesh.geometry == nil || mesh.material == nil {
			return
		}

		worldAABB := mesh.WorldBoundingBox()
		if !frustum.IntersectsAABB(worldAABB) {
			return
		}

		center := worldAABB.Center()
		dx := center.X - cameraPos.X
		dy := center.Y - cameraPos.Y
		dz := center.Z - cameraPos.Z

		meshIdx := len(r.meshTable)
		r.meshTable = append(r.meshTable, mesh)

		r.renderList.Add(render.DrawCall{
			PipelineKey: mesh.material.ShaderID(),
			MaterialID:  materialID(mesh.material),
			Distance:    dx*dx + dy*dy + dz*dz,
			Bucket:      mapBucket(mesh.material.RenderBucket()),
			MeshIndex:   meshIdx,
		})
	})

	r.renderList.Sort()
}

// buildFrameUniforms collects camera and light data into the frame uniform struct.
func (r *Renderer) buildFrameUniforms(scene *Scene, camera Camera, cameraPos Vec3) gpu.FrameUniformsData {
	var data gpu.FrameUniformsData
	data.ViewProjection = camera.ViewProjectionMatrix()
	data.CameraPosition = [4]float32{cameraPos.X, cameraPos.Y, cameraPos.Z, 1.0}

	lightIdx := uint32(0)
	scene.TraverseVisible(func(node *Node) {
		if lightIdx >= gpu.MaxLights {
			return
		}
		light := lightFromNode(node)
		if light == nil {
			return
		}
		lu := light.LightUniform()
		data.Lights[lightIdx] = gpu.LightData{
			Direction: lu.Direction,
			LightType: lu.Kind,
			Color:     lu.Color,
			Intensity: lu.Intensity,
		}
		lightIdx++
	})

	collectAmbientLights(scene, &data, &lightIdx)
	data.LightCount = lightIdx
	return data
}

// submitFrame creates GPU resources, records draw commands, and submits the
// command buffer for one frame.
func (r *Renderer) submitFrame(
	scene *Scene,
	targetView *wgpu.TextureView,
	frameData *gpu.FrameUniformsData,
) error {
	device := r.gpuState.Device()
	queue := r.gpuState.Queue()

	// Create frame uniform buffer.
	frameBytes := gpu.PackFrameUniforms(frameData)
	frameBuf, err := createMappedBuffer(device, "g3d_frame_uniforms",
		uint64(len(frameBytes)), wgpu.BufferUsageUniform|wgpu.BufferUsageCopyDst, frameBytes)
	if err != nil {
		return fmt.Errorf("g3d: create frame uniform buffer: %w", err)
	}
	defer frameBuf.Release()

	// Create command encoder and begin render pass.
	encoder, err := device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "g3d_frame"})
	if err != nil {
		return fmt.Errorf("g3d: create command encoder: %w", err)
	}

	bg := scene.Background
	renderPass, err := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "g3d_forward",
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:       targetView,
				LoadOp:     gputypes.LoadOpClear,
				StoreOp:    gputypes.StoreOpStore,
				ClearValue: gputypes.Color{R: float64(bg.R), G: float64(bg.G), B: float64(bg.B), A: float64(bg.A)},
			},
		},
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:            r.depthView,
			DepthLoadOp:     gputypes.LoadOpClear,
			DepthStoreOp:    gputypes.StoreOpStore,
			DepthClearValue: 1.0,
			StencilLoadOp:   gputypes.LoadOpClear,
			StencilStoreOp:  gputypes.StoreOpDiscard,
		},
	})
	if err != nil {
		return fmt.Errorf("g3d: begin render pass: %w", err)
	}

	// Per-frame resources to release after submit.
	var perFrameBuffers []*wgpu.Buffer
	var perFrameBindGroups []*wgpu.BindGroup

	// Draw opaque bucket.
	if drawErr := r.drawBucket(renderPass, r.renderList.Opaque(), device, frameBuf,
		&perFrameBuffers, &perFrameBindGroups); drawErr != nil {
		_ = renderPass.End()
		releaseResources(perFrameBindGroups, perFrameBuffers)
		return fmt.Errorf("g3d: draw opaque: %w", drawErr)
	}

	if err = renderPass.End(); err != nil {
		releaseResources(perFrameBindGroups, perFrameBuffers)
		return fmt.Errorf("g3d: end render pass: %w", err)
	}

	commands, err := encoder.Finish()
	if err != nil {
		releaseResources(perFrameBindGroups, perFrameBuffers)
		return fmt.Errorf("g3d: finish command encoder: %w", err)
	}

	_, err = queue.Submit(commands)
	releaseResources(perFrameBindGroups, perFrameBuffers)
	if err != nil {
		return fmt.Errorf("g3d: submit commands: %w", err)
	}

	return nil
}

// drawBucket records draw commands for a sorted slice of draw calls.
func (r *Renderer) drawBucket(
	rp *wgpu.RenderPassEncoder,
	calls []render.DrawCall,
	device *wgpu.Device,
	frameBuf *wgpu.Buffer,
	perFrameBuffers *[]*wgpu.Buffer,
	perFrameBindGroups *[]*wgpu.BindGroup,
) error {
	var currentPipelineKey gpu.PipelineKey
	var currentBundle *gpu.PipelineBundle
	var frameBindGroup *wgpu.BindGroup
	pipelineSet := false

	for i := range calls {
		dc := &calls[i]
		mesh := r.meshTable[dc.MeshIndex]
		mat := mesh.material
		geom := mesh.geometry

		// Resolve pipeline.
		pipeKey := gpu.PipelineKey{
			ShaderID:      mat.ShaderID(),
			SurfaceFormat: r.surfaceFormat,
			DepthFormat:   r.depthFormat,
			DoubleSided:   mat.DoubleSided(),
		}

		if !pipelineSet || pipeKey != currentPipelineKey {
			bundle, err := r.gpuState.Pipelines().Get(pipeKey)
			if err != nil {
				return fmt.Errorf("get pipeline %q: %w", pipeKey.ShaderID, err)
			}
			currentBundle = bundle
			currentPipelineKey = pipeKey
			pipelineSet = true
			rp.SetPipeline(bundle.Pipeline)

			// Create frame bind group for this pipeline's layout.
			if frameBindGroup != nil {
				*perFrameBindGroups = append(*perFrameBindGroups, frameBindGroup)
			}
			fbg, err := device.CreateBindGroup(&wgpu.BindGroupDescriptor{
				Label:  "g3d_frame_bg",
				Layout: bundle.FrameLayout,
				Entries: []wgpu.BindGroupEntry{
					{Binding: 0, Buffer: frameBuf, Size: gpu.FrameUniformsSize},
				},
			})
			if err != nil {
				return fmt.Errorf("create frame bind group: %w", err)
			}
			frameBindGroup = fbg
			rp.SetBindGroup(0, frameBindGroup, nil)
		}

		if err := r.recordDrawCall(rp, device, mesh, mat, geom,
			currentBundle, perFrameBuffers, perFrameBindGroups); err != nil {
			return err
		}
	}

	if frameBindGroup != nil {
		*perFrameBindGroups = append(*perFrameBindGroups, frameBindGroup)
	}
	return nil
}

// recordDrawCall uploads per-object uniforms and issues a single draw.
func (r *Renderer) recordDrawCall(
	rp *wgpu.RenderPassEncoder,
	device *wgpu.Device,
	mesh *Mesh,
	mat Material,
	geom Geometry,
	bundle *gpu.PipelineBundle,
	perFrameBuffers *[]*wgpu.Buffer,
	perFrameBindGroups *[]*wgpu.BindGroup,
) error {
	// Per-object uniform buffer (model + normal matrix).
	worldMat := mesh.node.WorldMatrix()
	normalMat := worldMat.Inverse().Transpose()

	objData := gpu.ObjectUniformsData{
		Model:       [16]float32(worldMat),
		NormalModel: [16]float32(normalMat),
	}
	objBytes := gpu.PackObjectUniforms(&objData)

	objBuf, err := createMappedBuffer(device, "g3d_object_uniforms",
		uint64(len(objBytes)), wgpu.BufferUsageUniform|wgpu.BufferUsageCopyDst, objBytes)
	if err != nil {
		return fmt.Errorf("create object uniform buffer: %w", err)
	}
	*perFrameBuffers = append(*perFrameBuffers, objBuf)

	// Per-material uniform buffer.
	matBytes := mat.UniformData()
	matBuf, err := createMappedBuffer(device, "g3d_material_uniforms",
		uint64(len(matBytes)), wgpu.BufferUsageUniform|wgpu.BufferUsageCopyDst, matBytes)
	if err != nil {
		return fmt.Errorf("create material uniform buffer: %w", err)
	}
	*perFrameBuffers = append(*perFrameBuffers, matBuf)

	// Object bind group (group 1: object + material).
	objBG, err := device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "g3d_object_bg",
		Layout: bundle.ObjectLayout,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, Buffer: objBuf, Size: uint64(len(objBytes))},
			{Binding: 1, Buffer: matBuf, Size: uint64(len(matBytes))},
		},
	})
	if err != nil {
		return fmt.Errorf("create object bind group: %w", err)
	}
	*perFrameBindGroups = append(*perFrameBindGroups, objBG)
	rp.SetBindGroup(1, objBG, nil)

	// Vertex and index buffers.
	vertBuf, err := createVertexBuffer(device, geom)
	if err != nil {
		return fmt.Errorf("create vertex buffer: %w", err)
	}
	*perFrameBuffers = append(*perFrameBuffers, vertBuf)

	idxBuf, idxCount, err := createIndexBuffer(device, geom)
	if err != nil {
		return fmt.Errorf("create index buffer: %w", err)
	}
	*perFrameBuffers = append(*perFrameBuffers, idxBuf)

	rp.SetVertexBuffer(0, vertBuf, 0)
	rp.SetIndexBuffer(idxBuf, gputypes.IndexFormatUint32, 0)
	rp.DrawIndexed(idxCount, 1, 0, 0, 0)

	return nil
}

// Release frees all GPU resources owned by the renderer.
// After this call the renderer cannot be used.
func (r *Renderer) Release() {
	r.releaseDepthTexture()
	r.gpuState.Close()
}

// recreateDepthTexture destroys the old depth texture and creates a new one
// at the current width/height.
func (r *Renderer) recreateDepthTexture() {
	r.releaseDepthTexture()

	if r.width == 0 || r.height == 0 {
		return
	}
	if !r.gpuState.IsReady() {
		return
	}

	device := r.gpuState.Device()

	tex, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label: "g3d_depth",
		Size: wgpu.Extent3D{
			Width:              r.width,
			Height:             r.height,
			DepthOrArrayLayers: 1,
		},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        r.depthFormat,
		Usage:         gputypes.TextureUsageRenderAttachment,
	})
	if err != nil {
		return
	}

	view, err := device.CreateTextureView(tex, nil)
	if err != nil {
		tex.Release()
		return
	}

	r.depthTexture = tex
	r.depthView = view
}

// releaseDepthTexture frees depth texture resources.
func (r *Renderer) releaseDepthTexture() {
	if r.depthView != nil {
		r.depthView.Release()
		r.depthView = nil
	}
	if r.depthTexture != nil {
		r.depthTexture.Release()
		r.depthTexture = nil
	}
}

// releaseResources frees per-frame bind groups and buffers after submit.
func releaseResources(bindGroups []*wgpu.BindGroup, buffers []*wgpu.Buffer) {
	for _, bg := range bindGroups {
		bg.Release()
	}
	for _, buf := range buffers {
		buf.Release()
	}
}

// createMappedBuffer creates a GPU buffer with MappedAtCreation, copies data
// into it via MappedRange, and unmaps it. The buffer is ready for use after
// this function returns.
func createMappedBuffer(
	device *wgpu.Device, label string, size uint64, usage wgpu.BufferUsage, data []byte,
) (*wgpu.Buffer, error) {
	buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label:            label,
		Size:             size,
		Usage:            usage,
		MappedAtCreation: true,
	})
	if err != nil {
		return nil, err
	}

	mapped, err := buf.MappedRange(0, size)
	if err != nil {
		buf.Release()
		return nil, err
	}

	copy(mapped.Bytes(), data)

	if err := buf.Unmap(); err != nil {
		buf.Release()
		return nil, err
	}

	return buf, nil
}

// createVertexBuffer creates a vertex buffer from geometry data using
// MappedAtCreation for zero-copy upload.
func createVertexBuffer(device *wgpu.Device, geom Geometry) (*wgpu.Buffer, error) {
	verts := geom.Vertices()
	size := uint64(len(verts)) * 4 // float32 = 4 bytes
	if size == 0 {
		return nil, fmt.Errorf("geometry has no vertices")
	}

	buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label:            "g3d_vertex",
		Size:             size,
		Usage:            wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
		MappedAtCreation: true,
	})
	if err != nil {
		return nil, err
	}

	mapped, err := buf.MappedRange(0, size)
	if err != nil {
		buf.Release()
		return nil, err
	}

	dst := mapped.Bytes()
	for i, v := range verts {
		binary.LittleEndian.PutUint32(dst[i*4:(i+1)*4], math.Float32bits(v))
	}

	if err := buf.Unmap(); err != nil {
		buf.Release()
		return nil, err
	}

	return buf, nil
}

// createIndexBuffer creates an index buffer from geometry index data.
// Returns the buffer and the index count as uint32.
func createIndexBuffer(device *wgpu.Device, geom Geometry) (*wgpu.Buffer, uint32, error) {
	indices := geom.Indices()
	count := len(indices)
	if count == 0 {
		return nil, 0, fmt.Errorf("geometry has no indices")
	}

	size := uint64(count) * 4 // uint32 = 4 bytes
	buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label:            "g3d_index",
		Size:             size,
		Usage:            wgpu.BufferUsageIndex | wgpu.BufferUsageCopyDst,
		MappedAtCreation: true,
	})
	if err != nil {
		return nil, 0, err
	}

	mapped, err := buf.MappedRange(0, size)
	if err != nil {
		buf.Release()
		return nil, 0, err
	}

	dst := mapped.Bytes()
	for i, idx := range indices {
		binary.LittleEndian.PutUint32(dst[i*4:(i+1)*4], idx)
	}

	if err := buf.Unmap(); err != nil {
		buf.Release()
		return nil, 0, err
	}

	return buf, uint32(count), nil //nolint:gosec // index count fits in uint32 (max ~4B indices)
}

// meshFromNode checks whether a Node's UserData holds a *Mesh. This is how
// the renderer discovers meshes during scene traversal. Meshes set themselves
// as UserData on their node.
func meshFromNode(node *Node) *Mesh {
	m, _ := node.UserData().(*Mesh)
	return m
}

// lightWithUniform is implemented by light types that can produce a LightUniform.
type lightWithUniform interface {
	LightUniform() LightUniform
}

// lightFromNode checks whether a Node's UserData holds a light with LightUniform support.
func lightFromNode(node *Node) lightWithUniform {
	l, _ := node.UserData().(lightWithUniform)
	return l
}

// collectAmbientLights scans the scene's direct children for AmbientLight instances
// that are stored as UserData. Ambient lights don't have their own nodes in some
// usage patterns, so we also accept them directly.
func collectAmbientLights(scene *Scene, data *gpu.FrameUniformsData, lightIdx *uint32) {
	for _, child := range scene.Children() {
		if *lightIdx >= gpu.MaxLights {
			return
		}
		if amb, ok := child.UserData().(*AmbientLight); ok {
			lu := amb.LightUniform()
			data.Lights[*lightIdx] = gpu.LightData{
				Direction: lu.Direction,
				LightType: lu.Kind,
				Color:     lu.Color,
				Intensity: lu.Intensity,
			}
			*lightIdx++
		}
	}
}

// mapBucket converts a g3d RenderBucket to an internal render.RenderBucket.
func mapBucket(b RenderBucket) render.RenderBucket {
	switch b {
	case RenderBucketBackground:
		return render.BucketBackground
	case RenderBucketOpaque:
		return render.BucketOpaque
	case RenderBucketTransmissive:
		return render.BucketTransmissive
	case RenderBucketTransparent:
		return render.BucketTransparent
	default:
		return render.BucketOpaque
	}
}

// materialID generates a stable identifier for a material instance. This is
// used for sort-key grouping — materials at the same address share bind groups.
//
// The ID is derived from the interface's data pointer, which is stable for a
// given pointer-based concrete type (*BasicMaterial, *StandardMaterial, etc.).
func materialID(mat Material) uint64 {
	// An interface value is {type, data} — we extract the data pointer.
	// This works because all Material implementations are pointer types.
	type iface struct {
		_    uintptr // type pointer
		data uintptr // data pointer
	}
	return uint64((*iface)(unsafe.Pointer(&mat)).data) //nolint:gosec // intentional: extract interface data pointer for stable material identity
}
