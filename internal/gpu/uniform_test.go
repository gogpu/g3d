package gpu

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestFrameUniformsSize(t *testing.T) {
	if FrameUniformsSize != 224 {
		t.Errorf("FrameUniformsSize = %d, want 224", FrameUniformsSize)
	}
}

func TestObjectUniformsSize(t *testing.T) {
	if ObjectUniformsSize != 128 {
		t.Errorf("ObjectUniformsSize = %d, want 128", ObjectUniformsSize)
	}
}

func TestPackFrameUniformsLength(t *testing.T) {
	data := &FrameUniformsData{}
	buf := PackFrameUniforms(data)
	if len(buf) != FrameUniformsSize {
		t.Errorf("PackFrameUniforms len = %d, want %d", len(buf), FrameUniformsSize)
	}
}

func TestPackObjectUniformsLength(t *testing.T) {
	data := &ObjectUniformsData{}
	buf := PackObjectUniforms(data)
	if len(buf) != ObjectUniformsSize {
		t.Errorf("PackObjectUniforms len = %d, want %d", len(buf), ObjectUniformsSize)
	}
}

func TestPackFrameUniformsViewProjection(t *testing.T) {
	data := &FrameUniformsData{}
	// Set identity-like VP (first element = 1.0).
	data.ViewProjection[0] = 1.0
	data.ViewProjection[5] = 2.0
	data.ViewProjection[10] = 3.0
	data.ViewProjection[15] = 4.0

	buf := PackFrameUniforms(data)

	// Check m[0] at offset 0.
	got := math.Float32frombits(binary.LittleEndian.Uint32(buf[0:4]))
	if got != 1.0 {
		t.Errorf("VP[0] = %f, want 1.0", got)
	}

	// Check m[5] at offset 20.
	got = math.Float32frombits(binary.LittleEndian.Uint32(buf[20:24]))
	if got != 2.0 {
		t.Errorf("VP[5] = %f, want 2.0", got)
	}

	// Check m[10] at offset 40.
	got = math.Float32frombits(binary.LittleEndian.Uint32(buf[40:44]))
	if got != 3.0 {
		t.Errorf("VP[10] = %f, want 3.0", got)
	}

	// Check m[15] at offset 60.
	got = math.Float32frombits(binary.LittleEndian.Uint32(buf[60:64]))
	if got != 4.0 {
		t.Errorf("VP[15] = %f, want 4.0", got)
	}
}

func TestPackFrameUniformsCameraPosition(t *testing.T) {
	data := &FrameUniformsData{}
	data.CameraPosition = [4]float32{10.0, 20.0, 30.0, 1.0}

	buf := PackFrameUniforms(data)

	// Camera position starts at offset 64.
	for i, want := range []float32{10.0, 20.0, 30.0, 1.0} {
		off := 64 + i*4
		got := math.Float32frombits(binary.LittleEndian.Uint32(buf[off : off+4]))
		if got != want {
			t.Errorf("CameraPosition[%d] = %f, want %f", i, got, want)
		}
	}
}

func TestPackFrameUniformsLights(t *testing.T) {
	data := &FrameUniformsData{}
	data.Lights[0] = LightData{
		Direction: [3]float32{0.0, -1.0, 0.0},
		LightType: 1, // Directional
		Color:     [3]float32{1.0, 0.5, 0.0},
		Intensity: 0.8,
	}
	data.LightCount = 1

	buf := PackFrameUniforms(data)

	// First light starts at offset 80.
	// Direction Y at offset 80+4 = 84.
	dirY := math.Float32frombits(binary.LittleEndian.Uint32(buf[84:88]))
	if dirY != -1.0 {
		t.Errorf("light[0].Direction.Y = %f, want -1.0", dirY)
	}

	// LightType at offset 80+12 = 92.
	lightType := binary.LittleEndian.Uint32(buf[92:96])
	if lightType != 1 {
		t.Errorf("light[0].LightType = %d, want 1", lightType)
	}

	// Color R at offset 80+16 = 96.
	colorR := math.Float32frombits(binary.LittleEndian.Uint32(buf[96:100]))
	if colorR != 1.0 {
		t.Errorf("light[0].Color.R = %f, want 1.0", colorR)
	}

	// Intensity at offset 80+28 = 108.
	intensity := math.Float32frombits(binary.LittleEndian.Uint32(buf[108:112]))
	if intensity != 0.8 {
		t.Errorf("light[0].Intensity = %f, want 0.8", intensity)
	}

	// LightCount at offset 208.
	lightCount := binary.LittleEndian.Uint32(buf[208:212])
	if lightCount != 1 {
		t.Errorf("LightCount = %d, want 1", lightCount)
	}
}

func TestPackFrameUniformsPaddingZero(t *testing.T) {
	data := &FrameUniformsData{}
	data.LightCount = 3

	buf := PackFrameUniforms(data)

	// Padding bytes at offsets 212, 216, 220 should be zero.
	for _, off := range []int{212, 216, 220} {
		val := binary.LittleEndian.Uint32(buf[off : off+4])
		if val != 0 {
			t.Errorf("padding at offset %d = %d, want 0", off, val)
		}
	}
}

func TestPackObjectUniformsModel(t *testing.T) {
	data := &ObjectUniformsData{}
	// Set a recognizable value.
	data.Model[0] = 1.0
	data.Model[5] = 2.0
	data.NormalModel[0] = 3.0
	data.NormalModel[5] = 4.0

	buf := PackObjectUniforms(data)

	// Model[0] at offset 0.
	got := math.Float32frombits(binary.LittleEndian.Uint32(buf[0:4]))
	if got != 1.0 {
		t.Errorf("Model[0] = %f, want 1.0", got)
	}

	// Model[5] at offset 20.
	got = math.Float32frombits(binary.LittleEndian.Uint32(buf[20:24]))
	if got != 2.0 {
		t.Errorf("Model[5] = %f, want 2.0", got)
	}

	// NormalModel[0] at offset 64.
	got = math.Float32frombits(binary.LittleEndian.Uint32(buf[64:68]))
	if got != 3.0 {
		t.Errorf("NormalModel[0] = %f, want 3.0", got)
	}

	// NormalModel[5] at offset 84.
	got = math.Float32frombits(binary.LittleEndian.Uint32(buf[84:88]))
	if got != 4.0 {
		t.Errorf("NormalModel[5] = %f, want 4.0", got)
	}
}
