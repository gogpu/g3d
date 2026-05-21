package g3d

import (
	"math"
	"testing"
)

const epsilon = 1e-5

func approxEqual(a, b, eps float32) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps
}

func approxEqualVec3(a, b Vec3, eps float32) bool {
	return approxEqual(a.X, b.X, eps) && approxEqual(a.Y, b.Y, eps) && approxEqual(a.Z, b.Z, eps)
}

func approxEqualVec4(a, b Vec4, eps float32) bool {
	return approxEqual(a.X, b.X, eps) && approxEqual(a.Y, b.Y, eps) &&
		approxEqual(a.Z, b.Z, eps) && approxEqual(a.W, b.W, eps)
}

// --- Vec2 Tests ---

func TestVec2Add(t *testing.T) {
	got := Vec2{1, 2}.Add(Vec2{3, 4})
	want := Vec2{4, 6}
	if got != want {
		t.Errorf("Vec2.Add() = %v, want %v", got, want)
	}
}

func TestVec2Sub(t *testing.T) {
	got := Vec2{5, 7}.Sub(Vec2{2, 3})
	want := Vec2{3, 4}
	if got != want {
		t.Errorf("Vec2.Sub() = %v, want %v", got, want)
	}
}

func TestVec2Scale(t *testing.T) {
	got := Vec2{2, 3}.Scale(2)
	want := Vec2{4, 6}
	if got != want {
		t.Errorf("Vec2.Scale() = %v, want %v", got, want)
	}
}

func TestVec2Dot(t *testing.T) {
	got := Vec2{1, 2}.Dot(Vec2{3, 4})
	want := float32(11) // 1*3 + 2*4
	if got != want {
		t.Errorf("Vec2.Dot() = %v, want %v", got, want)
	}
}

func TestVec2Length(t *testing.T) {
	v := Vec2{3, 4}
	got := v.Length()
	want := float32(5)
	if !approxEqual(got, want, epsilon) {
		t.Errorf("Vec2.Length() = %v, want %v", got, want)
	}
}

func TestVec2LengthSq(t *testing.T) {
	v := Vec2{3, 4}
	got := v.LengthSq()
	want := float32(25)
	if got != want {
		t.Errorf("Vec2.LengthSq() = %v, want %v", got, want)
	}
}

func TestVec2Normalize(t *testing.T) {
	tests := []struct {
		name string
		v    Vec2
		want Vec2
	}{
		{"unit X", Vec2{5, 0}, Vec2{1, 0}},
		{"unit Y", Vec2{0, 3}, Vec2{0, 1}},
		{"zero", Vec2{0, 0}, Vec2{0, 0}},
		{"diagonal", Vec2{1, 1}, Vec2{float32(1.0 / math.Sqrt(2)), float32(1.0 / math.Sqrt(2))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Normalize()
			if !approxEqual(got.X, tt.want.X, epsilon) || !approxEqual(got.Y, tt.want.Y, epsilon) {
				t.Errorf("Vec2.Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec2Lerp(t *testing.T) {
	a := Vec2{0, 0}
	b := Vec2{10, 20}
	got := a.Lerp(b, 0.5)
	want := Vec2{5, 10}
	if got != want {
		t.Errorf("Vec2.Lerp() = %v, want %v", got, want)
	}
}

// --- Vec3 Tests ---

func TestVec3Add(t *testing.T) {
	got := Vec3{1, 2, 3}.Add(Vec3{4, 5, 6})
	want := Vec3{5, 7, 9}
	if got != want {
		t.Errorf("Vec3.Add() = %v, want %v", got, want)
	}
}

func TestVec3Sub(t *testing.T) {
	got := Vec3{5, 7, 9}.Sub(Vec3{1, 2, 3})
	want := Vec3{4, 5, 6}
	if got != want {
		t.Errorf("Vec3.Sub() = %v, want %v", got, want)
	}
}

func TestVec3Scale(t *testing.T) {
	got := Vec3{1, 2, 3}.Scale(2)
	want := Vec3{2, 4, 6}
	if got != want {
		t.Errorf("Vec3.Scale() = %v, want %v", got, want)
	}
}

func TestVec3Mul(t *testing.T) {
	got := Vec3{2, 3, 4}.Mul(Vec3{5, 6, 7})
	want := Vec3{10, 18, 28}
	if got != want {
		t.Errorf("Vec3.Mul() = %v, want %v", got, want)
	}
}

func TestVec3Dot(t *testing.T) {
	got := Vec3{1, 2, 3}.Dot(Vec3{4, 5, 6})
	want := float32(32) // 1*4 + 2*5 + 3*6
	if got != want {
		t.Errorf("Vec3.Dot() = %v, want %v", got, want)
	}
}

func TestVec3Cross(t *testing.T) {
	tests := []struct {
		name string
		a, b Vec3
		want Vec3
	}{
		{"X cross Y = Z", Vec3{1, 0, 0}, Vec3{0, 1, 0}, Vec3{0, 0, 1}},
		{"Y cross Z = X", Vec3{0, 1, 0}, Vec3{0, 0, 1}, Vec3{1, 0, 0}},
		{"Z cross X = Y", Vec3{0, 0, 1}, Vec3{1, 0, 0}, Vec3{0, 1, 0}},
		{"Y cross X = -Z", Vec3{0, 1, 0}, Vec3{1, 0, 0}, Vec3{0, 0, -1}},
		{"parallel = zero", Vec3{1, 0, 0}, Vec3{2, 0, 0}, Vec3{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Cross(tt.b)
			if !approxEqualVec3(got, tt.want, epsilon) {
				t.Errorf("Vec3.Cross() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec3Length(t *testing.T) {
	tests := []struct {
		name string
		v    Vec3
		want float32
	}{
		{"unit X", Vec3{1, 0, 0}, 1},
		{"3-4-5", Vec3{3, 4, 0}, 5},
		{"zero", Vec3{0, 0, 0}, 0},
		{"all ones", Vec3{1, 1, 1}, float32(math.Sqrt(3))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Length()
			if !approxEqual(got, tt.want, epsilon) {
				t.Errorf("Vec3.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec3Normalize(t *testing.T) {
	tests := []struct {
		name string
		v    Vec3
		want Vec3
	}{
		{"unit X", Vec3{5, 0, 0}, Vec3{1, 0, 0}},
		{"unit Y", Vec3{0, 3, 0}, Vec3{0, 1, 0}},
		{"unit Z", Vec3{0, 0, 7}, Vec3{0, 0, 1}},
		{"zero", Vec3{0, 0, 0}, Vec3{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Normalize()
			if !approxEqualVec3(got, tt.want, epsilon) {
				t.Errorf("Vec3.Normalize() = %v, want %v", got, tt.want)
			}
			// Verify unit length (except zero vector)
			if tt.v != (Vec3{}) {
				l := got.Length()
				if !approxEqual(l, 1, epsilon) {
					t.Errorf("normalized length = %v, want 1", l)
				}
			}
		})
	}
}

func TestVec3Negate(t *testing.T) {
	got := Vec3{1, -2, 3}.Negate()
	want := Vec3{-1, 2, -3}
	if got != want {
		t.Errorf("Vec3.Negate() = %v, want %v", got, want)
	}
}

func TestVec3Distance(t *testing.T) {
	a := Vec3{1, 0, 0}
	b := Vec3{4, 0, 0}
	got := a.Distance(b)
	if !approxEqual(got, 3, epsilon) {
		t.Errorf("Vec3.Distance() = %v, want 3", got)
	}
}

func TestVec3DistanceSq(t *testing.T) {
	a := Vec3{1, 0, 0}
	b := Vec3{4, 0, 0}
	got := a.DistanceSq(b)
	if !approxEqual(got, 9, epsilon) {
		t.Errorf("Vec3.DistanceSq() = %v, want 9", got)
	}
}

func TestVec3Lerp(t *testing.T) {
	a := Vec3{0, 0, 0}
	b := Vec3{10, 20, 30}
	tests := []struct {
		name string
		t    float32
		want Vec3
	}{
		{"t=0", 0, Vec3{0, 0, 0}},
		{"t=0.5", 0.5, Vec3{5, 10, 15}},
		{"t=1", 1, Vec3{10, 20, 30}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.Lerp(b, tt.t)
			if !approxEqualVec3(got, tt.want, epsilon) {
				t.Errorf("Vec3.Lerp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec3MinMax(t *testing.T) {
	a := Vec3{1, 5, 3}
	b := Vec3{4, 2, 6}

	gotMin := a.Min(b)
	wantMin := Vec3{1, 2, 3}
	if !approxEqualVec3(gotMin, wantMin, epsilon) {
		t.Errorf("Vec3.Min() = %v, want %v", gotMin, wantMin)
	}

	gotMax := a.Max(b)
	wantMax := Vec3{4, 5, 6}
	if !approxEqualVec3(gotMax, wantMax, epsilon) {
		t.Errorf("Vec3.Max() = %v, want %v", gotMax, wantMax)
	}
}

// --- Vec4 Tests ---

func TestVec4Add(t *testing.T) {
	got := Vec4{1, 2, 3, 4}.Add(Vec4{5, 6, 7, 8})
	want := Vec4{6, 8, 10, 12}
	if got != want {
		t.Errorf("Vec4.Add() = %v, want %v", got, want)
	}
}

func TestVec4Sub(t *testing.T) {
	got := Vec4{5, 6, 7, 8}.Sub(Vec4{1, 2, 3, 4})
	want := Vec4{4, 4, 4, 4}
	if got != want {
		t.Errorf("Vec4.Sub() = %v, want %v", got, want)
	}
}

func TestVec4Scale(t *testing.T) {
	got := Vec4{1, 2, 3, 4}.Scale(2)
	want := Vec4{2, 4, 6, 8}
	if got != want {
		t.Errorf("Vec4.Scale() = %v, want %v", got, want)
	}
}

func TestVec4Dot(t *testing.T) {
	got := Vec4{1, 2, 3, 4}.Dot(Vec4{5, 6, 7, 8})
	want := float32(70) // 1*5 + 2*6 + 3*7 + 4*8
	if got != want {
		t.Errorf("Vec4.Dot() = %v, want %v", got, want)
	}
}

func TestVec4Length(t *testing.T) {
	v := Vec4{1, 0, 0, 0}
	if !approxEqual(v.Length(), 1, epsilon) {
		t.Errorf("Vec4{1,0,0,0}.Length() = %v, want 1", v.Length())
	}
}

func TestVec4Normalize(t *testing.T) {
	v := Vec4{3, 0, 0, 0}
	got := v.Normalize()
	want := Vec4{1, 0, 0, 0}
	if !approxEqualVec4(got, want, epsilon) {
		t.Errorf("Vec4.Normalize() = %v, want %v", got, want)
	}

	// Zero vector case
	zero := Vec4{0, 0, 0, 0}
	if zero.Normalize() != (Vec4{}) {
		t.Errorf("zero Vec4.Normalize() should return zero vector")
	}
}

func TestVec4Lerp(t *testing.T) {
	a := Vec4{0, 0, 0, 0}
	b := Vec4{4, 8, 12, 16}
	got := a.Lerp(b, 0.25)
	want := Vec4{1, 2, 3, 4}
	if !approxEqualVec4(got, want, epsilon) {
		t.Errorf("Vec4.Lerp() = %v, want %v", got, want)
	}
}

func TestVec4XYZ(t *testing.T) {
	v := Vec4{1, 2, 3, 4}
	got := v.XYZ()
	want := Vec3{1, 2, 3}
	if got != want {
		t.Errorf("Vec4.XYZ() = %v, want %v", got, want)
	}
}

// --- Utility Tests ---

func TestRadiansDegrees(t *testing.T) {
	tests := []struct {
		name    string
		degrees float32
	}{
		{"0", 0},
		{"90", 90},
		{"180", 180},
		{"360", 360},
		{"45", 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rad := Radians(tt.degrees)
			back := Degrees(rad)
			if !approxEqual(back, tt.degrees, epsilon) {
				t.Errorf("Degrees(Radians(%v)) = %v, want %v", tt.degrees, back, tt.degrees)
			}
		})
	}

	// Verify known conversions
	if !approxEqual(Radians(180), math.Pi, epsilon) {
		t.Errorf("Radians(180) = %v, want Pi", Radians(180))
	}
	if !approxEqual(Radians(90), math.Pi/2, epsilon) {
		t.Errorf("Radians(90) = %v, want Pi/2", Radians(90))
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name          string
		val, min, max float32
		want          float32
	}{
		{"within range", 5, 0, 10, 5},
		{"below min", -5, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
		{"at min", 0, 0, 10, 0},
		{"at max", 10, 0, 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clamp(tt.val, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("Clamp(%v, %v, %v) = %v, want %v", tt.val, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
