package g3d

import "math"

// Vec2 represents a 2D vector. Used for UV texture coordinates.
type Vec2 struct {
	X, Y float32
}

// Add returns the component-wise sum of v and other.
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

// Sub returns the component-wise difference of v and other.
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
}

// Scale returns v with each component multiplied by s.
func (v Vec2) Scale(s float32) Vec2 {
	return Vec2{v.X * s, v.Y * s}
}

// Dot returns the dot product of v and other.
func (v Vec2) Dot(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
}

// LengthSq returns the squared length of v. Avoids a sqrt call.
func (v Vec2) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y
}

// Length returns the Euclidean length of v.
func (v Vec2) Length() float32 {
	return float32(math.Sqrt(float64(v.LengthSq())))
}

// Normalize returns a unit-length vector in the same direction as v.
// Returns the zero vector if v has zero length.
func (v Vec2) Normalize() Vec2 {
	l := v.Length()
	if l == 0 {
		return Vec2{}
	}
	inv := 1.0 / l
	return Vec2{v.X * inv, v.Y * inv}
}

// Lerp linearly interpolates between v and other by t in [0,1].
func (v Vec2) Lerp(other Vec2, t float32) Vec2 {
	return Vec2{
		v.X + (other.X-v.X)*t,
		v.Y + (other.Y-v.Y)*t,
	}
}

// Vec3 represents a 3D vector. Used for positions, directions, and scale.
type Vec3 struct {
	X, Y, Z float32
}

// Add returns the component-wise sum of v and other.
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

// Sub returns the component-wise difference of v and other.
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

// Scale returns v with each component multiplied by s.
func (v Vec3) Scale(s float32) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// Mul returns the component-wise product of v and other (Hadamard product).
func (v Vec3) Mul(other Vec3) Vec3 {
	return Vec3{v.X * other.X, v.Y * other.Y, v.Z * other.Z}
}

// Dot returns the dot product of v and other.
func (v Vec3) Dot(other Vec3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross returns the cross product of v and other.
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		v.Y*other.Z - v.Z*other.Y,
		v.Z*other.X - v.X*other.Z,
		v.X*other.Y - v.Y*other.X,
	}
}

// LengthSq returns the squared length of v. Avoids a sqrt call.
func (v Vec3) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// Length returns the Euclidean length of v.
func (v Vec3) Length() float32 {
	return float32(math.Sqrt(float64(v.LengthSq())))
}

// Normalize returns a unit-length vector in the same direction as v.
// Returns the zero vector if v has zero length.
func (v Vec3) Normalize() Vec3 {
	l := v.Length()
	if l == 0 {
		return Vec3{}
	}
	inv := 1.0 / l
	return Vec3{v.X * inv, v.Y * inv, v.Z * inv}
}

// Negate returns -v.
func (v Vec3) Negate() Vec3 {
	return Vec3{-v.X, -v.Y, -v.Z}
}

// Distance returns the Euclidean distance between v and other.
func (v Vec3) Distance(other Vec3) float32 {
	return v.Sub(other).Length()
}

// DistanceSq returns the squared Euclidean distance between v and other.
func (v Vec3) DistanceSq(other Vec3) float32 {
	return v.Sub(other).LengthSq()
}

// Lerp linearly interpolates between v and other by t in [0,1].
func (v Vec3) Lerp(other Vec3, t float32) Vec3 {
	return Vec3{
		v.X + (other.X-v.X)*t,
		v.Y + (other.Y-v.Y)*t,
		v.Z + (other.Z-v.Z)*t,
	}
}

// Min returns a vector with the minimum of each component.
func (v Vec3) Min(other Vec3) Vec3 {
	return Vec3{
		float32(math.Min(float64(v.X), float64(other.X))),
		float32(math.Min(float64(v.Y), float64(other.Y))),
		float32(math.Min(float64(v.Z), float64(other.Z))),
	}
}

// Max returns a vector with the maximum of each component.
func (v Vec3) Max(other Vec3) Vec3 {
	return Vec3{
		float32(math.Max(float64(v.X), float64(other.X))),
		float32(math.Max(float64(v.Y), float64(other.Y))),
		float32(math.Max(float64(v.Z), float64(other.Z))),
	}
}

// Vec4 represents a 4D vector. Used for homogeneous coordinates and shader data.
type Vec4 struct {
	X, Y, Z, W float32
}

// Add returns the component-wise sum of v and other.
func (v Vec4) Add(other Vec4) Vec4 {
	return Vec4{v.X + other.X, v.Y + other.Y, v.Z + other.Z, v.W + other.W}
}

// Sub returns the component-wise difference of v and other.
func (v Vec4) Sub(other Vec4) Vec4 {
	return Vec4{v.X - other.X, v.Y - other.Y, v.Z - other.Z, v.W - other.W}
}

// Scale returns v with each component multiplied by s.
func (v Vec4) Scale(s float32) Vec4 {
	return Vec4{v.X * s, v.Y * s, v.Z * s, v.W * s}
}

// Dot returns the dot product of v and other.
func (v Vec4) Dot(other Vec4) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z + v.W*other.W
}

// LengthSq returns the squared length of v. Avoids a sqrt call.
func (v Vec4) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

// Length returns the Euclidean length of v.
func (v Vec4) Length() float32 {
	return float32(math.Sqrt(float64(v.LengthSq())))
}

// Normalize returns a unit-length vector in the same direction as v.
// Returns the zero vector if v has zero length.
func (v Vec4) Normalize() Vec4 {
	l := v.Length()
	if l == 0 {
		return Vec4{}
	}
	inv := 1.0 / l
	return Vec4{v.X * inv, v.Y * inv, v.Z * inv, v.W * inv}
}

// Lerp linearly interpolates between v and other by t in [0,1].
func (v Vec4) Lerp(other Vec4, t float32) Vec4 {
	return Vec4{
		v.X + (other.X-v.X)*t,
		v.Y + (other.Y-v.Y)*t,
		v.Z + (other.Z-v.Z)*t,
		v.W + (other.W-v.W)*t,
	}
}

// XYZ returns the first three components as a Vec3.
func (v Vec4) XYZ() Vec3 {
	return Vec3{v.X, v.Y, v.Z}
}

// Radians converts degrees to radians.
func Radians(degrees float32) float32 {
	return degrees * math.Pi / 180.0
}

// Degrees converts radians to degrees.
func Degrees(radians float32) float32 {
	return radians * 180.0 / math.Pi
}

// Clamp constrains val to the range [lo, hi].
func Clamp(val, lo, hi float32) float32 {
	if val < lo {
		return lo
	}
	if val > hi {
		return hi
	}
	return val
}
