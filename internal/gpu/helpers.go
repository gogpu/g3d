package gpu

import (
	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

const (
	shaderEntryVS = "vs_main"
	shaderEntryFS = "fs_main"
)

// OpaqueDepthStencil returns a DepthStencilState configured for standard opaque
// 3D rendering: depth test enabled (Less), depth write enabled, no stencil.
//
// g3d uses Depth24Plus (no stencil) unlike gg which uses Depth24PlusStencil8
// for stencil-then-cover path clipping. 3D opaque rendering does not need stencil.
func OpaqueDepthStencil(depthFormat gputypes.TextureFormat) *wgpu.DepthStencilState {
	return &wgpu.DepthStencilState{
		Format:            depthFormat,
		DepthWriteEnabled: true,
		DepthCompare:      gputypes.CompareFunctionLess,
		StencilFront: wgpu.StencilFaceState{
			Compare:     gputypes.CompareFunctionAlways,
			FailOp:      wgpu.StencilOperationKeep,
			DepthFailOp: wgpu.StencilOperationKeep,
			PassOp:      wgpu.StencilOperationKeep,
		},
		StencilBack: wgpu.StencilFaceState{
			Compare:     gputypes.CompareFunctionAlways,
			FailOp:      wgpu.StencilOperationKeep,
			DepthFailOp: wgpu.StencilOperationKeep,
			PassOp:      wgpu.StencilOperationKeep,
		},
		StencilReadMask:  0x00,
		StencilWriteMask: 0x00,
	}
}

// DefaultPrimitive returns a PrimitiveState for triangle list rendering.
// When doubleSided is true, culling is disabled; otherwise back-face culling
// is enabled with counter-clockwise front face (WebGPU default).
func DefaultPrimitive(doubleSided bool) gputypes.PrimitiveState {
	cullMode := gputypes.CullModeBack
	if doubleSided {
		cullMode = gputypes.CullModeNone
	}
	return gputypes.PrimitiveState{
		Topology: gputypes.PrimitiveTopologyTriangleList,
		CullMode: cullMode,
	}
}

// NoMultisample returns a MultisampleState with count=1 (no MSAA).
// Phase 1 does not use MSAA; it can be added in Phase 2.
func NoMultisample() gputypes.MultisampleState {
	return gputypes.MultisampleState{
		Count: 1,
		Mask:  0xFFFFFFFF,
	}
}

// StandardVertexLayout returns the vertex buffer layout for the standard
// 3D vertex format: position(vec3, loc 0) + normal(vec3, loc 1) + uv(vec2, loc 2).
// Stride: 32 bytes.
func StandardVertexLayout() []gputypes.VertexBufferLayout {
	return []gputypes.VertexBufferLayout{
		{
			ArrayStride: 32,
			StepMode:    gputypes.VertexStepModeVertex,
			Attributes: []gputypes.VertexAttribute{
				{Format: gputypes.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 0},  // position
				{Format: gputypes.VertexFormatFloat32x3, Offset: 12, ShaderLocation: 1}, // normal
				{Format: gputypes.VertexFormatFloat32x2, Offset: 24, ShaderLocation: 2}, // uv
			},
		},
	}
}
