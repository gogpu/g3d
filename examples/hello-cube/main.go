// Command hello-cube renders a lit, rotating cube using g3d + gogpu.
//
// This is the "hello world" of g3d -- a minimal example that demonstrates:
//   - Scene graph setup (scene, mesh, lights, camera)
//   - PBR material with functional options
//   - Forward renderer with depth testing
//   - dt-based animation (smooth on any refresh rate)
//   - Window resize handling (aspect ratio + depth buffer)
//   - Proper GPU resource lifecycle (defer Release)
//
// Run with different backends:
//
//	GOGPU_GRAPHICS_API=vulkan   go run .
//	GOGPU_GRAPHICS_API=dx12     go run .
//	GOGPU_GRAPHICS_API=gles     go run .
//	GOGPU_GRAPHICS_API=software go run .
package main

import (
	"log"

	"github.com/gogpu/g3d"
	"github.com/gogpu/gogpu"
)

func main() {
	// Configure the application window with continuous rendering (game loop).
	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("g3d Hello Cube").
		WithSize(800, 600).
		WithContinuousRender(true))

	// --- Scene setup (runs once before the first frame) ---

	// Scene: root of the object hierarchy, also sets the background clear color.
	scene := g3d.NewScene()
	scene.SetBackground(g3d.RGB(0.1, 0.1, 0.15))

	// Lights: ambient provides a base illumination so no face is pure black;
	// directional simulates sunlight from the upper-right.
	//
	// AmbientLight has no spatial properties, so we attach it as UserData
	// on a container node. The renderer discovers it during scene traversal.
	ambient := g3d.NewAmbientLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(0.3),
	)
	ambientNode := g3d.NewNode()
	ambientNode.SetName("AmbientLight")
	ambientNode.SetUserData(ambient)
	scene.Add(ambientNode)

	sun := g3d.NewDirectionalLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(1.0),
	)
	// Tilt the directional light 45 degrees down and 30 degrees to the right.
	// This ensures visible shading differences across cube faces.
	sun.LightNode().SetRotation(g3d.Euler{
		X: g3d.Radians(-45),
		Y: g3d.Radians(30),
	})
	scene.Add(sun.LightNode())

	// Mesh: a unit cube with PBR material (metallic-roughness model).
	cube := g3d.NewMesh(
		g3d.NewBoxGeometry(1, 1, 1),
		g3d.NewStandardMaterial(
			g3d.WithColor(g3d.RGB(0.4, 0.7, 1.0)),
			g3d.WithMetallic(0.3),
			g3d.WithRoughness(0.6),
		),
	)
	scene.Add(cube.MeshNode())

	// Camera: 75-degree vertical FOV, positioned 3 units back on the Z axis.
	camera := g3d.NewPerspectiveCamera(75, 800.0/600.0, 0.1, 1000)
	camera.CameraNode().SetPosition(g3d.Vec3{X: 0, Y: 0.5, Z: 3})

	// --- Renderer initialization (deferred until GPU is ready) ---

	// The g3d renderer is created inside OnDraw because GPUContextProvider()
	// is only valid after gogpu.App.Run() has initialized the GPU device.
	var renderer *g3d.Renderer

	// OnUpdate: called every frame with delta time in seconds.
	// Rotate the cube at a constant speed regardless of frame rate.
	// We use SetRotation (not direct field mutation) to trigger the dirty
	// flag that propagates through the scene graph transform cache.
	app.OnUpdate(func(dt float64) {
		r := cube.MeshNode().Rotation
		r.Y += float32(dt)
		r.X += float32(dt) * 0.3
		cube.MeshNode().SetRotation(r)
	})

	// OnResize: update camera aspect ratio when window changes size.
	// The renderer's SetSize is called in OnDraw with physical pixel dimensions.
	app.OnResize(func(width, height int) {
		if height > 0 {
			camera.SetAspect(float32(width) / float32(height))
		}
	})

	// OnDraw: acquire the surface texture and render the scene.
	app.OnDraw(func(ctx *gogpu.Context) {
		// Lazy initialization: create the renderer on the first frame when
		// the GPU device is available.
		if renderer == nil {
			provider := app.GPUContextProvider()
			if provider == nil {
				return // GPU not ready yet
			}
			var err error
			renderer, err = g3d.NewRenderer(provider)
			if err != nil {
				log.Fatalf("g3d: create renderer: %v", err)
			}
		}

		// Update the renderer's internal depth texture to match the current
		// framebuffer dimensions (no-op if size unchanged).
		fbW, fbH := ctx.FramebufferSize()
		renderer.SetSize(uint32(fbW), uint32(fbH))

		// Acquire the surface texture view for this frame and render.
		view := ctx.SurfaceView()
		if view == nil {
			return // Surface not available (minimized window)
		}

		if err := renderer.Render(scene, camera, view); err != nil {
			log.Printf("g3d: render error: %v", err)
		}
	})

	// Clean up GPU resources when the application exits.
	app.OnClose(func() {
		if renderer != nil {
			renderer.Release()
		}
	})

	// Run blocks until the window is closed.
	if err := app.Run(); err != nil {
		log.Fatalf("gogpu: %v", err)
	}
}
