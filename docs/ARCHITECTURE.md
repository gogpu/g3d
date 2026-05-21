# g3d Architecture

This document describes the architecture of the g3d 3D rendering library.

## Overview

g3d is a Pure Go 3D rendering library — not a game engine. It provides scene graph, PBR materials, cameras, lights, and a forward rendering pipeline. Zero CGO.

**Core principle: GPU-only via wgpu. One code path, all backends.**

```
                    ┌────────────────────────┐
                    │   User Application     │
                    │  (game, CAD, data viz) │
                    └───────────┬────────────┘
                                │
                         ┌──────▼──────┐
                         │     g3d     │
                         │ 3D Renderer │
                         └──────┬──────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                  │
       ┌──────▼──────┐  ┌──────▼──────┐   ┌──────▼──────┐
       │ Scene Graph │  │  Materials  │   │  Renderer   │
       │ Node, Mesh  │  │ Basic, PBR  │   │  Forward    │
       └─────────────┘  └─────────────┘   └──────┬──────┘
                                                  │
                                           ┌──────▼──────┐
                                           │    wgpu     │
                                           │  Pure Go    │
                                           └──────┬──────┘
                                                  │
                                ┌──────┬──────┬───┴───┬──────┐
                                │      │      │       │      │
                             ┌──▼──┐┌──▼──┐┌──▼──┐┌───▼──┐┌──▼──┐
                             │ Vk  ││DX12 ││Metal││ GLES ││Soft │
                             └─────┘└─────┘└─────┘└──────┘└─────┘
                                          wgpu/hal
```

## Package Structure

```
g3d/                              Root package — public API (~38 exported types)
│
├── Math types                    Vec2, Vec3, Vec4, Mat4, Quat, Euler, Color
├── Scene graph                   Node, Scene, Group, Mesh
├── Interfaces                    Geometry, Material, Camera, Light
├── Concrete types                PerspectiveCamera, OrthographicCamera,
│                                 BasicMaterial, StandardMaterial,
│                                 AmbientLight, DirectionalLight,
│                                 BoxGeometry, SphereGeometry, PlaneGeometry
├── Renderer                      Forward rendering facade
├── Options                       Functional options (WithColor, WithMetallic, etc.)
│
├── internal/geom/                Geometry generation (pure Go, no wgpu)
│   ├── box.go                    Box vertex/index generation
│   ├── sphere.go                 UV sphere generation
│   ├── plane.go                  Plane generation
│   └── layout.go                 Standard vertex stride (32 bytes)
│
├── internal/gpu/                 GPU pipeline management (imports wgpu)
│   ├── device.go                 GPUState, DeviceProvider integration
│   ├── pipeline.go               PipelineCache (lazy, keyed by format+shader)
│   ├── shader.go                 ShaderCache (embedded WGSL via //go:embed)
│   ├── uniform.go                FrameUniforms (224B) + ObjectUniforms (128B)
│   ├── helpers.go                Pipeline state helpers
│   └── shaders/
│       ├── basic.wgsl            Unlit vertex + fragment
│       └── standard.wgsl         Blinn-Phong lit vertex + fragment
│
├── internal/render/              Render list sorting (pure Go, no wgpu)
│   ├── list.go                   RenderList, DrawCall, 4-bucket sorting
│   └── bucket.go                 RenderBucket enum, sort strategies
│
└── examples/
    └── hello-cube/               Rotating lit cube (~20 lines user code)
```

### Dependency Flow (no circular dependencies)

```
gputypes ← gpucontext ← wgpu        (external ecosystem packages)
    ↑          ↑          ↑
    └──────────┼──────────┘
               │
         g3d (root)                   Public API — imports wgpu, gpucontext
          / | \
         ↓  ↓  ↓
  internal/ internal/ internal/       Private implementation
    geom/    gpu/      render/
     │        │          │
     │        ↓          │
     │    wgpu + types   │
     │                   │
    (pure Go)        (pure Go)        No wgpu dependency
```

**Rule:** internal packages NEVER import g3d root (prevents cycles).

## Rendering Pipeline

### Frame Flow

```
Renderer.Render(scene, camera, targetView)
    │
    ├── 1. scene.UpdateWorldTransforms()     Propagate dirty matrices top-down
    │
    ├── 2. Frustum extraction                 camera.ViewProjectionMatrix() → 6 planes
    │
    ├── 3. Scene traversal + cull             Depth-first, skip hidden nodes
    │       For each Mesh:
    │         if frustum.IntersectsAABB(mesh.WorldBoundingBox()):
    │           renderList.Add(drawCall)
    │
    ├── 4. renderList.Sort()                  4-bucket, 3-key opaque sort
    │       Opaque:      PipelineKey → MaterialID → Distance (front-to-back)
    │       Transparent: Distance (back-to-front)
    │
    ├── 5. Upload frame uniforms              ViewProjection + lights → GPU buffer
    │
    ├── 6. Record render pass                 For each DrawCall:
    │       SetPipeline (cached)                pipeline state changes
    │       SetBindGroup(0, frame)              per-frame uniforms
    │       SetBindGroup(1, object+material)    per-object uniforms
    │       SetVertexBuffer + SetIndexBuffer
    │       DrawIndexed
    │
    └── 7. queue.Submit()                     Send to GPU
```

### 4-Bucket Render Order (Three.js pattern)

| Bucket | Sort | Purpose |
|--------|------|---------|
| Background | none | Skybox, IBL (Phase 4) |
| Opaque | front-to-back | Early Z rejection, minimize overdraw |
| Transmissive | front-to-back | Transparent with depth-write (Phase 4) |
| Transparent | back-to-front | Correct alpha blending (Phase 2) |

### 3-Key Opaque Sort (validated against Three.js PR #15484)

1. **PipelineKey** (ascending) — minimize pipeline state changes (most expensive)
2. **MaterialID** (ascending) — minimize bind group changes
3. **Distance** (ascending) — front-to-back for early Z rejection

## GPU Integration

### DeviceProvider Pattern (same as gg)

```go
// With gogpu app framework:
renderer, err := g3d.NewRenderer(app.DeviceProvider())

// Standalone (bring your own device):
renderer, err := g3d.NewRendererFromDevice(device, queue, format)
```

g3d accepts `gpucontext.DeviceProvider`, type-asserts to `*wgpu.Device` internally. Pattern follows `gg/internal/gpu/gpu_shared.go:124-189` exactly.

### Bind Group Layout

| Group | Binding | Content | Size |
|-------|---------|---------|------|
| 0 | 0 | FrameUniforms (ViewProjection, camera pos, lights) | 224 bytes |
| 1 | 0 | ObjectUniforms (model matrix, normal matrix) | 128 bytes |
| 1 | 1 | MaterialUniforms (color, metallic, roughness, etc.) | 16-32 bytes |

### Shader Strategy

WGSL shaders embedded via `//go:embed`. naga compiles to target backend at runtime:
- Vulkan → SPIR-V
- Metal → MSL
- DX12 → HLSL (or direct DXIL)
- GLES → GLSL
- Software → SPIR-V (interpreted)

## Scene Graph

Three.js-inspired Node hierarchy with dirty flag propagation:

```go
type Node struct {
    Position Vec3    // public, user-facing
    Rotation Euler   // radians
    Scale    Vec3    // default {1,1,1}
    // ...
    localMatrix  Mat4   // cached, recomputed when dirty
    worldMatrix  Mat4   // parent.WorldMatrix * localMatrix
}
```

- **Dirty flags:** setting Position/Rotation/Scale marks local+world dirty
- **World dirty propagates DOWN** to all children recursively
- **WorldMatrix** = parent.WorldMatrix() × node.LocalMatrix()
- **LocalMatrix** = Translate(Position) × FromQuat(QuatFromEuler(Rotation)) × Scale(Scale)

## Key Design Decisions

| Decision | Rationale | Reference |
|----------|-----------|-----------|
| GPU-only (not dual CPU+GPU like gg) | 3D CPU = same algorithm as GPU. wgpu software backend exists | ADR-001, 0/9 enterprise refs use dual |
| Scene graph (not ECS) | Rendering library, not engine. Three.js pattern | ADR-001, Discussion #168 |
| Forward rendering | Works on ALL backends including GLES + Software | ADR-001, Kaiju pattern |
| Idiomatic Go API | Interfaces + functional options, not OOP | Discussion #168, @darkliquid |
| 4-bucket sorting | Three.js validated pattern | PHASE1-ARCHITECTURE-VALIDATION.md |
| Column-major Mat4 | Matches WGSL mat4x4<f32> memory layout | ADR-001 |
| WebGPU Z [0,1] | NOT OpenGL Z [-1,1]. c = -far/(far-near) | Three.js Matrix4, ADR-001 |
| Root package only (Phase 1) | wgpu=23K root, gg=38K root. Sub-packages Phase 2+ | Package structure research |

## Future Architecture (Phase 2+)

| Phase | Feature | Architecture Impact |
|-------|---------|-------------------|
| Phase 2 | PBR + Shadows | Cook-Torrance BRDF, shadow map render pass |
| Phase 3 | GLTF | `g3d/gltf` public sub-package |
| Phase 3 | Animation | `g3d/animation` public sub-package |
| Phase 4 | Post-processing | `g3d/postfx` public sub-package, render graph |
| Phase 5 | Instancing | 4-level draw grouping (Kaiju pattern, already in sort) |
