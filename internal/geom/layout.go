// Package geom implements geometry generation algorithms for g3d.
//
// This package operates on raw []float32 and []uint32 slices to avoid
// circular dependencies with the g3d root package. Vertex data is
// interleaved in standard layout: position(vec3) + normal(vec3) + uv(vec2).
package geom

// StandardVertexStride is the byte size of one interleaved vertex.
// Layout: position(12 bytes) + normal(12 bytes) + uv(8 bytes) = 32 bytes.
const StandardVertexStride = 32

// FloatsPerVertex is the number of float32 values per vertex.
// position(3) + normal(3) + uv(2) = 8 floats.
const FloatsPerVertex = 8

// Vertex attribute offsets in float32 units within a single vertex.
const (
	// OffsetPosition is the float32 offset of the position attribute (X, Y, Z).
	OffsetPosition = 0
	// OffsetNormal is the float32 offset of the normal attribute (X, Y, Z).
	OffsetNormal = 3
	// OffsetUV is the float32 offset of the UV attribute (U, V).
	OffsetUV = 6
)

// appendVertex appends a single interleaved vertex to the slice.
// The vertex consists of: position (px, py, pz), normal (nx, ny, nz), uv (u, v).
func appendVertex(verts []float32, px, py, pz, nx, ny, nz, u, v float32) []float32 {
	return append(verts, px, py, pz, nx, ny, nz, u, v)
}
