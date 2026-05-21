package g3d

import (
	"math"
	"testing"
)

func approxEqualMat4(a, b Mat4, eps float32) bool {
	for i := 0; i < 16; i++ {
		if !approxEqual(a[i], b[i], eps) {
			return false
		}
	}
	return true
}

func TestMat4Identity(t *testing.T) {
	m := Mat4Identity()
	want := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	if m != want {
		t.Errorf("Mat4Identity() = %v, want %v", m, want)
	}
}

func TestMat4ColumnMajorLayout(t *testing.T) {
	// Verify column-major layout: m[col*4+row]
	m := Mat4Identity()

	// Column 0
	if m[0] != 1 || m[1] != 0 || m[2] != 0 || m[3] != 0 {
		t.Error("Column 0 of identity should be (1,0,0,0)")
	}
	// Column 1
	if m[4] != 0 || m[5] != 1 || m[6] != 0 || m[7] != 0 {
		t.Error("Column 1 of identity should be (0,1,0,0)")
	}
	// Column 2
	if m[8] != 0 || m[9] != 0 || m[10] != 1 || m[11] != 0 {
		t.Error("Column 2 of identity should be (0,0,1,0)")
	}
	// Column 3
	if m[12] != 0 || m[13] != 0 || m[14] != 0 || m[15] != 1 {
		t.Error("Column 3 of identity should be (0,0,0,1)")
	}
}

func TestMat4Translate(t *testing.T) {
	m := Mat4Translate(Vec3{3, 4, 5})

	// Translation should be in column 3
	if m[12] != 3 || m[13] != 4 || m[14] != 5 {
		t.Errorf("Translation column = (%v, %v, %v), want (3, 4, 5)", m[12], m[13], m[14])
	}

	// Should transform a point correctly
	p := Vec4{1, 2, 3, 1}
	result := m.MulVec4(p)
	want := Vec4{4, 6, 8, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("Translate * point = %v, want %v", result, want)
	}

	// Direction (w=0) should be unaffected by translation
	d := Vec4{1, 0, 0, 0}
	result = m.MulVec4(d)
	if !approxEqualVec4(result, d, epsilon) {
		t.Errorf("Translate * direction = %v, want %v", result, d)
	}
}

func TestMat4TranslateExtract(t *testing.T) {
	v := Vec3{10, 20, 30}
	m := Mat4Translate(v)
	got := m.Translation()
	if !approxEqualVec3(got, v, epsilon) {
		t.Errorf("Translation() = %v, want %v", got, v)
	}
}

func TestMat4Scale(t *testing.T) {
	m := Mat4Scale(Vec3{2, 3, 4})
	p := Vec4{1, 1, 1, 1}
	result := m.MulVec4(p)
	want := Vec4{2, 3, 4, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("Scale * point = %v, want %v", result, want)
	}
}

func TestMat4RotateX(t *testing.T) {
	m := Mat4RotateX(Radians(90))
	// Rotating Y-axis unit vector by 90 degrees around X should give Z-axis
	p := Vec4{0, 1, 0, 1}
	result := m.MulVec4(p)
	want := Vec4{0, 0, 1, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("RotateX(90) * (0,1,0) = %v, want %v", result, want)
	}
}

func TestMat4RotateY(t *testing.T) {
	m := Mat4RotateY(Radians(90))
	// Rotating Z-axis unit vector by 90 degrees around Y should give X-axis
	p := Vec4{0, 0, 1, 1}
	result := m.MulVec4(p)
	want := Vec4{1, 0, 0, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("RotateY(90) * (0,0,1) = %v, want %v", result, want)
	}
}

func TestMat4RotateZ(t *testing.T) {
	m := Mat4RotateZ(Radians(90))
	// Rotating X-axis unit vector by 90 degrees around Z should give Y-axis
	p := Vec4{1, 0, 0, 1}
	result := m.MulVec4(p)
	want := Vec4{0, 1, 0, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("RotateZ(90) * (1,0,0) = %v, want %v", result, want)
	}
}

func TestMat4RotateIdentity(t *testing.T) {
	// Rotation by 0 should give identity
	tests := []struct {
		name string
		m    Mat4
	}{
		{"RotateX(0)", Mat4RotateX(0)},
		{"RotateY(0)", Mat4RotateY(0)},
		{"RotateZ(0)", Mat4RotateZ(0)},
	}
	id := Mat4Identity()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !approxEqualMat4(tt.m, id, epsilon) {
				t.Errorf("%s should equal identity, got %v", tt.name, tt.m)
			}
		})
	}
}

func TestMat4MulIdentity(t *testing.T) {
	id := Mat4Identity()
	m := Mat4Translate(Vec3{1, 2, 3})

	// Identity * M = M
	result := id.Mul(m)
	if !approxEqualMat4(result, m, epsilon) {
		t.Errorf("Identity * M != M")
	}

	// M * Identity = M
	result = m.Mul(id)
	if !approxEqualMat4(result, m, epsilon) {
		t.Errorf("M * Identity != M")
	}
}

func TestMat4MulTranslateScale(t *testing.T) {
	// TRS composition: translate after scale
	s := Mat4Scale(Vec3{2, 2, 2})
	tr := Mat4Translate(Vec3{1, 0, 0})
	m := tr.Mul(s) // first scale, then translate

	p := Vec4{1, 0, 0, 1}
	result := m.MulVec4(p)
	// Scale(2) * (1,0,0) = (2,0,0), then Translate(1,0,0) = (3,0,0)
	want := Vec4{3, 0, 0, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("Translate*Scale * point = %v, want %v", result, want)
	}
}

func TestMat4Transpose(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	tr := m.Transpose()

	// Verify double-transpose returns original
	if !approxEqualMat4(tr.Transpose(), m, epsilon) {
		t.Error("double transpose should return original matrix")
	}

	// Identity transpose = identity
	id := Mat4Identity()
	if !approxEqualMat4(id.Transpose(), id, epsilon) {
		t.Error("identity transpose should equal identity")
	}
}

func TestMat4Determinant(t *testing.T) {
	// Identity determinant = 1
	det := Mat4Identity().Determinant()
	if !approxEqual(det, 1, epsilon) {
		t.Errorf("identity determinant = %v, want 1", det)
	}

	// Scale matrix determinant = product of scales
	s := Mat4Scale(Vec3{2, 3, 4})
	det = s.Determinant()
	if !approxEqual(det, 24, epsilon) {
		t.Errorf("scale(2,3,4) determinant = %v, want 24", det)
	}

	// Translation matrix determinant = 1
	tr := Mat4Translate(Vec3{10, 20, 30})
	det = tr.Determinant()
	if !approxEqual(det, 1, epsilon) {
		t.Errorf("translation determinant = %v, want 1", det)
	}

	// Singular matrix determinant = 0
	var zero Mat4
	det = zero.Determinant()
	if det != 0 {
		t.Errorf("zero matrix determinant = %v, want 0", det)
	}
}

func TestMat4Inverse(t *testing.T) {
	tests := []struct {
		name string
		m    Mat4
	}{
		{"identity", Mat4Identity()},
		{"translate", Mat4Translate(Vec3{5, 10, 15})},
		{"scale", Mat4Scale(Vec3{2, 3, 4})},
		{"rotateX", Mat4RotateX(0.7)},
		{"rotateY", Mat4RotateY(1.2)},
		{"rotateZ", Mat4RotateZ(2.1)},
	}

	id := Mat4Identity()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := tt.m.Inverse()
			product := tt.m.Mul(inv)
			if !approxEqualMat4(product, id, 1e-4) {
				t.Errorf("M * M^-1 should equal identity for %s\nM = %v\nM^-1 = %v\nM*M^-1 = %v",
					tt.name, tt.m, inv, product)
			}
		})
	}

	// Verify singular matrix returns zero
	var zero Mat4
	inv := zero.Inverse()
	if inv != (Mat4{}) {
		t.Errorf("inverse of zero matrix should be zero matrix, got %v", inv)
	}
}

func TestMat4InverseTRS(t *testing.T) {
	// Composite TRS matrix
	m := Mat4Translate(Vec3{3, 4, 5}).
		Mul(Mat4RotateY(0.5)).
		Mul(Mat4Scale(Vec3{2, 2, 2}))

	inv := m.Inverse()
	product := m.Mul(inv)
	id := Mat4Identity()
	if !approxEqualMat4(product, id, 1e-4) {
		t.Errorf("TRS * TRS^-1 should equal identity")
	}
}

// --- Projection Matrix Tests (CRITICAL: WebGPU Z [0,1]) ---

func TestMat4Perspective_WebGPU_ClipSpace(t *testing.T) {
	fov := Radians(90)
	aspect := float32(1.0)
	near := float32(0.1)
	far := float32(100.0)

	m := Mat4Perspective(fov, aspect, near, far)

	// Verify key elements for WebGPU Z [0,1]:
	// m[10] = -far/(far-near) = -100/99.9
	// m[14] = -(far*near)/(far-near) = -10/99.9
	// m[11] = -1 (perspective divide)

	wantM10 := -far / (far - near)
	wantM14 := -(far * near) / (far - near)

	if !approxEqual(m[10], wantM10, epsilon) {
		t.Errorf("m[10] = %v, want %v (WebGPU Z [0,1])", m[10], wantM10)
	}
	if !approxEqual(m[14], wantM14, epsilon) {
		t.Errorf("m[14] = %v, want %v (WebGPU Z [0,1])", m[14], wantM14)
	}
	if m[11] != -1 {
		t.Errorf("m[11] = %v, want -1 (perspective divide)", m[11])
	}
}

func TestMat4Perspective_NearFarMapping(t *testing.T) {
	fov := Radians(60)
	aspect := float32(16.0 / 9.0)
	near := float32(0.1)
	far := float32(1000.0)

	m := Mat4Perspective(fov, aspect, near, far)

	// A point on the near plane (0,0,-near) should map to z_ndc = 0 after perspective divide
	nearPoint := Vec4{0, 0, -near, 1}
	clipNear := m.MulVec4(nearPoint)
	ndcNearZ := clipNear.Z / clipNear.W
	if !approxEqual(ndcNearZ, 0, 1e-4) {
		t.Errorf("near plane maps to z_ndc = %v, want 0 (WebGPU Z [0,1])", ndcNearZ)
	}

	// A point on the far plane (0,0,-far) should map to z_ndc = 1 after perspective divide
	farPoint := Vec4{0, 0, -far, 1}
	clipFar := m.MulVec4(farPoint)
	ndcFarZ := clipFar.Z / clipFar.W
	if !approxEqual(ndcFarZ, 1, 1e-4) {
		t.Errorf("far plane maps to z_ndc = %v, want 1 (WebGPU Z [0,1])", ndcFarZ)
	}
}

func TestMat4Perspective_NotOpenGL(t *testing.T) {
	// Verify we are NOT using OpenGL convention Z [-1,1]
	near := float32(1.0)
	far := float32(100.0)
	m := Mat4Perspective(Radians(90), 1, near, far)

	// OpenGL would have m[10] = -(far+near)/(far-near) = -101/99
	openGLM10 := -(far + near) / (far - near)
	if approxEqual(m[10], openGLM10, 0.01) {
		t.Errorf("m[10] = %v matches OpenGL formula %v — should be WebGPU!", m[10], openGLM10)
	}
}

func TestMat4Ortho_WebGPU_ClipSpace(t *testing.T) {
	left := float32(-10)
	right := float32(10)
	bottom := float32(-10)
	top := float32(10)
	near := float32(0.1)
	far := float32(100.0)

	m := Mat4Ortho(left, right, bottom, top, near, far)

	// Near plane (z=-near) should map to z_ndc = 0
	nearPoint := Vec4{0, 0, -near, 1}
	result := m.MulVec4(nearPoint)
	if !approxEqual(result.Z, 0, 1e-4) {
		t.Errorf("ortho near plane maps to z = %v, want 0 (WebGPU Z [0,1])", result.Z)
	}

	// Far plane (z=-far) should map to z_ndc = 1
	farPoint := Vec4{0, 0, -far, 1}
	result = m.MulVec4(farPoint)
	if !approxEqual(result.Z, 1, 1e-4) {
		t.Errorf("ortho far plane maps to z = %v, want 1 (WebGPU Z [0,1])", result.Z)
	}
}

func TestMat4Ortho_XYMapping(t *testing.T) {
	m := Mat4Ortho(-1, 1, -1, 1, 0.1, 100)

	// Center should map to (0,0)
	center := Vec4{0, 0, -1, 1}
	result := m.MulVec4(center)
	if !approxEqual(result.X, 0, epsilon) || !approxEqual(result.Y, 0, epsilon) {
		t.Errorf("ortho center maps to (%v, %v), want (0, 0)", result.X, result.Y)
	}

	// Left edge should map to X = -1
	leftEdge := Vec4{-1, 0, -1, 1}
	result = m.MulVec4(leftEdge)
	if !approxEqual(result.X, -1, epsilon) {
		t.Errorf("ortho left edge maps to X = %v, want -1", result.X)
	}

	// Right edge should map to X = 1
	rightEdge := Vec4{1, 0, -1, 1}
	result = m.MulVec4(rightEdge)
	if !approxEqual(result.X, 1, epsilon) {
		t.Errorf("ortho right edge maps to X = %v, want 1", result.X)
	}
}

func TestMat4LookAt(t *testing.T) {
	eye := Vec3{0, 0, 5}
	target := Vec3{0, 0, 0}
	up := Vec3{0, 1, 0}
	view := Mat4LookAt(eye, target, up)

	// The camera is at (0,0,5) looking at origin.
	// Origin should be at (0,0,-5) in camera space (right-handed, camera looks down -Z).
	originInView := view.MulVec4(Vec4{0, 0, 0, 1})
	if !approxEqual(originInView.Z, -5, epsilon) {
		t.Errorf("origin Z in view space = %v, want -5", originInView.Z)
	}
	if !approxEqual(originInView.X, 0, epsilon) || !approxEqual(originInView.Y, 0, epsilon) {
		t.Errorf("origin XY in view space = (%v, %v), want (0, 0)", originInView.X, originInView.Y)
	}

	// The camera position itself should map to origin in view space
	eyeInView := view.MulVec4(Vec4{eye.X, eye.Y, eye.Z, 1})
	if !approxEqualVec3(eyeInView.XYZ(), Vec3{0, 0, 0}, epsilon) {
		t.Errorf("eye in view space = %v, want (0,0,0)", eyeInView.XYZ())
	}
}

func TestMat4LookAt_UpAxis(t *testing.T) {
	// Looking along +X with Y up
	eye := Vec3{0, 0, 0}
	target := Vec3{1, 0, 0}
	up := Vec3{0, 1, 0}
	view := Mat4LookAt(eye, target, up)

	// A point at (1,0,0) should be at (0,0,-1) in view space
	p := Vec4{1, 0, 0, 1}
	result := view.MulVec4(p)
	if !approxEqual(result.Z, -1, epsilon) {
		t.Errorf("point along +X in view space Z = %v, want -1", result.Z)
	}
}

func TestMat4MulVec4(t *testing.T) {
	// Identity * v = v
	id := Mat4Identity()
	v := Vec4{1, 2, 3, 4}
	result := id.MulVec4(v)
	if !approxEqualVec4(result, v, epsilon) {
		t.Errorf("Identity * v = %v, want %v", result, v)
	}
}

func TestMat4FromQuat(t *testing.T) {
	// 90 degree rotation around Y axis
	q := QuatFromAxisAngle(Vec3{0, 1, 0}, Radians(90))
	m := Mat4FromQuat(q)

	// Should transform (0,0,1) to (1,0,0) approximately
	p := Vec4{0, 0, 1, 1}
	result := m.MulVec4(p)
	want := Vec4{1, 0, 0, 1}
	if !approxEqualVec4(result, want, epsilon) {
		t.Errorf("FromQuat(Y,90) * (0,0,1) = %v, want %v", result, want)
	}
}

func TestMat4Perspective_FOV(t *testing.T) {
	// 90-degree FOV with aspect 1:1 should map a point at 45 degrees to the edge
	m := Mat4Perspective(Radians(90), 1, 0.1, 100)

	// Point at (1, 1, -1) is on the edge of 90-degree FOV
	p := Vec4{1, 1, -1, 1}
	clip := m.MulVec4(p)
	ndcX := clip.X / clip.W
	ndcY := clip.Y / clip.W

	// At the exact edge, NDC should be +-1
	if !approxEqual(ndcX, 1, 0.01) {
		t.Errorf("90-deg FOV edge point NDC X = %v, want 1", ndcX)
	}
	if !approxEqual(ndcY, 1, 0.01) {
		t.Errorf("90-deg FOV edge point NDC Y = %v, want 1", ndcY)
	}
}

// --- Benchmarks ---

func BenchmarkMat4Mul(b *testing.B) {
	a := Mat4Translate(Vec3{1, 2, 3})
	c := Mat4RotateY(0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Mul(c)
	}
}

func BenchmarkMat4MulVec4(b *testing.B) {
	m := Mat4Perspective(Radians(60), 16.0/9.0, 0.1, 100)
	v := Vec4{1, 2, -3, 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.MulVec4(v)
	}
}

func BenchmarkMat4Inverse(b *testing.B) {
	m := Mat4Translate(Vec3{1, 2, 3}).
		Mul(Mat4RotateY(0.5)).
		Mul(Mat4Scale(Vec3{2, 2, 2}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Inverse()
	}
}

func BenchmarkMat4Perspective(b *testing.B) {
	fov := Radians(60)
	for i := 0; i < b.N; i++ {
		_ = Mat4Perspective(fov, 16.0/9.0, 0.1, 1000)
	}
}

// Verify that the perspective matrix size matches WGSL mat4x4<f32> (64 bytes)
func TestMat4Size(t *testing.T) {
	var m Mat4
	// [16]float32 = 16 * 4 bytes = 64 bytes
	sizeBytes := len(m) * 4
	if sizeBytes != 64 {
		t.Errorf("Mat4 size = %d bytes, want 64 (matches WGSL mat4x4<f32>)", sizeBytes)
	}
}

func TestMat4PerspectiveAspectRatio(t *testing.T) {
	// With 16:9 aspect, a square object should appear taller than wide
	m := Mat4Perspective(Radians(60), 16.0/9.0, 0.1, 100)

	// m[0] = f / aspect (X scale)
	// m[5] = f (Y scale)
	// Y scale should be larger than X scale for wide aspect ratios
	if m[5] <= m[0] {
		t.Errorf("Y scale (%v) should be > X scale (%v) for 16:9 aspect", m[5], m[0])
	}

	// For 1:1 aspect, scales should be equal
	m1 := Mat4Perspective(Radians(60), 1.0, 0.1, 100)
	if !approxEqual(m1[0], m1[5], epsilon) {
		t.Errorf("1:1 aspect: X scale (%v) should equal Y scale (%v)", m1[0], m1[5])
	}
}

func TestMat4Perspective_Symmetry(t *testing.T) {
	m := Mat4Perspective(Radians(60), 1.5, 0.1, 100)

	// Off-diagonal elements in the upper-left 3x3 should be zero for symmetric frustum
	if m[1] != 0 || m[2] != 0 || m[3] != 0 {
		t.Error("non-zero off-diagonal elements in column 0")
	}
	if m[4] != 0 || m[6] != 0 || m[7] != 0 {
		t.Error("non-zero off-diagonal elements in column 1")
	}
	if m[8] != 0 || m[9] != 0 {
		t.Error("non-zero off-diagonal elements in column 2 (rows 0-1)")
	}
}

func TestMat4RotateFullCircle(t *testing.T) {
	// Rotating by 2*PI should return approximately to identity
	id := Mat4Identity()
	full := float32(2 * math.Pi)

	tests := []struct {
		name string
		m    Mat4
	}{
		{"X full", Mat4RotateX(full)},
		{"Y full", Mat4RotateY(full)},
		{"Z full", Mat4RotateZ(full)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !approxEqualMat4(tt.m, id, 1e-4) {
				t.Errorf("%s should equal identity after full rotation", tt.name)
			}
		})
	}
}
