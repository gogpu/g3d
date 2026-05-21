package g3d

// MaterialOption configures material properties via functional options.
// Options that apply to both BasicMaterial and StandardMaterial implement
// both apply methods; standard-only options are no-ops on BasicMaterial.
type MaterialOption interface {
	applyBasic(m *BasicMaterial)
	applyStandard(m *StandardMaterial)
}

// --- Shared options (apply to both Basic and Standard) ---

type colorOption struct{ c Color }

func (o colorOption) applyBasic(m *BasicMaterial)       { m.color = o.c }
func (o colorOption) applyStandard(m *StandardMaterial) { m.color = o.c }

// WithColor sets the base color. Applies to both BasicMaterial and StandardMaterial.
func WithColor(c Color) MaterialOption { return colorOption{c} }

type opacityOption struct{ v float32 }

func (o opacityOption) applyBasic(m *BasicMaterial)       { m.opacity = o.v }
func (o opacityOption) applyStandard(m *StandardMaterial) { m.opacity = o.v }

// WithOpacity sets the opacity in [0,1]. Applies to both BasicMaterial and StandardMaterial.
// For BasicMaterial, opacity < 1 moves the material to the transparent render bucket.
func WithOpacity(v float32) MaterialOption { return opacityOption{v} }

type doubleSidedOption struct{ v bool }

func (o doubleSidedOption) applyBasic(m *BasicMaterial)       { m.doubleSided = o.v }
func (o doubleSidedOption) applyStandard(m *StandardMaterial) { m.doubleSided = o.v }

// WithDoubleSided disables back-face culling when true. Applies to both materials.
func WithDoubleSided(v bool) MaterialOption { return doubleSidedOption{v} }

type wireframeOption struct{ v bool }

func (o wireframeOption) applyBasic(m *BasicMaterial)       { m.wireframe = o.v }
func (o wireframeOption) applyStandard(m *StandardMaterial) { m.wireframe = o.v }

// WithWireframe enables wireframe rendering when true. Applies to both materials.
func WithWireframe(v bool) MaterialOption { return wireframeOption{v} }

// --- Standard-only options (no-op on BasicMaterial) ---

type metallicOption struct{ v float32 }

func (o metallicOption) applyBasic(*BasicMaterial)         {}
func (o metallicOption) applyStandard(m *StandardMaterial) { m.metallic = o.v }

// WithMetallic sets the metallic factor in [0,1]. 0 = dielectric, 1 = metal.
// Only affects StandardMaterial; ignored by BasicMaterial.
func WithMetallic(v float32) MaterialOption { return metallicOption{v} }

type roughnessOption struct{ v float32 }

func (o roughnessOption) applyBasic(*BasicMaterial)         {}
func (o roughnessOption) applyStandard(m *StandardMaterial) { m.roughness = o.v }

// WithRoughness sets the roughness factor in [0,1]. 0 = mirror, 1 = matte.
// Only affects StandardMaterial; ignored by BasicMaterial.
func WithRoughness(v float32) MaterialOption { return roughnessOption{v} }

type emissiveOption struct{ c Color }

func (o emissiveOption) applyBasic(*BasicMaterial)         {}
func (o emissiveOption) applyStandard(m *StandardMaterial) { m.emissive = o.c }

// WithEmissive sets the emissive color. Objects emit this color regardless of lighting.
// Only affects StandardMaterial; ignored by BasicMaterial.
func WithEmissive(c Color) MaterialOption { return emissiveOption{c} }

type alphaModeOption struct{ v AlphaMode }

func (o alphaModeOption) applyBasic(*BasicMaterial)         {}
func (o alphaModeOption) applyStandard(m *StandardMaterial) { m.alphaMode = o.v }

// WithAlphaMode sets the alpha handling mode (Opaque, Mask, or Blend).
// Only affects StandardMaterial; ignored by BasicMaterial.
func WithAlphaMode(mode AlphaMode) MaterialOption { return alphaModeOption{mode} }

type alphaCutoffOption struct{ v float32 }

func (o alphaCutoffOption) applyBasic(*BasicMaterial)         {}
func (o alphaCutoffOption) applyStandard(m *StandardMaterial) { m.alphaCutoff = o.v }

// WithAlphaCutoff sets the alpha test threshold for AlphaModeMask.
// Fragments with alpha below this value are discarded.
// Only affects StandardMaterial; ignored by BasicMaterial.
func WithAlphaCutoff(cutoff float32) MaterialOption { return alphaCutoffOption{cutoff} }
