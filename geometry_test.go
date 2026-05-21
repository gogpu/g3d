package g3d

import (
	"math"
	"testing"
)

// --- Geometry Interface Conformance ---

func TestBoxGeometry_ImplementsGeometry(t *testing.T) {
	var _ Geometry = (*BoxGeometry)(nil)
}

func TestSphereGeometry_ImplementsGeometry(t *testing.T) {
	var _ Geometry = (*SphereGeometry)(nil)
}

func TestPlaneGeometry_ImplementsGeometry(t *testing.T) {
	var _ Geometry = (*PlaneGeometry)(nil)
}

func TestBufferGeometry_ImplementsGeometry(t *testing.T) {
	var _ Geometry = (*BufferGeometry)(nil)
}

// --- BoxGeometry Tests ---

func TestNewBoxGeometry_Counts(t *testing.T) {
	box := NewBoxGeometry(1, 1, 1)

	if box.VertexCount() != 24 {
		t.Errorf("expected 24 vertices, got %d", box.VertexCount())
	}
	if len(box.Indices()) != 36 {
		t.Errorf("expected 36 indices, got %d", len(box.Indices()))
	}
}

func TestNewBoxGeometry_Accessors(t *testing.T) {
	box := NewBoxGeometry(2, 3, 4)

	if box.Width() != 2 {
		t.Errorf("expected width 2, got %f", box.Width())
	}
	if box.Height() != 3 {
		t.Errorf("expected height 3, got %f", box.Height())
	}
	if box.Depth() != 4 {
		t.Errorf("expected depth 4, got %f", box.Depth())
	}
}

func TestNewBoxGeometry_AABB(t *testing.T) {
	box := NewBoxGeometry(2, 4, 6)
	aabb := box.BoundingBox()

	if !approxF(aabb.Min.X, -1) || !approxF(aabb.Min.Y, -2) || !approxF(aabb.Min.Z, -3) {
		t.Errorf("unexpected AABB min: %+v", aabb.Min)
	}
	if !approxF(aabb.Max.X, 1) || !approxF(aabb.Max.Y, 2) || !approxF(aabb.Max.Z, 3) {
		t.Errorf("unexpected AABB max: %+v", aabb.Max)
	}
}

func TestNewBoxGeometry_NormalsOutward(t *testing.T) {
	box := NewBoxGeometry(1, 1, 1)
	verts := box.Vertices()

	for i := 0; i < box.VertexCount(); i++ {
		base := i * 8
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]

		dot := px*nx + py*ny + pz*nz
		if dot < 0 {
			t.Errorf("vertex %d: normal not outward (dot=%f)", i, dot)
		}
	}
}

func TestNewBoxGeometry_NonSquare(t *testing.T) {
	// A non-cube box should still have correct AABB.
	box := NewBoxGeometry(10, 0.5, 3)
	aabb := box.BoundingBox()

	size := aabb.Size()
	if !approxF(size.X, 10) || !approxF(size.Y, 0.5) || !approxF(size.Z, 3) {
		t.Errorf("AABB size mismatch: %+v", size)
	}
}

// --- SphereGeometry Tests ---

func TestNewSphereGeometry_DefaultSegments(t *testing.T) {
	s := NewSphereGeometry(1.0)

	if s.WidthSegments() != 32 {
		t.Errorf("expected 32 width segments, got %d", s.WidthSegments())
	}
	if s.HeightSegments() != 16 {
		t.Errorf("expected 16 height segments, got %d", s.HeightSegments())
	}
}

func TestNewSphereGeometry_CustomSegments(t *testing.T) {
	s := NewSphereGeometry(1.0, WithSegments(8, 4))

	if s.WidthSegments() != 8 {
		t.Errorf("expected 8 width segments, got %d", s.WidthSegments())
	}
	if s.HeightSegments() != 4 {
		t.Errorf("expected 4 height segments, got %d", s.HeightSegments())
	}
}

func TestNewSphereGeometry_VertexCount(t *testing.T) {
	s := NewSphereGeometry(1.0, WithSegments(8, 4))
	expected := (8 + 1) * (4 + 1) // 45
	if s.VertexCount() != expected {
		t.Errorf("expected %d vertices, got %d", expected, s.VertexCount())
	}
}

func TestNewSphereGeometry_Radius(t *testing.T) {
	radius := float32(2.5)
	s := NewSphereGeometry(radius, WithSegments(16, 8))

	if s.Radius() != radius {
		t.Errorf("expected radius %f, got %f", radius, s.Radius())
	}

	// All vertices should be at the specified radius from origin.
	verts := s.Vertices()
	for i := 0; i < s.VertexCount(); i++ {
		base := i * 8
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		dist := float32(math.Sqrt(float64(px*px + py*py + pz*pz)))
		if !approxF(dist, radius) {
			t.Errorf("vertex %d: distance from center = %f, expected %f", i, dist, radius)
		}
	}
}

func TestNewSphereGeometry_AABB(t *testing.T) {
	radius := float32(3.0)
	s := NewSphereGeometry(radius, WithSegments(32, 16))
	aabb := s.BoundingBox()

	// AABB should approximate [-radius, -radius, -radius] to [radius, radius, radius].
	// With 32x16 segments it won't be exactly at the corners, but the poles
	// and equator vertices reach the extremes along each axis.
	if aabb.Min.Y > -radius+0.01 || aabb.Max.Y < radius-0.01 {
		t.Errorf("AABB Y range not reaching poles: min=%f, max=%f", aabb.Min.Y, aabb.Max.Y)
	}
}

func TestNewSphereGeometry_NormalsOutward(t *testing.T) {
	s := NewSphereGeometry(1.0, WithSegments(16, 8))
	verts := s.Vertices()

	for i := 0; i < s.VertexCount(); i++ {
		base := i * 8
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]

		dot := px*nx + py*ny + pz*nz
		if dot < -0.001 {
			t.Errorf("vertex %d: normal not outward (dot=%f)", i, dot)
		}
	}
}

func TestNewSphereGeometry_IndicesInRange(t *testing.T) {
	s := NewSphereGeometry(1.0, WithSegments(16, 8))
	vertCount := uint32(s.VertexCount())

	for i, idx := range s.Indices() {
		if idx >= vertCount {
			t.Errorf("index %d out of range: %d >= %d", i, idx, vertCount)
		}
	}
}

// --- PlaneGeometry Tests ---

func TestNewPlaneGeometry_Counts(t *testing.T) {
	p := NewPlaneGeometry(1, 1)

	if p.VertexCount() != 4 {
		t.Errorf("expected 4 vertices, got %d", p.VertexCount())
	}
	if len(p.Indices()) != 6 {
		t.Errorf("expected 6 indices, got %d", len(p.Indices()))
	}
}

func TestNewPlaneGeometry_Accessors(t *testing.T) {
	p := NewPlaneGeometry(5, 3)

	if p.Width() != 5 {
		t.Errorf("expected width 5, got %f", p.Width())
	}
	if p.Height() != 3 {
		t.Errorf("expected height 3, got %f", p.Height())
	}
}

func TestNewPlaneGeometry_AABB(t *testing.T) {
	p := NewPlaneGeometry(4, 6)
	aabb := p.BoundingBox()

	if !approxF(aabb.Min.X, -2) || !approxF(aabb.Min.Z, -3) {
		t.Errorf("unexpected AABB min: %+v", aabb.Min)
	}
	if !approxF(aabb.Max.X, 2) || !approxF(aabb.Max.Z, 3) {
		t.Errorf("unexpected AABB max: %+v", aabb.Max)
	}
	// Y should be 0 (flat plane).
	if !approxF(aabb.Min.Y, 0) || !approxF(aabb.Max.Y, 0) {
		t.Errorf("plane Y should be 0: min.Y=%f, max.Y=%f", aabb.Min.Y, aabb.Max.Y)
	}
}

func TestNewPlaneGeometry_NormalUpY(t *testing.T) {
	p := NewPlaneGeometry(1, 1)
	verts := p.Vertices()

	for i := 0; i < p.VertexCount(); i++ {
		base := i * 8
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]
		if !approxF(nx, 0) || !approxF(ny, 1) || !approxF(nz, 0) {
			t.Errorf("vertex %d: expected normal (0,1,0), got (%f,%f,%f)", i, nx, ny, nz)
		}
	}
}

func TestNewPlaneGeometry_IndicesInRange(t *testing.T) {
	p := NewPlaneGeometry(1, 1)
	vertCount := uint32(p.VertexCount())

	for i, idx := range p.Indices() {
		if idx >= vertCount {
			t.Errorf("index %d out of range: %d >= %d", i, idx, vertCount)
		}
	}
}

// --- BufferGeometry Tests ---

func TestNewBufferGeometry_Empty(t *testing.T) {
	bg := NewBufferGeometry(nil, nil)

	if bg.VertexCount() != 0 {
		t.Errorf("expected 0 vertices, got %d", bg.VertexCount())
	}
	if bg.Indices() != nil {
		t.Error("expected nil indices")
	}
}

func TestNewBufferGeometry_CustomData(t *testing.T) {
	// Create a single triangle: 3 vertices * 8 floats.
	verts := []float32{
		// position       normal        uv
		0, 0, 0, 0, 1, 0, 0, 0,
		1, 0, 0, 0, 1, 0, 1, 0,
		0, 0, 1, 0, 1, 0, 0, 1,
	}
	indices := []uint32{0, 1, 2}

	bg := NewBufferGeometry(verts, indices)

	if bg.VertexCount() != 3 {
		t.Errorf("expected 3 vertices, got %d", bg.VertexCount())
	}
	if len(bg.Indices()) != 3 {
		t.Errorf("expected 3 indices, got %d", len(bg.Indices()))
	}

	aabb := bg.BoundingBox()
	if !approxF(aabb.Min.X, 0) || !approxF(aabb.Max.X, 1) {
		t.Errorf("AABB X: expected [0, 1], got [%f, %f]", aabb.Min.X, aabb.Max.X)
	}
}

// --- StandardVertexLayout Tests ---

func TestStandardVertexLayout(t *testing.T) {
	layout := StandardVertexLayout()

	if layout.Stride != 32 {
		t.Errorf("expected stride 32, got %d", layout.Stride)
	}
	if len(layout.Attributes) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(layout.Attributes))
	}

	tests := []struct {
		name       string
		offset     uint64
		floatCount int
	}{
		{"position", 0, 3},
		{"normal", 12, 3},
		{"uv", 24, 2},
	}

	for i, tc := range tests {
		attr := layout.Attributes[i]
		if attr.Name != tc.name {
			t.Errorf("attribute %d: expected name %q, got %q", i, tc.name, attr.Name)
		}
		if attr.Offset != tc.offset {
			t.Errorf("attribute %d: expected offset %d, got %d", i, tc.offset, attr.Offset)
		}
		if attr.FloatCount != tc.floatCount {
			t.Errorf("attribute %d: expected floatCount %d, got %d", i, tc.floatCount, attr.FloatCount)
		}
	}
}

// --- computeAABB Tests ---

func TestComputeAABB_Empty(t *testing.T) {
	aabb := computeAABB(nil)
	if aabb.Min != (Vec3{}) || aabb.Max != (Vec3{}) {
		t.Errorf("empty AABB should be zero: %+v", aabb)
	}
}

func TestComputeAABB_SingleVertex(t *testing.T) {
	verts := []float32{1, 2, 3, 0, 0, 0, 0, 0}
	aabb := computeAABB(verts)

	if !approxF(aabb.Min.X, 1) || !approxF(aabb.Min.Y, 2) || !approxF(aabb.Min.Z, 3) {
		t.Errorf("unexpected min: %+v", aabb.Min)
	}
	if !approxF(aabb.Max.X, 1) || !approxF(aabb.Max.Y, 2) || !approxF(aabb.Max.Z, 3) {
		t.Errorf("unexpected max: %+v", aabb.Max)
	}
}

// --- Helpers ---

func approxF(a, b float32) bool {
	const eps = 1e-4
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}
