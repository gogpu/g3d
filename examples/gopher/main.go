// Command gopher renders the Go Gopher mascot assembled from basic
// primitives (spheres, boxes) — demonstrating hierarchical scene graph,
// multiple materials, and grouping.
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

// Gopher colors.
var (
	gopherBlue = g3d.RGB(0.0, 0.6, 0.9)   // body
	skinBrown  = g3d.RGB(0.78, 0.6, 0.44) // nose, ears, feet
	darkBrown  = g3d.RGB(0.2, 0.15, 0.1)  // nose tip
	eyeWhite   = g3d.RGB(1, 1, 1)
	pupilBlack = g3d.RGB(0.05, 0.05, 0.05)
	toothWhite = g3d.RGB(0.95, 0.95, 0.90)
)

func main() {
	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("g3d — Go Gopher").
		WithSize(800, 600).
		WithContinuousRender(true))

	scene := g3d.NewScene()
	scene.SetBackground(g3d.RGB(0.15, 0.15, 0.2))

	setupLights(scene)
	gopher := buildGopher()
	scene.Add(gopher)

	camera := g3d.NewPerspectiveCamera(60, 800.0/600.0, 0.1, 100)
	camera.CameraNode().SetPosition(g3d.Vec3{X: 1.5, Y: 1.0, Z: 5})
	camera.CameraNode().LookAt(g3d.Vec3{X: 0, Y: 0.3, Z: 0})

	var renderer *g3d.Renderer

	app.OnUpdate(func(dt float64) {
		r := gopher.Rotation
		r.Y += float32(dt) * 0.5
		gopher.SetRotation(r)
	})

	app.OnResize(func(width, height int) {
		if height > 0 {
			camera.SetAspect(float32(width) / float32(height))
		}
	})

	app.OnDraw(func(ctx *gogpu.Context) {
		if renderer == nil {
			provider := app.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			renderer, err = g3d.NewRenderer(provider)
			if err != nil {
				log.Fatalf("g3d: create renderer: %v", err)
			}
		}

		fbW, fbH := ctx.FramebufferSize()
		renderer.SetSize(uint32(fbW), uint32(fbH))

		view := ctx.SurfaceView()
		if view == nil {
			return
		}
		if err := renderer.Render(scene, camera, view); err != nil {
			log.Printf("g3d: render error: %v", err)
		}
	})

	app.OnClose(func() {
		if renderer != nil {
			renderer.Release()
		}
	})

	if err := app.Run(); err != nil {
		log.Fatalf("gogpu: %v", err)
	}
}

func setupLights(scene *g3d.Scene) {
	ambient := g3d.NewAmbientLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(0.35),
	)
	an := g3d.NewNode()
	an.SetName("AmbientLight")
	an.SetUserData(ambient)
	scene.Add(an)

	sun := g3d.NewDirectionalLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(1.0),
	)
	sun.LightNode().SetRotation(g3d.Euler{
		X: g3d.Radians(-40),
		Y: g3d.Radians(30),
	})
	scene.Add(sun.LightNode())

	fill := g3d.NewDirectionalLight(
		g3d.WithLightColor(g3d.RGB(0.6, 0.7, 0.9)),
		g3d.WithLightIntensity(0.4),
	)
	fill.LightNode().SetRotation(g3d.Euler{
		X: g3d.Radians(-20),
		Y: g3d.Radians(-120),
	})
	scene.Add(fill.LightNode())
}

func buildGopher() *g3d.Node {
	root := g3d.NewGroup()
	root.SetName("Gopher")
	n := &root.Node

	bodyMat := g3d.NewStandardMaterial(
		g3d.WithColor(gopherBlue),
		g3d.WithMetallic(0.05),
		g3d.WithRoughness(0.7),
	)
	skinMat := g3d.NewStandardMaterial(
		g3d.WithColor(skinBrown),
		g3d.WithMetallic(0.0),
		g3d.WithRoughness(0.85),
	)
	whiteMat := g3d.NewStandardMaterial(
		g3d.WithColor(eyeWhite),
		g3d.WithMetallic(0.0),
		g3d.WithRoughness(0.3),
	)
	pupilMat := g3d.NewStandardMaterial(
		g3d.WithColor(pupilBlack),
		g3d.WithMetallic(0.0),
		g3d.WithRoughness(0.2),
	)
	toothMat := g3d.NewStandardMaterial(
		g3d.WithColor(toothWhite),
		g3d.WithMetallic(0.0),
		g3d.WithRoughness(0.4),
	)
	noseTipMat := g3d.NewStandardMaterial(
		g3d.WithColor(darkBrown),
		g3d.WithMetallic(0.0),
		g3d.WithRoughness(0.5),
	)

	hires := g3d.WithSegments(32, 24)
	lores := g3d.WithSegments(24, 16)

	// --- Body: egg-shaped (sphere scaled taller) ---
	body := g3d.NewMesh(g3d.NewSphereGeometry(1.0, hires), bodyMat)
	body.MeshNode().SetScale(g3d.Vec3{X: 1.0, Y: 1.25, Z: 0.9})
	body.MeshNode().SetName("Body")
	root.Add(body.MeshNode())

	// --- Eyes ---
	addEye := func(name string, xSign float32) {
		white := g3d.NewMesh(g3d.NewSphereGeometry(0.32, lores), whiteMat)
		white.MeshNode().SetPosition(g3d.Vec3{X: 0.28 * xSign, Y: 0.65, Z: 0.7})
		white.MeshNode().SetName(name + "White")
		root.Add(white.MeshNode())

		pupil := g3d.NewMesh(g3d.NewSphereGeometry(0.14, lores), pupilMat)
		pupil.MeshNode().SetPosition(g3d.Vec3{X: 0.3 * xSign, Y: 0.65, Z: 0.95})
		pupil.MeshNode().SetName(name + "Pupil")
		root.Add(pupil.MeshNode())
	}
	addEye("LeftEye", -1)
	addEye("RightEye", 1)

	// --- Nose ---
	nose := g3d.NewMesh(g3d.NewSphereGeometry(0.13, lores), skinMat)
	nose.MeshNode().SetPosition(g3d.Vec3{X: 0, Y: 0.32, Z: 0.88})
	nose.MeshNode().SetScale(g3d.Vec3{X: 1.2, Y: 0.9, Z: 1.0})
	nose.MeshNode().SetName("Nose")
	root.Add(nose.MeshNode())

	noseTip := g3d.NewMesh(g3d.NewSphereGeometry(0.06, lores), noseTipMat)
	noseTip.MeshNode().SetPosition(g3d.Vec3{X: 0, Y: 0.35, Z: 0.98})
	noseTip.MeshNode().SetName("NoseTip")
	root.Add(noseTip.MeshNode())

	// --- Teeth ---
	addTooth := func(name string, xOffset float32) {
		tooth := g3d.NewMesh(g3d.NewBoxGeometry(0.08, 0.14, 0.06), toothMat)
		tooth.MeshNode().SetPosition(g3d.Vec3{X: xOffset, Y: 0.15, Z: 0.88})
		tooth.MeshNode().SetName(name)
		root.Add(tooth.MeshNode())
	}
	addTooth("LeftTooth", -0.05)
	addTooth("RightTooth", 0.05)

	// --- Ears ---
	addEar := func(name string, xSign float32) {
		ear := g3d.NewMesh(g3d.NewSphereGeometry(0.12, lores), bodyMat)
		ear.MeshNode().SetPosition(g3d.Vec3{X: 0.75 * xSign, Y: 1.15, Z: 0.0})
		ear.MeshNode().SetScale(g3d.Vec3{X: 0.7, Y: 1.5, Z: 0.6})
		ear.MeshNode().SetName(name)
		root.Add(ear.MeshNode())
	}
	addEar("LeftEar", -1)
	addEar("RightEar", 1)

	// --- Arms (stubby) ---
	addArm := func(name string, xSign float32) {
		arm := g3d.NewMesh(g3d.NewSphereGeometry(0.15, lores), bodyMat)
		arm.MeshNode().SetPosition(g3d.Vec3{X: 0.95 * xSign, Y: -0.1, Z: 0.15})
		arm.MeshNode().SetScale(g3d.Vec3{X: 0.6, Y: 1.0, Z: 0.7})
		arm.MeshNode().SetName(name)
		root.Add(arm.MeshNode())

		hand := g3d.NewMesh(g3d.NewSphereGeometry(0.1, lores), skinMat)
		hand.MeshNode().SetPosition(g3d.Vec3{X: 1.0 * xSign, Y: -0.25, Z: 0.25})
		hand.MeshNode().SetName(name + "Hand")
		root.Add(hand.MeshNode())
	}
	addArm("LeftArm", -1)
	addArm("RightArm", 1)

	// --- Feet ---
	addFoot := func(name string, xSign float32) {
		foot := g3d.NewMesh(g3d.NewSphereGeometry(0.18, lores), skinMat)
		foot.MeshNode().SetPosition(g3d.Vec3{X: 0.35 * xSign, Y: -1.3, Z: 0.2})
		foot.MeshNode().SetScale(g3d.Vec3{X: 0.9, Y: 0.5, Z: 1.3})
		foot.MeshNode().SetName(name)
		root.Add(foot.MeshNode())
	}
	addFoot("LeftFoot", -1)
	addFoot("RightFoot", 1)

	// --- Tail ---
	tail := g3d.NewMesh(g3d.NewSphereGeometry(0.12, lores), bodyMat)
	tail.MeshNode().SetPosition(g3d.Vec3{X: 0, Y: -0.4, Z: -0.85})
	tail.MeshNode().SetScale(g3d.Vec3{X: 0.7, Y: 0.7, Z: 0.5})
	tail.MeshNode().SetName("Tail")
	root.Add(tail.MeshNode())

	return n
}
