package g3d

import "math"

// Plane represents a 3D plane defined by the equation: Normal . P + D = 0.
// The normal vector points toward the positive half-space.
type Plane struct {
	Normal Vec3
	D      float32
}

// DistanceToPoint returns the signed distance from the plane to point p.
// Positive means p is on the side the normal points to.
func (p Plane) DistanceToPoint(pt Vec3) float32 {
	return p.Normal.Dot(pt) + p.D
}

// Frustum represents a view frustum defined by 6 clipping planes.
// Planes are ordered: left, right, bottom, top, near, far.
// Each plane's normal points inward (toward the visible region).
type Frustum [6]Plane

// FrustumFromMat4 extracts 6 frustum planes from a view-projection matrix.
// The planes are normalized so that distance calculations are correct.
//
// This works for both perspective and orthographic projections.
//
// Uses WebGPU Z [0,1] clip space convention. The near plane is extracted from
// row2 alone (z >= 0), not row3+row2 which would be the OpenGL Z [-1,1] convention.
func FrustumFromMat4(vp Mat4) Frustum {
	var f Frustum

	// In column-major storage, vp[col*4+row]:
	// Row 0: vp[0], vp[4], vp[8],  vp[12]
	// Row 1: vp[1], vp[5], vp[9],  vp[13]
	// Row 2: vp[2], vp[6], vp[10], vp[14]
	// Row 3: vp[3], vp[7], vp[11], vp[15]

	// Left:   row3 + row0  (-w <= x condition)
	f[0] = normalizePlane(Plane{
		Normal: Vec3{vp[3] + vp[0], vp[7] + vp[4], vp[11] + vp[8]},
		D:      vp[15] + vp[12],
	})

	// Right:  row3 - row0  (x <= w condition)
	f[1] = normalizePlane(Plane{
		Normal: Vec3{vp[3] - vp[0], vp[7] - vp[4], vp[11] - vp[8]},
		D:      vp[15] - vp[12],
	})

	// Bottom: row3 + row1  (-w <= y condition)
	f[2] = normalizePlane(Plane{
		Normal: Vec3{vp[3] + vp[1], vp[7] + vp[5], vp[11] + vp[9]},
		D:      vp[15] + vp[13],
	})

	// Top:    row3 - row1  (y <= w condition)
	f[3] = normalizePlane(Plane{
		Normal: Vec3{vp[3] - vp[1], vp[7] - vp[5], vp[11] - vp[9]},
		D:      vp[15] - vp[13],
	})

	// Near: row2 only (WebGPU Z [0,1]: z >= 0, NOT z >= -w like OpenGL)
	f[4] = normalizePlane(Plane{
		Normal: Vec3{vp[2], vp[6], vp[10]},
		D:      vp[14],
	})

	// Far:   row3 - row2  (z <= w condition)
	f[5] = normalizePlane(Plane{
		Normal: Vec3{vp[3] - vp[2], vp[7] - vp[6], vp[11] - vp[10]},
		D:      vp[15] - vp[14],
	})

	return f
}

// normalizePlane scales the plane so that Normal has unit length.
func normalizePlane(p Plane) Plane {
	l := p.Normal.Length()
	if l == 0 {
		return p
	}
	inv := 1.0 / l
	return Plane{
		Normal: p.Normal.Scale(inv),
		D:      p.D * inv,
	}
}

// IntersectsAABB returns true if the AABB is at least partially inside the frustum.
// Uses the "positive vertex" test for each plane — fast and conservative.
func (f Frustum) IntersectsAABB(box AABB) bool {
	for i := 0; i < 6; i++ {
		// Find the "positive vertex" — the AABB corner most in the direction of the plane normal
		pv := Vec3{
			selectF(f[i].Normal.X >= 0, box.Max.X, box.Min.X),
			selectF(f[i].Normal.Y >= 0, box.Max.Y, box.Min.Y),
			selectF(f[i].Normal.Z >= 0, box.Max.Z, box.Min.Z),
		}
		// If the positive vertex is outside this plane, the entire AABB is outside
		if f[i].DistanceToPoint(pv) < 0 {
			return false
		}
	}
	return true
}

// ContainsPoint returns true if the point p is inside all 6 frustum planes.
func (f Frustum) ContainsPoint(p Vec3) bool {
	for i := 0; i < 6; i++ {
		if f[i].DistanceToPoint(p) < 0 {
			return false
		}
	}
	return true
}

// selectF returns a if cond is true, b otherwise. Branchless helper for AABB tests.
func selectF(cond bool, a, b float32) float32 {
	if cond {
		return a
	}
	return b
}

// AABB represents an axis-aligned bounding box defined by its minimum and maximum corners.
type AABB struct {
	Min, Max Vec3
}

// NewAABBFromPoints computes the smallest AABB that contains all given points.
// Returns a zero AABB if points is empty.
func NewAABBFromPoints(points []Vec3) AABB {
	if len(points) == 0 {
		return AABB{}
	}
	minV := points[0]
	maxV := points[0]
	for _, p := range points[1:] {
		minV = minV.Min(p)
		maxV = maxV.Max(p)
	}
	return AABB{Min: minV, Max: maxV}
}

// Transform returns a new AABB that encloses the original AABB after applying
// the transformation matrix m. The result is axis-aligned (not an OBB).
func (a AABB) Transform(m Mat4) AABB {
	// Transform all 8 corners and compute new AABB.
	// Optimized method from "Transforming Axis-Aligned Bounding Boxes" by James Arvo.
	// Instead of transforming 8 corners, we compute the contribution of each matrix
	// element to the min/max.
	translation := Vec3{m[12], m[13], m[14]}
	newMin := translation
	newMax := translation

	for col := 0; col < 3; col++ {
		for row := 0; row < 3; row++ {
			e := m[col*4+row]
			aVal := e * a.Min.component(col)
			bVal := e * a.Max.component(col)
			if aVal < bVal {
				setComponent(&newMin, row, componentOf(newMin, row)+aVal)
				setComponent(&newMax, row, componentOf(newMax, row)+bVal)
			} else {
				setComponent(&newMin, row, componentOf(newMin, row)+bVal)
				setComponent(&newMax, row, componentOf(newMax, row)+aVal)
			}
		}
	}

	return AABB{Min: newMin, Max: newMax}
}

// Merge returns the smallest AABB containing both a and other.
func (a AABB) Merge(other AABB) AABB {
	return AABB{
		Min: a.Min.Min(other.Min),
		Max: a.Max.Max(other.Max),
	}
}

// Center returns the center point of the AABB.
func (a AABB) Center() Vec3 {
	return Vec3{
		(a.Min.X + a.Max.X) * 0.5,
		(a.Min.Y + a.Max.Y) * 0.5,
		(a.Min.Z + a.Max.Z) * 0.5,
	}
}

// Size returns the dimensions (width, height, depth) of the AABB.
func (a AABB) Size() Vec3 {
	return Vec3{
		a.Max.X - a.Min.X,
		a.Max.Y - a.Min.Y,
		a.Max.Z - a.Min.Z,
	}
}

// IsEmpty returns true if the AABB has zero volume.
func (a AABB) IsEmpty() bool {
	return a.Min.X >= a.Max.X || a.Min.Y >= a.Max.Y || a.Min.Z >= a.Max.Z
}

// ContainsPoint returns true if the AABB contains point p.
func (a AABB) ContainsPoint(p Vec3) bool {
	return p.X >= a.Min.X && p.X <= a.Max.X &&
		p.Y >= a.Min.Y && p.Y <= a.Max.Y &&
		p.Z >= a.Min.Z && p.Z <= a.Max.Z
}

// HalfExtents returns the half-size of the AABB (distance from center to each face).
func (a AABB) HalfExtents() Vec3 {
	return a.Size().Scale(0.5)
}

// SurfaceArea returns the surface area of the AABB.
func (a AABB) SurfaceArea() float32 {
	s := a.Size()
	return 2.0 * (s.X*s.Y + s.Y*s.Z + s.Z*s.X)
}

// Volume returns the volume of the AABB.
func (a AABB) Volume() float32 {
	s := a.Size()
	return s.X * s.Y * s.Z
}

// ClosestPoint returns the closest point on or in the AABB to p.
func (a AABB) ClosestPoint(p Vec3) Vec3 {
	return Vec3{
		float32(math.Max(float64(a.Min.X), math.Min(float64(a.Max.X), float64(p.X)))),
		float32(math.Max(float64(a.Min.Y), math.Min(float64(a.Max.Y), float64(p.Y)))),
		float32(math.Max(float64(a.Min.Z), math.Min(float64(a.Max.Z), float64(p.Z)))),
	}
}

// component returns the i-th component of v (0=X, 1=Y, 2=Z).
func (v Vec3) component(i int) float32 {
	switch i {
	case 0:
		return v.X
	case 1:
		return v.Y
	default:
		return v.Z
	}
}

// componentOf returns the i-th component of v (0=X, 1=Y, 2=Z).
func componentOf(v Vec3, i int) float32 {
	return v.component(i)
}

// setComponent sets the i-th component of v (0=X, 1=Y, 2=Z).
func setComponent(v *Vec3, i int, val float32) {
	switch i {
	case 0:
		v.X = val
	case 1:
		v.Y = val
	default:
		v.Z = val
	}
}
