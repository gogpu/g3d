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
// GPU buffer lifecycle follows the gg/internal/gpu/render_session.go pattern:
// persistent buffers are created once, grown when needed, and updated each frame
// via queue.WriteBuffer(). This avoids the use-after-free bug where per-frame
// buffers are released before the GPU finishes executing the command buffer.
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

	// Persistent frame uniform buffer (group 0, binding 0). Created once,
	// updated each frame via queue.WriteBuffer(). Survives across frames.
	frameUniformBuf *wgpu.Buffer

	// Grow-only pools for per-object GPU resources.
	// Pools expand when more meshes are visible than the previous high-water mark.
	// Buffers are updated via queue.WriteBuffer(), never recreated per-frame.
	objectUniformBufs   []*wgpu.Buffer    // per-mesh model+normal matrix
	materialUniformBufs []*wgpu.Buffer    // per-mesh material properties
	objectBindGroups    []*wgpu.BindGroup // per-mesh bind group (group 1)

	// Cached geometry GPU buffers keyed by Geometry pointer identity.
	// Vertex/index data is static, so buffers are created on first use and
	// reused until the renderer is released.
	geomVertBufs  map[Geometry]*wgpu.Buffer
	geomIdxBufs   map[Geometry]*wgpu.Buffer
	geomIdxCounts map[Geometry]uint32

	// Frame bind groups (group 0) — one per distinct pipeline used in a frame.
	// Released at the start of the next frame, when the GPU has finished the
	// previous command buffer (VSync guarantees this for swapchain-based rendering).
	frameBindGroups []*wgpu.BindGroup

	// Bind groups pending release — deferred until after GPU completion.
	// WebGPU requires bind groups alive at submit time (wgpu-core track/mod.rs:631).
	// Same pattern as gg's pendingBindGroupRelease.
	pendingBindGroupRelease []*wgpu.BindGroup
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
// GPU buffers are persistent (created once, updated via queue.WriteBuffer).
// Pipeline creation is lazy and cached.
func (r *Renderer) Render(scene *Scene, camera Camera, targetView *wgpu.TextureView) error {
	if err := r.validateRenderState(); err != nil {
		return err
	}
	device := r.gpuState.Device()
	encoder, err := device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "g3d_frame"})
	if err != nil {
		return fmt.Errorf("g3d: create command encoder: %w", err)
	}

	if err := r.RenderTo(encoder, scene, camera, targetView); err != nil {
		encoder.DiscardEncoding()
		return err
	}

	commands, err := encoder.Finish()
	if err != nil {
		return fmt.Errorf("g3d: finish command encoder: %w", err)
	}
	if _, err := r.gpuState.Queue().Submit(commands); err != nil {
		return fmt.Errorf("g3d: submit commands: %w", err)
	}
	return nil
}

// RenderTo records the scene's render pass into a caller-owned command encoder.
// It does not finish, submit, or discard the encoder. This lets applications
// combine g3d with overlays and other renderers in one command buffer and one
// queue submission.
//
// The encoder and target view must belong to the renderer's device. The caller
// must keep them valid through submission and must not use the encoder
// concurrently.
func (r *Renderer) RenderTo(
	encoder *wgpu.CommandEncoder,
	scene *Scene,
	camera Camera,
	targetView *wgpu.TextureView,
) error {
	if encoder == nil {
		return fmt.Errorf("g3d: command encoder is nil")
	}
	if scene == nil {
		return fmt.Errorf("g3d: scene is nil")
	}
	if camera == nil {
		return fmt.Errorf("g3d: camera is nil")
	}
	if targetView == nil {
		return fmt.Errorf("g3d: target view is nil")
	}
	if err := r.validateRenderState(); err != nil {
		return err
	}

	// Release bind groups from the previous frame. By the time a new frame
	// begins, VSync guarantees the GPU has finished the previous command buffer.
	// This is the same deferred-release pattern as gg's pendingBindGroupRelease.
	r.releasePendingBindGroups()

	// Step 1-2: Update transforms, build render list.
	scene.UpdateWorldTransforms()
	cameraWorldPos := camera.CameraNode().WorldPosition()
	r.buildRenderList(scene, camera, cameraWorldPos)

	// Step 3: Collect lights, build frame uniform data.
	frameData := r.buildFrameUniforms(scene, camera, cameraWorldPos)

	// Step 4: Record draw calls. The encoder owner controls submission.
	return r.recordFrame(encoder, scene, targetView, &frameData)
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

	data.LightCount = lightIdx
	return data
}

// recordFrame uploads frame uniforms and records one render pass using
// persistent GPU buffers. Buffers are created once, grown when needed,
// and updated each frame via queue.WriteBuffer() -- never released
// before the GPU finishes the command buffer.
func (r *Renderer) recordFrame(
	encoder *wgpu.CommandEncoder,
	scene *Scene,
	targetView *wgpu.TextureView,
	frameData *gpu.FrameUniformsData,
) error {
	device := r.gpuState.Device()
	queue := r.gpuState.Queue()

	// Update the persistent frame uniform buffer via queue.WriteBuffer().
	// Buffer is created on first use and reused across frames.
	frameBytes := gpu.PackFrameUniforms(frameData)
	if err := r.ensureFrameUniformBuffer(device); err != nil {
		return fmt.Errorf("g3d: ensure frame uniform buffer: %w", err)
	}
	if err := queue.WriteBuffer(r.frameUniformBuf, 0, frameBytes); err != nil {
		return fmt.Errorf("g3d: write frame uniforms: %w", err)
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

	// Track the bind group pool index for this frame. The pool grows as needed
	// but bind groups are only released at the start of the NEXT frame.
	objPoolIdx := 0

	// Draw opaque bucket.
	if drawErr := r.drawBucket(renderPass, r.renderList.Opaque(), device, queue,
		&objPoolIdx); drawErr != nil {
		_ = renderPass.End()
		return fmt.Errorf("g3d: draw opaque: %w", drawErr)
	}

	if err = renderPass.End(); err != nil {
		return fmt.Errorf("g3d: end render pass: %w", err)
	}
	return nil
}

// drawBucket records draw commands for a sorted slice of draw calls.
// All GPU buffers are persistent -- created once, updated via queue.WriteBuffer().
func (r *Renderer) drawBucket(
	rp *wgpu.RenderPassEncoder,
	calls []render.DrawCall,
	device *wgpu.Device,
	queue *wgpu.Queue,
	objPoolIdx *int,
) error {
	var currentPipelineKey gpu.PipelineKey
	var currentBundle *gpu.PipelineBundle
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
			// Frame bind groups are deferred-released at the start of the next frame.
			fbg, err := device.CreateBindGroup(&wgpu.BindGroupDescriptor{
				Label:  "g3d_frame_bg",
				Layout: bundle.FrameLayout,
				Entries: []wgpu.BindGroupEntry{
					{Binding: 0, Buffer: r.frameUniformBuf, Size: gpu.FrameUniformsSize},
				},
			})
			if err != nil {
				return fmt.Errorf("create frame bind group: %w", err)
			}
			r.frameBindGroups = append(r.frameBindGroups, fbg)
			rp.SetBindGroup(0, fbg, nil)
		}

		if err := r.recordDrawCall(rp, device, queue, mesh, mat, geom,
			currentBundle, objPoolIdx); err != nil {
			return err
		}
	}

	return nil
}

// recordDrawCall uploads per-object uniforms and issues a single draw.
// Uses persistent buffer pools for uniforms and cached geometry buffers.
func (r *Renderer) recordDrawCall(
	rp *wgpu.RenderPassEncoder,
	device *wgpu.Device,
	queue *wgpu.Queue,
	mesh *Mesh,
	mat Material,
	geom Geometry,
	bundle *gpu.PipelineBundle,
	objPoolIdx *int,
) error {
	idx := *objPoolIdx
	*objPoolIdx++

	// Per-object uniform data (model + normal matrix).
	worldMat := mesh.node.WorldMatrix()
	normalMat := worldMat.Inverse().Transpose()

	objData := gpu.ObjectUniformsData{
		Model:       [16]float32(worldMat),
		NormalModel: [16]float32(normalMat),
	}
	objBytes := gpu.PackObjectUniforms(&objData)

	// Grow the object uniform buffer pool if needed.
	if err := r.ensureObjectPool(device, idx); err != nil {
		return fmt.Errorf("ensure object pool[%d]: %w", idx, err)
	}

	// Upload object uniforms via WriteBuffer (buffer persists across frames).
	if err := queue.WriteBuffer(r.objectUniformBufs[idx], 0, objBytes); err != nil {
		return fmt.Errorf("write object uniforms[%d]: %w", idx, err)
	}

	// Upload material uniforms via WriteBuffer.
	matBytes := mat.UniformData()
	if err := queue.WriteBuffer(r.materialUniformBufs[idx], 0, matBytes); err != nil {
		return fmt.Errorf("write material uniforms[%d]: %w", idx, err)
	}

	// Stale bind groups from the previous frame reference the same buffers with
	// different content. Create a new bind group each frame and defer its release
	// to the start of the next frame, when the GPU has finished.
	if idx < len(r.objectBindGroups) && r.objectBindGroups[idx] != nil {
		r.pendingBindGroupRelease = append(r.pendingBindGroupRelease, r.objectBindGroups[idx])
		r.objectBindGroups[idx] = nil
	}

	objBG, err := device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "g3d_object_bg",
		Layout: bundle.ObjectLayout,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, Buffer: r.objectUniformBufs[idx], Size: uint64(len(objBytes))},
			{Binding: 1, Buffer: r.materialUniformBufs[idx], Size: uint64(len(matBytes))},
		},
	})
	if err != nil {
		return fmt.Errorf("create object bind group: %w", err)
	}
	// Store in pool for deferred release next frame.
	for len(r.objectBindGroups) <= idx {
		r.objectBindGroups = append(r.objectBindGroups, nil)
	}
	r.objectBindGroups[idx] = objBG
	rp.SetBindGroup(1, objBG, nil)

	// Geometry buffers are cached by Geometry identity — vertex/index data is
	// static, so buffers are created on first use via MappedAtCreation and reused.
	vertBuf, err := r.ensureVertexBuffer(device, geom)
	if err != nil {
		return fmt.Errorf("create vertex buffer: %w", err)
	}

	if len(geom.Indices()) == 0 {
		vertexCount := geom.VertexCount()
		if vertexCount < 0 || uint64(vertexCount) > math.MaxUint32 {
			return fmt.Errorf("invalid vertex count %d", vertexCount)
		}
		rp.SetVertexBuffer(0, vertBuf, 0)
		rp.Draw(uint32(vertexCount), 1, 0, 0)
		return nil
	}

	idxBuf, idxCount, err := r.ensureIndexBuffer(device, geom)
	if err != nil {
		return fmt.Errorf("create index buffer: %w", err)
	}

	rp.SetVertexBuffer(0, vertBuf, 0)
	rp.SetIndexBuffer(idxBuf, gputypes.IndexFormatUint32, 0)
	rp.DrawIndexed(idxCount, 1, 0, 0, 0)

	return nil
}

// Release frees all GPU resources owned by the renderer.
// After this call the renderer cannot be used.
func (r *Renderer) Release() {
	// Release persistent uniform buffers and bind groups.
	r.releasePersistentResources()
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

// ensureFrameUniformBuffer creates the persistent frame uniform buffer on first
// use. The buffer is reused across frames and updated via queue.WriteBuffer().
func (r *Renderer) ensureFrameUniformBuffer(device *wgpu.Device) error {
	if r.frameUniformBuf != nil {
		return nil
	}
	buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "g3d_frame_uniforms",
		Size:  gpu.FrameUniformsSize,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}
	r.frameUniformBuf = buf
	return nil
}

// ensureObjectPool grows the per-object buffer pools (object uniforms,
// material uniforms) so that index idx is valid. New buffers are created
// with CopyDst usage for queue.WriteBuffer() updates.
func (r *Renderer) ensureObjectPool(device *wgpu.Device, idx int) error {
	for idx >= len(r.objectUniformBufs) {
		n := len(r.objectUniformBufs)
		buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
			Label: fmt.Sprintf("g3d_object_uniforms_%d", n),
			Size:  gpu.ObjectUniformsSize,
			Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		})
		if err != nil {
			return fmt.Errorf("create object uniform buffer[%d]: %w", n, err)
		}
		r.objectUniformBufs = append(r.objectUniformBufs, buf)
	}

	// Material uniform pool must match the object pool size.
	// Use 32 bytes (StandardMaterial size) as the max — BasicMaterial (16 bytes)
	// fits within 32 bytes with zero padding, and WriteBuffer only writes the
	// actual data length.
	const maxMaterialUniformSize = 32
	for idx >= len(r.materialUniformBufs) {
		n := len(r.materialUniformBufs)
		buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
			Label: fmt.Sprintf("g3d_material_uniforms_%d", n),
			Size:  maxMaterialUniformSize,
			Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		})
		if err != nil {
			return fmt.Errorf("create material uniform buffer[%d]: %w", n, err)
		}
		r.materialUniformBufs = append(r.materialUniformBufs, buf)
	}

	return nil
}

// ensureVertexBuffer returns a cached vertex buffer for the given geometry,
// creating it with MappedAtCreation on first use. Geometry data is static,
// so the buffer persists until the renderer is released.
func (r *Renderer) ensureVertexBuffer(device *wgpu.Device, geom Geometry) (*wgpu.Buffer, error) {
	if buf, ok := r.geomVertBufs[geom]; ok {
		return buf, nil
	}

	buf, err := createVertexBuffer(device, geom)
	if err != nil {
		return nil, err
	}

	if r.geomVertBufs == nil {
		r.geomVertBufs = make(map[Geometry]*wgpu.Buffer)
	}
	r.geomVertBufs[geom] = buf
	return buf, nil
}

// ensureIndexBuffer returns a cached index buffer for the given geometry,
// creating it with MappedAtCreation on first use.
func (r *Renderer) ensureIndexBuffer(device *wgpu.Device, geom Geometry) (*wgpu.Buffer, uint32, error) {
	if buf, ok := r.geomIdxBufs[geom]; ok {
		return buf, r.geomIdxCounts[geom], nil
	}

	buf, count, err := createIndexBuffer(device, geom)
	if err != nil {
		return nil, 0, err
	}

	if r.geomIdxBufs == nil {
		r.geomIdxBufs = make(map[Geometry]*wgpu.Buffer)
	}
	if r.geomIdxCounts == nil {
		r.geomIdxCounts = make(map[Geometry]uint32)
	}
	r.geomIdxBufs[geom] = buf
	r.geomIdxCounts[geom] = count
	return buf, count, nil
}

// releasePendingBindGroups frees bind groups from the previous frame.
// Called at the start of each new frame, when VSync guarantees the GPU
// has finished the previous command buffer.
func (r *Renderer) releasePendingBindGroups() {
	for _, bg := range r.pendingBindGroupRelease {
		bg.Release()
	}
	r.pendingBindGroupRelease = r.pendingBindGroupRelease[:0]

	// Frame bind groups are created per pipeline switch and also need
	// deferred release.
	for _, bg := range r.frameBindGroups {
		bg.Release()
	}
	r.frameBindGroups = r.frameBindGroups[:0]
}

// releasePersistentResources frees all persistent GPU resources owned by the
// renderer. Called from Release(). Order: bind groups, then buffers (bind
// groups reference buffers).
func (r *Renderer) releasePersistentResources() {
	// Release pending and current bind groups first.
	for _, bg := range r.pendingBindGroupRelease {
		bg.Release()
	}
	r.pendingBindGroupRelease = nil

	for _, bg := range r.frameBindGroups {
		bg.Release()
	}
	r.frameBindGroups = nil

	for _, bg := range r.objectBindGroups {
		if bg != nil {
			bg.Release()
		}
	}
	r.objectBindGroups = nil

	// Release uniform buffers.
	if r.frameUniformBuf != nil {
		r.frameUniformBuf.Release()
		r.frameUniformBuf = nil
	}
	for _, buf := range r.objectUniformBufs {
		buf.Release()
	}
	r.objectUniformBufs = nil

	for _, buf := range r.materialUniformBufs {
		buf.Release()
	}
	r.materialUniformBufs = nil

	// Release cached geometry buffers.
	for _, buf := range r.geomVertBufs {
		buf.Release()
	}
	r.geomVertBufs = nil

	for _, buf := range r.geomIdxBufs {
		buf.Release()
	}
	r.geomIdxBufs = nil
	r.geomIdxCounts = nil
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
