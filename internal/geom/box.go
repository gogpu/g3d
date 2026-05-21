package geom

// GenerateBox creates vertex and index data for a box with the given dimensions.
//
// The box is centered at the origin. Each face has 4 unique vertices with outward
// normals and UV coordinates in [0,1]. The result is 24 vertices and 36 indices.
//
// Vertex data is interleaved: [px,py,pz, nx,ny,nz, u,v, ...] with 8 floats per vertex.
// Winding order is counter-clockwise when viewed from outside (front-face).
func GenerateBox(width, height, depth float32) (vertices []float32, indices []uint32) {
	hw := width * 0.5
	hh := height * 0.5
	hd := depth * 0.5

	// Preallocate: 24 vertices * 8 floats = 192 floats, 36 indices.
	vertices = make([]float32, 0, 24*FloatsPerVertex)
	indices = make([]uint32, 0, 36)

	// Face definitions: 6 faces, each with 4 corners and a normal.
	// Order per face: bottom-left, bottom-right, top-right, top-left
	// (when viewed from outside, counter-clockwise winding).
	type face struct {
		// 4 corner positions (bl, br, tr, tl)
		p [4][3]float32
		// outward normal
		n [3]float32
	}

	faces := [6]face{
		// Front face (+Z): normal (0, 0, 1)
		{
			p: [4][3]float32{
				{-hw, -hh, hd}, {hw, -hh, hd}, {hw, hh, hd}, {-hw, hh, hd},
			},
			n: [3]float32{0, 0, 1},
		},
		// Back face (-Z): normal (0, 0, -1)
		{
			p: [4][3]float32{
				{hw, -hh, -hd}, {-hw, -hh, -hd}, {-hw, hh, -hd}, {hw, hh, -hd},
			},
			n: [3]float32{0, 0, -1},
		},
		// Right face (+X): normal (1, 0, 0)
		{
			p: [4][3]float32{
				{hw, -hh, hd}, {hw, -hh, -hd}, {hw, hh, -hd}, {hw, hh, hd},
			},
			n: [3]float32{1, 0, 0},
		},
		// Left face (-X): normal (-1, 0, 0)
		{
			p: [4][3]float32{
				{-hw, -hh, -hd}, {-hw, -hh, hd}, {-hw, hh, hd}, {-hw, hh, -hd},
			},
			n: [3]float32{-1, 0, 0},
		},
		// Top face (+Y): normal (0, 1, 0)
		{
			p: [4][3]float32{
				{-hw, hh, hd}, {hw, hh, hd}, {hw, hh, -hd}, {-hw, hh, -hd},
			},
			n: [3]float32{0, 1, 0},
		},
		// Bottom face (-Y): normal (0, -1, 0)
		{
			p: [4][3]float32{
				{-hw, -hh, -hd}, {hw, -hh, -hd}, {hw, -hh, hd}, {-hw, -hh, hd},
			},
			n: [3]float32{0, -1, 0},
		},
	}

	// UV coordinates per corner: bl(0,1), br(1,1), tr(1,0), tl(0,0)
	uvs := [4][2]float32{
		{0, 1}, {1, 1}, {1, 0}, {0, 0},
	}

	for i, f := range faces {
		base := uint32(i * 4)
		for j := 0; j < 4; j++ {
			vertices = appendVertex(vertices,
				f.p[j][0], f.p[j][1], f.p[j][2],
				f.n[0], f.n[1], f.n[2],
				uvs[j][0], uvs[j][1],
			)
		}
		// Two triangles per face (counter-clockwise): 0-1-2 and 0-2-3
		indices = append(indices,
			base+0, base+1, base+2,
			base+0, base+2, base+3,
		)
	}

	return vertices, indices
}
