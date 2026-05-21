package g3d

// Camera provides the view and projection matrices needed by the renderer.
// Every camera embeds a Node so it can be positioned and rotated within the
// scene graph. Use CameraNode to access the underlying Node for transform
// manipulation (Position, Rotation, LookAt, etc.).
//
// Two concrete implementations are provided:
//   - PerspectiveCamera — field-of-view perspective projection
//   - OrthographicCamera — parallel projection for CAD, 2D-in-3D, UI overlays
type Camera interface {
	// ViewMatrix returns the view (camera-space) matrix. It is the inverse of
	// the camera's world matrix — transforming world coordinates into the
	// camera's local space where the camera sits at the origin looking down -Z.
	ViewMatrix() Mat4

	// ProjectionMatrix returns the projection matrix that maps view-space
	// coordinates into WebGPU clip space (X/Y [-1,1], Z [0,1]).
	ProjectionMatrix() Mat4

	// ViewProjectionMatrix returns Projection * View, suitable for frustum
	// culling and transforming world-space positions directly to clip space.
	ViewProjectionMatrix() Mat4

	// Frustum returns the view frustum extracted from the ViewProjectionMatrix.
	// Used for frustum culling of scene objects.
	Frustum() Frustum

	// CameraNode returns the underlying Node for transform access.
	// Use this to set Position, Rotation, call LookAt, or attach the camera
	// to the scene graph.
	CameraNode() *Node
}
