package g3d

// DirectionalLight represents a light source that illuminates all objects from
// the same direction, like sunlight. It embeds Node for scene graph integration
// — the direction is derived from the node's world rotation.
//
// The default forward vector is (0, 0, -1). A freshly created DirectionalLight
// with no rotation shines along (0, 0, -1). Rotating the node changes the
// effective light direction accordingly.
//
// Typical usage:
//
//	sun := g3d.NewDirectionalLight(
//	    g3d.WithLightColor(g3d.White),
//	    g3d.WithLightIntensity(1.0),
//	)
//	sun.LightNode().SetRotation(g3d.Euler{X: g3d.Radians(-45), Y: 0, Z: 0})
//	scene.Add(sun.LightNode())
type DirectionalLight struct {
	node      Node
	color     Color
	intensity float32
}

// NewDirectionalLight creates a directional light with the given options.
// Defaults: color=White, intensity=1, direction=(0,0,-1) (from node identity rotation).
func NewDirectionalLight(opts ...LightOption) *DirectionalLight {
	l := &DirectionalLight{
		node:      *NewNode(),
		color:     White,
		intensity: 1,
	}
	l.node.SetName("DirectionalLight")
	// Store self as UserData so the renderer can discover lights during scene traversal.
	l.node.SetUserData(l)
	for _, opt := range opts {
		opt.applyDirectionalLight(l)
	}
	return l
}

// LightNode returns a pointer to the embedded Node for scene graph operations.
// Add this node to a scene or group to position the light in the hierarchy.
func (l *DirectionalLight) LightNode() *Node { return &l.node }

// LightType returns LightKindDirectional.
func (l *DirectionalLight) LightType() LightKind { return LightKindDirectional }

// LightColor returns the directional light color.
func (l *DirectionalLight) LightColor() Color { return l.color }

// LightIntensity returns the directional light intensity.
func (l *DirectionalLight) LightIntensity() float32 { return l.intensity }

// Color returns the directional light color.
func (l *DirectionalLight) Color() Color { return l.color }

// SetColor sets the directional light color.
func (l *DirectionalLight) SetColor(c Color) { l.color = c }

// Intensity returns the directional light intensity.
func (l *DirectionalLight) Intensity() float32 { return l.intensity }

// SetIntensity sets the directional light intensity.
func (l *DirectionalLight) SetIntensity(i float32) { l.intensity = i }

// Direction returns the world-space direction the light is shining towards.
// This is computed by rotating the default forward vector (0, 0, -1) by the
// node's world rotation (extracted from the world matrix).
func (l *DirectionalLight) Direction() Vec3 {
	wm := l.node.WorldMatrix()
	q := quatFromRotationMatrix(wm)
	// Default forward vector for a light: (0, 0, -1).
	forward := Vec3{0, 0, -1}
	return q.RotateVec3(forward).Normalize()
}

// LightUniform returns the GPU-side representation of this directional light.
// The direction is the normalized world-space direction from Direction().
func (l *DirectionalLight) LightUniform() LightUniform {
	dir := l.Direction()
	return LightUniform{
		Direction: [3]float32{dir.X, dir.Y, dir.Z},
		Kind:      uint32(LightKindDirectional),
		Color:     [3]float32{l.color.R, l.color.G, l.color.B},
		Intensity: l.intensity,
	}
}
