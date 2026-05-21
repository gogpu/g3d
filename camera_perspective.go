package g3d

import "math"

// PerspectiveCamera implements the Camera interface with a perspective
// projection. Objects farther from the camera appear smaller, matching human
// visual perception.
//
// The field of view is specified in degrees (user-friendly) but stored and
// used in radians internally. The projection uses WebGPU clip space (Z [0,1]).
//
// Position, rotation, and scene-graph hierarchy are managed by the embedded
// Node, accessible via CameraNode().
//
// Example:
//
//	cam := g3d.NewPerspectiveCamera(75, 16.0/9.0, 0.1, 1000)
//	cam.CameraNode().SetPosition(g3d.Vec3{0, 0, 5})
//	cam.CameraNode().LookAt(g3d.Vec3{0, 0, 0})
type PerspectiveCamera struct {
	node Node

	fovRadians float32 // vertical FOV stored in radians
	aspect     float32 // width / height
	near       float32
	far        float32

	projDirty bool
	projCache Mat4
}

// NewPerspectiveCamera creates a PerspectiveCamera with the given vertical
// field of view (in degrees), aspect ratio (width/height), and near/far
// clipping plane distances.
//
// The near plane must be > 0 and the far plane must be > near.
func NewPerspectiveCamera(fovDegrees, aspect, near, far float32) *PerspectiveCamera {
	c := &PerspectiveCamera{
		node:       *NewNode(),
		fovRadians: fovDegrees * math.Pi / 180.0,
		aspect:     aspect,
		near:       near,
		far:        far,
		projDirty:  true,
	}
	c.node.SetName("PerspectiveCamera")
	return c
}

// CameraNode returns the underlying Node for transform access.
func (c *PerspectiveCamera) CameraNode() *Node {
	return &c.node
}

// FOV returns the vertical field of view in degrees.
func (c *PerspectiveCamera) FOV() float32 {
	return c.fovRadians * 180.0 / math.Pi
}

// SetFOV sets the vertical field of view in degrees and invalidates the
// cached projection matrix.
func (c *PerspectiveCamera) SetFOV(degrees float32) {
	c.fovRadians = degrees * math.Pi / 180.0
	c.projDirty = true
}

// Aspect returns the aspect ratio (width / height).
func (c *PerspectiveCamera) Aspect() float32 {
	return c.aspect
}

// SetAspect sets the aspect ratio (width / height) and invalidates the
// cached projection matrix. Call this when the viewport is resized.
func (c *PerspectiveCamera) SetAspect(aspect float32) {
	c.aspect = aspect
	c.projDirty = true
}

// Near returns the near clipping plane distance.
func (c *PerspectiveCamera) Near() float32 {
	return c.near
}

// Far returns the far clipping plane distance.
func (c *PerspectiveCamera) Far() float32 {
	return c.far
}

// SetClipPlanes sets the near and far clipping plane distances and
// invalidates the cached projection matrix.
func (c *PerspectiveCamera) SetClipPlanes(near, far float32) {
	c.near = near
	c.far = far
	c.projDirty = true
}

// ViewMatrix returns the view matrix, which is the inverse of the camera's
// world matrix. This transforms world-space coordinates into camera-local
// space where the camera is at the origin looking down -Z.
func (c *PerspectiveCamera) ViewMatrix() Mat4 {
	return c.node.WorldMatrix().Inverse()
}

// ProjectionMatrix returns the perspective projection matrix using WebGPU
// clip space (Z [0,1]). The matrix is cached and only recomputed when FOV,
// aspect, near, or far changes.
func (c *PerspectiveCamera) ProjectionMatrix() Mat4 {
	if c.projDirty {
		c.projCache = Mat4Perspective(c.fovRadians, c.aspect, c.near, c.far)
		c.projDirty = false
	}
	return c.projCache
}

// ViewProjectionMatrix returns Projection * View. This single matrix
// transforms world-space positions directly to clip space.
func (c *PerspectiveCamera) ViewProjectionMatrix() Mat4 {
	return c.ProjectionMatrix().Mul(c.ViewMatrix())
}

// Frustum returns the view frustum planes extracted from the current
// view-projection matrix. Used for frustum culling.
func (c *PerspectiveCamera) Frustum() Frustum {
	return FrustumFromMat4(c.ViewProjectionMatrix())
}
