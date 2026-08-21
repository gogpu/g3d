package g3d

import (
	"math"
	"testing"
)

// Uses shared test helpers from math_test.go:
//   epsilon, approxEqual, approxEqualVec3

// --- Construction ---

func TestNewMeshDefaults(t *testing.T) {
	geom := NewBoxGeometry(1, 1, 1)
	mat := NewStandardMaterial()
	m := NewMesh(geom, mat)

	if m.MeshNode().Name() != "Mesh" {
		t.Errorf("expected name Mesh, got %q", m.MeshNode().Name())
	}
	if m.MeshNode().Scale != (Vec3{1, 1, 1}) {
		t.Errorf("expected scale {1,1,1}, got %v", m.MeshNode().Scale)
	}
	if !m.MeshNode().Visible() {
		t.Error("mesh should be visible by default")
	}
	if m.MeshNode().Parent() != nil {
		t.Error("new mesh should have no parent")
	}
	if m.Geometry() != geom {
		t.Error("geometry should match constructor argument")
	}
	if m.Material() != mat {
		t.Error("material should match constructor argument")
	}
}

func TestNewMeshNilGeometryAndMaterial(t *testing.T) {
	m := NewMesh(nil, nil)

	if m.Geometry() != nil {
		t.Error("expected nil geometry")
	}
	if m.Material() != nil {
		t.Error("expected nil material")
	}
	if m.MeshNode().Name() != "Mesh" {
		t.Errorf("expected name Mesh, got %q", m.MeshNode().Name())
	}
}

// --- Geometry access ---

func TestMeshSetGeometry(t *testing.T) {
	m := NewMesh(nil, nil)
	if m.Geometry() != nil {
		t.Fatal("expected nil geometry initially")
	}

	box := NewBoxGeometry(2, 2, 2)
	m.SetGeometry(box)
	if m.Geometry() != box {
		t.Error("SetGeometry should update geometry")
	}

	// Replace with a different geometry.
	sphere := NewSphereGeometry(1, WithSegments(16, 8))
	m.SetGeometry(sphere)
	if m.Geometry() != sphere {
		t.Error("SetGeometry should replace previous geometry")
	}

	// Clear geometry.
	m.SetGeometry(nil)
	if m.Geometry() != nil {
		t.Error("SetGeometry(nil) should clear geometry")
	}
}

// --- Material access ---

func TestMeshSetMaterial(t *testing.T) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewBasicMaterial())
	if m.Material().ShaderID() != "basic" {
		t.Errorf("expected basic material, got %q", m.Material().ShaderID())
	}

	standard := NewStandardMaterial()
	m.SetMaterial(standard)
	if m.Material().ShaderID() != "standard" {
		t.Errorf("expected standard material, got %q", m.Material().ShaderID())
	}

	m.SetMaterial(nil)
	if m.Material() != nil {
		t.Error("SetMaterial(nil) should clear material")
	}
}

// --- MeshNode identity ---

func TestMeshNodeReturnsSamePointer(t *testing.T) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	n1 := m.MeshNode()
	n2 := m.MeshNode()
	if n1 != n2 {
		t.Error("MeshNode() should return the same pointer on every call")
	}
}

func TestMeshNodeTransformPropagates(t *testing.T) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetPosition(Vec3{3, 4, 5})

	wp := m.MeshNode().WorldPosition()
	want := Vec3{3, 4, 5}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("WorldPosition after SetPosition: got %v, want %v", wp, want)
	}
}

// --- WorldBoundingBox ---

func TestMeshWorldBoundingBoxIdentity(t *testing.T) {
	// Unit cube centered at origin: AABB should be [-0.5, 0.5] on all axes.
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())

	bb := m.WorldBoundingBox()
	wantMin := Vec3{-0.5, -0.5, -0.5}
	wantMax := Vec3{0.5, 0.5, 0.5}
	if !approxEqualVec3(bb.Min, wantMin, epsilon) {
		t.Errorf("Min: got %v, want %v", bb.Min, wantMin)
	}
	if !approxEqualVec3(bb.Max, wantMax, epsilon) {
		t.Errorf("Max: got %v, want %v", bb.Max, wantMax)
	}
}

func TestMeshWorldBoundingBoxTranslated(t *testing.T) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetPosition(Vec3{10, 0, 0})

	bb := m.WorldBoundingBox()
	wantMin := Vec3{9.5, -0.5, -0.5}
	wantMax := Vec3{10.5, 0.5, 0.5}
	if !approxEqualVec3(bb.Min, wantMin, epsilon) {
		t.Errorf("Min: got %v, want %v", bb.Min, wantMin)
	}
	if !approxEqualVec3(bb.Max, wantMax, epsilon) {
		t.Errorf("Max: got %v, want %v", bb.Max, wantMax)
	}
}

func TestMeshWorldBoundingBoxScaled(t *testing.T) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetScale(Vec3{2, 3, 4})

	bb := m.WorldBoundingBox()
	wantMin := Vec3{-1, -1.5, -2}
	wantMax := Vec3{1, 1.5, 2}
	if !approxEqualVec3(bb.Min, wantMin, epsilon) {
		t.Errorf("Min: got %v, want %v", bb.Min, wantMin)
	}
	if !approxEqualVec3(bb.Max, wantMax, epsilon) {
		t.Errorf("Max: got %v, want %v", bb.Max, wantMax)
	}
}

func TestMeshWorldBoundingBoxRotated(t *testing.T) {
	// Rotate a unit cube 45 degrees around Y axis.
	// The AABB should grow because the rotated cube has a larger footprint.
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetRotation(Euler{0, math.Pi / 4, 0})

	bb := m.WorldBoundingBox()
	// After 45-degree Y rotation, the XZ extent grows by sqrt(2)/2.
	diag := float32(math.Sqrt(2)) / 2.0
	wantMin := Vec3{-diag, -0.5, -diag}
	wantMax := Vec3{diag, 0.5, diag}
	if !approxEqualVec3(bb.Min, wantMin, epsilon) {
		t.Errorf("Min: got %v, want %v", bb.Min, wantMin)
	}
	if !approxEqualVec3(bb.Max, wantMax, epsilon) {
		t.Errorf("Max: got %v, want %v", bb.Max, wantMax)
	}
}

func TestMeshWorldBoundingBoxWithParentTransform(t *testing.T) {
	// Parent group at (5,0,0), mesh is a child.
	scene := NewScene()
	group := NewGroup()
	group.SetPosition(Vec3{5, 0, 0})
	scene.Add(&group.Node)

	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	group.Add(m.MeshNode())

	scene.UpdateWorldTransforms()

	bb := m.WorldBoundingBox()
	wantMin := Vec3{4.5, -0.5, -0.5}
	wantMax := Vec3{5.5, 0.5, 0.5}
	if !approxEqualVec3(bb.Min, wantMin, epsilon) {
		t.Errorf("Min: got %v, want %v", bb.Min, wantMin)
	}
	if !approxEqualVec3(bb.Max, wantMax, epsilon) {
		t.Errorf("Max: got %v, want %v", bb.Max, wantMax)
	}
}

func TestMeshWorldBoundingBoxNilGeometry(t *testing.T) {
	m := NewMesh(nil, NewStandardMaterial())
	bb := m.WorldBoundingBox()
	zero := AABB{}
	if bb != zero {
		t.Errorf("expected zero AABB for nil geometry, got %v", bb)
	}
}

// --- Scene graph integration ---

func TestMeshAddToScene(t *testing.T) {
	scene := NewScene()
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	scene.Add(m.MeshNode())

	if scene.ChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", scene.ChildCount())
	}
	if m.MeshNode().Parent() != &scene.Node {
		t.Error("mesh node parent should be the scene node")
	}
}

func TestMeshTraversedInScene(t *testing.T) {
	scene := NewScene()
	m1 := NewMesh(NewBoxGeometry(1, 1, 1), NewBasicMaterial())
	m1.MeshNode().SetName("mesh1")
	m2 := NewMesh(NewBoxGeometry(2, 2, 2), NewStandardMaterial())
	m2.MeshNode().SetName("mesh2")
	scene.Add(m1.MeshNode())
	scene.Add(m2.MeshNode())

	var names []string
	scene.Traverse(func(n *Node) {
		names = append(names, n.Name())
	})

	if len(names) != 2 {
		t.Fatalf("expected 2 traversed nodes, got %d", len(names))
	}
	if names[0] != "mesh1" || names[1] != "mesh2" {
		t.Errorf("expected [mesh1, mesh2], got %v", names)
	}
}

func TestMeshInGroupHierarchy(t *testing.T) {
	scene := NewScene()
	group := NewGroup()
	group.SetPosition(Vec3{10, 0, 0})
	group.SetScale(Vec3{2, 2, 2})
	scene.Add(&group.Node)

	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetPosition(Vec3{0, 1, 0})
	group.Add(m.MeshNode())

	scene.UpdateWorldTransforms()

	// World position: group(10,0,0) + group.scale(2) * mesh.local(0,1,0) = (10,2,0)
	wp := m.MeshNode().WorldPosition()
	want := Vec3{10, 2, 0}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("WorldPosition: got %v, want %v", wp, want)
	}
}

func TestMeshVisibilitySkip(t *testing.T) {
	scene := NewScene()
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetVisible(false)
	scene.Add(m.MeshNode())

	count := 0
	scene.TraverseVisible(func(_ *Node) { count++ })
	if count != 0 {
		t.Errorf("hidden mesh should be skipped: got %d visits", count)
	}
}

func TestMeshRemoveFromScene(t *testing.T) {
	scene := NewScene()
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	scene.Add(m.MeshNode())
	scene.Remove(m.MeshNode())

	if scene.ChildCount() != 0 {
		t.Errorf("expected 0 children after removal, got %d", scene.ChildCount())
	}
	if m.MeshNode().Parent() != nil {
		t.Error("removed mesh should have nil parent")
	}
}

// --- Multiple meshes sharing geometry/material ---

func TestMeshesShareGeometry(t *testing.T) {
	shared := NewBoxGeometry(1, 1, 1)
	m1 := NewMesh(shared, NewBasicMaterial())
	m2 := NewMesh(shared, NewStandardMaterial())

	if m1.Geometry() != m2.Geometry() {
		t.Error("both meshes should reference the same geometry")
	}
}

func TestMeshesShareMaterial(t *testing.T) {
	shared := NewStandardMaterial()
	m1 := NewMesh(NewBoxGeometry(1, 1, 1), shared)
	m2 := NewMesh(NewBoxGeometry(2, 2, 2), shared)

	if m1.Material() != m2.Material() {
		t.Error("both meshes should reference the same material")
	}
}

// --- UserData on mesh node ---

func TestMeshNodeUserData(t *testing.T) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetUserData("my-mesh-id")

	if m.MeshNode().UserData() != "my-mesh-id" {
		t.Errorf("expected user data, got %v", m.MeshNode().UserData())
	}
}

// --- Benchmarks ---

func BenchmarkNewMesh(b *testing.B) {
	geom := NewBoxGeometry(1, 1, 1)
	mat := NewStandardMaterial()
	for range b.N {
		_ = NewMesh(geom, mat)
	}
}

func BenchmarkMeshWorldBoundingBox(b *testing.B) {
	m := NewMesh(NewBoxGeometry(1, 1, 1), NewStandardMaterial())
	m.MeshNode().SetPosition(Vec3{5, 3, 1})
	m.MeshNode().SetScale(Vec3{2, 2, 2})

	for range b.N {
		// Dirty the transform each iteration to measure actual recomputation.
		m.MeshNode().SetPosition(Vec3{float32(b.N), 3, 1})
		_ = m.WorldBoundingBox()
	}
}
