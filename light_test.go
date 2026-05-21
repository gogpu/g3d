package g3d

import (
	"encoding/binary"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// LightKind enum
// ---------------------------------------------------------------------------

func TestLightKindValues(t *testing.T) {
	if LightKindAmbient != 0 {
		t.Errorf("LightKindAmbient = %d, want 0", LightKindAmbient)
	}
	if LightKindDirectional != 1 {
		t.Errorf("LightKindDirectional = %d, want 1", LightKindDirectional)
	}
}

// ---------------------------------------------------------------------------
// LightUniform — byte layout
// ---------------------------------------------------------------------------

func TestLightUniformSize(t *testing.T) {
	if LightUniformSize != 32 {
		t.Errorf("LightUniformSize = %d, want 32", LightUniformSize)
	}
}

func TestLightUniformBytesLength(t *testing.T) {
	u := LightUniform{}
	data := u.Bytes()
	if len(data) != 32 {
		t.Fatalf("Bytes() length = %d, want 32", len(data))
	}
}

func TestLightUniformBytesLayout(t *testing.T) {
	u := LightUniform{
		Direction: [3]float32{0.1, 0.2, 0.3},
		Kind:      1,
		Color:     [3]float32{0.4, 0.5, 0.6},
		Intensity: 0.7,
	}
	data := u.Bytes()

	readF32 := func(offset int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
	}
	readU32 := func(offset int) uint32 {
		return binary.LittleEndian.Uint32(data[offset : offset+4])
	}

	// Direction at offsets 0, 4, 8
	if got := readF32(0); got != 0.1 {
		t.Errorf("direction.x at offset 0 = %v, want 0.1", got)
	}
	if got := readF32(4); got != 0.2 {
		t.Errorf("direction.y at offset 4 = %v, want 0.2", got)
	}
	if got := readF32(8); got != 0.3 {
		t.Errorf("direction.z at offset 8 = %v, want 0.3", got)
	}

	// Kind at offset 12
	if got := readU32(12); got != 1 {
		t.Errorf("kind at offset 12 = %d, want 1", got)
	}

	// Color at offsets 16, 20, 24
	if got := readF32(16); got != 0.4 {
		t.Errorf("color.r at offset 16 = %v, want 0.4", got)
	}
	if got := readF32(20); got != 0.5 {
		t.Errorf("color.g at offset 20 = %v, want 0.5", got)
	}
	if got := readF32(24); got != 0.6 {
		t.Errorf("color.b at offset 24 = %v, want 0.6", got)
	}

	// Intensity at offset 28
	if got := readF32(28); got != 0.7 {
		t.Errorf("intensity at offset 28 = %v, want 0.7", got)
	}
}

func TestLightUniformBytesAlignment(t *testing.T) {
	u := LightUniform{}
	data := u.Bytes()
	if len(data)%16 != 0 {
		t.Errorf("Bytes() size %d is not 16-byte aligned", len(data))
	}
}

func TestLightUniformBytesZeroAmbient(t *testing.T) {
	// Ambient light uniform: direction is zero, kind is 0.
	u := LightUniform{
		Direction: [3]float32{0, 0, 0},
		Kind:      0,
		Color:     [3]float32{1, 1, 1},
		Intensity: 0.3,
	}
	data := u.Bytes()

	readF32 := func(offset int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
	}
	readU32 := func(offset int) uint32 {
		return binary.LittleEndian.Uint32(data[offset : offset+4])
	}

	// Direction should be zero
	for i := 0; i < 3; i++ {
		if got := readF32(i * 4); got != 0 {
			t.Errorf("direction[%d] = %v, want 0", i, got)
		}
	}
	// Kind should be 0 (ambient)
	if got := readU32(12); got != 0 {
		t.Errorf("kind = %d, want 0", got)
	}
	// Color should be white
	for i := 0; i < 3; i++ {
		if got := readF32(16 + i*4); got != 1 {
			t.Errorf("color[%d] = %v, want 1", i, got)
		}
	}
	// Intensity
	if got := readF32(28); got != 0.3 {
		t.Errorf("intensity = %v, want 0.3", got)
	}
}

// ---------------------------------------------------------------------------
// AmbientLight — interface compliance
// ---------------------------------------------------------------------------

func TestAmbientLightImplementsLight(t *testing.T) {
	var _ Light = (*AmbientLight)(nil)
}

// ---------------------------------------------------------------------------
// AmbientLight — defaults
// ---------------------------------------------------------------------------

func TestAmbientLightDefaults(t *testing.T) {
	l := NewAmbientLight()

	if l.LightType() != LightKindAmbient {
		t.Errorf("LightType() = %d, want %d (Ambient)", l.LightType(), LightKindAmbient)
	}
	if l.LightColor() != White {
		t.Errorf("LightColor() = %+v, want White", l.LightColor())
	}
	if l.LightIntensity() != 1 {
		t.Errorf("LightIntensity() = %v, want 1", l.LightIntensity())
	}
	if l.Color() != White {
		t.Errorf("Color() = %+v, want White", l.Color())
	}
	if l.Intensity() != 1 {
		t.Errorf("Intensity() = %v, want 1", l.Intensity())
	}
}

// ---------------------------------------------------------------------------
// AmbientLight — functional options
// ---------------------------------------------------------------------------

func TestAmbientLightOptions(t *testing.T) {
	l := NewAmbientLight(
		WithLightColor(Red),
		WithLightIntensity(0.5),
	)

	if l.Color() != Red {
		t.Errorf("Color() = %+v, want Red", l.Color())
	}
	if l.Intensity() != 0.5 {
		t.Errorf("Intensity() = %v, want 0.5", l.Intensity())
	}
}

// ---------------------------------------------------------------------------
// AmbientLight — setters
// ---------------------------------------------------------------------------

func TestAmbientLightSetters(t *testing.T) {
	l := NewAmbientLight()

	l.SetColor(Green)
	if l.Color() != Green {
		t.Errorf("after SetColor(Green): Color() = %+v, want Green", l.Color())
	}

	l.SetIntensity(0.2)
	if l.Intensity() != 0.2 {
		t.Errorf("after SetIntensity(0.2): Intensity() = %v, want 0.2", l.Intensity())
	}
}

// ---------------------------------------------------------------------------
// AmbientLight — LightUniform
// ---------------------------------------------------------------------------

func TestAmbientLightUniform(t *testing.T) {
	l := NewAmbientLight(
		WithLightColor(RGB(0.2, 0.4, 0.6)),
		WithLightIntensity(0.3),
	)
	u := l.LightUniform()

	// Direction should be zero for ambient.
	if u.Direction != [3]float32{0, 0, 0} {
		t.Errorf("Direction = %v, want [0,0,0]", u.Direction)
	}
	if u.Kind != uint32(LightKindAmbient) {
		t.Errorf("Kind = %d, want %d", u.Kind, LightKindAmbient)
	}
	if u.Color != [3]float32{0.2, 0.4, 0.6} {
		t.Errorf("Color = %v, want [0.2,0.4,0.6]", u.Color)
	}
	if u.Intensity != 0.3 {
		t.Errorf("Intensity = %v, want 0.3", u.Intensity)
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight — interface compliance
// ---------------------------------------------------------------------------

func TestDirectionalLightImplementsLight(t *testing.T) {
	var _ Light = (*DirectionalLight)(nil)
}

// ---------------------------------------------------------------------------
// DirectionalLight — defaults
// ---------------------------------------------------------------------------

func TestDirectionalLightDefaults(t *testing.T) {
	l := NewDirectionalLight()

	if l.LightType() != LightKindDirectional {
		t.Errorf("LightType() = %d, want %d (Directional)", l.LightType(), LightKindDirectional)
	}
	if l.LightColor() != White {
		t.Errorf("LightColor() = %+v, want White", l.LightColor())
	}
	if l.LightIntensity() != 1 {
		t.Errorf("LightIntensity() = %v, want 1", l.LightIntensity())
	}
	if l.Color() != White {
		t.Errorf("Color() = %+v, want White", l.Color())
	}
	if l.Intensity() != 1 {
		t.Errorf("Intensity() = %v, want 1", l.Intensity())
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight — node
// ---------------------------------------------------------------------------

func TestDirectionalLightNode(t *testing.T) {
	l := NewDirectionalLight()
	n := l.LightNode()

	if n == nil {
		t.Fatal("LightNode() returned nil")
	}
	if n.Name() != "DirectionalLight" {
		t.Errorf("LightNode().Name() = %q, want %q", n.Name(), "DirectionalLight")
	}
	if !n.Visible() {
		t.Error("LightNode().Visible() = false, want true")
	}
	// Default scale should be identity.
	if n.Scale != (Vec3{1, 1, 1}) {
		t.Errorf("LightNode().Scale = %+v, want {1,1,1}", n.Scale)
	}
}

func TestDirectionalLightAddToScene(t *testing.T) {
	scene := NewScene()
	l := NewDirectionalLight()
	scene.Add(l.LightNode())

	if l.LightNode().Parent() != &scene.Node {
		t.Error("LightNode() should be child of scene after Add")
	}
	if scene.ChildCount() != 1 {
		t.Errorf("scene.ChildCount() = %d, want 1", scene.ChildCount())
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight — functional options
// ---------------------------------------------------------------------------

func TestDirectionalLightOptions(t *testing.T) {
	l := NewDirectionalLight(
		WithLightColor(Blue),
		WithLightIntensity(0.8),
	)

	if l.Color() != Blue {
		t.Errorf("Color() = %+v, want Blue", l.Color())
	}
	if l.Intensity() != 0.8 {
		t.Errorf("Intensity() = %v, want 0.8", l.Intensity())
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight — setters
// ---------------------------------------------------------------------------

func TestDirectionalLightSetters(t *testing.T) {
	l := NewDirectionalLight()

	l.SetColor(Yellow)
	if l.Color() != Yellow {
		t.Errorf("after SetColor(Yellow): Color() = %+v, want Yellow", l.Color())
	}

	l.SetIntensity(0.6)
	if l.Intensity() != 0.6 {
		t.Errorf("after SetIntensity(0.6): Intensity() = %v, want 0.6", l.Intensity())
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight — direction from rotation
// ---------------------------------------------------------------------------

func TestDirectionalLightDefaultDirection(t *testing.T) {
	l := NewDirectionalLight()
	dir := l.Direction()

	// Default rotation is identity, forward is (0, 0, -1).
	if !vec3Near(dir, Vec3{0, 0, -1}, 1e-6) {
		t.Errorf("default Direction() = %+v, want (0, 0, -1)", dir)
	}
}

func TestDirectionalLightDirectionRotatedX(t *testing.T) {
	l := NewDirectionalLight()
	// Rotate -90 degrees around X axis.
	// Rx(-90): y' = y*cos(-90) - z*sin(-90) = 0 - (-1)*(-1) = -1
	//          z' = y*sin(-90) + z*cos(-90) = 0 + 0 = 0
	// (0,0,-1) -> (0,-1,0): pointing down.
	l.LightNode().SetRotation(Euler{X: -math.Pi / 2, Y: 0, Z: 0})
	dir := l.Direction()

	if !vec3Near(dir, Vec3{0, -1, 0}, 1e-5) {
		t.Errorf("Direction() after -90 X rotation = %+v, want (0, -1, 0)", dir)
	}
}

func TestDirectionalLightDirectionRotatedY(t *testing.T) {
	l := NewDirectionalLight()
	// Rotate 90 degrees around Y axis: (0,0,-1) -> (-1,0,0).
	l.LightNode().SetRotation(Euler{X: 0, Y: math.Pi / 2, Z: 0})
	dir := l.Direction()

	if !vec3Near(dir, Vec3{-1, 0, 0}, 1e-5) {
		t.Errorf("Direction() after 90 Y rotation = %+v, want (-1, 0, 0)", dir)
	}
}

func TestDirectionalLightDirectionWithParent(t *testing.T) {
	scene := NewScene()
	group := NewGroup()
	// Rotate group 90 degrees around Y.
	group.SetRotation(Euler{X: 0, Y: math.Pi / 2, Z: 0})
	scene.Add(&group.Node)

	l := NewDirectionalLight()
	// Light has no local rotation, but parent rotates 90 deg around Y.
	group.Add(l.LightNode())

	dir := l.Direction()

	// (0,0,-1) rotated 90 deg around Y -> (-1,0,0)
	if !vec3Near(dir, Vec3{-1, 0, 0}, 1e-5) {
		t.Errorf("Direction() with parent rotation = %+v, want (-1, 0, 0)", dir)
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight — LightUniform
// ---------------------------------------------------------------------------

func TestDirectionalLightUniform(t *testing.T) {
	l := NewDirectionalLight(
		WithLightColor(RGB(0.8, 0.6, 0.4)),
		WithLightIntensity(1.5),
	)
	u := l.LightUniform()

	dir := l.Direction()
	if u.Direction != [3]float32{dir.X, dir.Y, dir.Z} {
		t.Errorf("Direction = %v, want %v", u.Direction, [3]float32{dir.X, dir.Y, dir.Z})
	}
	if u.Kind != uint32(LightKindDirectional) {
		t.Errorf("Kind = %d, want %d", u.Kind, LightKindDirectional)
	}
	if u.Color != [3]float32{0.8, 0.6, 0.4} {
		t.Errorf("Color = %v, want [0.8,0.6,0.4]", u.Color)
	}
	if u.Intensity != 1.5 {
		t.Errorf("Intensity = %v, want 1.5", u.Intensity)
	}
}

func TestDirectionalLightUniformRotated(t *testing.T) {
	l := NewDirectionalLight()
	// Rotate 180 degrees around Y: (0,0,-1) -> (0,0,1)
	l.LightNode().SetRotation(Euler{X: 0, Y: math.Pi, Z: 0})
	u := l.LightUniform()

	dir := Vec3{u.Direction[0], u.Direction[1], u.Direction[2]}
	if !vec3Near(dir, Vec3{0, 0, 1}, 1e-5) {
		t.Errorf("uniform Direction after 180 Y rotation = %+v, want (0, 0, 1)", dir)
	}
}

// ---------------------------------------------------------------------------
// LightUniform — roundtrip encode/decode
// ---------------------------------------------------------------------------

func TestLightUniformRoundtrip(t *testing.T) {
	original := LightUniform{
		Direction: [3]float32{-0.577, -0.577, -0.577},
		Kind:      1,
		Color:     [3]float32{1.0, 0.9, 0.8},
		Intensity: 2.5,
	}
	data := original.Bytes()

	readF32 := func(offset int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
	}
	readU32 := func(offset int) uint32 {
		return binary.LittleEndian.Uint32(data[offset : offset+4])
	}

	decoded := LightUniform{
		Direction: [3]float32{readF32(0), readF32(4), readF32(8)},
		Kind:      readU32(12),
		Color:     [3]float32{readF32(16), readF32(20), readF32(24)},
		Intensity: readF32(28),
	}

	if original != decoded {
		t.Errorf("roundtrip mismatch:\n  original: %+v\n  decoded:  %+v", original, decoded)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestAmbientLightZeroIntensity(t *testing.T) {
	l := NewAmbientLight(WithLightIntensity(0))
	if l.Intensity() != 0 {
		t.Errorf("Intensity() = %v, want 0", l.Intensity())
	}
	u := l.LightUniform()
	if u.Intensity != 0 {
		t.Errorf("uniform Intensity = %v, want 0", u.Intensity)
	}
}

func TestDirectionalLightHighIntensity(t *testing.T) {
	l := NewDirectionalLight(WithLightIntensity(100))
	if l.Intensity() != 100 {
		t.Errorf("Intensity() = %v, want 100", l.Intensity())
	}
	u := l.LightUniform()
	if u.Intensity != 100 {
		t.Errorf("uniform Intensity = %v, want 100", u.Intensity)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func vec3Near(a, b Vec3, eps float32) bool {
	return absf(a.X-b.X) < eps && absf(a.Y-b.Y) < eps && absf(a.Z-b.Z) < eps
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
