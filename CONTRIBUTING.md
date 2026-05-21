# Contributing to g3d

Thank you for your interest in contributing to g3d — the Pure Go 3D rendering library!

---

## Requirements

- **Go 1.25+**
- **golangci-lint** for code quality checks
- **GPU** (Vulkan, DX12, Metal, or GLES) for visual testing; software backend for CI

---

## Quick Start

```bash
# Clone the repository
git clone https://github.com/gogpu/g3d
cd g3d

# Build
go build ./...

# Run tests
go test ./...

# Run linter
golangci-lint run --timeout=5m

# Run hello cube example (from ecosystem workspace)
cd .. && go run ./g3d/examples/hello-cube/
```

---

## Development Workflow

### 1. Fork & Clone

```bash
git clone https://github.com/YOUR_USERNAME/g3d
cd g3d
```

### 2. Create Feature Branch

```bash
git checkout -b feat/your-feature
```

### 3. Make Changes

- Follow existing code patterns (read similar files first)
- Run `go fmt ./...` before committing
- Run `golangci-lint run --timeout=5m` — must be zero issues
- Add tests for new functionality

### 4. Test

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Coverage
go test -coverprofile=tmp/coverage.out ./...
```

### 5. Submit Pull Request

- One PR per feature
- Clear description of what and why
- Reference related issues

---

## Code Standards

### Go Conventions

- **Go 1.25+** features welcome
- **gofmt** — all code must be formatted
- **golangci-lint** — zero issues required (see `.golangci.yml`)
- **Naming**: `ID`, `URL`, `HTTP` uppercase; camelCase for unexported
- **Errors**: always handle or explicitly ignore with `_`

### g3d-Specific Rules

- **float32** — all math types use float32 (GPU standard), not float64
- **Column-major Mat4** — `[16]float32` where `m[col*4+row]`, matches WGSL `mat4x4<f32>`
- **WebGPU Z [0,1]** — projection matrices use Z range [0,1], NOT OpenGL [-1,1]
- **No type aliases** — `type X = Y` is forbidden in public API
- **No circular deps** — `internal/` packages never import root `g3d` package
- **MappedAtCreation** — use for GPU buffer upload, never legacy `Queue.ReadBuffer`
- **Uniform alignment** — WGSL `vec3<f32>` is 16-byte aligned; pad with `uint32` or use `vec4`

### Architecture

- **Root package** = public API only (types, interfaces, constructors, renderer facade)
- **`internal/geom/`** = geometry generation (pure Go, no wgpu dependency)
- **`internal/gpu/`** = wgpu pipeline management (DeviceProvider, pipeline cache, shaders)
- **`internal/render/`** = render list sorting (pure Go, no wgpu dependency)

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for full architecture documentation.

---

## Testing

### Unit Tests (no GPU required)

Math types, scene graph, geometry generation, render list sorting, material uniform packing — all testable without GPU.

```bash
go test ./...                    # all packages
go test ./internal/geom/...      # geometry only
go test ./internal/render/...    # sorting only
```

### Visual Tests (GPU required)

```bash
# Default backend (Vulkan on Linux/Windows, Metal on macOS)
go run ./examples/hello-cube/

# Specific backend
GOGPU_GRAPHICS_API=vulkan   go run ./examples/hello-cube/
GOGPU_GRAPHICS_API=dx12     go run ./examples/hello-cube/
GOGPU_GRAPHICS_API=gles     go run ./examples/hello-cube/
GOGPU_GRAPHICS_API=software go run ./examples/hello-cube/
```

### Cross-Platform Lint

```bash
golangci-lint run --timeout=5m
GOOS=linux GOARCH=amd64 golangci-lint run --timeout=5m
GOOS=darwin GOARCH=arm64 golangci-lint run --timeout=5m
```

---

## Priority Areas

We welcome contributions in these areas:

1. **GLTF loading** — parser and material mapping (Phase 3)
2. **Shader development** — PBR Cook-Torrance, shadows, post-processing in WGSL
3. **Geometry primitives** — Cylinder, Cone, Torus
4. **Examples** — showcase real-world usage
5. **Testing** — cross-platform GPU rendering tests
6. **Documentation** — tutorials, API guides

---

## Part of GoGPU Ecosystem

g3d is part of [GoGPU](https://github.com/gogpu) — see the [ecosystem CONTRIBUTING guide](https://github.com/gogpu/gogpu/blob/main/CONTRIBUTING.md) for ecosystem-wide conventions.

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
