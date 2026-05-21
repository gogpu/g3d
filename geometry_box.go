package g3d

import "github.com/gogpu/g3d/internal/geom"

// BoxGeometry is a rectangular cuboid geometry centered at the origin.
//
// Each face has 4 unique vertices with outward-pointing normals and UV coordinates
// mapped to [0,1]. Total: 24 vertices, 36 indices.
//
// Created via NewBoxGeometry:
//
//	cube := g3d.NewBoxGeometry(1, 1, 1)        // unit cube
//	box  := g3d.NewBoxGeometry(2, 0.5, 3)      // custom dimensions
type BoxGeometry struct {
	width, height, depth float32
	vertices             []float32
	indices              []uint32
	aabb                 AABB
}

// NewBoxGeometry creates a box geometry with the given width (X), height (Y), and depth (Z).
// The box is centered at the origin.
func NewBoxGeometry(width, height, depth float32) *BoxGeometry {
	verts, idxs := geom.GenerateBox(width, height, depth)
	return &BoxGeometry{
		width:    width,
		height:   height,
		depth:    depth,
		vertices: verts,
		indices:  idxs,
		aabb:     computeAABB(verts),
	}
}

// Vertices returns the interleaved vertex data.
func (g *BoxGeometry) Vertices() []float32 {
	return g.vertices
}

// Indices returns the index buffer.
func (g *BoxGeometry) Indices() []uint32 {
	return g.indices
}

// VertexCount returns the number of vertices (always 24 for a box).
func (g *BoxGeometry) VertexCount() int {
	return len(g.vertices) / geom.FloatsPerVertex
}

// BoundingBox returns the axis-aligned bounding box.
func (g *BoxGeometry) BoundingBox() AABB {
	return g.aabb
}

// Width returns the box width along the X axis.
func (g *BoxGeometry) Width() float32 { return g.width }

// Height returns the box height along the Y axis.
func (g *BoxGeometry) Height() float32 { return g.height }

// Depth returns the box depth along the Z axis.
func (g *BoxGeometry) Depth() float32 { return g.depth }
