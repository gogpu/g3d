package gpu

import (
	"encoding/binary"
	"math"
)

// Uniform buffer sizes must match WGSL struct layouts byte-for-byte.
const (
	// FrameUniformsSize is the byte size of the FrameUniforms struct (224 bytes).
	FrameUniformsSize = 224

	// ObjectUniformsSize is the byte size of the ObjectUniforms struct (128 bytes).
	ObjectUniformsSize = 128

	// MaxLights is the maximum number of lights per frame (matches WGSL array size).
	MaxLights = 4
)

// LightData is the GPU-side representation of a single light source.
// Layout matches the WGSL LightData struct (32 bytes, 16-byte aligned).
//
//	offset  0: Direction [3]float32 (12 bytes)
//	offset 12: LightType uint32     (4 bytes)
//	offset 16: Color     [3]float32 (12 bytes)
//	offset 28: Intensity float32    (4 bytes)
type LightData struct {
	Direction [3]float32
	LightType uint32
	Color     [3]float32
	Intensity float32
}

// FrameUniformsData holds per-frame data for GPU upload. Layout matches the
// WGSL FrameUniforms struct (224 bytes total).
//
//	offset   0: ViewProjection [16]float32 (64 bytes)
//	offset  64: CameraPosition [4]float32  (16 bytes — vec4, NOT vec3!)
//	offset  80: Lights         [4]LightData (128 bytes)
//	offset 208: LightCount     uint32       (4 bytes)
//	offset 212: Pad            [3]uint32    (12 bytes)
type FrameUniformsData struct {
	ViewProjection [16]float32
	CameraPosition [4]float32
	Lights         [MaxLights]LightData
	LightCount     uint32
	Pad            [3]uint32
}

// ObjectUniformsData holds per-object data for GPU upload. Layout matches
// the WGSL ObjectUniforms struct (128 bytes total).
//
//	offset  0: Model       [16]float32 (64 bytes)
//	offset 64: NormalModel  [16]float32 (64 bytes)
type ObjectUniformsData struct {
	Model       [16]float32
	NormalModel [16]float32
}

// PackFrameUniforms encodes FrameUniformsData as 224 bytes in little-endian
// format for GPU upload.
func PackFrameUniforms(data *FrameUniformsData) []byte {
	buf := make([]byte, FrameUniformsSize)
	off := 0

	// ViewProjection: 64 bytes (16 x float32)
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(data.ViewProjection[i]))
		off += 4
	}

	// CameraPosition: 16 bytes (4 x float32, vec4)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(data.CameraPosition[i]))
		off += 4
	}

	// Lights: 128 bytes (4 x 32 bytes)
	for i := 0; i < MaxLights; i++ {
		light := &data.Lights[i]
		// Direction: 12 bytes
		for j := 0; j < 3; j++ {
			binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(light.Direction[j]))
			off += 4
		}
		// LightType: 4 bytes
		binary.LittleEndian.PutUint32(buf[off:off+4], light.LightType)
		off += 4
		// Color: 12 bytes
		for j := 0; j < 3; j++ {
			binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(light.Color[j]))
			off += 4
		}
		// Intensity: 4 bytes
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(light.Intensity))
		off += 4
	}

	// LightCount: 4 bytes
	binary.LittleEndian.PutUint32(buf[off:off+4], data.LightCount)

	// Pad: 12 bytes (3 x uint32, zeroed) — buf is zero-initialized.

	return buf
}

// PackObjectUniforms encodes ObjectUniformsData as 128 bytes in little-endian
// format for GPU upload.
func PackObjectUniforms(data *ObjectUniformsData) []byte {
	buf := make([]byte, ObjectUniformsSize)
	off := 0

	// Model: 64 bytes
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(data.Model[i]))
		off += 4
	}

	// NormalModel: 64 bytes
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(data.NormalModel[i]))
		off += 4
	}

	return buf
}
