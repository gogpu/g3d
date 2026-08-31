package gpu

import (
	"fmt"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

// PipelineKey identifies a unique render pipeline configuration. Pipelines with
// the same key share a single GPU pipeline object.
//
// The key captures all state that affects pipeline compilation:
//   - ShaderID: which shader module ("basic" or "standard")
//   - SurfaceFormat: the color target format (e.g., BGRA8Unorm)
//   - DepthFormat: the depth attachment format (e.g., Depth24Plus)
//   - DoubleSided: whether back-face culling is disabled
type PipelineKey struct {
	ShaderID      string
	SurfaceFormat gputypes.TextureFormat
	DepthFormat   gputypes.TextureFormat
	DoubleSided   bool
}

// PipelineBundle holds a compiled render pipeline and its associated layouts.
// All resources in the bundle are released together via the PipelineCache.
type PipelineBundle struct {
	Pipeline       *wgpu.RenderPipeline
	FrameLayout    *wgpu.BindGroupLayout // group 0 — per-frame uniforms
	ObjectLayout   *wgpu.BindGroupLayout // group 1 — per-object + per-material uniforms
	PipelineLayout *wgpu.PipelineLayout
}

// PipelineCache creates and caches render pipelines keyed by PipelineKey.
// Pipelines are created on first request and reused for all subsequent frames.
type PipelineCache struct {
	device  *wgpu.Device
	shaders *ShaderCache
	cache   map[PipelineKey]*PipelineBundle
}

// NewPipelineCache creates a pipeline cache backed by the given device and shaders.
func NewPipelineCache(device *wgpu.Device, shaders *ShaderCache) *PipelineCache {
	return &PipelineCache{
		device:  device,
		shaders: shaders,
		cache:   make(map[PipelineKey]*PipelineBundle),
	}
}

// Get returns the cached pipeline bundle for the given key, creating it on first
// access. Returns an error if pipeline creation fails.
func (c *PipelineCache) Get(key PipelineKey) (*PipelineBundle, error) {
	if bundle, ok := c.cache[key]; ok {
		return bundle, nil
	}

	bundle, err := c.createPipeline(key)
	if err != nil {
		return nil, err
	}
	c.cache[key] = bundle
	return bundle, nil
}

// createPipeline compiles the render pipeline for the given key.
// Follows the gg/internal/gpu/convex_renderer.go:250-322 pattern.
func (c *PipelineCache) createPipeline(key PipelineKey) (*PipelineBundle, error) {
	shader := c.shaders.Get(key.ShaderID)
	if shader == nil {
		return nil, fmt.Errorf("g3d: unknown shader %q", key.ShaderID)
	}

	// Determine material uniform size from shader ID.
	// basic.wgsl MaterialUniforms = 16 bytes (vec4 color).
	// standard.wgsl MaterialUniforms = 32 bytes (vec4 color + metalness + roughness + alphaCutoff + pad).
	materialUniformSize := uint64(16)
	if key.ShaderID == ShaderStandard {
		materialUniformSize = 32
	}

	// Group 0: per-frame uniforms (FrameUniforms, 224 bytes).
	frameLayout, err := c.device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "g3d_frame_layout_" + key.ShaderID,
		Entries: []gputypes.BindGroupLayoutEntry{
			{
				Binding:    0,
				Visibility: gputypes.ShaderStageVertex | gputypes.ShaderStageFragment,
				Buffer:     &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: FrameUniformsSize},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("g3d: create frame bind group layout: %w", err)
	}

	// Group 1: per-object uniforms (ObjectUniforms binding 0, MaterialUniforms binding 1).
	objectLayout, err := c.device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "g3d_object_layout_" + key.ShaderID,
		Entries: []gputypes.BindGroupLayoutEntry{
			{
				Binding:    0,
				Visibility: gputypes.ShaderStageVertex,
				Buffer:     &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: ObjectUniformsSize},
			},
			{
				Binding:    1,
				Visibility: gputypes.ShaderStageFragment,
				Buffer:     &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: materialUniformSize},
			},
			{
				Binding:    2,
				Visibility: gputypes.ShaderStageFragment,
				Texture: &gputypes.TextureBindingLayout{
					SampleType:    gputypes.TextureSampleTypeFloat,
					ViewDimension: gputypes.TextureViewDimension2D,
				},
			},
			{
				Binding:    3,
				Visibility: gputypes.ShaderStageFragment,
				Sampler:    &gputypes.SamplerBindingLayout{Type: gputypes.SamplerBindingTypeFiltering},
			},
		},
	})
	if err != nil {
		frameLayout.Release()
		return nil, fmt.Errorf("g3d: create object bind group layout: %w", err)
	}

	pipeLayout, err := c.device.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "g3d_pipe_layout_" + key.ShaderID,
		BindGroupLayouts: []*wgpu.BindGroupLayout{frameLayout, objectLayout},
	})
	if err != nil {
		objectLayout.Release()
		frameLayout.Release()
		return nil, fmt.Errorf("g3d: create pipeline layout: %w", err)
	}

	pipeline, err := c.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "g3d_pipeline_" + key.ShaderID,
		Layout: pipeLayout,
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: shaderEntryVS,
			Buffers:    StandardVertexLayout(),
		},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: shaderEntryFS,
			Targets: []gputypes.ColorTargetState{
				{
					Format:    key.SurfaceFormat,
					WriteMask: gputypes.ColorWriteMaskAll,
				},
			},
		},
		Primitive:    DefaultPrimitive(key.DoubleSided),
		DepthStencil: OpaqueDepthStencil(key.DepthFormat),
		Multisample:  NoMultisample(),
	})
	if err != nil {
		pipeLayout.Release()
		objectLayout.Release()
		frameLayout.Release()
		return nil, fmt.Errorf("g3d: create render pipeline %q: %w", key.ShaderID, err)
	}

	return &PipelineBundle{
		Pipeline:       pipeline,
		FrameLayout:    frameLayout,
		ObjectLayout:   objectLayout,
		PipelineLayout: pipeLayout,
	}, nil
}

// Close releases all cached pipelines and layouts in reverse creation order.
func (c *PipelineCache) Close() {
	for key, bundle := range c.cache {
		bundle.Pipeline.Release()
		bundle.PipelineLayout.Release()
		bundle.ObjectLayout.Release()
		bundle.FrameLayout.Release()
		delete(c.cache, key)
	}
}
