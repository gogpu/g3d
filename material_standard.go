package g3d

import (
	"encoding/binary"
	"math"
)

// StandardMaterial is a PBR material following the GLTF 2.0 metallic-roughness model.
// Phase 1 uses Blinn-Phong approximation; full PBR in Phase 2.
//
// The GPU uniform layout is 32 bytes:
//
//	offset  0: color RGBA        [4]float32  (16 bytes)
//	offset 16: metallic          float32     (4 bytes)
//	offset 20: roughness         float32     (4 bytes)
//	offset 24: alpha_cutoff      float32     (4 bytes)
//	offset 28: _padding          float32     (4 bytes)
//	                                         Total: 32 bytes
type StandardMaterial struct {
	color       Color
	metallic    float32
	roughness   float32
	emissive    Color
	opacity     float32
	alphaMode   AlphaMode
	alphaCutoff float32
	doubleSided bool
	wireframe   bool
}

// NewStandardMaterial creates a PBR material with the given options.
// Defaults: color=White, metallic=0, roughness=0.5, emissive=Black,
// opacity=1, alphaMode=Opaque, alphaCutoff=0.5.
func NewStandardMaterial(opts ...MaterialOption) *StandardMaterial {
	m := &StandardMaterial{
		color:       White,
		metallic:    0,
		roughness:   0.5,
		emissive:    Black,
		opacity:     1,
		alphaMode:   AlphaModeOpaque,
		alphaCutoff: 0.5,
	}
	for _, opt := range opts {
		opt.applyStandard(m)
	}
	return m
}

// ShaderID returns "standard", identifying the PBR shader pipeline.
func (m *StandardMaterial) ShaderID() string { return shaderStandard }

// RenderBucket returns the appropriate bucket based on AlphaMode:
//   - AlphaModeOpaque, AlphaModeMask -> RenderBucketOpaque
//   - AlphaModeBlend -> RenderBucketTransparent
func (m *StandardMaterial) RenderBucket() RenderBucket {
	if m.alphaMode == AlphaModeBlend {
		return RenderBucketTransparent
	}
	return RenderBucketOpaque
}

// DoubleSided reports whether back-face culling is disabled.
func (m *StandardMaterial) DoubleSided() bool { return m.doubleSided }

// Wireframe reports whether the material renders as wireframe.
func (m *StandardMaterial) Wireframe() bool { return m.wireframe }

// Color returns the base color.
func (m *StandardMaterial) Color() Color { return m.color }

// Metallic returns the metallic factor.
func (m *StandardMaterial) Metallic() float32 { return m.metallic }

// Roughness returns the roughness factor.
func (m *StandardMaterial) Roughness() float32 { return m.roughness }

// Emissive returns the emissive color.
func (m *StandardMaterial) Emissive() Color { return m.emissive }

// Opacity returns the opacity value.
func (m *StandardMaterial) Opacity() float32 { return m.opacity }

// AlphaMode returns the alpha handling mode.
func (m *StandardMaterial) AlphaMode() AlphaMode { return m.alphaMode }

// AlphaCutoff returns the alpha test threshold for AlphaModeMask.
func (m *StandardMaterial) AlphaCutoff() float32 { return m.alphaCutoff }

// UniformData returns 32 bytes encoding the material properties for GPU upload.
// The alpha component of color is pre-multiplied with opacity.
//
// Layout (matches WGSL MaterialUniforms struct):
//
//	offset  0: R             float32
//	offset  4: G             float32
//	offset  8: B             float32
//	offset 12: A             float32 (color.A * opacity)
//	offset 16: metallic      float32
//	offset 20: roughness     float32
//	offset 24: alpha_cutoff  float32
//	offset 28: _padding      float32 (zero)
func (m *StandardMaterial) UniformData() []byte {
	var buf [32]byte
	// Color RGBA (vec4<f32>, 16 bytes)
	binary.LittleEndian.PutUint32(buf[0:4], math.Float32bits(m.color.R))
	binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(m.color.G))
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(m.color.B))
	binary.LittleEndian.PutUint32(buf[12:16], math.Float32bits(m.color.A*m.opacity))
	// Scalar properties (3 x float32 + padding)
	binary.LittleEndian.PutUint32(buf[16:20], math.Float32bits(m.metallic))
	binary.LittleEndian.PutUint32(buf[20:24], math.Float32bits(m.roughness))
	binary.LittleEndian.PutUint32(buf[24:28], math.Float32bits(m.alphaCutoff))
	// buf[28:32] is zero-initialized (padding)
	return buf[:]
}
