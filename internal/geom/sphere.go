package geom

import "math"

// DefaultWidthSegments is the default number of longitudinal segments for a UV sphere.
const DefaultWidthSegments = 32

// DefaultHeightSegments is the default number of latitudinal segments for a UV sphere.
const DefaultHeightSegments = 16

// GenerateSphere creates vertex and index data for a UV sphere.
//
// The sphere is centered at the origin with the given radius.
// widthSegments controls the number of longitudinal divisions (min 3).
// heightSegments controls the number of latitudinal divisions (min 2).
//
// Poles are handled as single shared vertices with triangles (not degenerate quads).
// The equator and intermediate rings use quad strips decomposed into triangle pairs.
//
// Vertex data is interleaved: [px,py,pz, nx,ny,nz, u,v, ...] with 8 floats per vertex.
func GenerateSphere(radius float32, widthSegments, heightSegments int) (vertices []float32, indices []uint32) {
	if widthSegments < 3 {
		widthSegments = 3
	}
	if heightSegments < 2 {
		heightSegments = 2
	}

	// Vertex count: (widthSegments+1) * (heightSegments+1)
	// This includes duplicated vertices at the seam (u=0 and u=1) for correct UV mapping.
	vertCount := (widthSegments + 1) * (heightSegments + 1)
	vertices = make([]float32, 0, vertCount*FloatsPerVertex)

	// Index count: poles contribute widthSegments * 3 each (triangles),
	// body rings contribute widthSegments * 6 each (two triangles per quad).
	idxCount := widthSegments*3*2 + widthSegments*6*(heightSegments-2)
	indices = make([]uint32, 0, idxCount)

	// Generate vertices row by row from top pole (phi=0) to bottom pole (phi=pi).
	for y := 0; y <= heightSegments; y++ {
		phi := float32(y) * math.Pi / float32(heightSegments) // [0, pi]
		sinPhi := float32(math.Sin(float64(phi)))
		cosPhi := float32(math.Cos(float64(phi)))

		for x := 0; x <= widthSegments; x++ {
			theta := float32(x) * 2.0 * math.Pi / float32(widthSegments) // [0, 2*pi]
			sinTheta := float32(math.Sin(float64(theta)))
			cosTheta := float32(math.Cos(float64(theta)))

			// Position on unit sphere, then scale by radius.
			nx := cosTheta * sinPhi
			ny := cosPhi
			nz := sinTheta * sinPhi

			px := nx * radius
			py := ny * radius
			pz := nz * radius

			// UV: u goes from 0 (left) to 1 (right), v from 0 (top) to 1 (bottom).
			u := float32(x) / float32(widthSegments)
			v := float32(y) / float32(heightSegments)

			vertices = appendVertex(vertices, px, py, pz, nx, ny, nz, u, v)
		}
	}

	// Generate indices.
	// Each row has (widthSegments+1) vertices.
	stride := uint32(widthSegments + 1)

	for y := 0; y < heightSegments; y++ {
		for x := 0; x < widthSegments; x++ {
			// Four corners of the current quad.
			a := uint32(y)*stride + uint32(x)   // top-left
			b := a + 1                          // top-right
			c := uint32(y+1)*stride + uint32(x) // bottom-left
			d := c + 1                          // bottom-right

			switch y {
			case 0:
				// Top pole: single triangle (a is the pole vertex).
				indices = append(indices, a, c, d)
			case heightSegments - 1:
				// Bottom pole: single triangle (c is the pole vertex).
				indices = append(indices, a, c, b)
			default:
				// Body: two triangles forming a quad.
				indices = append(indices,
					a, c, d,
					a, d, b,
				)
			}
		}
	}

	return vertices, indices
}
