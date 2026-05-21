package g3d

import "github.com/gogpu/g3d/internal/geom"

// PlaneGeometry is a flat rectangular plane in the XZ plane.
//
// The plane is centered at the origin with its normal pointing in the +Y direction.
// This matches the standard ground plane convention used by Three.js and most 3D engines.
// Total: 4 vertices, 6 indices.
//
// Created via NewPlaneGeometry:
//
//	ground := g3d.NewPlaneGeometry(10, 10)    // 10x10 ground plane
//	wall   := g3d.NewPlaneGeometry(5, 3)      // 5x3 plane
type PlaneGeometry struct {
	width, height float32
	vertices      []float32
	indices       []uint32
	aabb          AABB
}

// NewPlaneGeometry creates a plane in the XZ plane with the given width (X) and height (Z).
// The plane is centered at the origin with its normal pointing up (+Y).
func NewPlaneGeometry(width, height float32) *PlaneGeometry {
	verts, idxs := geom.GeneratePlane(width, height)
	return &PlaneGeometry{
		width:    width,
		height:   height,
		vertices: verts,
		indices:  idxs,
		aabb:     computeAABB(verts),
	}
}

// Vertices returns the interleaved vertex data.
func (g *PlaneGeometry) Vertices() []float32 {
	return g.vertices
}

// Indices returns the index buffer.
func (g *PlaneGeometry) Indices() []uint32 {
	return g.indices
}

// VertexCount returns the number of vertices (always 4 for a plane).
func (g *PlaneGeometry) VertexCount() int {
	return len(g.vertices) / geom.FloatsPerVertex
}

// BoundingBox returns the axis-aligned bounding box.
func (g *PlaneGeometry) BoundingBox() AABB {
	return g.aabb
}

// Width returns the plane width along the X axis.
func (g *PlaneGeometry) Width() float32 { return g.width }

// Height returns the plane height along the Z axis.
func (g *PlaneGeometry) Height() float32 { return g.height }
