package g3d

// AmbientLight provides uniform illumination across the entire scene regardless
// of position or direction. It does not embed Node because ambient light has no
// spatial properties.
//
// Ambient light contributes equally to all surfaces:
//
//	finalColor += ambientColor * ambientIntensity * surfaceColor
//
// Typical usage:
//
//	ambient := g3d.NewAmbientLight(
//	    g3d.WithLightColor(g3d.White),
//	    g3d.WithLightIntensity(0.3),
//	)
type AmbientLight struct {
	color     Color
	intensity float32
}

// NewAmbientLight creates an ambient light with the given options.
// Defaults: color=White, intensity=1.
func NewAmbientLight(opts ...LightOption) *AmbientLight {
	l := &AmbientLight{
		color:     White,
		intensity: 1,
	}
	for _, opt := range opts {
		opt.applyAmbientLight(l)
	}
	return l
}

// LightType returns LightKindAmbient.
func (l *AmbientLight) LightType() LightKind { return LightKindAmbient }

// LightColor returns the ambient light color.
func (l *AmbientLight) LightColor() Color { return l.color }

// LightIntensity returns the ambient light intensity.
func (l *AmbientLight) LightIntensity() float32 { return l.intensity }

// Color returns the ambient light color.
func (l *AmbientLight) Color() Color { return l.color }

// SetColor sets the ambient light color.
func (l *AmbientLight) SetColor(c Color) { l.color = c }

// Intensity returns the ambient light intensity.
func (l *AmbientLight) Intensity() float32 { return l.intensity }

// SetIntensity sets the ambient light intensity.
func (l *AmbientLight) SetIntensity(i float32) { l.intensity = i }

// LightUniform returns the GPU-side representation of this ambient light.
// Direction is zero (unused for ambient); Kind is LightKindAmbient (0).
func (l *AmbientLight) LightUniform() LightUniform {
	return LightUniform{
		Direction: [3]float32{0, 0, 0},
		Kind:      uint32(LightKindAmbient),
		Color:     [3]float32{l.color.R, l.color.G, l.color.B},
		Intensity: l.intensity,
	}
}
