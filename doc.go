// Package g3d is a Pure Go 3D rendering library built on gogpu/wgpu.
//
// g3d provides a scene graph, PBR materials, cameras, lights, GLTF loading,
// and a forward rendering pipeline with zero CGO dependencies. It is designed
// as a reusable foundation — not a game engine — so that game engines, CAD viewers,
// data visualizers, and AR/VR applications can build on top of it.
//
// Quick start:
//
//	scene := g3d.NewScene()
//	scene.Add(g3d.NewAmbientLight(g3d.White, 0.3))
//	scene.Add(g3d.NewDirectionalLight(g3d.White, 1.0))
//
//	cube := g3d.NewMesh(g3d.NewBoxGeometry(1, 1, 1), g3d.NewStandardMaterial())
//	scene.Add(cube)
//
//	camera := g3d.NewPerspectiveCamera(75, aspect, 0.1, 1000)
//	camera.Position = g3d.Vec3{0, 0, 3}
//
//	renderer.Render(scene, camera)
//
// g3d depends on gogpu/wgpu for GPU abstraction (Vulkan, Metal, DX12, GLES, Software)
// and gogpu/naga for shader compilation (WGSL to SPIR-V/MSL/GLSL/HLSL).
//
// Part of the GoGPU ecosystem: https://github.com/gogpu
package g3d
