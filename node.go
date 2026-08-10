package g3d

import "math"

// Node is the base building block of the g3d scene graph. Every object in a
// scene — meshes, cameras, lights, groups — embeds a Node. It holds a local
// transform (Position, Rotation, Scale), a cached local and world matrix with
// dirty-flag propagation, and parent-child relationships.
//
// Setting Position, Rotation, or Scale through the provided setters marks the
// local matrix dirty. Direct mutations of those public fields are detected
// lazily by LocalMatrix and WorldMatrix; setters are preferred when a caller
// needs dirty propagation to happen immediately. A dirty local matrix
// automatically dirties the world matrix of the node and all of its
// descendants, ensuring that WorldMatrix always returns a correct result
// without manual intervention.
//
// Design: Three.js Object3D + Kaiju Transform dirty propagation + idiomatic Go.
type Node struct {
	// Position is the local translation relative to the parent.
	Position Vec3
	// Rotation is the local rotation in radians (intrinsic XYZ order).
	Rotation Euler
	// Scale is the local scale. Default is {1,1,1}.
	Scale Vec3

	name     string
	visible  bool
	parent   *Node
	children []*Node

	// Cached matrices, lazily recomputed on read.
	localMatrix Mat4
	worldMatrix Mat4
	localDirty  bool
	worldDirty  bool

	// Transform snapshots let the lazy matrix accessors detect callers that
	// mutate the public transform fields directly instead of using the setters.
	// The fields remain public for compatibility, so a setter cannot be the only
	// source of dirty state.
	lastPosition      Vec3
	lastRotation      Euler
	lastScale         Vec3
	transformSnapshot bool

	// UserData holds arbitrary application-specific data attached to this node.
	userData any
}

// NewNode returns a new Node with identity transform (Scale = {1,1,1}),
// visible, no parent, and no children.
func NewNode() *Node {
	return &Node{
		Scale:      Vec3{1, 1, 1},
		visible:    true,
		localDirty: true,
		worldDirty: true,
	}
}

// Name returns the node's name. The empty string is the default.
func (n *Node) Name() string { return n.name }

// SetName sets the node's name. Names are not required to be unique.
func (n *Node) SetName(name string) { n.name = name }

// Visible returns whether the node is visible. Invisible nodes (and their
// entire subtree) are skipped during TraverseVisible.
func (n *Node) Visible() bool { return n.visible }

// SetVisible sets the node's visibility flag.
func (n *Node) SetVisible(v bool) { n.visible = v }

// Parent returns the parent node, or nil if this node is a root.
func (n *Node) Parent() *Node { return n.parent }

// UserData returns the application-specific data attached to this node.
func (n *Node) UserData() any { return n.userData }

// SetUserData attaches arbitrary application-specific data to this node.
func (n *Node) SetUserData(data any) { n.userData = data }

// Children returns a shallow copy of the children slice. The caller may modify
// the returned slice without affecting the node's internal state.
func (n *Node) Children() []*Node {
	if len(n.children) == 0 {
		return nil
	}
	out := make([]*Node, len(n.children))
	copy(out, n.children)
	return out
}

// ChildCount returns the number of direct children.
func (n *Node) ChildCount() int { return len(n.children) }

// Add appends child to this node's children list. If child already has a
// parent, it is first removed from that parent. Circular references (adding
// a node to itself or to one of its descendants) are silently ignored.
func (n *Node) Add(child *Node) {
	if child == nil || child == n {
		return
	}
	// Prevent cycles: if n is a descendant of child, skip.
	if n.isDescendantOf(child) {
		return
	}
	// Detach from old parent.
	if child.parent != nil {
		child.parent.removeChild(child)
	}
	child.parent = n
	n.children = append(n.children, child)
	child.markWorldDirty()
}

// Remove detaches child from this node's children list. If child is not a
// direct child, this is a no-op.
func (n *Node) Remove(child *Node) {
	if child == nil {
		return
	}
	for i, c := range n.children {
		if c == child {
			// Order-preserving removal.
			n.children = append(n.children[:i], n.children[i+1:]...)
			child.parent = nil
			child.markWorldDirty()
			return
		}
	}
}

// SetPosition sets the local position and marks the transform dirty.
func (n *Node) SetPosition(p Vec3) {
	n.Position = p
	n.markLocalDirty()
}

// SetRotation sets the local rotation (radians, intrinsic XYZ) and marks the
// transform dirty.
func (n *Node) SetRotation(r Euler) {
	n.Rotation = r
	n.markLocalDirty()
}

// SetScale sets the local scale and marks the transform dirty.
func (n *Node) SetScale(s Vec3) {
	n.Scale = s
	n.markLocalDirty()
}

// LocalMatrix returns the local transformation matrix, lazily recomputed from
// Position, Rotation, and Scale when dirty.
//
// Local = Translate(Position) * Rotate(QuatFromEuler(Rotation)) * Scale(Scale)
func (n *Node) LocalMatrix() Mat4 {
	n.syncTransformState()
	if n.localDirty {
		t := Mat4Translate(n.Position)
		r := Mat4FromQuat(QuatFromEuler(n.Rotation))
		s := Mat4Scale(n.Scale)
		n.localMatrix = t.Mul(r).Mul(s)
		n.lastPosition = n.Position
		n.lastRotation = n.Rotation
		n.lastScale = n.Scale
		n.transformSnapshot = true
		n.localDirty = false
	}
	return n.localMatrix
}

// WorldMatrix returns the world (model) transformation matrix. It is computed
// as parent.WorldMatrix() * node.LocalMatrix(). If there is no parent, the
// world matrix equals the local matrix. Recomputed lazily when dirty.
func (n *Node) WorldMatrix() Mat4 {
	n.syncTransformState()

	// A parent can also have been changed through its public fields. Resolve
	// it before returning a cached child matrix so that the parent's dirty
	// propagation is observed even when the child itself is otherwise clean.
	var parentWorld Mat4
	if n.parent != nil {
		parentWorld = n.parent.WorldMatrix()
	}

	if n.worldDirty {
		local := n.LocalMatrix()
		if n.parent != nil {
			n.worldMatrix = parentWorld.Mul(local)
		} else {
			n.worldMatrix = local
		}
		n.worldDirty = false
	}
	return n.worldMatrix
}

// syncTransformState detects direct mutations of Position, Rotation, and
// Scale. The public fields predate the setters and cannot be made private
// without breaking callers, so matrix access must validate them before using
// cached values. Dirty propagation still happens only when a value actually
// differs from the last computed local matrix.
func (n *Node) syncTransformState() {
	if n.transformSnapshot &&
		sameVec3Bits(n.Position, n.lastPosition) &&
		sameEulerBits(n.Rotation, n.lastRotation) &&
		sameVec3Bits(n.Scale, n.lastScale) {
		return
	}
	n.markLocalDirty()
}

// sameVec3Bits compares the exact float representation, including NaN
// payloads. Go's == treats NaN as unequal to itself, which would otherwise
// dirty and recompute a matrix on every accessor after a caller assigns an
// unchanged NaN transform.
func sameVec3Bits(a, b Vec3) bool {
	return math.Float32bits(a.X) == math.Float32bits(b.X) &&
		math.Float32bits(a.Y) == math.Float32bits(b.Y) &&
		math.Float32bits(a.Z) == math.Float32bits(b.Z)
}

func sameEulerBits(a, b Euler) bool {
	return math.Float32bits(a.X) == math.Float32bits(b.X) &&
		math.Float32bits(a.Y) == math.Float32bits(b.Y) &&
		math.Float32bits(a.Z) == math.Float32bits(b.Z)
}

// WorldPosition returns the world-space position by extracting column 3 of the
// world matrix. This is equivalent to the translation component after all
// parent transforms have been applied.
func (n *Node) WorldPosition() Vec3 {
	return n.WorldMatrix().Translation()
}

// LookAt orients the node so that its local -Z axis points at the target
// position. The node's world position is used as the eye. The up vector is
// {0,1,0}. This method works correctly only for nodes without rotated parents;
// for arbitrary hierarchies the caller should account for the parent transform.
func (n *Node) LookAt(target Vec3) {
	eye := n.WorldPosition()
	view := Mat4LookAt(eye, target, Vec3{0, 1, 0})
	// LookAt returns a view matrix (camera-space). The model rotation is its
	// inverse. For a pure rotation+translation matrix, the inverse of the
	// rotation is the transpose of the upper-left 3x3 block.
	// Extract the rotation quaternion from the inverted view matrix.
	inv := view.Inverse()
	q := quatFromRotationMatrix(inv)
	n.SetRotation(q.ToEuler())
}

// markLocalDirty marks both local and world matrices as needing recomputation
// and propagates world-dirty to all descendants.
func (n *Node) markLocalDirty() {
	n.localDirty = true
	n.markWorldDirty()
}

// markWorldDirty marks the world matrix as needing recomputation and propagates
// to all descendants. Early-exits if already dirty to avoid redundant traversal.
func (n *Node) markWorldDirty() {
	if n.worldDirty {
		return
	}
	n.worldDirty = true
	for _, child := range n.children {
		child.markWorldDirty()
	}
}

// removeChild removes the given child from the children slice (order-preserving).
func (n *Node) removeChild(child *Node) {
	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			child.parent = nil
			return
		}
	}
}

// isDescendantOf reports whether n is a descendant of (or the same node as) other.
func (n *Node) isDescendantOf(other *Node) bool {
	cur := n
	for cur != nil {
		if cur == other {
			return true
		}
		cur = cur.parent
	}
	return false
}

// quatFromRotationMatrix extracts a quaternion from the upper-left 3x3 rotation
// part of a column-major Mat4. This is the reverse of Quat.ToMat4().
func quatFromRotationMatrix(m Mat4) Quat {
	// m[col*4+row]: m00=m[0], m11=m[5], m22=m[10]
	// m10=m[1], m01=m[4], m20=m[2], m02=m[8], m21=m[9], m12=m[6]
	trace := m[0] + m[5] + m[10]
	var q Quat
	switch {
	case trace > 0:
		s := 0.5 / sqrtf(trace+1.0)
		q.W = 0.25 / s
		q.X = (m[6] - m[9]) * s // (m12 - m21)
		q.Y = (m[8] - m[2]) * s // (m20 - m02)
		q.Z = (m[1] - m[4]) * s // (m01 - m10)
	case m[0] > m[5] && m[0] > m[10]:
		s := 2.0 * sqrtf(1.0+m[0]-m[5]-m[10])
		q.W = (m[6] - m[9]) / s
		q.X = 0.25 * s
		q.Y = (m[4] + m[1]) / s
		q.Z = (m[8] + m[2]) / s
	case m[5] > m[10]:
		s := 2.0 * sqrtf(1.0+m[5]-m[0]-m[10])
		q.W = (m[8] - m[2]) / s
		q.X = (m[4] + m[1]) / s
		q.Y = 0.25 * s
		q.Z = (m[9] + m[6]) / s
	default:
		s := 2.0 * sqrtf(1.0+m[10]-m[0]-m[5])
		q.W = (m[1] - m[4]) / s
		q.X = (m[8] + m[2]) / s
		q.Y = (m[9] + m[6]) / s
		q.Z = 0.25 * s
	}
	return q
}

// sqrtf computes the square root of a float32 value.
func sqrtf(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}
