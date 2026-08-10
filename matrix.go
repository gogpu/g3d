package g3d

import "math"

// Mat4 is a 4x4 transformation matrix stored in column-major order.
// This matches the WGSL mat4x4<f32> memory layout for direct GPU upload.
//
// Storage: m[col*4+row]
//
//	m[0..3]   = column 0 (X-axis + translation.x)
//	m[4..7]   = column 1 (Y-axis + translation.y)
//	m[8..11]  = column 2 (Z-axis + translation.z)
//	m[12..15] = column 3 (translation + w=1)
//
// Index layout (row, col):
//
//	[0]  [4]  [8]  [12]     (row 0)
//	[1]  [5]  [9]  [13]     (row 1)
//	[2]  [6]  [10] [14]     (row 2)
//	[3]  [7]  [11] [15]     (row 3)
type Mat4 [16]float32

// Mat4Identity returns the 4x4 identity matrix.
func Mat4Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0, // column 0
		0, 1, 0, 0, // column 1
		0, 0, 1, 0, // column 2
		0, 0, 0, 1, // column 3
	}
}

// Mat4Translate returns a translation matrix that moves by v.
func Mat4Translate(v Vec3) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		v.X, v.Y, v.Z, 1,
	}
}

// Mat4Scale returns a scale matrix that scales by v.
func Mat4Scale(v Vec3) Mat4 {
	return Mat4{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1,
	}
}

// Mat4RotateX returns a rotation matrix around the X axis by radians.
func Mat4RotateX(radians float32) Mat4 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Mat4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

// Mat4RotateY returns a rotation matrix around the Y axis by radians.
func Mat4RotateY(radians float32) Mat4 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Mat4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

// Mat4RotateZ returns a rotation matrix around the Z axis by radians.
func Mat4RotateZ(radians float32) Mat4 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Mat4Perspective returns a perspective projection matrix.
//
// CRITICAL: Uses WebGPU clip space Z [0,1], NOT OpenGL Z [-1,1].
// Formula validated against Three.js Matrix4.makePerspective(WebGPUCoordinateSystem).
//
// Parameters:
//   - fovYRadians: vertical field of view in radians
//   - aspect: width / height ratio
//   - near: near clipping plane distance (must be > 0)
//   - far: far clipping plane distance (must be > near)
func Mat4Perspective(fovYRadians, aspect, near, far float32) Mat4 {
	f := float32(1.0 / math.Tan(float64(fovYRadians)*0.5))
	rangeInv := 1.0 / (far - near)

	// WebGPU Z [0,1] mapping:
	// m[10] = -far / (far - near)         = maps far plane to z=1
	// m[14] = -(far * near) / (far - near) = maps near plane to z=0
	var m Mat4
	m[0] = f / aspect // X scale
	m[5] = f          // Y scale
	m[10] = -far * rangeInv
	m[11] = -1 // perspective divide
	m[14] = -(far * near) * rangeInv
	return m
}

// Mat4Ortho returns an orthographic projection matrix.
//
// Uses WebGPU clip space Z [0,1], NOT OpenGL Z [-1,1].
func Mat4Ortho(left, right, bottom, top, near, far float32) Mat4 {
	rl := 1.0 / (right - left)
	tb := 1.0 / (top - bottom)
	fn := 1.0 / (far - near)

	var m Mat4
	m[0] = 2.0 * rl
	m[5] = 2.0 * tb
	m[10] = -fn // WebGPU Z [0,1]: -1/(far-near), NOT -2/(far-near)
	m[12] = -(right + left) * rl
	m[13] = -(top + bottom) * tb
	m[14] = -near * fn // WebGPU Z [0,1]: -near/(far-near), NOT -(far+near)/(far-near)
	m[15] = 1
	return m
}

// Mat4LookAt returns a view matrix looking from eye toward target with the given up direction.
func Mat4LookAt(eye, target, up Vec3) Mat4 {
	// Forward vector (camera looks down -Z in right-handed coordinates). A
	// coincident eye and target do not define a direction; keep the view valid
	// by retaining the conventional -Z camera forward direction.
	direction := target.Sub(eye)
	if direction == (Vec3{}) {
		direction = Vec3{0, 0, -1}
	}
	f := direction.Normalize()
	// Normalize the caller's up vector before comparing its cross product with
	// the view direction. Parallel detection must not depend on the up vector's
	// magnitude (for example, a scaled world-up vector).
	up = up.Normalize()
	if up == (Vec3{}) {
		// A zero up vector carries no roll information. Preserve the API's
		// conventional world +Y orientation rather than selecting an arbitrary
		// axis for ordinary view directions.
		up = Vec3{0, 1, 0}
	}

	// Right vector. When the requested up direction is parallel (or nearly
	// parallel) to the view direction, its cross product cannot define a stable
	// roll. Select a cardinal fallback axis that is safely non-parallel with f.
	r := f.Cross(up)
	if r.LengthSq() <= lookAtEpsilon*lookAtEpsilon {
		fallbackUp := Vec3{0, 1, 0}
		if absFloat32(f.Y) >= 1-lookAtEpsilon {
			// For top-down views, world +Z is the screen-up direction. This
			// gives a deterministic and useful roll for both +/-Y views.
			fallbackUp = Vec3{0, 0, 1}
		}
		r = f.Cross(fallbackUp)
	}
	r = r.Normalize()
	// Recalculated up vector. Both f and r are unit and perpendicular, so the
	// cross product preserves a right-handed, orthonormal view basis.
	u := r.Cross(f).Normalize()

	return Mat4{
		r.X, u.X, -f.X, 0,
		r.Y, u.Y, -f.Y, 0,
		r.Z, u.Z, -f.Z, 0,
		-r.Dot(eye), -u.Dot(eye), f.Dot(eye), 1,
	}
}

const lookAtEpsilon float32 = 1e-6

func absFloat32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// Mat4FromQuat returns a rotation matrix equivalent to the given quaternion.
func Mat4FromQuat(q Quat) Mat4 {
	return q.ToMat4()
}

// Mul returns the matrix product m * other.
// Matrix multiplication is associative: (A*B)*C = A*(B*C).
func (m Mat4) Mul(other Mat4) Mat4 {
	var result Mat4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			var sum float32
			for k := 0; k < 4; k++ {
				sum += m[k*4+row] * other[col*4+k]
			}
			result[col*4+row] = sum
		}
	}
	return result
}

// MulVec4 returns the product m * v (matrix-vector multiplication).
func (m Mat4) MulVec4(v Vec4) Vec4 {
	return Vec4{
		m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]*v.W,
		m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]*v.W,
		m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]*v.W,
		m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]*v.W,
	}
}

// Transpose returns the transpose of m.
func (m Mat4) Transpose() Mat4 {
	return Mat4{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
}

// Translation extracts the translation component from column 3.
func (m Mat4) Translation() Vec3 {
	return Vec3{m[12], m[13], m[14]}
}

// Determinant returns the determinant of m.
func (m Mat4) Determinant() float32 {
	// Expand along the first row using cofactors
	a00, a01, a02, a03 := m[0], m[4], m[8], m[12]
	a10, a11, a12, a13 := m[1], m[5], m[9], m[13]
	a20, a21, a22, a23 := m[2], m[6], m[10], m[14]
	a30, a31, a32, a33 := m[3], m[7], m[11], m[15]

	b00 := a00*a11 - a01*a10
	b01 := a00*a12 - a02*a10
	b02 := a00*a13 - a03*a10
	b03 := a01*a12 - a02*a11
	b04 := a01*a13 - a03*a11
	b05 := a02*a13 - a03*a12
	b06 := a20*a31 - a21*a30
	b07 := a20*a32 - a22*a30
	b08 := a20*a33 - a23*a30
	b09 := a21*a32 - a22*a31
	b10 := a21*a33 - a23*a31
	b11 := a22*a33 - a23*a32

	return b00*b11 - b01*b10 + b02*b09 + b03*b08 - b04*b07 + b05*b06
}

// Inverse returns the inverse of m. If m is singular (determinant = 0),
// returns the zero matrix.
func (m Mat4) Inverse() Mat4 {
	a00, a01, a02, a03 := m[0], m[4], m[8], m[12]
	a10, a11, a12, a13 := m[1], m[5], m[9], m[13]
	a20, a21, a22, a23 := m[2], m[6], m[10], m[14]
	a30, a31, a32, a33 := m[3], m[7], m[11], m[15]

	b00 := a00*a11 - a01*a10
	b01 := a00*a12 - a02*a10
	b02 := a00*a13 - a03*a10
	b03 := a01*a12 - a02*a11
	b04 := a01*a13 - a03*a11
	b05 := a02*a13 - a03*a12
	b06 := a20*a31 - a21*a30
	b07 := a20*a32 - a22*a30
	b08 := a20*a33 - a23*a30
	b09 := a21*a32 - a22*a31
	b10 := a21*a33 - a23*a31
	b11 := a22*a33 - a23*a32

	det := b00*b11 - b01*b10 + b02*b09 + b03*b08 - b04*b07 + b05*b06
	if det == 0 {
		return Mat4{}
	}

	invDet := 1.0 / det

	return Mat4{
		(a11*b11 - a12*b10 + a13*b09) * invDet,
		(-a10*b11 + a12*b08 - a13*b07) * invDet,
		(a10*b10 - a11*b08 + a13*b06) * invDet,
		(-a10*b09 + a11*b07 - a12*b06) * invDet,
		(-a01*b11 + a02*b10 - a03*b09) * invDet,
		(a00*b11 - a02*b08 + a03*b07) * invDet,
		(-a00*b10 + a01*b08 - a03*b06) * invDet,
		(a00*b09 - a01*b07 + a02*b06) * invDet,
		(a31*b05 - a32*b04 + a33*b03) * invDet,
		(-a30*b05 + a32*b02 - a33*b01) * invDet,
		(a30*b04 - a31*b02 + a33*b00) * invDet,
		(-a30*b03 + a31*b01 - a32*b00) * invDet,
		(-a21*b05 + a22*b04 - a23*b03) * invDet,
		(a20*b05 - a22*b02 + a23*b01) * invDet,
		(-a20*b04 + a21*b02 - a23*b00) * invDet,
		(a20*b03 - a21*b01 + a22*b00) * invDet,
	}
}
