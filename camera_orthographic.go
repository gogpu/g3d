package g3d

// OrthographicCamera implements the Camera interface with an orthographic
// (parallel) projection. Objects retain their size regardless of distance
// from the camera, making this suitable for CAD viewers, 2D-in-3D overlays,
// isometric games, and technical visualization.
//
// The viewing volume is defined by six planes: left, right, bottom, top,
// near, and far. The projection uses WebGPU clip space (Z [0,1]).
//
// Position, rotation, and scene-graph hierarchy are managed by the embedded
// Node, accessible via CameraNode().
//
// Example:
//
//	cam := g3d.NewOrthographicCamera(-10, 10, -10, 10, 0.1, 100)
//	cam.CameraNode().SetPosition(g3d.Vec3{0, 10, 0})
//	cam.CameraNode().LookAt(g3d.Vec3{0, 0, 0})
type OrthographicCamera struct {
	node Node

	left   float32
	right  float32
	bottom float32
	top    float32
	near   float32
	far    float32

	projDirty bool
	projCache Mat4
}

// NewOrthographicCamera creates an OrthographicCamera with the given viewing
// volume bounds and near/far clipping plane distances. The parameters define
// the edges of the orthographic projection box in camera space.
func NewOrthographicCamera(left, right, bottom, top, near, far float32) *OrthographicCamera {
	c := &OrthographicCamera{
		node:      *NewNode(),
		left:      left,
		right:     right,
		bottom:    bottom,
		top:       top,
		near:      near,
		far:       far,
		projDirty: true,
	}
	c.node.SetName("OrthographicCamera")
	return c
}

// CameraNode returns the underlying Node for transform access.
func (c *OrthographicCamera) CameraNode() *Node {
	return &c.node
}

// Left returns the left edge of the viewing volume.
func (c *OrthographicCamera) Left() float32 { return c.left }

// Right returns the right edge of the viewing volume.
func (c *OrthographicCamera) Right() float32 { return c.right }

// Bottom returns the bottom edge of the viewing volume.
func (c *OrthographicCamera) Bottom() float32 { return c.bottom }

// Top returns the top edge of the viewing volume.
func (c *OrthographicCamera) Top() float32 { return c.top }

// Near returns the near clipping plane distance.
func (c *OrthographicCamera) Near() float32 { return c.near }

// Far returns the far clipping plane distance.
func (c *OrthographicCamera) Far() float32 { return c.far }

// SetBounds sets all four edges of the viewing volume and invalidates the
// cached projection matrix.
func (c *OrthographicCamera) SetBounds(left, right, bottom, top float32) {
	c.left = left
	c.right = right
	c.bottom = bottom
	c.top = top
	c.projDirty = true
}

// SetClipPlanes sets the near and far clipping plane distances and
// invalidates the cached projection matrix.
func (c *OrthographicCamera) SetClipPlanes(near, far float32) {
	c.near = near
	c.far = far
	c.projDirty = true
}

// ViewMatrix returns the view matrix, which is the inverse of the camera's
// world matrix. This transforms world-space coordinates into camera-local
// space where the camera is at the origin looking down -Z.
func (c *OrthographicCamera) ViewMatrix() Mat4 {
	return c.node.WorldMatrix().Inverse()
}

// ProjectionMatrix returns the orthographic projection matrix using WebGPU
// clip space (Z [0,1]). The matrix is cached and only recomputed when the
// bounds or clip planes change.
func (c *OrthographicCamera) ProjectionMatrix() Mat4 {
	if c.projDirty {
		c.projCache = Mat4Ortho(c.left, c.right, c.bottom, c.top, c.near, c.far)
		c.projDirty = false
	}
	return c.projCache
}

// ViewProjectionMatrix returns Projection * View. This single matrix
// transforms world-space positions directly to clip space.
func (c *OrthographicCamera) ViewProjectionMatrix() Mat4 {
	return c.ProjectionMatrix().Mul(c.ViewMatrix())
}

// Frustum returns the view frustum planes extracted from the current
// view-projection matrix. Used for frustum culling.
func (c *OrthographicCamera) Frustum() Frustum {
	return FrustumFromMat4(c.ViewProjectionMatrix())
}
