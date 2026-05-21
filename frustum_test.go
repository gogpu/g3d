package g3d

import (
	"math"
	"testing"
)

// --- AABB Tests ---

func TestNewAABBFromPoints(t *testing.T) {
	tests := []struct {
		name   string
		points []Vec3
		want   AABB
	}{
		{
			"empty",
			nil,
			AABB{},
		},
		{
			"single point",
			[]Vec3{{1, 2, 3}},
			AABB{Min: Vec3{1, 2, 3}, Max: Vec3{1, 2, 3}},
		},
		{
			"two corners",
			[]Vec3{{-1, -2, -3}, {4, 5, 6}},
			AABB{Min: Vec3{-1, -2, -3}, Max: Vec3{4, 5, 6}},
		},
		{
			"multiple points",
			[]Vec3{{0, 0, 0}, {-1, 5, -3}, {4, -2, 6}, {2, 2, 2}},
			AABB{Min: Vec3{-1, -2, -3}, Max: Vec3{4, 5, 6}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAABBFromPoints(tt.points)
			if !approxEqualVec3(got.Min, tt.want.Min, epsilon) || !approxEqualVec3(got.Max, tt.want.Max, epsilon) {
				t.Errorf("NewAABBFromPoints() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAABBCenter(t *testing.T) {
	box := AABB{Min: Vec3{-1, -2, -3}, Max: Vec3{3, 4, 5}}
	got := box.Center()
	want := Vec3{1, 1, 1}
	if !approxEqualVec3(got, want, epsilon) {
		t.Errorf("AABB.Center() = %v, want %v", got, want)
	}
}

func TestAABBSize(t *testing.T) {
	box := AABB{Min: Vec3{-1, -2, -3}, Max: Vec3{3, 4, 5}}
	got := box.Size()
	want := Vec3{4, 6, 8}
	if !approxEqualVec3(got, want, epsilon) {
		t.Errorf("AABB.Size() = %v, want %v", got, want)
	}
}

func TestAABBMerge(t *testing.T) {
	a := AABB{Min: Vec3{0, 0, 0}, Max: Vec3{1, 1, 1}}
	b := AABB{Min: Vec3{-1, 2, -1}, Max: Vec3{0.5, 3, 0.5}}
	got := a.Merge(b)
	want := AABB{Min: Vec3{-1, 0, -1}, Max: Vec3{1, 3, 1}}
	if !approxEqualVec3(got.Min, want.Min, epsilon) || !approxEqualVec3(got.Max, want.Max, epsilon) {
		t.Errorf("AABB.Merge() = %+v, want %+v", got, want)
	}
}

func TestAABBContainsPoint(t *testing.T) {
	box := AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}}
	tests := []struct {
		name string
		p    Vec3
		want bool
	}{
		{"center", Vec3{0, 0, 0}, true},
		{"corner", Vec3{1, 1, 1}, true},
		{"min corner", Vec3{-1, -1, -1}, true},
		{"outside X", Vec3{2, 0, 0}, false},
		{"outside Y", Vec3{0, -2, 0}, false},
		{"outside Z", Vec3{0, 0, 5}, false},
		{"edge", Vec3{1, 0, 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := box.ContainsPoint(tt.p)
			if got != tt.want {
				t.Errorf("AABB.ContainsPoint(%v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestAABBIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		box  AABB
		want bool
	}{
		{"normal", AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}}, false},
		{"flat X", AABB{Min: Vec3{0, -1, -1}, Max: Vec3{0, 1, 1}}, true},
		{"inverted", AABB{Min: Vec3{1, 1, 1}, Max: Vec3{-1, -1, -1}}, true},
		{"zero", AABB{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.box.IsEmpty()
			if got != tt.want {
				t.Errorf("AABB.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAABBTransform(t *testing.T) {
	box := AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}}

	// Identity transform should not change the AABB
	id := Mat4Identity()
	got := box.Transform(id)
	if !approxEqualVec3(got.Min, box.Min, epsilon) || !approxEqualVec3(got.Max, box.Max, epsilon) {
		t.Errorf("identity transform changed AABB: %+v", got)
	}

	// Translation should shift both min and max
	tr := Mat4Translate(Vec3{5, 0, 0})
	got = box.Transform(tr)
	wantMin := Vec3{4, -1, -1}
	wantMax := Vec3{6, 1, 1}
	if !approxEqualVec3(got.Min, wantMin, epsilon) || !approxEqualVec3(got.Max, wantMax, epsilon) {
		t.Errorf("translate transform: got %+v, want min=%v max=%v", got, wantMin, wantMax)
	}

	// Uniform scale
	sc := Mat4Scale(Vec3{2, 2, 2})
	got = box.Transform(sc)
	wantMin = Vec3{-2, -2, -2}
	wantMax = Vec3{2, 2, 2}
	if !approxEqualVec3(got.Min, wantMin, epsilon) || !approxEqualVec3(got.Max, wantMax, epsilon) {
		t.Errorf("scale transform: got %+v, want min=%v max=%v", got, wantMin, wantMax)
	}

	// 90-degree rotation around Z: unit cube should still be unit cube
	rz := Mat4RotateZ(Radians(90))
	got = box.Transform(rz)
	// After 90-degree Z rotation, the AABB should still be [-1,1] in all axes (symmetric cube)
	if !approxEqualVec3(got.Min, Vec3{-1, -1, -1}, 1e-4) || !approxEqualVec3(got.Max, Vec3{1, 1, 1}, 1e-4) {
		t.Errorf("90-deg Z rotation of unit cube: got %+v, want [-1,-1,-1] to [1,1,1]", got)
	}
}

func TestAABBHalfExtents(t *testing.T) {
	box := AABB{Min: Vec3{-2, -3, -4}, Max: Vec3{2, 3, 4}}
	got := box.HalfExtents()
	want := Vec3{2, 3, 4}
	if !approxEqualVec3(got, want, epsilon) {
		t.Errorf("AABB.HalfExtents() = %v, want %v", got, want)
	}
}

func TestAABBVolume(t *testing.T) {
	box := AABB{Min: Vec3{0, 0, 0}, Max: Vec3{2, 3, 4}}
	got := box.Volume()
	want := float32(24) // 2*3*4
	if !approxEqual(got, want, epsilon) {
		t.Errorf("AABB.Volume() = %v, want %v", got, want)
	}
}

func TestAABBSurfaceArea(t *testing.T) {
	box := AABB{Min: Vec3{0, 0, 0}, Max: Vec3{2, 3, 4}}
	// SA = 2*(2*3 + 3*4 + 4*2) = 2*(6+12+8) = 52
	got := box.SurfaceArea()
	want := float32(52)
	if !approxEqual(got, want, epsilon) {
		t.Errorf("AABB.SurfaceArea() = %v, want %v", got, want)
	}
}

func TestAABBClosestPoint(t *testing.T) {
	box := AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}}
	tests := []struct {
		name string
		p    Vec3
		want Vec3
	}{
		{"inside", Vec3{0.5, 0.5, 0.5}, Vec3{0.5, 0.5, 0.5}},
		{"outside X", Vec3{5, 0, 0}, Vec3{1, 0, 0}},
		{"outside negative", Vec3{-5, -5, -5}, Vec3{-1, -1, -1}},
		{"on face", Vec3{0, 0, 1}, Vec3{0, 0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := box.ClosestPoint(tt.p)
			if !approxEqualVec3(got, tt.want, epsilon) {
				t.Errorf("ClosestPoint(%v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

// --- Plane Tests ---

func TestPlaneDistanceToPoint(t *testing.T) {
	// XZ plane at Y=0, normal pointing up
	p := Plane{Normal: Vec3{0, 1, 0}, D: 0}

	above := Vec3{0, 5, 0}
	if !approxEqual(p.DistanceToPoint(above), 5, epsilon) {
		t.Errorf("distance to (0,5,0) = %v, want 5", p.DistanceToPoint(above))
	}

	below := Vec3{0, -3, 0}
	if !approxEqual(p.DistanceToPoint(below), -3, epsilon) {
		t.Errorf("distance to (0,-3,0) = %v, want -3", p.DistanceToPoint(below))
	}

	onPlane := Vec3{5, 0, 10}
	if !approxEqual(p.DistanceToPoint(onPlane), 0, epsilon) {
		t.Errorf("distance to point on plane = %v, want 0", p.DistanceToPoint(onPlane))
	}
}

// --- Frustum Tests ---

func TestFrustumFromMat4_Perspective(t *testing.T) {
	view := Mat4LookAt(Vec3{0, 0, 5}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Perspective(Radians(90), 1, 0.1, 100)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)

	// All planes should have normalized normals (length ~1)
	for i := 0; i < 6; i++ {
		l := f[i].Normal.Length()
		if !approxEqual(l, 1, 0.01) {
			t.Errorf("plane[%d] normal length = %v, want 1", i, l)
		}
	}
}

func TestFrustumContainsPoint(t *testing.T) {
	// Camera at (0,0,5) looking at origin, 90-degree FOV, near=0.1, far=100
	view := Mat4LookAt(Vec3{0, 0, 5}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Perspective(Radians(90), 1, 0.1, 100)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)

	tests := []struct {
		name string
		p    Vec3
		want bool
	}{
		{"origin (in front)", Vec3{0, 0, 0}, true},
		{"near camera", Vec3{0, 0, 4.8}, true},
		{"behind camera", Vec3{0, 0, 10}, false},
		{"far away along Z", Vec3{0, 0, -90}, true},
		{"beyond far plane (distance > 100 from camera)", Vec3{0, 0, -96}, false},
		{"way beyond far", Vec3{0, 0, -200}, false},
		{"to the side", Vec3{100, 0, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.ContainsPoint(tt.p)
			if got != tt.want {
				t.Errorf("Frustum.ContainsPoint(%v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestFrustumIntersectsAABB(t *testing.T) {
	// Camera at (0,0,10) looking at origin
	view := Mat4LookAt(Vec3{0, 0, 10}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Perspective(Radians(60), 1.5, 1, 100)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)

	tests := []struct {
		name string
		box  AABB
		want bool
	}{
		{
			"at origin (inside)",
			AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}},
			true,
		},
		{
			"far behind camera",
			AABB{Min: Vec3{-1, -1, 20}, Max: Vec3{1, 1, 22}},
			false,
		},
		{
			"way to the right",
			AABB{Min: Vec3{500, 500, -5}, Max: Vec3{502, 502, -3}},
			false,
		},
		{
			"beyond far plane",
			AABB{Min: Vec3{-1, -1, -200}, Max: Vec3{1, 1, -198}},
			false,
		},
		{
			"straddling near plane",
			AABB{Min: Vec3{-0.5, -0.5, 8.5}, Max: Vec3{0.5, 0.5, 9.5}},
			true,
		},
		{
			"large box overlapping",
			AABB{Min: Vec3{-100, -100, -100}, Max: Vec3{100, 100, 100}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.IntersectsAABB(tt.box)
			if got != tt.want {
				t.Errorf("Frustum.IntersectsAABB(%+v) = %v, want %v", tt.box, got, tt.want)
			}
		})
	}
}

func TestFrustumIntersectsAABB_Orthographic(t *testing.T) {
	view := Mat4LookAt(Vec3{0, 0, 5}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Ortho(-10, 10, -10, 10, 0.1, 100)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)

	tests := []struct {
		name string
		box  AABB
		want bool
	}{
		{
			"inside",
			AABB{Min: Vec3{-1, -1, 0}, Max: Vec3{1, 1, 2}},
			true,
		},
		{
			"outside lateral",
			AABB{Min: Vec3{15, 15, 0}, Max: Vec3{20, 20, 2}},
			false,
		},
		{
			"behind camera",
			AABB{Min: Vec3{-1, -1, 10}, Max: Vec3{1, 1, 12}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.IntersectsAABB(tt.box)
			if got != tt.want {
				t.Errorf("Ortho Frustum.IntersectsAABB(%+v) = %v, want %v", tt.box, got, tt.want)
			}
		})
	}
}

// --- Benchmarks ---

func BenchmarkFrustumIntersectsAABB(b *testing.B) {
	view := Mat4LookAt(Vec3{0, 0, 10}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Perspective(Radians(60), 1.5, 0.1, 1000)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)
	box := AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.IntersectsAABB(box)
	}
}

func BenchmarkFrustumFromMat4(b *testing.B) {
	view := Mat4LookAt(Vec3{0, 0, 10}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Perspective(Radians(60), 1.5, 0.1, 1000)
	vp := proj.Mul(view)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FrustumFromMat4(vp)
	}
}

func BenchmarkAABBTransform(b *testing.B) {
	box := AABB{Min: Vec3{-1, -1, -1}, Max: Vec3{1, 1, 1}}
	m := Mat4Translate(Vec3{5, 0, 0}).Mul(Mat4RotateY(0.5)).Mul(Mat4Scale(Vec3{2, 2, 2}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = box.Transform(m)
	}
}

// Verify frustum planes are correct by testing a known geometry
func TestFrustumFromMat4_PlaneCount(t *testing.T) {
	view := Mat4Identity()
	proj := Mat4Perspective(Radians(60), 1, 0.1, 100)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)

	// Should have exactly 6 planes
	if len(f) != 6 {
		t.Errorf("frustum has %d planes, want 6", len(f))
	}

	// All normals should be non-zero
	for i := 0; i < 6; i++ {
		if f[i].Normal.LengthSq() < 0.01 {
			t.Errorf("plane[%d] has near-zero normal", i)
		}
	}
}

// Edge case: degenerate (zero) matrix
func TestFrustumFromMat4_ZeroMatrix(t *testing.T) {
	var zero Mat4
	f := FrustumFromMat4(zero)
	// Should not panic. Planes will be degenerate but that is acceptable.
	_ = f
}

// Test that frustum culling is conservative (no false negatives)
func TestFrustumIntersectsAABB_NoFalseNegatives(t *testing.T) {
	view := Mat4LookAt(Vec3{0, 0, 5}, Vec3{0, 0, 0}, Vec3{0, 1, 0})
	proj := Mat4Perspective(Radians(90), 1, 0.1, 100)
	vp := proj.Mul(view)
	f := FrustumFromMat4(vp)

	// A box centered at origin with small size should definitely be inside
	smallBox := AABB{Min: Vec3{-0.1, -0.1, -0.1}, Max: Vec3{0.1, 0.1, 0.1}}
	if !f.IntersectsAABB(smallBox) {
		t.Error("small box at origin should be inside frustum (false negative)")
	}

	// Points distributed inside the frustum should all pass
	for z := float32(-80); z <= 4; z += 10 {
		box := AABB{
			Min: Vec3{-0.01, -0.01, z - 0.01},
			Max: Vec3{0.01, 0.01, z + 0.01},
		}
		if !f.IntersectsAABB(box) {
			t.Errorf("box near Z=%v should be inside frustum (false negative)", z)
		}
	}
}

// Edge case: selectF helper
func TestSelectF(t *testing.T) {
	if selectF(true, 1, 2) != 1 {
		t.Error("selectF(true, 1, 2) should return 1")
	}
	if selectF(false, 1, 2) != 2 {
		t.Error("selectF(false, 1, 2) should return 2")
	}
}

// Vec3 component helpers
func TestVec3Component(t *testing.T) {
	v := Vec3{1, 2, 3}
	if v.component(0) != 1 || v.component(1) != 2 || v.component(2) != 3 {
		t.Errorf("Vec3.component() failed: %v %v %v", v.component(0), v.component(1), v.component(2))
	}
}

func TestSetComponent(t *testing.T) {
	var v Vec3
	setComponent(&v, 0, 10)
	setComponent(&v, 1, 20)
	setComponent(&v, 2, 30)
	want := Vec3{10, 20, 30}
	if v != want {
		t.Errorf("setComponent: got %v, want %v", v, want)
	}
}

func TestNormalizePlane(t *testing.T) {
	p := Plane{Normal: Vec3{0, 3, 0}, D: 6}
	n := normalizePlane(p)
	if !approxEqual(n.Normal.Length(), 1, epsilon) {
		t.Errorf("normalized plane normal length = %v, want 1", n.Normal.Length())
	}
	if !approxEqual(n.D, 2, epsilon) {
		t.Errorf("normalized plane D = %v, want 2", n.D)
	}

	// Zero normal should not crash
	zero := Plane{Normal: Vec3{0, 0, 0}, D: 5}
	nz := normalizePlane(zero)
	if nz.D != 5 {
		t.Errorf("zero normal plane D changed: %v", nz.D)
	}
}

// Suppress unused import warning
var _ = math.Pi
