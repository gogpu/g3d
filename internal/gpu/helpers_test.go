package gpu

import (
	"testing"

	"github.com/gogpu/gputypes"
)

func TestOpaqueDepthStencilFormat(t *testing.T) {
	ds := OpaqueDepthStencil(gputypes.TextureFormatDepth24Plus)
	if ds.Format != gputypes.TextureFormatDepth24Plus {
		t.Errorf("Format = %v, want Depth24Plus", ds.Format)
	}
}

func TestOpaqueDepthStencilWriteEnabled(t *testing.T) {
	ds := OpaqueDepthStencil(gputypes.TextureFormatDepth24Plus)
	if !ds.DepthWriteEnabled {
		t.Error("DepthWriteEnabled should be true for opaque rendering")
	}
}

func TestOpaqueDepthStencilCompare(t *testing.T) {
	ds := OpaqueDepthStencil(gputypes.TextureFormatDepth24Plus)
	if ds.DepthCompare != gputypes.CompareFunctionLess {
		t.Errorf("DepthCompare = %v, want Less", ds.DepthCompare)
	}
}

func TestOpaqueDepthStencilNoStencilWrite(t *testing.T) {
	ds := OpaqueDepthStencil(gputypes.TextureFormatDepth24Plus)
	if ds.StencilWriteMask != 0 {
		t.Errorf("StencilWriteMask = %x, want 0 (no stencil in Phase 1)", ds.StencilWriteMask)
	}
}

func TestDefaultPrimitiveCullBack(t *testing.T) {
	ps := DefaultPrimitive(false)
	if ps.CullMode != gputypes.CullModeBack {
		t.Errorf("CullMode = %v, want CullModeBack", ps.CullMode)
	}
	if ps.Topology != gputypes.PrimitiveTopologyTriangleList {
		t.Errorf("Topology = %v, want TriangleList", ps.Topology)
	}
}

func TestDefaultPrimitiveDoubleSided(t *testing.T) {
	ps := DefaultPrimitive(true)
	if ps.CullMode != gputypes.CullModeNone {
		t.Errorf("CullMode = %v, want CullModeNone for double-sided", ps.CullMode)
	}
}

func TestNoMultisample(t *testing.T) {
	ms := NoMultisample()
	if ms.Count != 1 {
		t.Errorf("Count = %d, want 1 (no MSAA)", ms.Count)
	}
	if ms.Mask != 0xFFFFFFFF {
		t.Errorf("Mask = %x, want 0xFFFFFFFF", ms.Mask)
	}
}

func TestStandardVertexLayoutStride(t *testing.T) {
	vbl := StandardVertexLayout()
	if len(vbl) != 1 {
		t.Fatalf("expected 1 vertex buffer layout, got %d", len(vbl))
	}
	if vbl[0].ArrayStride != 32 {
		t.Errorf("ArrayStride = %d, want 32", vbl[0].ArrayStride)
	}
}

func TestStandardVertexLayoutAttributes(t *testing.T) {
	vbl := StandardVertexLayout()
	attrs := vbl[0].Attributes
	if len(attrs) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(attrs))
	}

	// Position: vec3, offset 0, location 0.
	if attrs[0].Format != gputypes.VertexFormatFloat32x3 {
		t.Errorf("position format = %v, want Float32x3", attrs[0].Format)
	}
	if attrs[0].Offset != 0 {
		t.Errorf("position offset = %d, want 0", attrs[0].Offset)
	}
	if attrs[0].ShaderLocation != 0 {
		t.Errorf("position location = %d, want 0", attrs[0].ShaderLocation)
	}

	// Normal: vec3, offset 12, location 1.
	if attrs[1].Format != gputypes.VertexFormatFloat32x3 {
		t.Errorf("normal format = %v, want Float32x3", attrs[1].Format)
	}
	if attrs[1].Offset != 12 {
		t.Errorf("normal offset = %d, want 12", attrs[1].Offset)
	}
	if attrs[1].ShaderLocation != 1 {
		t.Errorf("normal location = %d, want 1", attrs[1].ShaderLocation)
	}

	// UV: vec2, offset 24, location 2.
	if attrs[2].Format != gputypes.VertexFormatFloat32x2 {
		t.Errorf("uv format = %v, want Float32x2", attrs[2].Format)
	}
	if attrs[2].Offset != 24 {
		t.Errorf("uv offset = %d, want 24", attrs[2].Offset)
	}
	if attrs[2].ShaderLocation != 2 {
		t.Errorf("uv location = %d, want 2", attrs[2].ShaderLocation)
	}
}

func TestShaderEntryPoints(t *testing.T) {
	if shaderEntryVS != "vs_main" {
		t.Errorf("shaderEntryVS = %q, want %q", shaderEntryVS, "vs_main")
	}
	if shaderEntryFS != "fs_main" {
		t.Errorf("shaderEntryFS = %q, want %q", shaderEntryFS, "fs_main")
	}
}
