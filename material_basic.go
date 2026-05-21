package g3d

import (
	"encoding/binary"
	"math"
)

// BasicMaterial is an unlit material that renders a solid color.
// Suitable for prototyping, data visualization, wireframe overlays,
// and any situation where lighting calculations are not needed.
//
// The GPU uniform layout is 16 bytes:
//
//	offset 0:  color RGBA  [4]float32
type BasicMaterial struct {
	color       Color
	opacity     float32
	doubleSided bool
	wireframe   bool
}

// NewBasicMaterial creates an unlit material with the given options.
// Defaults: color=White, opacity=1.
func NewBasicMaterial(opts ...MaterialOption) *BasicMaterial {
	m := &BasicMaterial{
		color:   White,
		opacity: 1,
	}
	for _, opt := range opts {
		opt.applyBasic(m)
	}
	return m
}

// ShaderID returns "basic", identifying the unlit shader pipeline.
func (m *BasicMaterial) ShaderID() string { return shaderBasic }

// RenderBucket returns RenderBucketTransparent if opacity < 1,
// otherwise RenderBucketOpaque.
func (m *BasicMaterial) RenderBucket() RenderBucket {
	if m.opacity < 1 {
		return RenderBucketTransparent
	}
	return RenderBucketOpaque
}

// DoubleSided reports whether back-face culling is disabled.
func (m *BasicMaterial) DoubleSided() bool { return m.doubleSided }

// Wireframe reports whether the material renders as wireframe.
func (m *BasicMaterial) Wireframe() bool { return m.wireframe }

// Color returns the base color.
func (m *BasicMaterial) Color() Color { return m.color }

// Opacity returns the opacity value.
func (m *BasicMaterial) Opacity() float32 { return m.opacity }

// UniformData returns 16 bytes encoding the material color as RGBA float32 values.
// The alpha component is pre-multiplied with opacity.
//
// Layout (matches WGSL vec4<f32>):
//
//	offset 0:  R  float32
//	offset 4:  G  float32
//	offset 8:  B  float32
//	offset 12: A  float32 (color.A * opacity)
func (m *BasicMaterial) UniformData() []byte {
	var buf [16]byte
	binary.LittleEndian.PutUint32(buf[0:4], math.Float32bits(m.color.R))
	binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(m.color.G))
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(m.color.B))
	binary.LittleEndian.PutUint32(buf[12:16], math.Float32bits(m.color.A*m.opacity))
	return buf[:]
}
