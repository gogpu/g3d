package g3d

import (
	"encoding/binary"
	"math"
)

// LightKind identifies the type of a light source. The value is stored in
// the GPU uniform struct so the shader can branch on light type.
type LightKind uint32

const (
	// LightKindAmbient is a uniform light with no direction or position.
	LightKindAmbient LightKind = 0

	// LightKindDirectional is a parallel light from an infinite distance.
	LightKindDirectional LightKind = 1

	// Phase 2:
	// LightKindPoint LightKind = 2
	// LightKindSpot  LightKind = 3
)

// Light provides illumination data for the renderer. Every light type
// (ambient, directional, and future point/spot) implements this interface.
// The renderer collects all lights from the scene and packs their data into
// a uniform buffer each frame.
type Light interface {
	// LightType returns the kind discriminant for the GPU uniform.
	LightType() LightKind

	// LightColor returns the light's color.
	LightColor() Color

	// LightIntensity returns the light's intensity multiplier.
	LightIntensity() float32
}

// LightOption configures light properties via functional options.
type LightOption interface {
	applyAmbientLight(l *AmbientLight)
	applyDirectionalLight(l *DirectionalLight)
}

// --- Shared light options ---

type lightColorOption struct{ c Color }

func (o lightColorOption) applyAmbientLight(l *AmbientLight)         { l.color = o.c }
func (o lightColorOption) applyDirectionalLight(l *DirectionalLight) { l.color = o.c }

// WithLightColor sets the light color. Applies to all light types.
func WithLightColor(c Color) LightOption { return lightColorOption{c} }

type lightIntensityOption struct{ v float32 }

func (o lightIntensityOption) applyAmbientLight(l *AmbientLight)         { l.intensity = o.v }
func (o lightIntensityOption) applyDirectionalLight(l *DirectionalLight) { l.intensity = o.v }

// WithLightIntensity sets the light intensity. Applies to all light types.
func WithLightIntensity(v float32) LightOption { return lightIntensityOption{v} }

// LightUniform is the GPU-side representation of a single light source.
// The struct layout matches the WGSL DirectionalLight struct exactly (32 bytes,
// 16-byte aligned). Fields are ordered to satisfy WGSL alignment rules:
//
//	offset  0: Direction [3]float32  (12 bytes — vec3<f32>)
//	offset 12: Kind      uint32      (4 bytes  — padding slot + discriminant)
//	offset 16: Color     [3]float32  (12 bytes — vec3<f32>)
//	offset 28: Intensity float32     (4 bytes)
//	Total: 32 bytes, 16-byte aligned
//
// The Direction+Kind pair fills a full 16-byte row, and Color+Intensity fills
// the second 16-byte row, satisfying WGSL vec3<f32> alignment requirements.
type LightUniform struct {
	Direction [3]float32 // world-space direction (unused for ambient)
	Kind      uint32     // LightKind discriminant
	Color     [3]float32 // linear RGB color
	Intensity float32    // intensity multiplier
}

// LightUniformSize is the byte size of a single LightUniform (32 bytes).
const LightUniformSize = 32

// Bytes encodes the LightUniform as 32 bytes in little-endian format for GPU
// upload. The layout matches the WGSL struct byte-for-byte.
func (u LightUniform) Bytes() []byte {
	var buf [LightUniformSize]byte
	binary.LittleEndian.PutUint32(buf[0:4], math.Float32bits(u.Direction[0]))
	binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(u.Direction[1]))
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(u.Direction[2]))
	binary.LittleEndian.PutUint32(buf[12:16], u.Kind)
	binary.LittleEndian.PutUint32(buf[16:20], math.Float32bits(u.Color[0]))
	binary.LittleEndian.PutUint32(buf[20:24], math.Float32bits(u.Color[1]))
	binary.LittleEndian.PutUint32(buf[24:28], math.Float32bits(u.Color[2]))
	binary.LittleEndian.PutUint32(buf[28:32], math.Float32bits(u.Intensity))
	return buf[:]
}
