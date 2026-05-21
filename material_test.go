package g3d

import (
	"encoding/binary"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// AlphaMode enum
// ---------------------------------------------------------------------------

func TestAlphaModeValues(t *testing.T) {
	// Ensure iota ordering is stable (serialization depends on it).
	if AlphaModeOpaque != 0 {
		t.Errorf("AlphaModeOpaque = %d, want 0", AlphaModeOpaque)
	}
	if AlphaModeMask != 1 {
		t.Errorf("AlphaModeMask = %d, want 1", AlphaModeMask)
	}
	if AlphaModeBlend != 2 {
		t.Errorf("AlphaModeBlend = %d, want 2", AlphaModeBlend)
	}
}

// ---------------------------------------------------------------------------
// RenderBucket enum
// ---------------------------------------------------------------------------

func TestRenderBucketValues(t *testing.T) {
	if RenderBucketBackground != 0 {
		t.Errorf("RenderBucketBackground = %d, want 0", RenderBucketBackground)
	}
	if RenderBucketOpaque != 1 {
		t.Errorf("RenderBucketOpaque = %d, want 1", RenderBucketOpaque)
	}
	if RenderBucketTransmissive != 2 {
		t.Errorf("RenderBucketTransmissive = %d, want 2", RenderBucketTransmissive)
	}
	if RenderBucketTransparent != 3 {
		t.Errorf("RenderBucketTransparent = %d, want 3", RenderBucketTransparent)
	}
}

func TestRenderBucketOrder(t *testing.T) {
	// Render order must be Background < Opaque < Transmissive < Transparent.
	if RenderBucketBackground >= RenderBucketOpaque {
		t.Error("Background must sort before Opaque")
	}
	if RenderBucketOpaque >= RenderBucketTransmissive {
		t.Error("Opaque must sort before Transmissive")
	}
	if RenderBucketTransmissive >= RenderBucketTransparent {
		t.Error("Transmissive must sort before Transparent")
	}
}

// ---------------------------------------------------------------------------
// BasicMaterial — defaults
// ---------------------------------------------------------------------------

func TestBasicMaterialDefaults(t *testing.T) {
	m := NewBasicMaterial()

	if m.ShaderID() != "basic" {
		t.Errorf("ShaderID() = %q, want %q", m.ShaderID(), "basic")
	}
	if m.RenderBucket() != RenderBucketOpaque {
		t.Errorf("RenderBucket() = %d, want %d (Opaque)", m.RenderBucket(), RenderBucketOpaque)
	}
	if m.DoubleSided() {
		t.Error("DoubleSided() = true, want false")
	}
	if m.Wireframe() {
		t.Error("Wireframe() = true, want false")
	}
	if m.Color() != White {
		t.Errorf("Color() = %+v, want White", m.Color())
	}
	if m.Opacity() != 1 {
		t.Errorf("Opacity() = %v, want 1", m.Opacity())
	}
}

// ---------------------------------------------------------------------------
// BasicMaterial — options
// ---------------------------------------------------------------------------

func TestBasicMaterialOptions(t *testing.T) {
	m := NewBasicMaterial(
		WithColor(Red),
		WithOpacity(0.5),
		WithDoubleSided(true),
		WithWireframe(true),
	)

	if m.Color() != Red {
		t.Errorf("Color() = %+v, want Red", m.Color())
	}
	if m.Opacity() != 0.5 {
		t.Errorf("Opacity() = %v, want 0.5", m.Opacity())
	}
	if !m.DoubleSided() {
		t.Error("DoubleSided() = false, want true")
	}
	if !m.Wireframe() {
		t.Error("Wireframe() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// BasicMaterial — RenderBucket by opacity
// ---------------------------------------------------------------------------

func TestBasicMaterialRenderBucket(t *testing.T) {
	tests := []struct {
		name    string
		opacity float32
		want    RenderBucket
	}{
		{"opaque", 1.0, RenderBucketOpaque},
		{"half transparent", 0.5, RenderBucketTransparent},
		{"fully transparent", 0.0, RenderBucketTransparent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBasicMaterial(WithOpacity(tt.opacity))
			if got := m.RenderBucket(); got != tt.want {
				t.Errorf("RenderBucket() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BasicMaterial — UniformData
// ---------------------------------------------------------------------------

func TestBasicMaterialUniformDataSize(t *testing.T) {
	m := NewBasicMaterial()
	data := m.UniformData()
	if len(data) != 16 {
		t.Fatalf("UniformData length = %d, want 16", len(data))
	}
}

func TestBasicMaterialUniformDataLayout(t *testing.T) {
	m := NewBasicMaterial(WithColor(RGBA(0.2, 0.4, 0.6, 0.8)), WithOpacity(0.5))
	data := m.UniformData()

	r := math.Float32frombits(binary.LittleEndian.Uint32(data[0:4]))
	g := math.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))
	b := math.Float32frombits(binary.LittleEndian.Uint32(data[8:12]))
	a := math.Float32frombits(binary.LittleEndian.Uint32(data[12:16]))

	if r != 0.2 {
		t.Errorf("R = %v, want 0.2", r)
	}
	if g != 0.4 {
		t.Errorf("G = %v, want 0.4", g)
	}
	if b != 0.6 {
		t.Errorf("B = %v, want 0.6", b)
	}
	wantA := float32(0.8) * float32(0.5)
	if a != wantA {
		t.Errorf("A = %v, want %v (color.A * opacity)", a, wantA)
	}
}

func TestBasicMaterialUniformDataWhiteOpaque(t *testing.T) {
	m := NewBasicMaterial()
	data := m.UniformData()

	for i := 0; i < 4; i++ {
		v := math.Float32frombits(binary.LittleEndian.Uint32(data[i*4 : i*4+4]))
		if v != 1.0 {
			t.Errorf("component[%d] = %v, want 1.0", i, v)
		}
	}
}

// ---------------------------------------------------------------------------
// BasicMaterial — interface compliance
// ---------------------------------------------------------------------------

func TestBasicMaterialImplementsMaterial(t *testing.T) {
	var _ Material = (*BasicMaterial)(nil)
}

// ---------------------------------------------------------------------------
// StandardMaterial — defaults
// ---------------------------------------------------------------------------

func TestStandardMaterialDefaults(t *testing.T) {
	m := NewStandardMaterial()

	if m.ShaderID() != "standard" {
		t.Errorf("ShaderID() = %q, want %q", m.ShaderID(), "standard")
	}
	if m.RenderBucket() != RenderBucketOpaque {
		t.Errorf("RenderBucket() = %d, want %d (Opaque)", m.RenderBucket(), RenderBucketOpaque)
	}
	if m.DoubleSided() {
		t.Error("DoubleSided() = true, want false")
	}
	if m.Wireframe() {
		t.Error("Wireframe() = true, want false")
	}
	if m.Color() != White {
		t.Errorf("Color() = %+v, want White", m.Color())
	}
	if m.Metallic() != 0 {
		t.Errorf("Metallic() = %v, want 0", m.Metallic())
	}
	if m.Roughness() != 0.5 {
		t.Errorf("Roughness() = %v, want 0.5", m.Roughness())
	}
	if m.Emissive() != Black {
		t.Errorf("Emissive() = %+v, want Black", m.Emissive())
	}
	if m.Opacity() != 1 {
		t.Errorf("Opacity() = %v, want 1", m.Opacity())
	}
	if m.AlphaMode() != AlphaModeOpaque {
		t.Errorf("AlphaMode() = %d, want %d (Opaque)", m.AlphaMode(), AlphaModeOpaque)
	}
	if m.AlphaCutoff() != 0.5 {
		t.Errorf("AlphaCutoff() = %v, want 0.5", m.AlphaCutoff())
	}
}

// ---------------------------------------------------------------------------
// StandardMaterial — options
// ---------------------------------------------------------------------------

func TestStandardMaterialOptions(t *testing.T) {
	m := NewStandardMaterial(
		WithColor(Blue),
		WithMetallic(0.8),
		WithRoughness(0.2),
		WithEmissive(Red),
		WithOpacity(0.7),
		WithAlphaMode(AlphaModeBlend),
		WithAlphaCutoff(0.3),
		WithDoubleSided(true),
		WithWireframe(true),
	)

	if m.Color() != Blue {
		t.Errorf("Color() = %+v, want Blue", m.Color())
	}
	if m.Metallic() != 0.8 {
		t.Errorf("Metallic() = %v, want 0.8", m.Metallic())
	}
	if m.Roughness() != 0.2 {
		t.Errorf("Roughness() = %v, want 0.2", m.Roughness())
	}
	if m.Emissive() != Red {
		t.Errorf("Emissive() = %+v, want Red", m.Emissive())
	}
	if m.Opacity() != 0.7 {
		t.Errorf("Opacity() = %v, want 0.7", m.Opacity())
	}
	if m.AlphaMode() != AlphaModeBlend {
		t.Errorf("AlphaMode() = %d, want %d (Blend)", m.AlphaMode(), AlphaModeBlend)
	}
	if m.AlphaCutoff() != 0.3 {
		t.Errorf("AlphaCutoff() = %v, want 0.3", m.AlphaCutoff())
	}
	if !m.DoubleSided() {
		t.Error("DoubleSided() = false, want true")
	}
	if !m.Wireframe() {
		t.Error("Wireframe() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// StandardMaterial — RenderBucket by AlphaMode
// ---------------------------------------------------------------------------

func TestStandardMaterialRenderBucket(t *testing.T) {
	tests := []struct {
		name string
		mode AlphaMode
		want RenderBucket
	}{
		{"opaque", AlphaModeOpaque, RenderBucketOpaque},
		{"mask", AlphaModeMask, RenderBucketOpaque},
		{"blend", AlphaModeBlend, RenderBucketTransparent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewStandardMaterial(WithAlphaMode(tt.mode))
			if got := m.RenderBucket(); got != tt.want {
				t.Errorf("RenderBucket() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StandardMaterial — UniformData
// ---------------------------------------------------------------------------

func TestStandardMaterialUniformDataSize(t *testing.T) {
	m := NewStandardMaterial()
	data := m.UniformData()
	if len(data) != 32 {
		t.Fatalf("UniformData length = %d, want 32", len(data))
	}
}

func TestStandardMaterialUniformDataLayout(t *testing.T) {
	m := NewStandardMaterial(
		WithColor(RGBA(0.1, 0.2, 0.3, 0.8)),
		WithOpacity(0.5),
		WithMetallic(0.9),
		WithRoughness(0.4),
		WithAlphaCutoff(0.6),
	)
	data := m.UniformData()

	readF32 := func(offset int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
	}

	// Color RGBA at offset 0-15
	if got := readF32(0); got != 0.1 {
		t.Errorf("R at offset 0 = %v, want 0.1", got)
	}
	if got := readF32(4); got != 0.2 {
		t.Errorf("G at offset 4 = %v, want 0.2", got)
	}
	if got := readF32(8); got != 0.3 {
		t.Errorf("B at offset 8 = %v, want 0.3", got)
	}
	wantA := float32(0.8) * float32(0.5)
	if got := readF32(12); got != wantA {
		t.Errorf("A at offset 12 = %v, want %v", got, wantA)
	}

	// Metallic at offset 16
	if got := readF32(16); got != 0.9 {
		t.Errorf("metallic at offset 16 = %v, want 0.9", got)
	}

	// Roughness at offset 20
	if got := readF32(20); got != 0.4 {
		t.Errorf("roughness at offset 20 = %v, want 0.4", got)
	}

	// AlphaCutoff at offset 24
	if got := readF32(24); got != 0.6 {
		t.Errorf("alpha_cutoff at offset 24 = %v, want 0.6", got)
	}

	// Padding at offset 28 should be zero
	if got := readF32(28); got != 0 {
		t.Errorf("padding at offset 28 = %v, want 0", got)
	}
}

func TestStandardMaterialUniformDataDefaults(t *testing.T) {
	m := NewStandardMaterial()
	data := m.UniformData()

	readF32 := func(offset int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
	}

	// Default: White (RGBA 1,1,1,1), opacity=1
	if got := readF32(0); got != 1.0 {
		t.Errorf("default R = %v, want 1.0", got)
	}
	if got := readF32(4); got != 1.0 {
		t.Errorf("default G = %v, want 1.0", got)
	}
	if got := readF32(8); got != 1.0 {
		t.Errorf("default B = %v, want 1.0", got)
	}
	if got := readF32(12); got != 1.0 {
		t.Errorf("default A = %v, want 1.0", got)
	}

	// Default metallic=0
	if got := readF32(16); got != 0 {
		t.Errorf("default metallic = %v, want 0", got)
	}

	// Default roughness=0.5
	if got := readF32(20); got != 0.5 {
		t.Errorf("default roughness = %v, want 0.5", got)
	}

	// Default alphaCutoff=0.5
	if got := readF32(24); got != 0.5 {
		t.Errorf("default alpha_cutoff = %v, want 0.5", got)
	}

	// Padding = 0
	if got := readF32(28); got != 0 {
		t.Errorf("default padding = %v, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// StandardMaterial — interface compliance
// ---------------------------------------------------------------------------

func TestStandardMaterialImplementsMaterial(t *testing.T) {
	var _ Material = (*StandardMaterial)(nil)
}

// ---------------------------------------------------------------------------
// Functional options — standard-only options are no-op on BasicMaterial
// ---------------------------------------------------------------------------

func TestStandardOnlyOptionsIgnoredByBasic(t *testing.T) {
	m := NewBasicMaterial(
		WithMetallic(0.9),
		WithRoughness(0.1),
		WithEmissive(Red),
		WithAlphaMode(AlphaModeBlend),
		WithAlphaCutoff(0.3),
	)
	// BasicMaterial should still have its defaults — standard-only options are no-ops.
	if m.Color() != White {
		t.Errorf("Color() = %+v, want White (unchanged)", m.Color())
	}
	if m.Opacity() != 1 {
		t.Errorf("Opacity() = %v, want 1 (unchanged)", m.Opacity())
	}
}

// ---------------------------------------------------------------------------
// Shared options work on both material types
// ---------------------------------------------------------------------------

func TestSharedOptionsApplyToBoth(t *testing.T) {
	color := RGBA(0.3, 0.6, 0.9, 1.0)
	opts := []MaterialOption{
		WithColor(color),
		WithOpacity(0.75),
		WithDoubleSided(true),
		WithWireframe(true),
	}

	basic := NewBasicMaterial(opts...)
	standard := NewStandardMaterial(opts...)

	// Both should have the same shared properties.
	if basic.Color() != color {
		t.Errorf("BasicMaterial.Color() = %+v, want %+v", basic.Color(), color)
	}
	if standard.Color() != color {
		t.Errorf("StandardMaterial.Color() = %+v, want %+v", standard.Color(), color)
	}
	if basic.Opacity() != 0.75 {
		t.Errorf("BasicMaterial.Opacity() = %v, want 0.75", basic.Opacity())
	}
	if standard.Opacity() != 0.75 {
		t.Errorf("StandardMaterial.Opacity() = %v, want 0.75", standard.Opacity())
	}
	if !basic.DoubleSided() {
		t.Error("BasicMaterial.DoubleSided() = false, want true")
	}
	if !standard.DoubleSided() {
		t.Error("StandardMaterial.DoubleSided() = false, want true")
	}
	if !basic.Wireframe() {
		t.Error("BasicMaterial.Wireframe() = false, want true")
	}
	if !standard.Wireframe() {
		t.Error("StandardMaterial.Wireframe() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Pipeline sharing: same ShaderID = shared pipeline
// ---------------------------------------------------------------------------

func TestPipelineSharingSameShaderID(t *testing.T) {
	a := NewStandardMaterial(WithColor(Red))
	b := NewStandardMaterial(WithColor(Blue), WithMetallic(1.0))

	if a.ShaderID() != b.ShaderID() {
		t.Errorf("same type should share ShaderID: %q != %q", a.ShaderID(), b.ShaderID())
	}
}

func TestPipelineSharingDifferentShaderID(t *testing.T) {
	basic := NewBasicMaterial()
	standard := NewStandardMaterial()

	if basic.ShaderID() == standard.ShaderID() {
		t.Errorf("different types should have different ShaderIDs: both %q", basic.ShaderID())
	}
}

// ---------------------------------------------------------------------------
// UniformData byte alignment
// ---------------------------------------------------------------------------

func TestBasicMaterialUniformAlignment(t *testing.T) {
	data := NewBasicMaterial().UniformData()
	// 16 bytes = exactly one vec4<f32>, 16-byte aligned.
	if len(data)%16 != 0 {
		t.Errorf("BasicMaterial UniformData size %d is not 16-byte aligned", len(data))
	}
}

func TestStandardMaterialUniformAlignment(t *testing.T) {
	data := NewStandardMaterial().UniformData()
	// 32 bytes = two vec4<f32>, 16-byte aligned.
	if len(data)%16 != 0 {
		t.Errorf("StandardMaterial UniformData size %d is not 16-byte aligned", len(data))
	}
}
