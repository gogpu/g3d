package g3d

// Group is an empty transform container in the scene graph. It has no visual
// representation — its sole purpose is to group children under a shared local
// transform. Moving/rotating/scaling a Group applies that transform to all of
// its descendants.
//
// Typical uses:
//   - Grouping related objects (e.g., all parts of a robot arm)
//   - Applying a shared offset or scale to a set of meshes
//   - Organizing the scene tree for logical structure
type Group struct {
	Node
}

// NewGroup returns a new Group with identity transform (Scale = {1,1,1}).
func NewGroup() *Group {
	g := &Group{
		Node: *NewNode(),
	}
	g.SetName("Group")
	return g
}
