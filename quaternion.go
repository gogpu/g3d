package g3d

import "math"

// Quat represents a quaternion rotation. W is the scalar component,
// X/Y/Z are the vector components. The identity quaternion is {0,0,0,1}.
//
// Quaternions avoid gimbal lock and provide smooth interpolation (Slerp).
type Quat struct {
	X, Y, Z, W float32
}

// Euler represents rotation in radians using intrinsic XYZ rotation order.
// All angles are in radians.
type Euler struct {
	X, Y, Z float32
}

// QuatIdentity returns the identity quaternion (no rotation).
func QuatIdentity() Quat {
	return Quat{0, 0, 0, 1}
}

// QuatFromAxisAngle creates a quaternion representing rotation around axis by radians.
// The axis should be normalized.
func QuatFromAxisAngle(axis Vec3, radians float32) Quat {
	halfAngle := float64(radians) * 0.5
	s := float32(math.Sin(halfAngle))
	c := float32(math.Cos(halfAngle))
	return Quat{
		X: axis.X * s,
		Y: axis.Y * s,
		Z: axis.Z * s,
		W: c,
	}
}

// QuatFromEuler creates a quaternion from Euler angles (intrinsic XYZ rotation order).
// Angles are in radians.
func QuatFromEuler(e Euler) Quat {
	hx := float64(e.X) * 0.5
	hy := float64(e.Y) * 0.5
	hz := float64(e.Z) * 0.5

	cx := float32(math.Cos(hx))
	sx := float32(math.Sin(hx))
	cy := float32(math.Cos(hy))
	sy := float32(math.Sin(hy))
	cz := float32(math.Cos(hz))
	sz := float32(math.Sin(hz))

	// Intrinsic XYZ: Qz * Qy * Qx applied right-to-left
	return Quat{
		X: sx*cy*cz + cx*sy*sz,
		Y: cx*sy*cz - sx*cy*sz,
		Z: cx*cy*sz + sx*sy*cz,
		W: cx*cy*cz - sx*sy*sz,
	}
}

// ToMat4 converts the quaternion to a 4x4 rotation matrix (column-major).
func (q Quat) ToMat4() Mat4 {
	x2 := q.X + q.X
	y2 := q.Y + q.Y
	z2 := q.Z + q.Z

	xx := q.X * x2
	xy := q.X * y2
	xz := q.X * z2
	yy := q.Y * y2
	yz := q.Y * z2
	zz := q.Z * z2
	wx := q.W * x2
	wy := q.W * y2
	wz := q.W * z2

	return Mat4{
		1 - (yy + zz), xy + wz, xz - wy, 0,
		xy - wz, 1 - (xx + zz), yz + wx, 0,
		xz + wy, yz - wx, 1 - (xx + yy), 0,
		0, 0, 0, 1,
	}
}

// ToEuler converts the quaternion to Euler angles (intrinsic XYZ rotation order).
// Returns angles in radians.
func (q Quat) ToEuler() Euler {
	// Derive from rotation matrix to extract XYZ intrinsic Euler angles
	m := q.ToMat4()

	// m[2] = col0,row2 = xz + wy = sin(Y) detection
	// For column-major: row=r, col=c => m[c*4+r]
	m02 := m[8]  // col2, row0 = xz - wy ... but we need the matrix form
	m12 := m[9]  // col2, row1 = yz + wx
	m22 := m[10] // col2, row2 = 1 - (xx + yy)
	m00 := m[0]  // col0, row0 = 1 - (yy + zz)
	m01 := m[4]  // col1, row0 = xy - wz
	m10 := m[1]  // col0, row1 = xy + wz
	m11 := m[5]  // col1, row1 = 1 - (xx + zz)

	// XYZ intrinsic: extract Y from m[0][2] = sin(Y)
	sy := Clamp(m02, -1.0, 1.0)
	y := float32(math.Asin(float64(sy)))

	var x, z float32
	if float32(math.Abs(float64(sy))) < 0.9999999 {
		// Standard case: cos(Y) != 0
		x = float32(math.Atan2(float64(-m12), float64(m22)))
		z = float32(math.Atan2(float64(-m01), float64(m00)))
	} else {
		// Gimbal lock: Y = +/-90 degrees
		x = float32(math.Atan2(float64(m10), float64(m11)))
		z = 0
	}

	return Euler{X: x, Y: y, Z: z}
}

// Mul returns the Hamilton product q * other.
// This combines rotations: q.Mul(other) applies other first, then q.
func (q Quat) Mul(other Quat) Quat {
	return Quat{
		X: q.W*other.X + q.X*other.W + q.Y*other.Z - q.Z*other.Y,
		Y: q.W*other.Y - q.X*other.Z + q.Y*other.W + q.Z*other.X,
		Z: q.W*other.Z + q.X*other.Y - q.Y*other.X + q.Z*other.W,
		W: q.W*other.W - q.X*other.X - q.Y*other.Y - q.Z*other.Z,
	}
}

// Dot returns the dot product of q and other.
func (q Quat) Dot(other Quat) float32 {
	return q.X*other.X + q.Y*other.Y + q.Z*other.Z + q.W*other.W
}

// LengthSq returns the squared length of the quaternion.
func (q Quat) LengthSq() float32 {
	return q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
}

// Length returns the length (norm) of the quaternion.
func (q Quat) Length() float32 {
	return float32(math.Sqrt(float64(q.LengthSq())))
}

// Normalize returns a unit quaternion in the same direction.
// Returns the identity quaternion if q has zero length.
func (q Quat) Normalize() Quat {
	l := q.Length()
	if l == 0 {
		return QuatIdentity()
	}
	inv := 1.0 / l
	return Quat{q.X * inv, q.Y * inv, q.Z * inv, q.W * inv}
}

// Conjugate returns the conjugate of q (negated vector part).
func (q Quat) Conjugate() Quat {
	return Quat{-q.X, -q.Y, -q.Z, q.W}
}

// Inverse returns the inverse of q. For unit quaternions, this equals the conjugate.
func (q Quat) Inverse() Quat {
	lenSq := q.LengthSq()
	if lenSq == 0 {
		return QuatIdentity()
	}
	inv := 1.0 / lenSq
	return Quat{-q.X * inv, -q.Y * inv, -q.Z * inv, q.W * inv}
}

// RotateVec3 rotates v by the rotation represented by q.
// Equivalent to q * (0,v,0) * q^-1, optimized to avoid full quaternion multiply.
func (q Quat) RotateVec3(v Vec3) Vec3 {
	// Optimized formula: v' = v + 2*w*(qv x v) + 2*(qv x (qv x v))
	// where qv = (q.X, q.Y, q.Z), w = q.W
	qv := Vec3{q.X, q.Y, q.Z}
	t := qv.Cross(v).Scale(2)
	return v.Add(t.Scale(q.W)).Add(qv.Cross(t))
}

// Slerp performs spherical linear interpolation between q and other by t in [0,1].
// Produces constant-speed rotation along the shortest arc.
func (q Quat) Slerp(other Quat, t float32) Quat {
	// Handle edge cases
	if t <= 0 {
		return q
	}
	if t >= 1 {
		return other
	}

	// Compute cosine of angle between quaternions
	cosHalfTheta := q.Dot(other)

	// If negative dot, negate one quaternion to take shortest path
	target := other
	if cosHalfTheta < 0 {
		target = Quat{-other.X, -other.Y, -other.Z, -other.W}
		cosHalfTheta = -cosHalfTheta
	}

	// If quaternions are very close, use linear interpolation to avoid division by zero
	if cosHalfTheta >= 1.0 {
		return q
	}

	sqrSinHalfTheta := 1.0 - cosHalfTheta*cosHalfTheta
	if sqrSinHalfTheta <= 1e-12 {
		// Nearly identical rotations — linear interpolation + normalize
		s := 1.0 - t
		return Quat{
			X: s*q.X + t*target.X,
			Y: s*q.Y + t*target.Y,
			Z: s*q.Z + t*target.Z,
			W: s*q.W + t*target.W,
		}.Normalize()
	}

	sinHalfTheta := float32(math.Sqrt(float64(sqrSinHalfTheta)))
	halfTheta := float32(math.Atan2(float64(sinHalfTheta), float64(cosHalfTheta)))
	ratioA := float32(math.Sin(float64((1.0-t)*halfTheta))) / sinHalfTheta
	ratioB := float32(math.Sin(float64(t*halfTheta))) / sinHalfTheta

	return Quat{
		X: q.X*ratioA + target.X*ratioB,
		Y: q.Y*ratioA + target.Y*ratioB,
		Z: q.Z*ratioA + target.Z*ratioB,
		W: q.W*ratioA + target.W*ratioB,
	}
}
