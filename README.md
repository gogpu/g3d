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
  Scene graph, PBR materials, forward rendering. Zero CGO.<br>
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

g3d is a **3D rendering library** — not a game engine. It provides the building blocks (scene graph, cameras, lights, materials, geometry primitives) that game engines, CAD viewers, data visualizers, and AR/VR applications build upon.

Think of it like [Three.js](https://threejs.org/) for Go: simple API, powerful rendering, zero opinion about your application architecture.

```go
package main

import (
    "log"
    "math"

    "github.com/gogpu/g3d"
    "github.com/gogpu/gogpu"
)

func main() {
    app := gogpu.NewApp(gogpu.WithTitle("g3d Hello Cube"), gogpu.WithSize(800, 600))

    scene := g3d.NewScene()

    // Lights
    ambientNode := g3d.NewNode()
    ambientNode.SetUserData(g3d.NewAmbientLight(g3d.White, 0.3))
    scene.Add(ambientNode)

    dirLight := g3d.NewDirectionalLight(g3d.White, 1.0)
    scene.Add(dirLight.LightNode())

    // Cube mesh with PBR material
    cube := g3d.NewMesh(
        g3d.NewBoxGeometry(1, 1, 1),
        g3d.NewStandardMaterial(g3d.WithColor(g3d.Color{0.4, 0.7, 1.0, 1.0})),
    )
    scene.Add(cube.MeshNode())

    // Camera
    camera := g3d.NewPerspectiveCamera(75, 800.0/600.0, 0.1, 1000)
    camera.CameraNode().SetPosition(g3d.Vec3{0, 1, 3})

    var renderer *g3d.Renderer
    var angle float32

    app.OnDraw(func(ctx *gogpu.Context) {
        if renderer == nil {
            var err error
            renderer, err = g3d.NewRenderer(ctx.GPUContextProvider())
            if err != nil {
                log.Fatal(err)
            }
            renderer.SetSize(uint32(ctx.Width()), uint32(ctx.Height()))
        }
        angle += float32(ctx.DeltaTime())
        cube.MeshNode().SetRotation(g3d.Euler{0, angle, 0})
        _ = renderer.Render(scene, camera, ctx.SurfaceView())
    })
    app.Run()
}
```

## Features

### Core (v0.1.0)
- **Scene graph** — hierarchical Node tree with parent-child transform propagation and dirty flags
- **Cameras** — Perspective and Orthographic with frustum extraction
- **Geometries** — Box, Sphere, Plane + custom BufferGeometry
- **Forward renderer** — 4-bucket sorting (background, opaque, transmissive, transparent)
- **Frustum culling** — automatic AABB visibility testing against camera frustum

### Materials (v0.1.0)
- **BasicMaterial** — unlit, for prototyping and data visualization
- **StandardMaterial** — PBR metallic-roughness with Blinn-Phong shading

### Lighting (v0.1.0)
- **AmbientLight** — uniform environment lighting
- **DirectionalLight** — sun-like parallel light

### Performance (v0.1.0)
- **Zero-alloc render path** — no GC pressure during frame rendering
- **Pipeline cache** — compile shader variants once, reuse forever
- **3-key opaque sort** — PipelineKey → MaterialID → Distance (minimizes GPU state changes)
- **MappedAtCreation** — zero-copy GPU buffer upload (WebGPU compliant)

### Planned
- **Full PBR** — Cook-Torrance BRDF, shadow mapping, normal maps (Phase 2)
- **GLTF 2.0** — binary (.glb) and JSON (.gltf) with PBR materials, animations (Phase 3)
- **Instance batching** — thousands of objects with minimal draw calls (Phase 4)
- **Post-processing** — bloom, tone mapping, FXAA (Phase 4)

## Not a Game Engine

g3d deliberately does **not** include:

| Feature | Why Not | Where to Get It |
|---------|---------|----------------|
| Entity Component System | Game engine concern | Build on top, or use external ECS |
| Physics | Simulation concern | Integrate Bullet, ODE, or Pure Go physics |
| Audio | Unrelated to rendering | Use [gogpu/audio](https://github.com/gogpu/audio) or Oto |
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
| **Software** | All | Fallback (CI/testing) |

All backends are Pure Go — zero CGO, single binary deployment.

```bash
# Select backend via environment variable
GOGPU_GRAPHICS_API=vulkan   go run ./examples/hello-cube/
GOGPU_GRAPHICS_API=dx12     go run ./examples/hello-cube/
GOGPU_GRAPHICS_API=software go run ./examples/hello-cube/
```

## Standalone Usage

g3d works without the gogpu application framework. Bring your own window and GPU device:

```go
// Use g3d with any wgpu.Device — no gogpu dependency required
renderer, err := g3d.NewRendererFromDevice(device, queue, surfaceFormat)

scene := g3d.NewScene()
// ... build your scene
renderer.Render(scene, camera, targetView)
```

## Architecture

```
Your Application (game engine, CAD viewer, data viz, AR/VR)
         |
    gogpu/g3d  — Scene Graph + Materials + Render Pipeline
         |
    gogpu/wgpu — Pure Go WebGPU (Vulkan/Metal/DX12/GLES/Software)
         |
    gogpu/naga — Shader Compiler (WGSL → SPIR-V/MSL/GLSL/HLSL)
```

g3d depends **down** (wgpu, naga), never **up** (gogpu, gg, ui). This ensures it can be used in any context.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for full architecture documentation.

## Installation

```bash
go get github.com/gogpu/g3d
```

**Requirements:** Go 1.25+

## Roadmap

| Phase | Features | Status |
|-------|----------|--------|
| **Phase 1** | Scene graph, cameras, materials, box/sphere/plane, forward renderer | **Complete** |
| **Phase 2** | Full PBR (Cook-Torrance), shadows, normal maps, textures | Planned |
| **Phase 3** | GLTF 2.0 loader, skeletal animation, morph targets | Planned |
| **Phase 4** | Instance batching, environment maps, post-processing, skybox | Planned |
| **Phase 5** | Frustum culling BVH, LOD, SIMD math | Planned |

## Design Principles

1. **Simple API** — rotating lit cube in ~20 lines. Progressive complexity.
2. **Zero CGO** — Pure Go on all platforms. Single binary deployment.
3. **Reusable** — rendering library, not a framework. No opinions about your architecture.
4. **PBR from day one** — metallic-roughness workflow, GLTF standard.
5. **Zero-alloc rendering** — no GC pressure in the hot path.
6. **All GPU backends** — Vulkan, Metal, DX12, GLES, Software through wgpu.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development workflow, code standards, and priority areas.

## Part of the GoGPU Ecosystem

g3d is part of [GoGPU](https://github.com/gogpu) — a Pure Go GPU computing ecosystem with 800K+ lines of code.

| Library | Purpose |
|:--------|:--------|
| [gogpu](https://github.com/gogpu/gogpu) | Application framework, windowing |
| [wgpu](https://github.com/gogpu/wgpu) | Pure Go WebGPU (Vulkan/Metal/DX12/GLES) |
| [naga](https://github.com/gogpu/naga) | Shader compiler (WGSL → SPIR-V/MSL/GLSL/HLSL) |
| [gg](https://github.com/gogpu/gg) | 2D graphics with GPU acceleration |
| **[g3d](https://github.com/gogpu/g3d)** | **3D rendering (this library)** |
| [ui](https://github.com/gogpu/ui) | GUI toolkit (22+ widgets, 4 themes) |
| [systray](https://github.com/gogpu/systray) | System tray (Win32/macOS/Linux) |
| [audio](https://github.com/gogpu/audio) | Pure Go audio engine (WASAPI) |

## License

MIT License — see [LICENSE](LICENSE) for details.
