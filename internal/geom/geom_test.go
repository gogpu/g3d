package geom

import (
	"math"
	"testing"
)

// --- Box Tests ---

func TestGenerateBox_VertexCount(t *testing.T) {
	verts, _ := GenerateBox(1, 1, 1)
	got := len(verts) / FloatsPerVertex
	if got != 24 {
		t.Errorf("expected 24 vertices, got %d", got)
	}
}

func TestGenerateBox_IndexCount(t *testing.T) {
	_, indices := GenerateBox(1, 1, 1)
	if len(indices) != 36 {
		t.Errorf("expected 36 indices, got %d", len(indices))
	}
}

func TestGenerateBox_Dimensions(t *testing.T) {
	width, height, depth := float32(2), float32(3), float32(4)
	verts, _ := GenerateBox(width, height, depth)

	minX, minY, minZ := float32(math.MaxFloat32), float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY, maxZ := float32(-math.MaxFloat32), float32(-math.MaxFloat32), float32(-math.MaxFloat32)

	count := len(verts) / FloatsPerVertex
	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		if px < minX {
			minX = px
		}
		if py < minY {
			minY = py
		}
		if pz < minZ {
			minZ = pz
		}
		if px > maxX {
			maxX = px
		}
		if py > maxY {
			maxY = py
		}
		if pz > maxZ {
			maxZ = pz
		}
	}

	if !approx(maxX-minX, width) {
		t.Errorf("width: expected %f, got %f", width, maxX-minX)
	}
	if !approx(maxY-minY, height) {
		t.Errorf("height: expected %f, got %f", height, maxY-minY)
	}
	if !approx(maxZ-minZ, depth) {
		t.Errorf("depth: expected %f, got %f", depth, maxZ-minZ)
	}
}

func TestGenerateBox_NormalsOutward(t *testing.T) {
	verts, _ := GenerateBox(2, 2, 2)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]

		// The normal should point away from the center (origin).
		// For a cube centered at origin, dot(position, normal) > 0 for outward normals.
		dot := px*nx + py*ny + pz*nz
		if dot < 0 {
			t.Errorf("vertex %d: normal not outward (pos=[%f,%f,%f], normal=[%f,%f,%f], dot=%f)",
				i, px, py, pz, nx, ny, nz, dot)
		}
	}
}

func TestGenerateBox_NormalsUnitLength(t *testing.T) {
	verts, _ := GenerateBox(1, 1, 1)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]
		length := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if !approx(length, 1.0) {
			t.Errorf("vertex %d: normal not unit length: %f", i, length)
		}
	}
}

func TestGenerateBox_UVRange(t *testing.T) {
	verts, _ := GenerateBox(1, 1, 1)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		u, v := verts[base+OffsetUV], verts[base+OffsetUV+1]
		if u < 0 || u > 1 || v < 0 || v > 1 {
			t.Errorf("vertex %d: UV out of [0,1] range: (%f, %f)", i, u, v)
		}
	}
}

func TestGenerateBox_IndicesInRange(t *testing.T) {
	verts, indices := GenerateBox(1, 1, 1)
	vertCount := uint32(len(verts) / FloatsPerVertex)

	for i, idx := range indices {
		if idx >= vertCount {
			t.Errorf("index %d out of range: %d >= %d", i, idx, vertCount)
		}
	}
}

func TestGenerateBox_Centered(t *testing.T) {
	verts, _ := GenerateBox(4, 6, 8)
	cx, cy, cz := centroid(verts)

	if !approx(cx, 0) || !approx(cy, 0) || !approx(cz, 0) {
		t.Errorf("box not centered at origin: centroid=(%f, %f, %f)", cx, cy, cz)
	}
}

// --- Sphere Tests ---

func TestGenerateSphere_VertexCount(t *testing.T) {
	verts, _ := GenerateSphere(1, 32, 16)
	got := len(verts) / FloatsPerVertex
	expected := (32 + 1) * (16 + 1)
	if got != expected {
		t.Errorf("expected %d vertices, got %d", expected, got)
	}
}

func TestGenerateSphere_SmallSegments(t *testing.T) {
	// Minimum viable sphere: 3 width, 2 height.
	verts, indices := GenerateSphere(1, 3, 2)
	vertCount := len(verts) / FloatsPerVertex
	if vertCount == 0 {
		t.Fatal("sphere with minimum segments produced no vertices")
	}
	if len(indices) == 0 {
		t.Fatal("sphere with minimum segments produced no indices")
	}

	// Verify all indices in range.
	for i, idx := range indices {
		if idx >= uint32(vertCount) {
			t.Errorf("index %d out of range: %d >= %d", i, idx, vertCount)
		}
	}
}

func TestGenerateSphere_SegmentClamping(t *testing.T) {
	// Segments below minimum should be clamped.
	verts1, _ := GenerateSphere(1, 1, 1) // below min
	verts2, _ := GenerateSphere(1, 3, 2) // at min

	count1 := len(verts1) / FloatsPerVertex
	count2 := len(verts2) / FloatsPerVertex
	if count1 != count2 {
		t.Errorf("clamped sphere should match minimum: %d vs %d", count1, count2)
	}
}

func TestGenerateSphere_NormalsOutward(t *testing.T) {
	verts, _ := GenerateSphere(2, 16, 8)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]

		// For a sphere centered at origin, the normal at any surface point
		// should point outward: dot(position, normal) >= 0.
		dot := px*nx + py*ny + pz*nz
		if dot < -0.001 {
			t.Errorf("vertex %d: normal not outward (pos=[%f,%f,%f], normal=[%f,%f,%f], dot=%f)",
				i, px, py, pz, nx, ny, nz, dot)
		}
	}
}

func TestGenerateSphere_NormalsUnitLength(t *testing.T) {
	verts, _ := GenerateSphere(1, 16, 8)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]
		length := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if !approx(length, 1.0) {
			t.Errorf("vertex %d: normal not unit length: %f", i, length)
		}
	}
}

func TestGenerateSphere_Radius(t *testing.T) {
	radius := float32(3.0)
	verts, _ := GenerateSphere(radius, 32, 16)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		px, py, pz := verts[base], verts[base+1], verts[base+2]
		dist := float32(math.Sqrt(float64(px*px + py*py + pz*pz)))
		if !approx(dist, radius) {
			t.Errorf("vertex %d: distance from center = %f, expected %f", i, dist, radius)
		}
	}
}

func TestGenerateSphere_Poles(t *testing.T) {
	radius := float32(1.0)
	verts, _ := GenerateSphere(radius, 16, 8)

	// Top pole is the first row of vertices. All should be at (0, radius, 0).
	for x := 0; x <= 16; x++ {
		base := x * FloatsPerVertex
		py := verts[base+1]
		if !approx(py, radius) {
			t.Errorf("top pole vertex %d: y=%f, expected %f", x, py, radius)
		}
	}

	// Bottom pole is the last row.
	lastRow := 8 * (16 + 1)
	for x := 0; x <= 16; x++ {
		base := (lastRow + x) * FloatsPerVertex
		py := verts[base+1]
		if !approx(py, -radius) {
			t.Errorf("bottom pole vertex %d: y=%f, expected %f", x, py, -radius)
		}
	}
}

func TestGenerateSphere_UVRange(t *testing.T) {
	verts, _ := GenerateSphere(1, 32, 16)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		u, v := verts[base+OffsetUV], verts[base+OffsetUV+1]
		if u < -0.001 || u > 1.001 || v < -0.001 || v > 1.001 {
			t.Errorf("vertex %d: UV out of [0,1] range: (%f, %f)", i, u, v)
		}
	}
}

func TestGenerateSphere_IndicesInRange(t *testing.T) {
	verts, indices := GenerateSphere(1, 32, 16)
	vertCount := uint32(len(verts) / FloatsPerVertex)

	for i, idx := range indices {
		if idx >= vertCount {
			t.Errorf("index %d out of range: %d >= %d", i, idx, vertCount)
		}
	}
}

func TestGenerateSphere_NoDegenerateTriangles(t *testing.T) {
	verts, indices := GenerateSphere(1, 8, 4)

	for i := 0; i+2 < len(indices); i += 3 {
		a, b, c := indices[i], indices[i+1], indices[i+2]
		if a == b || b == c || a == c {
			t.Errorf("degenerate triangle at index %d: [%d, %d, %d]", i, a, b, c)
			continue
		}

		// Also check for zero-area triangles (all three positions identical).
		ba := a * uint32(FloatsPerVertex)
		bb := b * uint32(FloatsPerVertex)
		bc := c * uint32(FloatsPerVertex)

		sameAB := approx(verts[ba], verts[bb]) && approx(verts[ba+1], verts[bb+1]) && approx(verts[ba+2], verts[bb+2])
		sameBC := approx(verts[bb], verts[bc]) && approx(verts[bb+1], verts[bc+1]) && approx(verts[bb+2], verts[bc+2])
		if sameAB && sameBC {
			t.Errorf("zero-area triangle at index %d: all vertices at same position", i)
		}
	}
}

// --- Plane Tests ---

func TestGeneratePlane_VertexCount(t *testing.T) {
	verts, _ := GeneratePlane(1, 1)
	got := len(verts) / FloatsPerVertex
	if got != 4 {
		t.Errorf("expected 4 vertices, got %d", got)
	}
}

func TestGeneratePlane_IndexCount(t *testing.T) {
	_, indices := GeneratePlane(1, 1)
	if len(indices) != 6 {
		t.Errorf("expected 6 indices, got %d", len(indices))
	}
}

func TestGeneratePlane_Dimensions(t *testing.T) {
	width, height := float32(5), float32(3)
	verts, _ := GeneratePlane(width, height)

	minX, minZ := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxZ := float32(-math.MaxFloat32), float32(-math.MaxFloat32)

	count := len(verts) / FloatsPerVertex
	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		px, pz := verts[base], verts[base+2]
		if px < minX {
			minX = px
		}
		if pz < minZ {
			minZ = pz
		}
		if px > maxX {
			maxX = px
		}
		if pz > maxZ {
			maxZ = pz
		}
	}

	if !approx(maxX-minX, width) {
		t.Errorf("width: expected %f, got %f", width, maxX-minX)
	}
	if !approx(maxZ-minZ, height) {
		t.Errorf("height: expected %f, got %f", height, maxZ-minZ)
	}
}

func TestGeneratePlane_NormalUpY(t *testing.T) {
	verts, _ := GeneratePlane(1, 1)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		nx, ny, nz := verts[base+3], verts[base+4], verts[base+5]
		if !approx(nx, 0) || !approx(ny, 1) || !approx(nz, 0) {
			t.Errorf("vertex %d: expected normal (0,1,0), got (%f,%f,%f)", i, nx, ny, nz)
		}
	}
}

func TestGeneratePlane_YIsZero(t *testing.T) {
	verts, _ := GeneratePlane(1, 1)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		py := verts[base+1]
		if !approx(py, 0) {
			t.Errorf("vertex %d: Y position should be 0, got %f", i, py)
		}
	}
}

func TestGeneratePlane_Centered(t *testing.T) {
	verts, _ := GeneratePlane(4, 6)
	cx, _, cz := centroidXZ(verts)

	if !approx(cx, 0) || !approx(cz, 0) {
		t.Errorf("plane not centered at origin: centroid=(%f, _, %f)", cx, cz)
	}
}

func TestGeneratePlane_UVRange(t *testing.T) {
	verts, _ := GeneratePlane(1, 1)
	count := len(verts) / FloatsPerVertex

	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		u, v := verts[base+OffsetUV], verts[base+OffsetUV+1]
		if u < 0 || u > 1 || v < 0 || v > 1 {
			t.Errorf("vertex %d: UV out of [0,1] range: (%f, %f)", i, u, v)
		}
	}
}

func TestGeneratePlane_IndicesInRange(t *testing.T) {
	verts, indices := GeneratePlane(1, 1)
	vertCount := uint32(len(verts) / FloatsPerVertex)

	for i, idx := range indices {
		if idx >= vertCount {
			t.Errorf("index %d out of range: %d >= %d", i, idx, vertCount)
		}
	}
}

// --- Layout Tests ---

func TestStandardVertexStride(t *testing.T) {
	// position(12) + normal(12) + uv(8) = 32
	if StandardVertexStride != 32 {
		t.Errorf("expected stride 32, got %d", StandardVertexStride)
	}
}

func TestFloatsPerVertex(t *testing.T) {
	// 32 bytes / 4 bytes per float32 = 8
	if FloatsPerVertex != 8 {
		t.Errorf("expected 8 floats per vertex, got %d", FloatsPerVertex)
	}
	if FloatsPerVertex*4 != StandardVertexStride {
		t.Error("FloatsPerVertex * 4 should equal StandardVertexStride")
	}
}

// --- Winding Order Tests ---

func TestGenerateBox_WindingOrder(t *testing.T) {
	verts, indices := GenerateBox(1, 1, 1)
	verifyWindingOrder(t, "box", verts, indices)
}

func TestGeneratePlane_WindingOrder(t *testing.T) {
	verts, indices := GeneratePlane(1, 1)
	verifyWindingOrder(t, "plane", verts, indices)
}

// verifyWindingOrder checks that triangle normals (from cross product) align
// with the stored vertex normals (indicating counter-clockwise front-face winding).
func verifyWindingOrder(t *testing.T, name string, verts []float32, indices []uint32) {
	t.Helper()

	for i := 0; i+2 < len(indices); i += 3 {
		a, b, c := indices[i], indices[i+1], indices[i+2]

		ax := verts[a*uint32(FloatsPerVertex)]
		ay := verts[a*uint32(FloatsPerVertex)+1]
		az := verts[a*uint32(FloatsPerVertex)+2]
		bx := verts[b*uint32(FloatsPerVertex)]
		by := verts[b*uint32(FloatsPerVertex)+1]
		bz := verts[b*uint32(FloatsPerVertex)+2]
		cx := verts[c*uint32(FloatsPerVertex)]
		cy := verts[c*uint32(FloatsPerVertex)+1]
		cz := verts[c*uint32(FloatsPerVertex)+2]

		// Edge vectors.
		e1x, e1y, e1z := bx-ax, by-ay, bz-az
		e2x, e2y, e2z := cx-ax, cy-ay, cz-az

		// Cross product (face normal from winding).
		fnx := e1y*e2z - e1z*e2y
		fny := e1z*e2x - e1x*e2z
		fnz := e1x*e2y - e1y*e2x

		// Average stored normal of the triangle's vertices.
		avgNX := (verts[a*uint32(FloatsPerVertex)+3] + verts[b*uint32(FloatsPerVertex)+3] + verts[c*uint32(FloatsPerVertex)+3]) / 3
		avgNY := (verts[a*uint32(FloatsPerVertex)+4] + verts[b*uint32(FloatsPerVertex)+4] + verts[c*uint32(FloatsPerVertex)+4]) / 3
		avgNZ := (verts[a*uint32(FloatsPerVertex)+5] + verts[b*uint32(FloatsPerVertex)+5] + verts[c*uint32(FloatsPerVertex)+5]) / 3

		// Dot product of face normal and average stored normal should be positive.
		dot := fnx*avgNX + fny*avgNY + fnz*avgNZ
		if dot < 0 {
			t.Errorf("%s triangle %d [%d,%d,%d]: winding order reversed (dot=%f)", name, i/3, a, b, c, dot)
		}
	}
}

// --- Helpers ---

func approx(a, b float32) bool {
	const eps = 1e-4
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}

func centroid(verts []float32) (cx, cy, cz float32) {
	count := len(verts) / FloatsPerVertex
	if count == 0 {
		return
	}
	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		cx += verts[base]
		cy += verts[base+1]
		cz += verts[base+2]
	}
	n := float32(count)
	return cx / n, cy / n, cz / n
}

func centroidXZ(verts []float32) (cx, cy, cz float32) {
	count := len(verts) / FloatsPerVertex
	if count == 0 {
		return
	}
	for i := 0; i < count; i++ {
		base := i * FloatsPerVertex
		cx += verts[base]
		cz += verts[base+2]
	}
	n := float32(count)
	return cx / n, 0, cz / n
}
