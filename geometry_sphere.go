package g3d

import "github.com/gogpu/g3d/internal/geom"

// SphereGeometry is a UV sphere geometry centered at the origin.
//
// The sphere uses latitudinal/longitudinal subdivision with proper pole handling
// (single vertex at each pole, triangle fans instead of degenerate quads).
//
// Created via NewSphereGeometry with optional segment configuration:
//
//	sphere := g3d.NewSphereGeometry(1.0)                              // default 32x16 segments
//	sphere := g3d.NewSphereGeometry(0.5, g3d.WithSegments(64, 32))   // high detail
type SphereGeometry struct {
	radius                        float32
	widthSegments, heightSegments int
	vertices                      []float32
	indices                       []uint32
	aabb                          AABB
}

// SphereOption configures sphere geometry generation.
type SphereOption func(*sphereConfig)

type sphereConfig struct {
	widthSegments  int
	heightSegments int
}

// WithSegments sets the number of longitudinal (width) and latitudinal (height) segments.
// Width segments must be at least 3, height segments at least 2.
func WithSegments(width, height int) SphereOption {
	return func(c *sphereConfig) {
		c.widthSegments = width
		c.heightSegments = height
	}
}

// NewSphereGeometry creates a UV sphere with the given radius and optional segment configuration.
// Default segments are 32 (width) and 16 (height) if no options are provided.
func NewSphereGeometry(radius float32, opts ...SphereOption) *SphereGeometry {
	cfg := sphereConfig{
		widthSegments:  geom.DefaultWidthSegments,
		heightSegments: geom.DefaultHeightSegments,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	verts, idxs := geom.GenerateSphere(radius, cfg.widthSegments, cfg.heightSegments)
	return &SphereGeometry{
		radius:         radius,
		widthSegments:  cfg.widthSegments,
		heightSegments: cfg.heightSegments,
		vertices:       verts,
		indices:        idxs,
		aabb:           computeAABB(verts),
	}
}

// Vertices returns the interleaved vertex data.
func (g *SphereGeometry) Vertices() []float32 {
	return g.vertices
}

// Indices returns the index buffer.
func (g *SphereGeometry) Indices() []uint32 {
	return g.indices
}

// VertexCount returns the number of vertices.
func (g *SphereGeometry) VertexCount() int {
	return len(g.vertices) / geom.FloatsPerVertex
}

// BoundingBox returns the axis-aligned bounding box.
func (g *SphereGeometry) BoundingBox() AABB {
	return g.aabb
}

// Radius returns the sphere radius.
func (g *SphereGeometry) Radius() float32 { return g.radius }

// WidthSegments returns the number of longitudinal segments.
func (g *SphereGeometry) WidthSegments() int { return g.widthSegments }

// HeightSegments returns the number of latitudinal segments.
func (g *SphereGeometry) HeightSegments() int { return g.heightSegments }
