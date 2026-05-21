package g3d

// Scene is the root of a g3d scene graph. It embeds Node and adds a background
// clear color used by the renderer. All objects (meshes, lights, cameras) are
// added as children of the Scene using Add.
//
// Scene provides traversal methods for walking the entire node hierarchy:
//   - Traverse visits every descendant in depth-first order.
//   - TraverseVisible visits only visible descendants, skipping hidden subtrees.
//   - UpdateWorldTransforms forces a top-down recomputation of all dirty world
//     matrices.
type Scene struct {
	Node

	// Background is the clear color used when rendering this scene.
	Background Color
}

// NewScene returns a new Scene with a black background and an identity
// transform. The scene itself is always visible.
func NewScene() *Scene {
	s := &Scene{
		Node:       *NewNode(),
		Background: Black,
	}
	s.SetName("Scene")
	return s
}

// SetBackground sets the background clear color for the scene.
func (s *Scene) SetBackground(c Color) {
	s.Background = c
}

// Traverse walks every descendant node in depth-first, pre-order. The scene's
// own Node is NOT passed to fn — only its descendants are visited.
func (s *Scene) Traverse(fn func(*Node)) {
	traverseDepthFirst(&s.Node, fn)
}

// TraverseVisible walks only visible descendant nodes in depth-first, pre-order.
// If a node is invisible, its entire subtree is skipped. The scene's own Node
// is NOT passed to fn.
func (s *Scene) TraverseVisible(fn func(*Node)) {
	traverseVisibleDepthFirst(&s.Node, fn)
}

// UpdateWorldTransforms forces a top-down traversal that recomputes every
// dirty world matrix. After this call, WorldMatrix() on any node in the scene
// returns the correct value without further lazy computation.
//
// This is useful before rendering to ensure all matrices are up-to-date in a
// single pass rather than on-demand per node.
func (s *Scene) UpdateWorldTransforms() {
	updateWorldTransformsTopDown(&s.Node)
}

// traverseDepthFirst visits all descendants of node in depth-first pre-order.
func traverseDepthFirst(node *Node, fn func(*Node)) {
	for _, child := range node.children {
		fn(child)
		traverseDepthFirst(child, fn)
	}
}

// traverseVisibleDepthFirst visits visible descendants of node in depth-first
// pre-order. If a child is not visible, it and its entire subtree are skipped.
func traverseVisibleDepthFirst(node *Node, fn func(*Node)) {
	for _, child := range node.children {
		if !child.visible {
			continue
		}
		fn(child)
		traverseVisibleDepthFirst(child, fn)
	}
}

// updateWorldTransformsTopDown recursively forces world matrix recomputation
// for all nodes starting from node. This is a breadth-like top-down walk:
// parent matrices are computed before children, ensuring correctness.
func updateWorldTransformsTopDown(node *Node) {
	// Force recomputation of this node's world matrix.
	// WorldMatrix() is lazy — calling it triggers recompute if dirty.
	_ = node.WorldMatrix()
	for _, child := range node.children {
		updateWorldTransformsTopDown(child)
	}
}
