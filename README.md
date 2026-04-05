<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png">
    <img src="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png" alt="GoGPU Logo" width="100" />
  </picture>
</p>

<h1 align="center">g3d</h1>

<p align="center">
  <strong>Pure Go 3D rendering library</strong><br>
  Scene graph, PBR materials, GLTF loading. Zero CGO.<br>
  Built on <a href="https://github.com/gogpu/wgpu">gogpu/wgpu</a> (Vulkan/Metal/DX12/GLES/Software).
</p>

<p align="center">
  <a href="https://github.com/gogpu/g3d/actions"><img src="https://github.com/gogpu/g3d/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/gogpu/g3d"><img src="https://pkg.go.dev/badge/github.com/gogpu/g3d.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/gogpu/g3d"><img src="https://goreportcard.com/badge/github.com/gogpu/g3d" alt="Go Report Card"></a>
  <a href="https://github.com/gogpu/g3d/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/gogpu/g3d"><img src="https://img.shields.io/badge/Pure_Go-Zero_CGO-brightgreen" alt="Zero CGO"></a>
</p>

---

## What is g3d?

g3d is a **3D rendering library** — not a game engine. It provides the building blocks (scene graph, cameras, lights, materials, mesh loading) that game engines, CAD viewers, data visualizers, and AR/VR applications build upon.

Think of it like [Three.js](https://threejs.org/) for Go: simple API, powerful rendering, zero opinion about your application architecture.

```go
package main

import (
    "github.com/gogpu/gogpu"
    "github.com/gogpu/g3d"
)

func main() {
    app := gogpu.NewApp(gogpu.Config{Title: "g3d Hello", Width: 800, Height: 600})
    renderer := g3d.NewRenderer(app)

    scene := g3d.NewScene()
    scene.Add(g3d.NewAmbientLight(g3d.White, 0.3))
    scene.Add(g3d.NewDirectionalLight(g3d.White, 1.0))

    cube := g3d.NewMesh(g3d.NewBoxGeometry(1, 1, 1), g3d.NewStandardMaterial())
    scene.Add(cube)

    camera := g3d.NewPerspectiveCamera(75, 800.0/600.0, 0.1, 1000)
    camera.Position = g3d.Vec3{0, 0, 3}

    app.OnUpdate(func(dt float64) {
        cube.Rotation.Y += float32(dt)
    })
    app.OnDraw(func(_ *gogpu.Context) {
        renderer.Render(scene, camera)
    })
    app.Run()
}
```

## Features

### Core
- **Scene graph** — hierarchical Node tree with parent-child transform propagation
- **Cameras** — Perspective and Orthographic with configurable clip planes
- **Geometries** — Box, Sphere, Plane, Cylinder, Cone, Torus + custom BufferGeometry
- **Renderer** — Forward rendering with 3-bucket sorting (opaque, alpha mask, transparent)
- **Frustum culling** — automatic visibility testing against camera frustum

### Materials
- **BasicMaterial** — unlit, for prototyping and data visualization
- **StandardMaterial** — PBR metallic-roughness (GLTF standard), production quality
- **ShaderMaterial** — custom WGSL shaders for advanced users

### Lighting
- **DirectionalLight** — sun-like parallel light with shadow mapping
- **PointLight** — omnidirectional light with distance attenuation
- **SpotLight** — focused cone light with soft edges
- **AmbientLight** — uniform environment lighting

### Loading
- **GLTF 2.0** — binary (.glb) and JSON (.gltf) with PBR materials, animations, scene hierarchy
- **Textures** — PNG, JPEG, HDR for environment maps

### Performance
- **Zero-alloc render path** — no GC pressure during frame rendering
- **Instance batching** — thousands of objects with minimal draw calls
- **Pipeline specialization cache** — compile shader variants once, reuse forever
- **4-level draw grouping** — RenderPass → Pipeline → Material → Mesh → Instances

## Not a Game Engine

g3d deliberately does **not** include:

| Feature | Why Not | Where to Get It |
|---------|---------|----------------|
| Entity Component System | Game engine concern | Build on top, or use external ECS |
| Physics | Simulation concern | Integrate Bullet, ODE, or Pure Go physics |
| Audio | Unrelated to rendering | Use Oto, Beep, or platform audio |
| Networking | Unrelated to rendering | Use net/http, gRPC, WebSocket |
| Scripting | Engine concern | Use Lua/Wasm/Yaegi on top |
| Scene editor | Tool concern | Build with gogpu/ui + g3d |

This separation means g3d is **reusable everywhere** — game engines, CAD tools, scientific visualizations, AR/VR, data dashboards.

## GPU Backends

g3d renders through [gogpu/wgpu](https://github.com/gogpu/wgpu), which supports:

| Backend | Platforms | Status |
|---------|-----------|--------|
| **Vulkan** | Windows, Linux | Stable |
| **Metal** | macOS | Stable |
| **DirectX 12** | Windows | Stable |
| **OpenGL ES** | Windows, Linux | Stable |
| **Software** | All | Fallback |

All backends are Pure Go — zero CGO, single binary deployment.

## Standalone Usage

g3d works without the gogpu application framework. Bring your own window and GPU device:

```go
// Use g3d with any wgpu.Device — no gogpu dependency required
device, queue := myCustomGPUSetup()
renderer := g3d.NewRendererFromDevice(device, queue)

scene := g3d.NewScene()
// ... build your scene
renderer.Render(scene, camera)
```

## Architecture

```
Your Application (game engine, CAD viewer, data viz, AR/VR)
         |
    gogpu/g3d  — Scene Graph + PBR + GLTF + Render Pipeline
         |
    gogpu/wgpu — Pure Go WebGPU (Vulkan/Metal/DX12/GLES)
         |
    gogpu/naga — Shader Compiler (WGSL → SPIR-V/MSL/GLSL/HLSL)
```

g3d depends **down** (wgpu, naga), never **up** (gogpu, gg, ui). This ensures it can be used in any context.

## Installation

```bash
go get github.com/gogpu/g3d
```

**Requirements:** Go 1.25+

## Roadmap

| Phase | Features | Status |
|-------|----------|--------|
| **Phase 1** | Scene graph, cameras, basic materials, box/sphere/plane, renderer | Planned |
| **Phase 2** | PBR lighting, shadows, StandardMaterial, normal maps | Planned |
| **Phase 3** | GLTF 2.0 loader, skeletal animation, morph targets | Planned |
| **Phase 4** | Instance batching, environment maps, post-processing, skybox | Planned |
| **Phase 5** | Frustum culling BVH, LOD, SIMD math, pipeline cache | Planned |

## Design Principles

1. **Simple API** — 15 lines for a lit rotating cube. Progressive complexity.
2. **Zero CGO** — Pure Go on all platforms. Single binary deployment.
3. **Reusable** — rendering library, not a framework. No opinions about your architecture.
4. **PBR from day one** — metallic-roughness workflow, GLTF standard.
5. **Zero-alloc rendering** — no GC pressure in the hot path.
6. **All GPU backends** — Vulkan, Metal, DX12, GLES, Software through wgpu.

## Contributing

We welcome contributions! Priority areas:

1. **GLTF loading** — parser and material mapping
2. **Shader development** — PBR, shadows, post-processing in WGSL
3. **Geometry primitives** — additional built-in shapes
4. **Examples** — showcase real-world usage
5. **Testing** — cross-platform GPU rendering tests

## Part of the GoGPU Ecosystem

g3d is part of [GoGPU](https://github.com/gogpu) — a Pure Go GPU computing ecosystem with 632K+ lines of code.

| Library | Purpose |
|:--------|:--------|
| [gogpu](https://github.com/gogpu/gogpu) | Application framework, windowing |
| [wgpu](https://github.com/gogpu/wgpu) | Pure Go WebGPU (Vulkan/Metal/DX12/GLES) |
| [naga](https://github.com/gogpu/naga) | Shader compiler (WGSL → SPIR-V/MSL/GLSL/HLSL) |
| [gg](https://github.com/gogpu/gg) | 2D graphics with GPU acceleration |
| **[g3d](https://github.com/gogpu/g3d)** | **3D rendering (this library)** |
| [ui](https://github.com/gogpu/ui) | GUI toolkit (22+ widgets, 4 themes) |
| [systray](https://github.com/gogpu/systray) | System tray (Win32/macOS/Linux) |

## License

MIT License — see [LICENSE](LICENSE) for details.
