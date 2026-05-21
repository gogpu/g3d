# Security Policy

## Supported Versions

g3d is currently in early development (v0.x.x).

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1.0 | :x:                |

## Reporting a Vulnerability

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues via:

1. **Private Security Advisory** (preferred):
   https://github.com/gogpu/g3d/security/advisories/new

2. **GitHub Discussions**:
   https://github.com/gogpu/gogpu/discussions

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact

### Response Timeline

- **Initial Response**: Within 72 hours
- **Fix & Disclosure**: Coordinated with reporter

## Security Considerations

g3d uses the gogpu/wgpu Pure Go WebGPU implementation. Users should be aware of:

1. **GPU Memory** — ensure proper resource cleanup (`Renderer.Release()`) to avoid GPU memory leaks
2. **Shader Code** — WGSL shaders are compiled by naga and executed on GPU hardware
3. **Buffer Mapping** — mapped GPU buffers expose raw memory; unmap after use
4. **No CGO** — g3d is Pure Go with zero C dependencies, reducing native attack surface

## Security Contact

- **GitHub Security Advisory**: https://github.com/gogpu/g3d/security/advisories/new
- **Public Issues**: https://github.com/gogpu/g3d/issues

---

**Thank you for helping keep g3d secure!**
