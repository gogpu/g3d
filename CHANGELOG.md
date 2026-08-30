# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.10] - 2026-08-30

### Fixed
- **Stencil operations for depth-only formats** — configure `StencilLoadOp`/`StencilStoreOp` based on depth format: `Undefined` + `StencilReadOnly: true` for `Depth24Plus` (no stencil aspect), normal stencil ops for `Depth24PlusStencil8` and `Depth32FloatStencil8`. Fixes WebGPU browser validation rejection ([#38](https://github.com/gogpu/g3d/pull/38), contributed by [@tarmo888](https://github.com/tarmo888))
- **Writable mapped-range API** — migrated `createVertexBuffer`/`createIndexBuffer` from `Bytes()` to `BytesMut()` + `Flush()` for cross-backend compatibility (browser backend requires explicit flush) ([#38](https://github.com/gogpu/g3d/pull/38), contributed by [@tarmo888](https://github.com/tarmo888))

### Changed
- **deps:** wgpu v0.31.4 → v0.33.0, gpucontext v0.28.0 → v0.31.2, gputypes v0.5.2 → v0.7.0, naga v0.18.0 → v0.19.0
- **deps:** removed stale `gogpu` direct dependency (g3d only depends down: wgpu, gpucontext, gputypes)

## [0.1.9] - 2026-08-14

### Added
- **gopher example** — 3D Go Gopher mascot assembled from basic primitives (spheres + boxes), demonstrating hierarchical scene graph with `Group`, multiple PBR materials, 3-light setup, and `dt`-based rotation animation

### Changed
- **deps:** wgpu v0.31.2 → v0.31.4, gpucontext v0.27.0 → v0.28.0

## [0.1.8] - 2026-08-11

### Fixed
- **Duplicate ambient lights** — `collectAmbientLights` redundantly scanned direct children after `TraverseVisible` already discovered them, causing double ambient contribution ([#26](https://github.com/gogpu/g3d/pull/26))
- **Non-indexed geometry rendering** — `recordDrawCall` unconditionally called `DrawIndexed`, now branches to `Draw` when `Indices()` returns nil/empty ([#31](https://github.com/gogpu/g3d/pull/31))
- **LookAt with parallel up vectors** — `Mat4LookAt` produced NaN/Inf when camera looked directly along the up vector (top-down/bottom-up views), now falls back to perpendicular axis ([#29](https://github.com/gogpu/g3d/pull/29))
- **Stale matrices after direct transform edits** — public field mutation (`node.Position.X = 10`) bypassed dirty flags, now detected via transform snapshots on every matrix read ([#27](https://github.com/gogpu/g3d/pull/27))
- **Directional light distorted by scale** — `Direction()` extracted quaternion from world matrix including scale, now composes quaternions directly via `worldRotation()` ([#28](https://github.com/gogpu/g3d/pull/28))

### Changed
- **deps:** wgpu v0.31.0 → v0.31.2, gputypes v0.5.1 → v0.5.2
- **deps (examples):** gogpu v0.52.1, gg v0.52.2, gpucontext v0.27.0, ui v0.1.53, wgpu v0.31.2

## [0.1.7] - 2026-08-10

### Changed
- **Damage source registration** (ADR-065) — migrated from `SetDamageRects` + `MarkExternalContent` to `RegisterDamageSource` + `MarkPreserveContent`. g3d registers as named damage source ("g3d"), reports full-viewport damage each frame. Compositor unions all sources at present time.
- **deps:** gpucontext v0.24.0 → v0.27.0 (SurfaceCompositor, DamageSource interfaces), wgpu v0.30.37 → v0.31.0 (core resource tracker, inline present barrier)

## [0.1.6] - 2026-08-07

### Changed
- **deps:** wgpu v0.30.36 → v0.30.37 (Vulkan present layout alignment — `COLOR_ATTACHMENT_OPTIMAL` instead of `PRESENT_SRC_KHR` for intermediate render passes, fixes TBDR glitch on Asahi Linux [#22](https://github.com/gogpu/g3d/issues/22))
- **deps:** examples updated to gogpu v0.50.2, gg v0.50.14, ui v0.1.51

## [0.1.5] - 2026-08-06

### Fixed
- **fullscreen-overlay Vulkan present race** — switched example from `Render()` (own encoder + submit) to `RenderTo()` + shared framework encoder, eliminating a dual-submit race on TBDR GPUs (Apple Silicon via Mesa Asahi). Screen capture showed correct buffer, but physical display showed stale frames due to present semaphore consumed by first submit only ([#22](https://github.com/gogpu/g3d/issues/22))

### Changed
- **deps:** wgpu v0.30.34 → v0.30.36 (Vulkan accumulated present semaphores — multi-submit sync fix)

## [0.1.4] - 2026-08-02

### Added
- **fullscreen-overlay example** — g3d + gg compositing: rotating PBR cube with 2D HUD overlay (title, FPS counter, crosshair, status bar). Demonstrates enterprise multi-pass rendering: g3d pass (LoadOp::Clear + depth) → MarkExternalContent → gg pass (LoadOp::Load + alpha blend)

### Fixed
- **GPU buffer lifecycle** — uniform buffers (frame, object, material) were released via `defer` before async `queue.Submit()` completed on GPU, causing "command buffer references released buffer" errors. Replaced per-frame MappedAtCreation+Release with persistent buffers + `queue.WriteBuffer()` following gg's enterprise pattern. Geometry buffers cached by identity. Bind groups deferred-released next frame

### Changed
- **viewport3d example:** `ui/core/viewport3d` → `ui/core/gpuview` (follow ui rename)
- **deps:** wgpu v0.30.29 → v0.30.34, gpucontext v0.23.0 → v0.24.0, naga v0.17.16 → v0.18.0, goffi v0.6.2 → v0.6.3

## [0.1.3] - 2026-07-28

### Added

- **`Renderer.RenderTo`** — record g3d render passes into a caller-owned `wgpu.CommandEncoder` without finishing or submitting, enabling one-submit composition with 2D overlays and other renderers ([#12](https://github.com/gogpu/g3d/pull/12))
- **viewport3d example** — g3d + ui integration: rotating PBR cube rendered by g3d inside a Viewport3D ui widget with Material 3 theme and UI controls ([#10](https://github.com/gogpu/g3d/pull/10))

### Fixed
- **hello-cube smooth animation** — added `WithContinuousRender(true)` for 60fps game-loop rendering instead of event-driven IDLE mode ([#11](https://github.com/gogpu/g3d/pull/11))

### Changed
- **deps:** wgpu v0.30.23 → v0.30.27, gpucontext v0.21.1 → v0.23.0

## [0.1.2] - 2026-07-26

### Fixed
- **Metal shader compilation** — naga v0.17.16 fixes `writeSwizzle` parenthesization for binary expressions (`(mat4 * vec4).xyz`). Fixes [#7](https://github.com/gogpu/g3d/issues/7).

### Changed
- **deps:** wgpu v0.30.2 → v0.30.23, naga v0.17.15 → v0.17.16, gpucontext v0.21.0 → v0.21.1, gputypes v0.5.0 → v0.5.1, goffi v0.5.5 → v0.6.2, webgpu v0.5.2 → v0.5.4

## [0.1.1] - 2026-06-14

### Fixed
- **gpucontext v0.21+ compatibility** — `Device`/`Adapter` changed from interface to opaque struct (ADR-018). Fixed nil checks (`IsNil()`) and type assertions (`wgpu.DeviceFromHandle()`/`wgpu.AdapterFromHandle()`). Fixes [#3](https://github.com/gogpu/g3d/issues/3).

### Changed
- **deps:** wgpu v0.29.15 → v0.30.2, gpucontext v0.19.0 → v0.21.0, goffi v0.5.3 → v0.5.5

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

[Unreleased]: https://github.com/gogpu/g3d/compare/v0.1.10...HEAD
[0.1.10]: https://github.com/gogpu/g3d/compare/v0.1.9...v0.1.10
[0.1.9]: https://github.com/gogpu/g3d/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/gogpu/g3d/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/gogpu/g3d/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/gogpu/g3d/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/gogpu/g3d/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/gogpu/g3d/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/gogpu/g3d/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/gogpu/g3d/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/gogpu/g3d/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/gogpu/g3d/releases/tag/v0.1.0
