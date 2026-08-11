package g3d

import (
	"math"
	"testing"
)

// Uses shared test helpers from math_test.go:
//   epsilon, approxEqual, approxEqualVec3
// Uses shared test helper from matrix_test.go:
//   approxEqualMat4

// --- NewNode defaults ---

func TestNewNodeDefaults(t *testing.T) {
	n := NewNode()
	if n.Position != (Vec3{}) {
		t.Errorf("expected zero position, got %v", n.Position)
	}
	if n.Rotation != (Euler{}) {
		t.Errorf("expected zero rotation, got %v", n.Rotation)
	}
	if n.Scale != (Vec3{1, 1, 1}) {
		t.Errorf("expected scale {1,1,1}, got %v", n.Scale)
	}
	if !n.Visible() {
		t.Error("expected visible=true")
	}
	if n.Parent() != nil {
		t.Error("expected nil parent")
	}
	if n.ChildCount() != 0 {
		t.Error("expected 0 children")
	}
	if n.Name() != "" {
		t.Errorf("expected empty name, got %q", n.Name())
	}
	if n.UserData() != nil {
		t.Error("expected nil userData")
	}
}

func TestNewNodeScaleNotZero(t *testing.T) {
	n := NewNode()
	// Scale must be {1,1,1}, not {0,0,0}. Zero scale collapses geometry.
	if n.Scale == (Vec3{}) {
		t.Fatal("CRITICAL: Scale default is {0,0,0} — must be {1,1,1}")
	}
}

// --- Name, Visible, UserData ---

func TestNodeSetName(t *testing.T) {
	n := NewNode()
	n.SetName("Cube")
	if n.Name() != "Cube" {
		t.Errorf("expected name Cube, got %q", n.Name())
	}
}

func TestNodeSetVisible(t *testing.T) {
	n := NewNode()
	n.SetVisible(false)
	if n.Visible() {
		t.Error("expected visible=false")
	}
	n.SetVisible(true)
	if !n.Visible() {
		t.Error("expected visible=true")
	}
}

func TestNodeUserData(t *testing.T) {
	n := NewNode()
	n.SetUserData("hello")
	if n.UserData() != "hello" {
		t.Errorf("expected userData hello, got %v", n.UserData())
	}
}

// --- Parent-Child relationships ---

func TestNodeAddChild(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	parent.Add(child)

	if child.Parent() != parent {
		t.Error("child parent should be parent")
	}
	if parent.ChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", parent.ChildCount())
	}
}

func TestNodeAddChildReparents(t *testing.T) {
	oldParent := NewNode()
	newParent := NewNode()
	child := NewNode()

	oldParent.Add(child)
	if child.Parent() != oldParent {
		t.Error("child should be under oldParent")
	}

	newParent.Add(child)
	if child.Parent() != newParent {
		t.Error("child should be reparented to newParent")
	}
	if oldParent.ChildCount() != 0 {
		t.Error("oldParent should have 0 children after reparent")
	}
	if newParent.ChildCount() != 1 {
		t.Error("newParent should have 1 child")
	}
}

func TestNodeRemoveChild(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	parent.Add(child)
	parent.Remove(child)

	if child.Parent() != nil {
		t.Error("child parent should be nil after remove")
	}
	if parent.ChildCount() != 0 {
		t.Error("parent should have 0 children")
	}
}

func TestNodeRemoveNonChild(t *testing.T) {
	parent := NewNode()
	other := NewNode()
	// Should not panic.
	parent.Remove(other)
	if parent.ChildCount() != 0 {
		t.Error("no children should have been affected")
	}
}

func TestNodeAddNil(t *testing.T) {
	n := NewNode()
	n.Add(nil) // Should not panic.
	if n.ChildCount() != 0 {
		t.Error("adding nil should be a no-op")
	}
}

func TestNodeAddSelf(t *testing.T) {
	n := NewNode()
	n.Add(n) // Should not panic or create a cycle.
	if n.ChildCount() != 0 {
		t.Error("adding self should be a no-op")
	}
}

func TestNodeAddCircular(t *testing.T) {
	a := NewNode()
	b := NewNode()
	c := NewNode()
	a.Add(b)
	b.Add(c)

	// c is a descendant of a. Adding a as child of c would create a cycle.
	c.Add(a)
	if a.Parent() == c {
		t.Error("adding ancestor as child should be prevented")
	}
}

func TestNodeRemoveNil(t *testing.T) {
	n := NewNode()
	n.Remove(nil) // Should not panic.
}

// --- Children() returns a copy ---

func TestNodeChildrenReturnsCopy(t *testing.T) {
	parent := NewNode()
	child1 := NewNode()
	child2 := NewNode()
	parent.Add(child1)
	parent.Add(child2)

	children := parent.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	// Mutate the returned slice — should NOT affect the node.
	children[0] = nil
	if parent.Children()[0] == nil {
		t.Error("Children() must return a copy, not the internal slice")
	}
}

func TestNodeChildrenEmptyReturnsNil(t *testing.T) {
	n := NewNode()
	if n.Children() != nil {
		t.Error("Children() on a node with no children should return nil")
	}
}

// --- Dirty flag propagation ---

func TestNodeSetPositionDirties(t *testing.T) {
	n := NewNode()
	_ = n.LocalMatrix() // Force clean.
	_ = n.WorldMatrix()

	n.SetPosition(Vec3{1, 2, 3})
	if !n.localDirty {
		t.Error("SetPosition should dirty localMatrix")
	}
	if !n.worldDirty {
		t.Error("SetPosition should dirty worldMatrix")
	}
}

func TestNodeSetRotationDirties(t *testing.T) {
	n := NewNode()
	_ = n.LocalMatrix()
	_ = n.WorldMatrix()

	n.SetRotation(Euler{0.1, 0, 0})
	if !n.localDirty {
		t.Error("SetRotation should dirty localMatrix")
	}
}

func TestNodeSetScaleDirties(t *testing.T) {
	n := NewNode()
	_ = n.LocalMatrix()
	_ = n.WorldMatrix()

	n.SetScale(Vec3{2, 2, 2})
	if !n.localDirty {
		t.Error("SetScale should dirty localMatrix")
	}
}

func TestNodeDirtyPropagesToChildren(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	grandchild := NewNode()
	parent.Add(child)
	child.Add(grandchild)

	// Force all clean.
	_ = parent.WorldMatrix()
	_ = child.WorldMatrix()
	_ = grandchild.WorldMatrix()

	// Now dirty the parent.
	parent.SetPosition(Vec3{5, 0, 0})

	if !child.worldDirty {
		t.Error("parent SetPosition should propagate worldDirty to child")
	}
	if !grandchild.worldDirty {
		t.Error("parent SetPosition should propagate worldDirty to grandchild")
	}
}

func TestNodeDirtyPropagationSkipsAlreadyDirty(t *testing.T) {
	// When a node is already worldDirty, the propagation should early-exit
	// to avoid O(N^2) re-traversal. We verify this indirectly: no panic,
	// and the flags are correct.
	parent := NewNode()
	child := NewNode()
	parent.Add(child)

	// Both start dirty (from NewNode + Add). Set one clean, then re-dirty.
	_ = parent.WorldMatrix()
	_ = child.WorldMatrix()

	parent.SetPosition(Vec3{1, 0, 0})
	parent.SetPosition(Vec3{2, 0, 0}) // Second dirty should be fine.

	if !child.worldDirty {
		t.Error("child should be worldDirty")
	}
}

// --- LocalMatrix correctness ---

func TestNodeLocalMatrixIdentity(t *testing.T) {
	n := NewNode()
	got := n.LocalMatrix()
	want := Mat4Identity()
	if got != want {
		t.Errorf("default LocalMatrix should be identity\n  got:  %v\n  want: %v", got, want)
	}
}

func TestNodeLocalMatrixTranslation(t *testing.T) {
	n := NewNode()
	n.SetPosition(Vec3{3, 4, 5})
	m := n.LocalMatrix()

	// Column 3 (indices 12,13,14) holds translation.
	if !approxEqual(m[12], 3, epsilon) || !approxEqual(m[13], 4, epsilon) || !approxEqual(m[14], 5, epsilon) {
		t.Errorf("translation should be (3,4,5), got (%v,%v,%v)", m[12], m[13], m[14])
	}
}

func TestNodeLocalMatrixScale(t *testing.T) {
	n := NewNode()
	n.SetScale(Vec3{2, 3, 4})
	m := n.LocalMatrix()

	// For pure scale, diagonal entries are (sx, sy, sz, 1).
	if !approxEqual(m[0], 2, epsilon) || !approxEqual(m[5], 3, epsilon) || !approxEqual(m[10], 4, epsilon) {
		t.Errorf("scale diagonal should be (2,3,4), got (%v,%v,%v)", m[0], m[5], m[10])
	}
}

func TestNodeLocalMatrixRotation90Y(t *testing.T) {
	n := NewNode()
	n.SetRotation(Euler{0, math.Pi / 2, 0}) // 90 degrees around Y.
	m := n.LocalMatrix()

	// After 90-deg Y rotation:
	// col0 (X-axis) should point toward +Z: (0, 0, -1)
	// col2 (Z-axis) should point toward +X: (1, 0, 0)
	// (Column-major: col0 = m[0..3], col2 = m[8..11])
	if !approxEqual(m[0], 0, epsilon) || !approxEqual(m[2], -1, epsilon) {
		t.Errorf("Y-rotation 90deg: col0 should be ~(0,0,-1), got (%v,%v,%v)", m[0], m[1], m[2])
	}
	if !approxEqual(m[8], 1, epsilon) || !approxEqual(m[10], 0, epsilon) {
		t.Errorf("Y-rotation 90deg: col2 should be ~(1,0,0), got (%v,%v,%v)", m[8], m[9], m[10])
	}
}

func TestNodeLocalMatrixTRS(t *testing.T) {
	// Verify that LocalMatrix = T * R * S
	n := NewNode()
	n.SetPosition(Vec3{1, 2, 3})
	n.SetRotation(Euler{0.3, 0.5, 0.7})
	n.SetScale(Vec3{2, 3, 4})

	got := n.LocalMatrix()

	tr := Mat4Translate(Vec3{1, 2, 3})
	rot := Mat4FromQuat(QuatFromEuler(Euler{0.3, 0.5, 0.7}))
	sc := Mat4Scale(Vec3{2, 3, 4})
	want := tr.Mul(rot).Mul(sc)

	if !approxEqualMat4(got, want, epsilon) {
		t.Errorf("LocalMatrix should equal T*R*S\n  got:  %v\n  want: %v", got, want)
	}
}

func TestNodeLocalMatrixCachedUntilDirty(t *testing.T) {
	n := NewNode()
	m1 := n.LocalMatrix()
	m2 := n.LocalMatrix()
	if m1 != m2 {
		t.Error("LocalMatrix should return the same cached value")
	}
	if n.localDirty {
		t.Error("localDirty should be false after reading LocalMatrix")
	}
}

func TestNodeDirectPositionMutationRefreshesLocalMatrix(t *testing.T) {
	n := NewNode()
	_ = n.LocalMatrix()

	n.Position = Vec3{3, 4, 5}
	got := n.LocalMatrix().Translation()
	want := Vec3{3, 4, 5}
	if !approxEqualVec3(got, want, epsilon) {
		t.Errorf("direct Position mutation: got %v, want %v", got, want)
	}
}

func TestNodeDirectRotationMutationRefreshesLocalMatrix(t *testing.T) {
	n := NewNode()
	_ = n.LocalMatrix()

	n.Rotation = Euler{0, math.Pi / 2, 0}
	m := n.LocalMatrix()
	if !approxEqual(m[0], 0, epsilon) || !approxEqual(m[2], -1, epsilon) {
		t.Errorf("direct Rotation mutation should rotate +X toward -Z, got (%v,%v)", m[0], m[2])
	}
}

func TestNodeDirectScaleMutationRefreshesLocalMatrix(t *testing.T) {
	n := NewNode()
	_ = n.LocalMatrix()

	n.Scale = Vec3{2, 3, 4}
	m := n.LocalMatrix()
	if !approxEqual(m[0], 2, epsilon) || !approxEqual(m[5], 3, epsilon) || !approxEqual(m[10], 4, epsilon) {
		t.Errorf("direct Scale mutation should update diagonal, got (%v,%v,%v)", m[0], m[5], m[10])
	}
}

func TestNodeDirectParentMutationRefreshesChildWorldMatrix(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	child.SetPosition(Vec3{0, 5, 0})
	parent.Add(child)

	if got := child.WorldPosition(); !approxEqualVec3(got, Vec3{0, 5, 0}, epsilon) {
		t.Fatalf("initial child WorldPosition: got %v", got)
	}

	// Mutating the parent's public field must invalidate a cached child world
	// matrix even though no setter was called on either node.
	parent.Position = Vec3{10, 0, 0}
	got := child.WorldPosition()
	want := Vec3{10, 5, 0}
	if !approxEqualVec3(got, want, epsilon) {
		t.Errorf("after direct parent Position mutation: got %v, want %v", got, want)
	}
}

func TestNodeDirectGrandparentMutationRefreshesDescendantWorldMatrix(t *testing.T) {
	grandparent := NewNode()
	parent := NewNode()
	child := NewNode()
	parent.SetPosition(Vec3{0, 5, 0})
	child.SetPosition(Vec3{0, 0, 2})
	grandparent.Add(parent)
	parent.Add(child)

	if got := child.WorldPosition(); !approxEqualVec3(got, Vec3{0, 5, 2}, epsilon) {
		t.Fatalf("initial child WorldPosition: got %v", got)
	}

	grandparent.Position = Vec3{10, 0, 0}
	got := child.WorldPosition()
	want := Vec3{10, 5, 2}
	if !approxEqualVec3(got, want, epsilon) {
		t.Errorf("after direct grandparent Position mutation: got %v, want %v", got, want)
	}
}

func TestNodeCachedMatricesStayCleanWithoutTransformChanges(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	parent.Add(child)
	first := child.WorldMatrix()
	if parent.localDirty || parent.worldDirty || child.localDirty || child.worldDirty {
		t.Fatal("matrix reads should leave a node clean")
	}

	second := child.WorldMatrix()
	if first != second {
		t.Error("unchanged transforms should return the cached world matrix")
	}
	if parent.localDirty || parent.worldDirty || child.localDirty || child.worldDirty {
		t.Error("unchanged transform reads should not dirty either matrix")
	}
}

func TestNodeUnchangedNaNTransformDoesNotDirtyDescendants(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	parent.Add(child)
	nan := math.Float32frombits(0x7fc12345)
	parent.Position = Vec3{nan, nan, nan}
	parent.Rotation = Euler{nan, nan, nan}
	parent.Scale = Vec3{nan, nan, nan}

	// The first read records the exact public-field representation, including
	// the NaN payload, in the transform snapshot.
	_ = child.WorldMatrix()
	if parent.localDirty || parent.worldDirty || child.localDirty || child.worldDirty {
		t.Fatal("initial NaN transform read should leave the hierarchy clean")
	}

	// Re-reading unchanged NaN fields must not look like a new mutation. The
	// dirty flags make this observable without relying on benchmark timing.
	_ = parent.LocalMatrix()
	if parent.localDirty || parent.worldDirty || child.worldDirty {
		t.Error("unchanged NaN transform should not dirty or recompute descendants")
	}
}

// --- WorldMatrix correctness ---

func TestNodeWorldMatrixNoParent(t *testing.T) {
	n := NewNode()
	n.SetPosition(Vec3{1, 2, 3})

	if n.WorldMatrix() != n.LocalMatrix() {
		t.Error("WorldMatrix with no parent should equal LocalMatrix")
	}
}

func TestNodeWorldMatrixParentChild(t *testing.T) {
	parent := NewNode()
	parent.SetPosition(Vec3{10, 0, 0})

	child := NewNode()
	child.SetPosition(Vec3{0, 5, 0})
	parent.Add(child)

	wm := child.WorldMatrix()
	want := parent.LocalMatrix().Mul(child.LocalMatrix())
	if !approxEqualMat4(wm, want, epsilon) {
		t.Errorf("WorldMatrix should be parent.World * child.Local\n  got:  %v\n  want: %v", wm, want)
	}
}

func TestNodeWorldMatrixThreeLevels(t *testing.T) {
	root := NewNode()
	root.SetPosition(Vec3{1, 0, 0})

	mid := NewNode()
	mid.SetPosition(Vec3{0, 2, 0})
	root.Add(mid)

	leaf := NewNode()
	leaf.SetPosition(Vec3{0, 0, 3})
	mid.Add(leaf)

	// Expected: root.Local * mid.Local * leaf.Local
	want := root.LocalMatrix().Mul(mid.LocalMatrix()).Mul(leaf.LocalMatrix())
	got := leaf.WorldMatrix()

	if !approxEqualMat4(got, want, epsilon) {
		t.Errorf("3-level WorldMatrix mismatch\n  got:  %v\n  want: %v", got, want)
	}
}

func TestNodeWorldMatrixTranslationAccumulates(t *testing.T) {
	root := NewNode()
	root.SetPosition(Vec3{1, 0, 0})

	child := NewNode()
	child.SetPosition(Vec3{0, 2, 0})
	root.Add(child)

	grandchild := NewNode()
	grandchild.SetPosition(Vec3{0, 0, 3})
	child.Add(grandchild)

	pos := grandchild.WorldPosition()
	want := Vec3{1, 2, 3}
	if !approxEqualVec3(pos, want, epsilon) {
		t.Errorf("WorldPosition should accumulate: got %v, want %v", pos, want)
	}
}

func TestNodeWorldMatrixWithScaleAndRotation(t *testing.T) {
	parent := NewNode()
	parent.SetScale(Vec3{2, 2, 2})
	parent.SetRotation(Euler{0, math.Pi / 2, 0})

	child := NewNode()
	child.SetPosition(Vec3{1, 0, 0})
	parent.Add(child)

	wp := child.WorldPosition()
	// Parent: scale(2) + rotateY(90deg). Child at (1,0,0).
	// Scale: (1,0,0) -> (2,0,0). RotateY(90): (2,0,0) -> (0,0,-2).
	want := Vec3{0, 0, -2}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("WorldPosition with scale+rotation: got %v, want %v", wp, want)
	}
}

func TestNodeWorldMatrixRecomputesAfterParentChange(t *testing.T) {
	parent := NewNode()
	parent.SetPosition(Vec3{10, 0, 0})

	child := NewNode()
	child.SetPosition(Vec3{0, 5, 0})
	parent.Add(child)

	pos1 := child.WorldPosition()
	if !approxEqualVec3(pos1, Vec3{10, 5, 0}, epsilon) {
		t.Fatalf("initial WorldPosition: got %v", pos1)
	}

	// Move parent — child's world should update.
	parent.SetPosition(Vec3{20, 0, 0})
	pos2 := child.WorldPosition()
	if !approxEqualVec3(pos2, Vec3{20, 5, 0}, epsilon) {
		t.Errorf("after parent move: got %v, want {20,5,0}", pos2)
	}
}

// --- WorldPosition ---

func TestNodeWorldPosition(t *testing.T) {
	n := NewNode()
	n.SetPosition(Vec3{7, 8, 9})
	wp := n.WorldPosition()
	if !approxEqualVec3(wp, Vec3{7, 8, 9}, epsilon) {
		t.Errorf("WorldPosition for root node: got %v", wp)
	}
}

// --- Add/Remove edge cases ---

func TestNodeAddMultipleChildren(t *testing.T) {
	parent := NewNode()
	c1 := NewNode()
	c2 := NewNode()
	c3 := NewNode()
	parent.Add(c1)
	parent.Add(c2)
	parent.Add(c3)

	if parent.ChildCount() != 3 {
		t.Errorf("expected 3 children, got %d", parent.ChildCount())
	}
}

func TestNodeRemoveMiddleChild(t *testing.T) {
	parent := NewNode()
	c1 := NewNode()
	c1.SetName("c1")
	c2 := NewNode()
	c2.SetName("c2")
	c3 := NewNode()
	c3.SetName("c3")
	parent.Add(c1)
	parent.Add(c2)
	parent.Add(c3)

	parent.Remove(c2)
	if parent.ChildCount() != 2 {
		t.Errorf("expected 2 children after remove, got %d", parent.ChildCount())
	}
	children := parent.Children()
	if children[0].Name() != "c1" || children[1].Name() != "c3" {
		t.Errorf("order should be [c1,c3], got [%s,%s]", children[0].Name(), children[1].Name())
	}
}

func TestNodeAddSameChildTwice(t *testing.T) {
	parent := NewNode()
	child := NewNode()
	parent.Add(child)
	parent.Add(child) // Should not duplicate.

	if parent.ChildCount() != 1 {
		t.Errorf("expected 1 child (no duplicate), got %d", parent.ChildCount())
	}
}

// --- LookAt ---

func TestNodeLookAtPositiveZ(t *testing.T) {
	n := NewNode()
	n.SetPosition(Vec3{0, 0, 0})
	n.LookAt(Vec3{0, 0, -10})

	// Looking at -Z from origin should produce roughly identity rotation
	// (camera convention: -Z is forward).
	wp := n.WorldPosition()
	if !approxEqualVec3(wp, Vec3{0, 0, 0}, epsilon) {
		t.Errorf("position should be unchanged, got %v", wp)
	}
}

func TestNodeLookAtParallelAndAntiParallelUp(t *testing.T) {
	tests := []struct {
		name string
		eye  Vec3
	}{
		{name: "parallel", eye: Vec3{0, -10, 0}},
		{name: "anti-parallel", eye: Vec3{0, 10, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNode()
			n.SetPosition(tt.eye)
			n.LookAt(Vec3{})

			world := n.WorldMatrix()
			if !approxEqual(world.Determinant(), 1, 1e-4) {
				t.Fatalf("LookAt world determinant = %v, want 1", world.Determinant())
			}
			view := world.Inverse()
			targetInView := view.MulVec4(Vec4{0, 0, 0, 1})
			if !approxEqual(targetInView.X, 0, 1e-4) || !approxEqual(targetInView.Y, 0, 1e-4) {
				t.Errorf("target in view space = %v, want zero X/Y", targetInView)
			}
			if !approxEqual(targetInView.Z, -10, 1e-4) {
				t.Errorf("target in view Z = %v, want -10", targetInView.Z)
			}
		})
	}
}

func TestNodeLookAtWithTranslatedParent(t *testing.T) {
	parent := NewNode()
	parent.SetPosition(Vec3{5, 2, -3})
	child := NewNode()
	child.SetPosition(Vec3{0, 10, 0})
	parent.Add(child)

	// A translated (but unrotated) parent is supported by Node.LookAt: the
	// child eye is computed in world space before deriving its local rotation.
	child.LookAt(parent.WorldPosition())

	view := child.WorldMatrix().Inverse()
	targetInView := view.MulVec4(Vec4{parent.Position.X, parent.Position.Y, parent.Position.Z, 1})
	if !approxEqual(targetInView.X, 0, 1e-4) || !approxEqual(targetInView.Y, 0, 1e-4) {
		t.Errorf("parent target in child view = %v, want zero X/Y", targetInView)
	}
	if !approxEqual(targetInView.Z, -10, 1e-4) {
		t.Errorf("parent target in child view Z = %v, want -10", targetInView.Z)
	}
}

// --- Deep hierarchy benchmark ---

func TestNodeDeepHierarchyWorldMatrix(t *testing.T) {
	// Build a 10-level chain, each with translation (1,0,0).
	// Leaf WorldPosition should be (10,0,0).
	const depth = 10
	nodes := make([]*Node, depth)
	for i := range depth {
		nodes[i] = NewNode()
		nodes[i].SetPosition(Vec3{1, 0, 0})
		if i > 0 {
			nodes[i-1].Add(nodes[i])
		}
	}

	wp := nodes[depth-1].WorldPosition()
	want := Vec3{float32(depth), 0, 0}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("10-level chain WorldPosition: got %v, want %v", wp, want)
	}
}

// --- Benchmarks ---

func BenchmarkNodeLocalMatrix(b *testing.B) {
	n := NewNode()
	n.SetPosition(Vec3{1, 2, 3})
	n.SetRotation(Euler{0.3, 0.5, 0.7})
	n.SetScale(Vec3{2, 3, 4})

	for range b.N {
		n.localDirty = true
		_ = n.LocalMatrix()
	}
}

func BenchmarkNodeWorldMatrix3Levels(b *testing.B) {
	root := NewNode()
	root.SetPosition(Vec3{1, 0, 0})
	mid := NewNode()
	mid.SetPosition(Vec3{0, 2, 0})
	root.Add(mid)
	leaf := NewNode()
	leaf.SetPosition(Vec3{0, 0, 3})
	mid.Add(leaf)

	for range b.N {
		root.markLocalDirty()
		_ = leaf.WorldMatrix()
	}
}

func BenchmarkNodeAddRemove(b *testing.B) {
	parent := NewNode()
	child := NewNode()

	for range b.N {
		parent.Add(child)
		parent.Remove(child)
	}
}
