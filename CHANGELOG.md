# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-21

### Added

**Math Types**
- `Vec2`, `Vec3`, `Vec4` — 2D/3D/4D vectors with full arithmetic (Add, Sub, Scale, Dot, Cross, Normalize, Lerp)
- `Mat4` — 4x4 column-major matrix (Identity, Translate, Scale, Rotate, Perspective, Ortho, LookAt, Inverse, Mul)
- `Quat` — quaternion rotation (FromAxisAngle, FromEuler, Slerp, RotateVec3, ToMat4)
- `Euler` — human-readable rotation in radians
- `Frustum` — 6-plane frustum with AABB intersection test
- `AABB` — axis-aligned bounding box with Arvo transform
- `Color` — RGBA float32 with named constants (White, Black, Red, etc.)
- WebGPU clip space Z [0,1] in all projection matrices

**Scene Graph**
- `Node` — base scene graph type with Position/Rotation/Scale, dirty flag propagation
- `Scene` — root node with Traverse, TraverseVisible, UpdateWorldTransforms
- `Group` — empty transform container
- `Mesh` — renderable combining Geometry + Material
- Parent-child hierarchy with world matrix propagation

**Cameras**
- `Camera` interface with ViewMatrix, ProjectionMatrix, Frustum
- `PerspectiveCamera` — FOV (degrees API), aspect ratio, clip planes
- `OrthographicCamera` — bounds-based projection

**Lights**
- `Light` interface with LightKind enum
- `AmbientLight` — uniform color illumination
- `DirectionalLight` — parallel light with direction from scene graph rotation
- `LightUniform` — 32-byte GPU struct matching WGSL layout

**Materials**
- `Material` interface with ShaderID, RenderBucket, UniformData
- `BasicMaterial` — unlit solid color (16-byte uniform)
- `StandardMaterial` — PBR metallic-roughness with Blinn-Phong shading (32-byte uniform)
- `AlphaMode` — Opaque, Mask, Blend
- `RenderBucket` — Background, Opaque, Transmissive, Transparent (4-bucket Three.js pattern)
- Functional options: WithColor, WithMetallic, WithRoughness, WithEmissive, etc.

**Geometry**
- `Geometry` interface with Vertices, Indices, VertexCount, BoundingBox
- `BoxGeometry` — 24 vertices, 36 indices, outward normals, [0,1] UVs
- `SphereGeometry` — UV sphere with configurable segments, pole handling
- `PlaneGeometry` — XZ plane, +Y normal
- `BufferGeometry` — custom vertex data
- Standard vertex layout: position(vec3) + normal(vec3) + uv(vec2) = 32 bytes

**Forward Renderer**
- `Renderer` with DeviceProvider integration (gogpu) and standalone mode
- 4-bucket render list with 3-key opaque sort (PipelineKey → MaterialID → Distance)
- Pipeline cache with lazy creation
- Shader cache with embedded WGSL (basic + standard Blinn-Phong)
- Per-frame uniform upload via MappedAtCreation (WebGPU compliant)
- Depth24Plus depth buffer
- Frustum culling during scene traversal

**WGSL Shaders**
- `basic.wgsl` — unlit vertex + fragment shader
- `standard.wgsl` — Blinn-Phong lit shader with ambient + directional lights

**Infrastructure**
- GitHub Actions CI (build + test + lint + formatting + deps, 3 platforms)
- golangci-lint v2 configuration with enterprise linter set
- Codecov configuration with GPU code exclusions

**Example**
- `examples/hello-cube/` — rotating lit cube demonstrating full g3d API

[Unreleased]: https://github.com/gogpu/g3d/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/gogpu/g3d/releases/tag/v0.1.0
