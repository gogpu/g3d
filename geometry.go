package g3d

import "github.com/gogpu/g3d/internal/geom"

// Geometry provides vertex and index data for rendering a mesh.
//
// Implementations include BoxGeometry, SphereGeometry, and PlaneGeometry.
// Custom geometry can be created by implementing this interface directly
// or by using NewBufferGeometry with raw vertex/index data.
//
// Vertex data is interleaved in standard layout:
//
//	position(vec3, 12 bytes) + normal(vec3, 12 bytes) + uv(vec2, 8 bytes) = 32 bytes per vertex
//
// This layout matches the WGSL shader attribute locations:
//
//	@location(0) position: vec3<f32>  (offset 0)
//	@location(1) normal:   vec3<f32>  (offset 12)
//	@location(2) uv:       vec2<f32>  (offset 24)
type Geometry interface {
	// Vertices returns the interleaved vertex data as float32 values.
	// Each vertex has 8 floats: position(3) + normal(3) + uv(2).
	Vertices() []float32

	// Indices returns the index buffer for indexed drawing.
	// Returns nil for non-indexed geometry.
	Indices() []uint32

	// VertexCount returns the number of vertices.
	VertexCount() int

	// BoundingBox returns the axis-aligned bounding box enclosing all vertices.
	BoundingBox() AABB
}

// VertexLayout describes the memory layout of interleaved vertex data.
// This is used when creating GPU vertex buffer layouts for the render pipeline.
type VertexLayout struct {
	// Stride is the number of bytes between consecutive vertices.
	Stride uint64

	// Attributes describes each vertex attribute within the stride.
	Attributes []VertexAttribute
}

// VertexAttribute describes a single attribute within a vertex.
type VertexAttribute struct {
	// Name identifies the attribute (e.g. "position", "normal", "uv").
	Name string

	// Offset is the byte offset of this attribute within a vertex.
	Offset uint64

	// FloatCount is the number of float32 components (e.g. 3 for vec3, 2 for vec2).
	FloatCount int
}

// StandardVertexLayout returns the vertex layout used by all built-in geometries.
//
// Layout: position(vec3, offset 0) + normal(vec3, offset 12) + uv(vec2, offset 24).
// Stride: 32 bytes.
func StandardVertexLayout() VertexLayout {
	return VertexLayout{
		Stride: uint64(geom.StandardVertexStride),
		Attributes: []VertexAttribute{
			{Name: "position", Offset: 0, FloatCount: 3},
			{Name: "normal", Offset: 12, FloatCount: 3},
			{Name: "uv", Offset: 24, FloatCount: 2},
		},
	}
}

// BufferGeometry is a concrete Geometry holding precomputed vertex and index data.
//
// Use NewBufferGeometry to create custom geometry from raw data, or use the
// built-in constructors (NewBoxGeometry, NewSphereGeometry, NewPlaneGeometry).
type BufferGeometry struct {
	vertices []float32
	indices  []uint32
	aabb     AABB
}

// NewBufferGeometry creates a Geometry from raw interleaved vertex data and indices.
//
// The vertices slice must contain interleaved data with 8 floats per vertex:
// position(3) + normal(3) + uv(2). The indices slice may be nil for non-indexed geometry.
//
// The bounding box is computed from the vertex positions.
func NewBufferGeometry(vertices []float32, indices []uint32) *BufferGeometry {
	return &BufferGeometry{
		vertices: vertices,
		indices:  indices,
		aabb:     computeAABB(vertices),
	}
}

// Vertices returns the interleaved vertex data.
func (g *BufferGeometry) Vertices() []float32 {
	return g.vertices
}

// Indices returns the index buffer.
func (g *BufferGeometry) Indices() []uint32 {
	return g.indices
}

// VertexCount returns the number of vertices.
func (g *BufferGeometry) VertexCount() int {
	return len(g.vertices) / geom.FloatsPerVertex
}

// BoundingBox returns the axis-aligned bounding box.
func (g *BufferGeometry) BoundingBox() AABB {
	return g.aabb
}

// computeAABB extracts position data from interleaved vertices and computes an AABB.
func computeAABB(vertices []float32) AABB {
	count := len(vertices) / geom.FloatsPerVertex
	if count == 0 {
		return AABB{}
	}

	// Extract the first position to initialize min/max.
	minX := vertices[0]
	minY := vertices[1]
	minZ := vertices[2]
	maxX := minX
	maxY := minY
	maxZ := minZ

	for i := 1; i < count; i++ {
		base := i * geom.FloatsPerVertex
		px := vertices[base]
		py := vertices[base+1]
		pz := vertices[base+2]
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

	return AABB{
		Min: Vec3{minX, minY, minZ},
		Max: Vec3{maxX, maxY, maxZ},
	}
}
