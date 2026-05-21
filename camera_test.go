package g3d

import (
	"math"
	"testing"
)

// approxEq compares two float32 values within a tolerance.
func approxEq(a, b, tol float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= tol
}

const testTol = 1e-5

// --- Interface compliance ---

func TestPerspectiveCameraImplementsCamera(t *testing.T) {
	var _ Camera = (*PerspectiveCamera)(nil)
}

func TestOrthographicCameraImplementsCamera(t *testing.T) {
	var _ Camera = (*OrthographicCamera)(nil)
}

// --- PerspectiveCamera construction ---

func TestNewPerspectiveCamera(t *testing.T) {
	cam := NewPerspectiveCamera(75, 16.0/9.0, 0.1, 1000)

	if !approxEq(cam.FOV(), 75, testTol) {
		t.Errorf("FOV: got %f, want 75", cam.FOV())
	}
	if !approxEq(cam.Aspect(), 16.0/9.0, testTol) {
		t.Errorf("Aspect: got %f, want %f", cam.Aspect(), 16.0/9.0)
	}
	if !approxEq(cam.Near(), 0.1, testTol) {
		t.Errorf("Near: got %f, want 0.1", cam.Near())
	}
	if !approxEq(cam.Far(), 1000, testTol) {
		t.Errorf("Far: got %f, want 1000", cam.Far())
	}
	if cam.CameraNode().Name() != "PerspectiveCamera" {
		t.Errorf("Name: got %q, want %q", cam.CameraNode().Name(), "PerspectiveCamera")
	}
}

// --- PerspectiveCamera projection ---

func TestPerspectiveProjectionIdentityView(t *testing.T) {
	// Camera at origin with identity transform. 90-degree FOV, 1:1 aspect.
	cam := NewPerspectiveCamera(90, 1.0, 1.0, 100.0)

	proj := cam.ProjectionMatrix()

	// At 90-degree FOV, tan(45) = 1 => f = 1/tan(45) = 1
	// m[0] = f/aspect = 1/1 = 1
	// m[5] = f = 1
	if !approxEq(proj[0], 1.0, testTol) {
		t.Errorf("proj[0] (X scale): got %f, want 1.0", proj[0])
	}
	if !approxEq(proj[5], 1.0, testTol) {
		t.Errorf("proj[5] (Y scale): got %f, want 1.0", proj[5])
	}

	// m[11] = -1 (perspective divide: w = -z)
	if !approxEq(proj[11], -1.0, testTol) {
		t.Errorf("proj[11] (perspective divide): got %f, want -1.0", proj[11])
	}

	// m[15] = 0 (perspective matrix)
	if !approxEq(proj[15], 0.0, testTol) {
		t.Errorf("proj[15]: got %f, want 0.0", proj[15])
	}
}

func TestPerspectiveProjectionNearFarMapping(t *testing.T) {
	// WebGPU Z [0,1]: near plane maps to z=0, far plane maps to z=1 in NDC.
	cam := NewPerspectiveCamera(60, 1.0, 0.5, 100.0)
	proj := cam.ProjectionMatrix()

	// Transform a point on the near plane (0, 0, -near, 1) in view space.
	nearPoint := Vec4{0, 0, -cam.Near(), 1}
	clipNear := proj.MulVec4(nearPoint)
	// After perspective divide: z_ndc = clip.z / clip.w
	ndcNearZ := clipNear.Z / clipNear.W
	if !approxEq(ndcNearZ, 0.0, testTol) {
		t.Errorf("near plane NDC Z: got %f, want 0.0", ndcNearZ)
	}

	// Transform a point on the far plane (0, 0, -far, 1) in view space.
	farPoint := Vec4{0, 0, -cam.Far(), 1}
	clipFar := proj.MulVec4(farPoint)
	ndcFarZ := clipFar.Z / clipFar.W
	if !approxEq(ndcFarZ, 1.0, testTol) {
		t.Errorf("far plane NDC Z: got %f, want 1.0", ndcFarZ)
	}
}

func TestPerspectiveProjectionAspect(t *testing.T) {
	// Widescreen: aspect > 1 should compress X scale.
	cam := NewPerspectiveCamera(90, 2.0, 0.1, 100)
	proj := cam.ProjectionMatrix()

	// f = 1/tan(45) = 1, m[0] = f/aspect = 0.5
	if !approxEq(proj[0], 0.5, testTol) {
		t.Errorf("X scale with aspect 2: got %f, want 0.5", proj[0])
	}
	// m[5] = f = 1 (Y unchanged by aspect)
	if !approxEq(proj[5], 1.0, testTol) {
		t.Errorf("Y scale with aspect 2: got %f, want 1.0", proj[5])
	}
}

func TestPerspectiveProjectionCaching(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.5, 0.1, 100)

	p1 := cam.ProjectionMatrix()
	p2 := cam.ProjectionMatrix()

	if p1 != p2 {
		t.Error("expected cached projection to return identical result")
	}

	// Change FOV — cache should be invalidated.
	cam.SetFOV(90)
	p3 := cam.ProjectionMatrix()
	if p1 == p3 {
		t.Error("projection should change after SetFOV")
	}
}

func TestPerspectiveSetAspectInvalidatesCache(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	p1 := cam.ProjectionMatrix()

	cam.SetAspect(2.0)
	p2 := cam.ProjectionMatrix()

	if p1 == p2 {
		t.Error("projection should change after SetAspect")
	}
}

func TestPerspectiveSetClipPlanesInvalidatesCache(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	p1 := cam.ProjectionMatrix()

	cam.SetClipPlanes(1.0, 500.0)
	p2 := cam.ProjectionMatrix()

	if p1 == p2 {
		t.Error("projection should change after SetClipPlanes")
	}
	if !approxEq(cam.Near(), 1.0, testTol) {
		t.Errorf("Near: got %f, want 1.0", cam.Near())
	}
	if !approxEq(cam.Far(), 500.0, testTol) {
		t.Errorf("Far: got %f, want 500.0", cam.Far())
	}
}

// --- PerspectiveCamera view matrix ---

func TestPerspectiveViewMatrixAtOrigin(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	// Camera at origin with default rotation => view = identity inverse = identity.
	view := cam.ViewMatrix()
	ident := Mat4Identity()

	for i := 0; i < 16; i++ {
		if !approxEq(view[i], ident[i], testTol) {
			t.Errorf("view[%d]: got %f, want %f", i, view[i], ident[i])
		}
	}
}

func TestPerspectiveViewMatrixTranslated(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	cam.CameraNode().SetPosition(Vec3{0, 0, 5})

	view := cam.ViewMatrix()

	// A world-space point at (0, 0, 5) should map to origin in camera space.
	worldPoint := Vec4{0, 0, 5, 1}
	cameraPoint := view.MulVec4(worldPoint)
	if !approxEq(cameraPoint.X, 0, testTol) || !approxEq(cameraPoint.Y, 0, testTol) || !approxEq(cameraPoint.Z, 0, testTol) {
		t.Errorf("camera origin in camera space: got (%f, %f, %f), want (0, 0, 0)",
			cameraPoint.X, cameraPoint.Y, cameraPoint.Z)
	}

	// A world-space point at origin should be at (0, 0, -5) in camera space.
	originPoint := Vec4{0, 0, 0, 1}
	camOrigin := view.MulVec4(originPoint)
	if !approxEq(camOrigin.X, 0, testTol) || !approxEq(camOrigin.Y, 0, testTol) || !approxEq(camOrigin.Z, -5, testTol) {
		t.Errorf("world origin in camera space: got (%f, %f, %f), want (0, 0, -5)",
			camOrigin.X, camOrigin.Y, camOrigin.Z)
	}
}

func TestPerspectiveViewProjectionMatrix(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	cam.CameraNode().SetPosition(Vec3{0, 0, 5})

	vp := cam.ViewProjectionMatrix()

	// VP should equal Projection * View.
	expected := cam.ProjectionMatrix().Mul(cam.ViewMatrix())
	for i := 0; i < 16; i++ {
		if !approxEq(vp[i], expected[i], testTol) {
			t.Errorf("VP[%d]: got %f, want %f", i, vp[i], expected[i])
		}
	}
}

// --- PerspectiveCamera frustum ---

func TestPerspectiveFrustumContainsVisiblePoint(t *testing.T) {
	cam := NewPerspectiveCamera(90, 1.0, 1.0, 100)
	cam.CameraNode().SetPosition(Vec3{0, 0, 5})

	frustum := cam.Frustum()

	// A point directly in front of the camera within the near/far range should
	// be inside the frustum. The camera is at z=5 looking down -Z, so z=0 is
	// at distance 5 from camera — within [1, 100].
	if !frustum.ContainsPoint(Vec3{0, 0, 0}) {
		t.Error("expected point (0,0,0) to be inside frustum of camera at (0,0,5)")
	}
}

func TestPerspectiveFrustumRejectsBehindCamera(t *testing.T) {
	cam := NewPerspectiveCamera(90, 1.0, 1.0, 100)
	cam.CameraNode().SetPosition(Vec3{0, 0, 5})

	frustum := cam.Frustum()

	// A point behind the camera (z > 5 when camera looks down -Z) should be
	// outside the frustum.
	if frustum.ContainsPoint(Vec3{0, 0, 10}) {
		t.Error("expected point (0,0,10) to be outside frustum (behind camera at z=5)")
	}
}

func TestPerspectiveFrustumRejectsBeyondFar(t *testing.T) {
	cam := NewPerspectiveCamera(90, 1.0, 1.0, 10)
	cam.CameraNode().SetPosition(Vec3{0, 0, 5})

	frustum := cam.Frustum()

	// Camera at z=5 looking down -Z, far=10, so anything beyond z=-5 is out.
	if frustum.ContainsPoint(Vec3{0, 0, -100}) {
		t.Error("expected point (0,0,-100) to be outside frustum (beyond far plane)")
	}
}

// --- PerspectiveCamera LookAt via Node ---

func TestPerspectiveLookAtTarget(t *testing.T) {
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	cam.CameraNode().SetPosition(Vec3{0, 0, 5})
	cam.CameraNode().LookAt(Vec3{0, 0, 0})

	view := cam.ViewMatrix()

	// After LookAt, a point at the target (0,0,0) should project to
	// approximately the center of the view (x=0, y=0 in camera space).
	target := Vec4{0, 0, 0, 1}
	camTarget := view.MulVec4(target)
	if !approxEq(camTarget.X, 0, 0.01) || !approxEq(camTarget.Y, 0, 0.01) {
		t.Errorf("LookAt target in camera space: got (%f, %f), want (~0, ~0)",
			camTarget.X, camTarget.Y)
	}
	// Z should be negative (in front of camera).
	if camTarget.Z > 0 {
		t.Errorf("LookAt target Z should be negative (in front): got %f", camTarget.Z)
	}
}

// --- PerspectiveCamera FOV degrees-radians ---

func TestPerspectiveFOVRoundTrip(t *testing.T) {
	tests := []float32{30, 45, 60, 75, 90, 120}
	for _, fov := range tests {
		cam := NewPerspectiveCamera(fov, 1.0, 0.1, 100)
		if !approxEq(cam.FOV(), fov, testTol) {
			t.Errorf("FOV round-trip: set %f, got %f", fov, cam.FOV())
		}

		cam.SetFOV(fov + 10)
		if !approxEq(cam.FOV(), fov+10, testTol) {
			t.Errorf("SetFOV round-trip: set %f, got %f", fov+10, cam.FOV())
		}
	}
}

// --- OrthographicCamera construction ---

func TestNewOrthographicCamera(t *testing.T) {
	cam := NewOrthographicCamera(-10, 10, -5, 5, 0.1, 100)

	if !approxEq(cam.Left(), -10, testTol) {
		t.Errorf("Left: got %f, want -10", cam.Left())
	}
	if !approxEq(cam.Right(), 10, testTol) {
		t.Errorf("Right: got %f, want 10", cam.Right())
	}
	if !approxEq(cam.Bottom(), -5, testTol) {
		t.Errorf("Bottom: got %f, want -5", cam.Bottom())
	}
	if !approxEq(cam.Top(), 5, testTol) {
		t.Errorf("Top: got %f, want 5", cam.Top())
	}
	if !approxEq(cam.Near(), 0.1, testTol) {
		t.Errorf("Near: got %f, want 0.1", cam.Near())
	}
	if !approxEq(cam.Far(), 100, testTol) {
		t.Errorf("Far: got %f, want 100", cam.Far())
	}
	if cam.CameraNode().Name() != "OrthographicCamera" {
		t.Errorf("Name: got %q, want %q", cam.CameraNode().Name(), "OrthographicCamera")
	}
}

// --- OrthographicCamera projection ---

func TestOrthographicProjectionNearFarMapping(t *testing.T) {
	cam := NewOrthographicCamera(-1, 1, -1, 1, 1.0, 100.0)
	proj := cam.ProjectionMatrix()

	// Near plane point (0, 0, -near) should map to NDC Z = 0.
	nearPoint := Vec4{0, 0, -cam.Near(), 1}
	clipNear := proj.MulVec4(nearPoint)
	// Orthographic: w stays 1, so ndc_z = clip_z.
	if !approxEq(clipNear.Z, 0.0, testTol) {
		t.Errorf("ortho near plane NDC Z: got %f, want 0.0", clipNear.Z)
	}

	// Far plane point (0, 0, -far) should map to NDC Z = 1.
	farPoint := Vec4{0, 0, -cam.Far(), 1}
	clipFar := proj.MulVec4(farPoint)
	if !approxEq(clipFar.Z, 1.0, testTol) {
		t.Errorf("ortho far plane NDC Z: got %f, want 1.0", clipFar.Z)
	}
}

func TestOrthographicProjectionBoundsMapping(t *testing.T) {
	cam := NewOrthographicCamera(-10, 10, -5, 5, 0.1, 100)
	proj := cam.ProjectionMatrix()

	// Left edge should map to NDC X = -1.
	leftPoint := Vec4{-10, 0, -1, 1}
	clipLeft := proj.MulVec4(leftPoint)
	if !approxEq(clipLeft.X, -1.0, testTol) {
		t.Errorf("left edge NDC X: got %f, want -1.0", clipLeft.X)
	}

	// Right edge should map to NDC X = +1.
	rightPoint := Vec4{10, 0, -1, 1}
	clipRight := proj.MulVec4(rightPoint)
	if !approxEq(clipRight.X, 1.0, testTol) {
		t.Errorf("right edge NDC X: got %f, want 1.0", clipRight.X)
	}

	// Top edge should map to NDC Y = +1.
	topPoint := Vec4{0, 5, -1, 1}
	clipTop := proj.MulVec4(topPoint)
	if !approxEq(clipTop.Y, 1.0, testTol) {
		t.Errorf("top edge NDC Y: got %f, want 1.0", clipTop.Y)
	}

	// Bottom edge should map to NDC Y = -1.
	bottomPoint := Vec4{0, -5, -1, 1}
	clipBottom := proj.MulVec4(bottomPoint)
	if !approxEq(clipBottom.Y, -1.0, testTol) {
		t.Errorf("bottom edge NDC Y: got %f, want -1.0", clipBottom.Y)
	}
}

func TestOrthographicProjectionCaching(t *testing.T) {
	cam := NewOrthographicCamera(-1, 1, -1, 1, 0.1, 100)

	p1 := cam.ProjectionMatrix()
	p2 := cam.ProjectionMatrix()
	if p1 != p2 {
		t.Error("expected cached ortho projection to return identical result")
	}

	cam.SetBounds(-2, 2, -2, 2)
	p3 := cam.ProjectionMatrix()
	if p1 == p3 {
		t.Error("ortho projection should change after SetBounds")
	}
}

func TestOrthographicSetClipPlanesInvalidatesCache(t *testing.T) {
	cam := NewOrthographicCamera(-1, 1, -1, 1, 0.1, 100)
	p1 := cam.ProjectionMatrix()

	cam.SetClipPlanes(1.0, 500.0)
	p2 := cam.ProjectionMatrix()
	if p1 == p2 {
		t.Error("ortho projection should change after SetClipPlanes")
	}
}

// --- OrthographicCamera view matrix ---

func TestOrthographicViewMatrixTranslated(t *testing.T) {
	cam := NewOrthographicCamera(-10, 10, -10, 10, 0.1, 100)
	cam.CameraNode().SetPosition(Vec3{0, 5, 0})

	view := cam.ViewMatrix()

	// A point at the camera position should map to origin in camera space.
	camPos := Vec4{0, 5, 0, 1}
	result := view.MulVec4(camPos)
	if !approxEq(result.X, 0, testTol) || !approxEq(result.Y, 0, testTol) || !approxEq(result.Z, 0, testTol) {
		t.Errorf("camera position in camera space: got (%f, %f, %f), want (0, 0, 0)",
			result.X, result.Y, result.Z)
	}
}

// --- OrthographicCamera frustum ---

func TestOrthographicFrustumContainsVisiblePoint(t *testing.T) {
	cam := NewOrthographicCamera(-10, 10, -10, 10, 1, 100)
	// Camera at origin looking down -Z.
	frustum := cam.Frustum()

	// Point directly in front, within bounds.
	if !frustum.ContainsPoint(Vec3{0, 0, -5}) {
		t.Error("expected point (0,0,-5) to be inside ortho frustum")
	}
}

func TestOrthographicFrustumRejectsOutside(t *testing.T) {
	cam := NewOrthographicCamera(-10, 10, -10, 10, 1, 100)
	frustum := cam.Frustum()

	// Point outside left boundary.
	if frustum.ContainsPoint(Vec3{-20, 0, -5}) {
		t.Error("expected point (-20,0,-5) to be outside ortho frustum")
	}

	// Point beyond far plane.
	if frustum.ContainsPoint(Vec3{0, 0, -200}) {
		t.Error("expected point (0,0,-200) to be outside ortho frustum (beyond far)")
	}
}

// --- Cross-cutting: Camera with scene graph ---

func TestCameraInSceneGraph(t *testing.T) {
	scene := NewScene()
	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)

	// Add camera node to scene.
	scene.Add(cam.CameraNode())

	if cam.CameraNode().Parent() != &scene.Node {
		t.Error("expected camera node parent to be scene node")
	}

	// Move scene — camera should inherit transform.
	scene.SetPosition(Vec3{10, 0, 0})
	scene.UpdateWorldTransforms()

	worldPos := cam.CameraNode().WorldPosition()
	if !approxEq(worldPos.X, 10, testTol) || !approxEq(worldPos.Y, 0, testTol) || !approxEq(worldPos.Z, 0, testTol) {
		t.Errorf("camera world position: got (%f, %f, %f), want (10, 0, 0)",
			worldPos.X, worldPos.Y, worldPos.Z)
	}
}

func TestCameraViewReflectsSceneGraphTransform(t *testing.T) {
	scene := NewScene()
	group := NewGroup()
	group.SetPosition(Vec3{0, 0, 10})
	scene.Add(&group.Node)

	cam := NewPerspectiveCamera(60, 1.0, 0.1, 100)
	group.Add(cam.CameraNode())
	scene.UpdateWorldTransforms()

	// Camera is child of group at (0,0,10), so camera world position = (0,0,10).
	view := cam.ViewMatrix()
	origin := Vec4{0, 0, 10, 1}
	result := view.MulVec4(origin)
	if !approxEq(result.X, 0, testTol) || !approxEq(result.Y, 0, testTol) || !approxEq(result.Z, 0, testTol) {
		t.Errorf("group-positioned camera: world origin at camera pos should map to (0,0,0), got (%f, %f, %f)",
			result.X, result.Y, result.Z)
	}
}

// --- Three.js validation: 75-degree FOV standard camera ---

func TestPerspectiveMatchesThreeJSStandard(t *testing.T) {
	// Three.js: new THREE.PerspectiveCamera(75, 16/9, 0.1, 1000)
	// with WebGPU coordinate system.
	cam := NewPerspectiveCamera(75, 16.0/9.0, 0.1, 1000)
	proj := cam.ProjectionMatrix()

	// f = 1/tan(37.5 degrees) = 1/tan(0.6545 rad) ≈ 1.3032
	expectedF := float32(1.0 / math.Tan(float64(Radians(75))*0.5))
	expectedXScale := expectedF / (16.0 / 9.0)

	if !approxEq(proj[0], expectedXScale, testTol) {
		t.Errorf("Three.js standard X scale: got %f, want %f", proj[0], expectedXScale)
	}
	if !approxEq(proj[5], expectedF, testTol) {
		t.Errorf("Three.js standard Y scale: got %f, want %f", proj[5], expectedF)
	}
}
