package geom

// GeneratePlane creates vertex and index data for a flat rectangular plane.
//
// The plane lies in the XZ plane, centered at the origin, with the normal pointing
// in the +Y direction. This matches the standard ground plane convention used by
// Three.js and most 3D engines.
//
// The result is 4 vertices and 6 indices (2 triangles).
// Vertex data is interleaved: [px,py,pz, nx,ny,nz, u,v, ...] with 8 floats per vertex.
// Winding order is counter-clockwise when viewed from above (+Y looking down).
func GeneratePlane(width, height float32) (vertices []float32, indices []uint32) {
	hw := width * 0.5
	hh := height * 0.5

	// Normal: straight up (+Y).
	const nx, ny, nz float32 = 0, 1, 0

	// 4 vertices with UV mapping:
	// V0 (-X, -Z) = UV(0, 1)  — back-left
	// V1 (+X, -Z) = UV(1, 1)  — back-right
	// V2 (+X, +Z) = UV(1, 0)  — front-right
	// V3 (-X, +Z) = UV(0, 0)  — front-left
	vertices = make([]float32, 0, 4*FloatsPerVertex)
	vertices = appendVertex(vertices, -hw, 0, -hh, nx, ny, nz, 0, 1) // V0
	vertices = appendVertex(vertices, hw, 0, -hh, nx, ny, nz, 1, 1)  // V1
	vertices = appendVertex(vertices, hw, 0, hh, nx, ny, nz, 1, 0)   // V2
	vertices = appendVertex(vertices, -hw, 0, hh, nx, ny, nz, 0, 0)  // V3

	// Two triangles. Winding order produces +Y face normal via cross product:
	//   E1 = V2-V0 = (+w, 0, +h), E2 = V1-V0 = (+w, 0, 0)
	//   cross(E1, E2) = (0, +wh, 0) = +Y
	// Triangle 1: V0 → V2 → V1
	// Triangle 2: V0 → V3 → V2
	indices = []uint32{0, 2, 1, 0, 3, 2}

	return vertices, indices
}
