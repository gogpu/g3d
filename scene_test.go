package g3d

import (
	"math"
	"testing"
)

// Uses shared test helpers from math_test.go:
//   epsilon, approxEqual, approxEqualVec3

// --- NewScene defaults ---

func TestNewSceneDefaults(t *testing.T) {
	s := NewScene()
	if s.Name() != "Scene" {
		t.Errorf("expected name Scene, got %q", s.Name())
	}
	if s.Background != Black {
		t.Errorf("expected black background, got %v", s.Background)
	}
	if !s.Visible() {
		t.Error("scene should be visible by default")
	}
	if s.ChildCount() != 0 {
		t.Error("expected 0 children")
	}
}

func TestNewSceneIsNode(t *testing.T) {
	s := NewScene()
	// Scene embeds Node — verify Position/Scale work.
	s.SetPosition(Vec3{1, 2, 3})
	if s.Position != (Vec3{1, 2, 3}) {
		t.Error("Scene should behave as a Node")
	}
	if s.Scale != (Vec3{1, 1, 1}) {
		t.Errorf("Scene Scale should be {1,1,1}, got %v", s.Scale)
	}
}

// --- SetBackground ---

func TestSceneSetBackground(t *testing.T) {
	s := NewScene()
	s.SetBackground(RGB(0.2, 0.3, 0.4))
	if !approxEqual(s.Background.R, 0.2, epsilon) {
		t.Errorf("Background.R should be 0.2, got %v", s.Background.R)
	}
}

// --- Traverse ---

func TestSceneTraverseEmpty(t *testing.T) {
	s := NewScene()
	count := 0
	s.Traverse(func(_ *Node) { count++ })
	if count != 0 {
		t.Errorf("traverse empty scene: expected 0 visits, got %d", count)
	}
}

func TestSceneTraverseFlat(t *testing.T) {
	s := NewScene()
	c1 := NewNode()
	c1.SetName("c1")
	c2 := NewNode()
	c2.SetName("c2")
	c3 := NewNode()
	c3.SetName("c3")
	s.Add(c1)
	s.Add(c2)
	s.Add(c3)

	var names []string
	s.Traverse(func(n *Node) {
		names = append(names, n.Name())
	})
	if len(names) != 3 {
		t.Fatalf("expected 3 visits, got %d", len(names))
	}
	if names[0] != "c1" || names[1] != "c2" || names[2] != "c3" {
		t.Errorf("expected [c1,c2,c3], got %v", names)
	}
}

func TestSceneTraverseDepthFirst(t *testing.T) {
	s := NewScene()
	a := NewNode()
	a.SetName("a")
	b := NewNode()
	b.SetName("b")
	c := NewNode()
	c.SetName("c")
	d := NewNode()
	d.SetName("d")

	// Tree:
	//   Scene
	//   +-- a
	//   |   +-- b
	//   |   +-- c
	//   +-- d
	s.Add(a)
	a.Add(b)
	a.Add(c)
	s.Add(d)

	var order []string
	s.Traverse(func(n *Node) {
		order = append(order, n.Name())
	})
	// Depth-first pre-order: a, b, c, d
	want := []string{"a", "b", "c", "d"}
	if len(order) != len(want) {
		t.Fatalf("expected %d visits, got %d: %v", len(want), len(order), order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("visit[%d]: expected %q, got %q", i, want[i], order[i])
		}
	}
}

func TestSceneTraverseDoesNotIncludeScene(t *testing.T) {
	s := NewScene()
	child := NewNode()
	child.SetName("child")
	s.Add(child)

	visited := false
	s.Traverse(func(n *Node) {
		if n.Name() == "Scene" {
			visited = true
		}
	})
	if visited {
		t.Error("Traverse should NOT include the scene node itself")
	}
}

// --- TraverseVisible ---

func TestSceneTraverseVisibleSkipsHidden(t *testing.T) {
	s := NewScene()
	a := NewNode()
	a.SetName("a")
	b := NewNode()
	b.SetName("b")
	c := NewNode()
	c.SetName("c")

	s.Add(a)
	a.Add(b)
	s.Add(c)

	// Hide a — should skip a and its child b.
	a.SetVisible(false)

	var names []string
	s.TraverseVisible(func(n *Node) {
		names = append(names, n.Name())
	})
	if len(names) != 1 || names[0] != "c" {
		t.Errorf("expected [c], got %v", names)
	}
}

func TestSceneTraverseVisibleAllVisible(t *testing.T) {
	s := NewScene()
	a := NewNode()
	a.SetName("a")
	b := NewNode()
	b.SetName("b")
	s.Add(a)
	a.Add(b)

	var count int
	s.TraverseVisible(func(_ *Node) { count++ })
	if count != 2 {
		t.Errorf("all visible: expected 2 visits, got %d", count)
	}
}

func TestSceneTraverseVisibleHiddenSubtree(t *testing.T) {
	s := NewScene()
	parent := NewNode()
	parent.SetName("parent")
	child := NewNode()
	child.SetName("child")
	grandchild := NewNode()
	grandchild.SetName("grandchild")

	s.Add(parent)
	parent.Add(child)
	child.Add(grandchild)

	// Hide child — grandchild should also be skipped even if visible.
	child.SetVisible(false)

	var names []string
	s.TraverseVisible(func(n *Node) {
		names = append(names, n.Name())
	})
	if len(names) != 1 || names[0] != "parent" {
		t.Errorf("expected [parent], got %v", names)
	}
}

// --- UpdateWorldTransforms ---

func TestSceneUpdateWorldTransforms(t *testing.T) {
	s := NewScene()
	s.SetPosition(Vec3{10, 0, 0})

	child := NewNode()
	child.SetPosition(Vec3{0, 5, 0})
	s.Add(child)

	grandchild := NewNode()
	grandchild.SetPosition(Vec3{0, 0, 3})
	child.Add(grandchild)

	s.UpdateWorldTransforms()

	// After UpdateWorldTransforms, all world matrices should be clean.
	if s.worldDirty {
		t.Error("scene worldDirty should be false after UpdateWorldTransforms")
	}
	if child.worldDirty {
		t.Error("child worldDirty should be false after UpdateWorldTransforms")
	}
	if grandchild.worldDirty {
		t.Error("grandchild worldDirty should be false after UpdateWorldTransforms")
	}

	// Check correctness: worldPosition of grandchild = (10, 5, 3).
	wp := grandchild.WorldPosition()
	want := Vec3{10, 5, 3}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("grandchild WorldPosition: got %v, want %v", wp, want)
	}
}

func TestSceneUpdateWorldTransformsWithScaleRotation(t *testing.T) {
	s := NewScene()
	parent := NewNode()
	parent.SetScale(Vec3{2, 2, 2})
	s.Add(parent)

	child := NewNode()
	child.SetPosition(Vec3{1, 0, 0})
	parent.Add(child)

	s.UpdateWorldTransforms()

	wp := child.WorldPosition()
	want := Vec3{2, 0, 0} // Scale(2) applied to local position (1,0,0).
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("scaled child WorldPosition: got %v, want %v", wp, want)
	}
}

// --- Group ---

func TestNewGroupDefaults(t *testing.T) {
	g := NewGroup()
	if g.Name() != "Group" {
		t.Errorf("expected name Group, got %q", g.Name())
	}
	if g.Scale != (Vec3{1, 1, 1}) {
		t.Errorf("expected scale {1,1,1}, got %v", g.Scale)
	}
	if !g.Visible() {
		t.Error("group should be visible by default")
	}
}

func TestGroupAsTransformContainer(t *testing.T) {
	s := NewScene()
	g := NewGroup()
	g.SetPosition(Vec3{5, 0, 0})
	s.Add(&g.Node)

	child := NewNode()
	child.SetPosition(Vec3{0, 3, 0})
	g.Add(child)

	s.UpdateWorldTransforms()

	wp := child.WorldPosition()
	want := Vec3{5, 3, 0}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("child through Group: got %v, want %v", wp, want)
	}
}

func TestGroupRotationPropagates(t *testing.T) {
	s := NewScene()
	g := NewGroup()
	g.SetRotation(Euler{0, math.Pi / 2, 0}) // 90 deg Y.
	s.Add(&g.Node)

	child := NewNode()
	child.SetPosition(Vec3{1, 0, 0})
	g.Add(child)

	s.UpdateWorldTransforms()

	// rotateY(90) transforms (1,0,0) -> (0,0,-1).
	wp := child.WorldPosition()
	want := Vec3{0, 0, -1}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("rotated group child: got %v, want %v", wp, want)
	}
}

func TestGroupScalePropagates(t *testing.T) {
	g := NewGroup()
	g.SetScale(Vec3{3, 3, 3})

	child := NewNode()
	child.SetPosition(Vec3{2, 0, 0})
	g.Add(child)

	wp := child.WorldPosition()
	want := Vec3{6, 0, 0}
	if !approxEqualVec3(wp, want, epsilon) {
		t.Errorf("scaled group child: got %v, want %v", wp, want)
	}
}

// --- Complex scene graph tests ---

func TestSceneComplexHierarchy(t *testing.T) {
	// Build a scene:
	//   Scene(offset 1,0,0)
	//   +-- Group(scale 2x)
	//   |   +-- node_a(pos 0,1,0)
	//   |   +-- node_b(pos 0,0,1)
	//   +-- node_c(pos 0,0,0)
	s := NewScene()
	s.SetPosition(Vec3{1, 0, 0})

	g := NewGroup()
	g.SetScale(Vec3{2, 2, 2})
	s.Add(&g.Node)

	nodeA := NewNode()
	nodeA.SetName("a")
	nodeA.SetPosition(Vec3{0, 1, 0})
	g.Add(nodeA)

	nodeB := NewNode()
	nodeB.SetName("b")
	nodeB.SetPosition(Vec3{0, 0, 1})
	g.Add(nodeB)

	nodeC := NewNode()
	nodeC.SetName("c")
	s.Add(nodeC)

	s.UpdateWorldTransforms()

	// node_a: scene(1,0,0) * group(scale 2) * node_a(0,1,0) = (1, 2, 0)
	wpA := nodeA.WorldPosition()
	if !approxEqualVec3(wpA, Vec3{1, 2, 0}, epsilon) {
		t.Errorf("node_a: got %v, want {1,2,0}", wpA)
	}

	// node_b: scene(1,0,0) * group(scale 2) * node_b(0,0,1) = (1, 0, 2)
	wpB := nodeB.WorldPosition()
	if !approxEqualVec3(wpB, Vec3{1, 0, 2}, epsilon) {
		t.Errorf("node_b: got %v, want {1,0,2}", wpB)
	}

	// node_c: scene(1,0,0) * identity = (1, 0, 0)
	wpC := nodeC.WorldPosition()
	if !approxEqualVec3(wpC, Vec3{1, 0, 0}, epsilon) {
		t.Errorf("node_c: got %v, want {1,0,0}", wpC)
	}

	// Count total nodes via traverse.
	count := 0
	s.Traverse(func(_ *Node) { count++ })
	// g(Group), node_a, node_b, node_c = 4 total descendants
	if count != 4 {
		t.Errorf("traverse count: got %d, want 4", count)
	}
}

func TestSceneAddRemovePreservesIntegrity(t *testing.T) {
	s := NewScene()
	a := NewNode()
	b := NewNode()
	c := NewNode()
	s.Add(a)
	s.Add(b)
	s.Add(c)

	s.Remove(b)
	if s.ChildCount() != 2 {
		t.Errorf("expected 2 children, got %d", s.ChildCount())
	}

	// Re-add b under a.
	a.Add(b)
	count := 0
	s.Traverse(func(_ *Node) { count++ })
	// a, b (under a), c = 3
	if count != 3 {
		t.Errorf("expected 3 nodes in traversal, got %d", count)
	}
}

func TestSceneReparentDuringTraversal(t *testing.T) {
	// Collect node names, then reparent outside the traversal.
	// (Modifying tree during traversal is undefined — just verify post-reparent state.)
	s := NewScene()
	a := NewNode()
	a.SetName("a")
	b := NewNode()
	b.SetName("b")
	s.Add(a)
	s.Add(b)

	// Reparent b under a.
	a.Add(b)

	var names []string
	s.Traverse(func(n *Node) {
		names = append(names, n.Name())
	})
	// Should be: a, b (depth-first)
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("after reparent: expected [a, b], got %v", names)
	}
	if s.ChildCount() != 1 {
		t.Errorf("scene should have 1 direct child, got %d", s.ChildCount())
	}
}

// --- Benchmarks ---

func BenchmarkSceneTraverse100(b *testing.B) {
	s := NewScene()
	for range 100 {
		n := NewNode()
		s.Add(n)
	}

	for range b.N {
		s.Traverse(func(_ *Node) {})
	}
}

func BenchmarkSceneUpdateWorldTransforms100(b *testing.B) {
	s := NewScene()
	for range 100 {
		n := NewNode()
		n.SetPosition(Vec3{1, 2, 3})
		s.Add(n)
	}

	for range b.N {
		// Dirty all nodes.
		s.SetPosition(Vec3{float32(b.N), 0, 0})
		s.UpdateWorldTransforms()
	}
}
