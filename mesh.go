package g3d

// Mesh combines a Geometry and a Material into a renderable scene object.
// It participates in the scene graph via an internal Node (accessed through
// MeshNode) and provides the data the Renderer needs to issue draw calls.
//
// Mesh is a pure data container — it does not manage GPU resources. Buffer
// creation and pipeline binding are the Renderer's responsibility (TASK-G3D-007).
//
// The Node is stored by value (not embedded) to avoid method name conflicts
// between Mesh and Node. Use MeshNode() to access the underlying Node for
// transform manipulation (SetPosition, SetRotation, Add children, etc.).
//
// Example:
//
//	cube := g3d.NewMesh(g3d.NewBoxGeometry(1, 1, 1), g3d.NewStandardMaterial())
//	scene.Add(cube.MeshNode())
//	cube.MeshNode().SetPosition(g3d.Vec3{2, 0, 0})
type Mesh struct {
	node     Node
	geometry Geometry
	material Material
}

// NewMesh creates a Mesh from a Geometry and a Material. The internal Node is
// initialized with identity transform (Scale = {1,1,1}) and named "Mesh".
//
// Both geometry and material may be nil at construction time, allowing deferred
// setup. The Renderer will skip meshes that have a nil geometry or material.
func NewMesh(geometry Geometry, material Material) *Mesh {
	m := &Mesh{
		node:     *NewNode(),
		geometry: geometry,
		material: material,
	}
	m.node.SetName("Mesh")
	// Store self as UserData so the renderer can discover meshes during scene traversal.
	m.node.SetUserData(m)
	return m
}

// MeshNode returns a pointer to the underlying Node for scene graph operations.
// Use this to set Position, Rotation, Scale, call LookAt, or add the mesh to
// a scene:
//
//	scene.Add(mesh.MeshNode())
//	mesh.MeshNode().SetPosition(g3d.Vec3{1, 2, 3})
func (m *Mesh) MeshNode() *Node {
	return &m.node
}

// Geometry returns the current Geometry, or nil if none is set.
func (m *Mesh) Geometry() Geometry {
	return m.geometry
}

// Material returns the current Material, or nil if none is set.
func (m *Mesh) Material() Material {
	return m.material
}

// SetGeometry replaces the mesh's Geometry. Pass nil to clear.
func (m *Mesh) SetGeometry(g Geometry) {
	m.geometry = g
}

// SetMaterial replaces the mesh's Material. Changing the material may change
// the ShaderID and thus the render pipeline used for this mesh.
func (m *Mesh) SetMaterial(mat Material) {
	m.material = mat
}

// WorldBoundingBox returns the geometry's axis-aligned bounding box transformed
// into world space by the node's world matrix. This is used for frustum culling.
//
// If geometry is nil, returns a zero AABB.
func (m *Mesh) WorldBoundingBox() AABB {
	if m.geometry == nil {
		return AABB{}
	}
	return m.geometry.BoundingBox().Transform(m.node.WorldMatrix())
}
