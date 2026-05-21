package g3d

import (
	"math"
	"testing"
)

func approxEqualQuat(a, b Quat, eps float32) bool {
	return approxEqual(a.X, b.X, eps) && approxEqual(a.Y, b.Y, eps) &&
		approxEqual(a.Z, b.Z, eps) && approxEqual(a.W, b.W, eps)
}

func TestQuatIdentity(t *testing.T) {
	q := QuatIdentity()
	if q.X != 0 || q.Y != 0 || q.Z != 0 || q.W != 1 {
		t.Errorf("QuatIdentity() = %+v, want {0,0,0,1}", q)
	}
}

func TestQuatFromAxisAngle(t *testing.T) {
	tests := []struct {
		name    string
		axis    Vec3
		radians float32
		want    Quat
	}{
		{
			"90 deg around Y",
			Vec3{0, 1, 0},
			Radians(90),
			Quat{0, float32(math.Sin(math.Pi / 4)), 0, float32(math.Cos(math.Pi / 4))},
		},
		{
			"zero rotation",
			Vec3{1, 0, 0},
			0,
			Quat{0, 0, 0, 1},
		},
		{
			"180 deg around Z",
			Vec3{0, 0, 1},
			Radians(180),
			Quat{0, 0, 1, 0}, // sin(90)=1, cos(90)=0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuatFromAxisAngle(tt.axis, tt.radians)
			if !approxEqualQuat(got, tt.want, 1e-5) {
				t.Errorf("QuatFromAxisAngle(%v, %v) = %+v, want %+v", tt.axis, tt.radians, got, tt.want)
			}
		})
	}
}

func TestQuatNormalize(t *testing.T) {
	q := Quat{1, 2, 3, 4}
	n := q.Normalize()
	l := n.Length()
	if !approxEqual(l, 1.0, epsilon) {
		t.Errorf("Normalize().Length() = %v, want 1", l)
	}

	// Zero quaternion should return identity
	zero := Quat{0, 0, 0, 0}
	got := zero.Normalize()
	if !approxEqualQuat(got, QuatIdentity(), epsilon) {
		t.Errorf("zero Quat.Normalize() = %+v, want identity", got)
	}
}

func TestQuatLength(t *testing.T) {
	q := QuatIdentity()
	if !approxEqual(q.Length(), 1, epsilon) {
		t.Errorf("identity length = %v, want 1", q.Length())
	}

	q2 := Quat{1, 0, 0, 0}
	if !approxEqual(q2.Length(), 1, epsilon) {
		t.Errorf("{1,0,0,0} length = %v, want 1", q2.Length())
	}
}

func TestQuatMul(t *testing.T) {
	// Multiplying by identity should return the same quaternion
	q := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(45))
	id := QuatIdentity()

	result := q.Mul(id)
	if !approxEqualQuat(result, q, epsilon) {
		t.Errorf("q * identity = %+v, want %+v", result, q)
	}

	result = id.Mul(q)
	if !approxEqualQuat(result, q, epsilon) {
		t.Errorf("identity * q = %+v, want %+v", result, q)
	}
}

func TestQuatMulComposition(t *testing.T) {
	// Two 90-degree rotations around Y should equal one 180-degree rotation
	q90 := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(90))
	q180 := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(180))

	composed := q90.Mul(q90)
	// Quaternions q and -q represent the same rotation
	if !approxEqualQuat(composed, q180, 1e-4) &&
		!approxEqualQuat(composed, Quat{-q180.X, -q180.Y, -q180.Z, -q180.W}, 1e-4) {
		t.Errorf("90+90 = %+v, want %+v (or negated)", composed, q180)
	}
}

func TestQuatConjugate(t *testing.T) {
	q := Quat{1, 2, 3, 4}
	c := q.Conjugate()
	want := Quat{-1, -2, -3, 4}
	if c != want {
		t.Errorf("Conjugate() = %+v, want %+v", c, want)
	}
}

func TestQuatInverse(t *testing.T) {
	q := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(45)).Normalize()
	inv := q.Inverse()

	// q * q^-1 should equal identity
	result := q.Mul(inv)
	id := QuatIdentity()
	if !approxEqualQuat(result, id, 1e-4) {
		t.Errorf("q * q^-1 = %+v, want identity", result)
	}

	// Zero quaternion inverse should return identity
	zero := Quat{0, 0, 0, 0}
	got := zero.Inverse()
	if !approxEqualQuat(got, QuatIdentity(), epsilon) {
		t.Errorf("zero Quat.Inverse() = %+v, want identity", got)
	}
}

func TestQuatRotateVec3(t *testing.T) {
	tests := []struct {
		name string
		q    Quat
		v    Vec3
		want Vec3
	}{
		{
			"identity rotation",
			QuatIdentity(),
			Vec3{1, 2, 3},
			Vec3{1, 2, 3},
		},
		{
			"90 deg Y: (0,0,1) -> (1,0,0)",
			QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(90)),
			Vec3{0, 0, 1},
			Vec3{1, 0, 0},
		},
		{
			"90 deg X: (0,1,0) -> (0,0,1)",
			QuatFromAxisAngle(Vec3{1, 0, 0}, Radians(90)),
			Vec3{0, 1, 0},
			Vec3{0, 0, 1},
		},
		{
			"90 deg Z: (1,0,0) -> (0,1,0)",
			QuatFromAxisAngle(Vec3{0, 0, 1}, Radians(90)),
			Vec3{1, 0, 0},
			Vec3{0, 1, 0},
		},
		{
			"180 deg Y: (1,0,0) -> (-1,0,0)",
			QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(180)),
			Vec3{1, 0, 0},
			Vec3{-1, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.q.RotateVec3(tt.v)
			if !approxEqualVec3(got, tt.want, 1e-4) {
				t.Errorf("RotateVec3(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestQuatToMat4_ConsistentWithRotateVec3(t *testing.T) {
	q := QuatFromAxisAngle(Vec3{1, 1, 0}.Normalize(), Radians(60))
	v := Vec3{1, 2, 3}

	// Both paths should give same result
	byQuat := q.RotateVec3(v)
	m := q.ToMat4()
	byMat := m.MulVec4(Vec4{v.X, v.Y, v.Z, 1}).XYZ()

	if !approxEqualVec3(byQuat, byMat, 1e-4) {
		t.Errorf("RotateVec3 = %v, ToMat4*v = %v — should match", byQuat, byMat)
	}
}

func TestQuatToMat4_Identity(t *testing.T) {
	q := QuatIdentity()
	m := q.ToMat4()
	id := Mat4Identity()
	if !approxEqualMat4(m, id, epsilon) {
		t.Errorf("QuatIdentity().ToMat4() should equal Mat4Identity()")
	}
}

func TestQuatDot(t *testing.T) {
	a := Quat{1, 0, 0, 0}
	b := Quat{0, 1, 0, 0}
	if a.Dot(b) != 0 {
		t.Errorf("orthogonal quaternion dot = %v, want 0", a.Dot(b))
	}

	c := a.Dot(a)
	if !approxEqual(c, 1, epsilon) {
		t.Errorf("self dot = %v, want 1", c)
	}
}

// --- Euler round-trip tests ---

func TestQuatEulerRoundtrip(t *testing.T) {
	tests := []struct {
		name  string
		euler Euler
	}{
		{"zero", Euler{0, 0, 0}},
		{"X only", Euler{Radians(45), 0, 0}},
		{"Y only", Euler{0, Radians(30), 0}},
		{"Z only", Euler{0, 0, Radians(60)}},
		{"XYZ", Euler{Radians(20), Radians(30), Radians(40)}},
		{"negative", Euler{Radians(-45), Radians(-30), Radians(-60)}},
		{"small angles", Euler{0.01, 0.02, 0.03}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := QuatFromEuler(tt.euler)
			back := q.ToEuler()
			qBack := QuatFromEuler(back)

			// The rotation should be equivalent even if Euler angles differ
			// (Euler angles are not unique). Check by rotating a test vector.
			testVec := Vec3{1, 2, 3}
			v1 := q.RotateVec3(testVec)
			v2 := qBack.RotateVec3(testVec)
			if !approxEqualVec3(v1, v2, 1e-3) {
				t.Errorf("Euler roundtrip: original rotation(%v) = %v, roundtrip rotation(%v) = %v",
					tt.euler, v1, back, v2)
			}
		})
	}
}

func TestQuatFromEuler_KnownValues(t *testing.T) {
	// Pure X rotation of 90 degrees
	q := QuatFromEuler(Euler{Radians(90), 0, 0})
	v := q.RotateVec3(Vec3{0, 1, 0})
	want := Vec3{0, 0, 1}
	if !approxEqualVec3(v, want, 1e-4) {
		t.Errorf("Euler(90,0,0) rotating (0,1,0) = %v, want %v", v, want)
	}
}

// --- Slerp tests ---

func TestQuatSlerp(t *testing.T) {
	a := QuatIdentity()
	b := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(90))

	tests := []struct {
		name string
		t    float32
	}{
		{"t=0", 0},
		{"t=0.25", 0.25},
		{"t=0.5", 0.5},
		{"t=0.75", 0.75},
		{"t=1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.Slerp(b, tt.t)

			// Result should be a unit quaternion
			if !approxEqual(result.Length(), 1, 1e-4) {
				t.Errorf("Slerp result length = %v, want 1", result.Length())
			}

			// At endpoints, should match inputs
			if tt.t == 0 {
				if !approxEqualQuat(result, a, epsilon) {
					t.Errorf("Slerp(0) = %+v, want %+v", result, a)
				}
			}
			if tt.t == 1 {
				if !approxEqualQuat(result, b, epsilon) {
					t.Errorf("Slerp(1) = %+v, want %+v", result, b)
				}
			}
		})
	}
}

func TestQuatSlerp_ShortestPath(t *testing.T) {
	// Slerp should take the shortest path even when quaternions are negated
	a := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(10))
	b := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(350))

	// b represents the same rotation as QuatFromAxisAngle(Y, -10) but via long way
	mid := a.Slerp(b, 0.5)

	// The midpoint should be approximately at 0 degrees (short path)
	// or at 180 degrees (long path). Short path is correct.
	testVec := Vec3{0, 0, 1}
	rotated := mid.RotateVec3(testVec)

	// If taking short path, rotated.Z should be close to 1 (small rotation)
	// If taking long path, rotated.Z would be close to -1 (180 deg)
	if rotated.Z < 0 {
		t.Errorf("Slerp appears to take long path: rotated Z = %v", rotated.Z)
	}
}

func TestQuatSlerp_IdenticalQuats(t *testing.T) {
	q := QuatFromAxisAngle(Vec3{1, 0, 0}, Radians(45))
	result := q.Slerp(q, 0.5)
	if !approxEqualQuat(result, q, epsilon) {
		t.Errorf("Slerp of identical quaternions = %+v, want %+v", result, q)
	}
}

func TestQuatSlerp_OppositeQuats(t *testing.T) {
	// Nearly opposite quaternions (180 degrees apart)
	a := QuatIdentity()
	b := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(179))

	result := a.Slerp(b, 0.5)
	// Should not crash or produce NaN
	if math.IsNaN(float64(result.X)) || math.IsNaN(float64(result.W)) {
		t.Errorf("Slerp of nearly-opposite quaternions produced NaN: %+v", result)
	}
	// Should be unit length
	if !approxEqual(result.Length(), 1, 0.01) {
		t.Errorf("Slerp result length = %v, want 1", result.Length())
	}
}

// --- Benchmarks ---

func BenchmarkQuatMul(b *testing.B) {
	a := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(45))
	c := QuatFromAxisAngle(Vec3{1, 0, 0}, Radians(30))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Mul(c)
	}
}

func BenchmarkQuatRotateVec3(b *testing.B) {
	q := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(45))
	v := Vec3{1, 2, 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.RotateVec3(v)
	}
}

func BenchmarkQuatSlerp(b *testing.B) {
	a := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(10))
	c := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(80))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Slerp(c, 0.5)
	}
}

func BenchmarkQuatToMat4(b *testing.B) {
	q := QuatFromAxisAngle(Vec3{1, 1, 0}.Normalize(), Radians(60))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.ToMat4()
	}
}
