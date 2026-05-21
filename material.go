package g3d

// Shader ID constants. These must match the keys in internal/gpu.ShaderCache.
const (
	shaderBasic    = "basic"
	shaderStandard = "standard"
)

// AlphaMode determines how alpha values are handled during rendering.
// Matches GLTF 2.0 alphaMode and Three.js material side conventions.
type AlphaMode uint8

const (
	// AlphaModeOpaque renders the material fully opaque, ignoring alpha.
	AlphaModeOpaque AlphaMode = iota

	// AlphaModeMask discards fragments with alpha below the cutoff threshold.
	AlphaModeMask

	// AlphaModeBlend enables standard alpha blending (src-over compositing).
	AlphaModeBlend
)

// RenderBucket determines the render order for a material.
// Four buckets follow the Three.js validated pattern (see PHASE1-ARCHITECTURE-VALIDATION.md).
// Within each bucket, draw calls are sorted by pipeline key, material ID, then distance.
type RenderBucket uint8

const (
	// RenderBucketBackground is drawn first. Used for skyboxes and IBL.
	RenderBucketBackground RenderBucket = iota

	// RenderBucketOpaque is drawn second with depth testing and no blending.
	// Front-to-back sorted to minimize overdraw.
	RenderBucketOpaque

	// RenderBucketTransmissive is drawn third. Used for transparent objects
	// that still write to the depth buffer (e.g., refraction).
	RenderBucketTransmissive

	// RenderBucketTransparent is drawn last with alpha blending enabled.
	// Back-to-front sorted for correct compositing.
	RenderBucketTransparent
)

// Material describes the surface appearance of a mesh. Materials provide
// the shader identifier for pipeline caching, the render bucket for draw
// ordering, and the serialized uniform data for GPU upload.
//
// Two materials with the same ShaderID share a render pipeline.
// The UniformData layout must match the corresponding WGSL material struct.
type Material interface {
	// ShaderID returns an identifier used for render pipeline caching.
	// Materials with the same ShaderID share a GPU pipeline.
	ShaderID() string

	// RenderBucket determines which render pass draws this material.
	RenderBucket() RenderBucket

	// DoubleSided reports whether back-face culling is disabled.
	DoubleSided() bool

	// UniformData returns the material-specific uniform bytes for GPU upload.
	// The byte layout must match the WGSL material uniform struct exactly.
	UniformData() []byte
}
