package main

import "github.com/gogpu/g3d"

func updateScene(cube *g3d.Mesh, dt float64) {
	r := cube.MeshNode().Rotation
	r.Y += float32(dt)
	r.X += float32(dt) * 0.3
	cube.MeshNode().SetRotation(r)
}

func buildScene() (*g3d.Scene, *g3d.PerspectiveCamera, *g3d.Mesh) {
	scene := g3d.NewScene()
	scene.SetBackground(g3d.RGB(0.1, 0.1, 0.15))

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
	sun.LightNode().SetRotation(g3d.Euler{
		X: g3d.Radians(-45),
		Y: g3d.Radians(30),
	})
	scene.Add(sun.LightNode())

	cube := g3d.NewMesh(
		g3d.NewBoxGeometry(1, 1, 1),
		g3d.NewStandardMaterial(
			g3d.WithColor(g3d.RGB(0.4, 0.7, 1.0)),
			g3d.WithMetallic(0.3),
			g3d.WithRoughness(0.6),
		),
	)
	scene.Add(cube.MeshNode())

	camera := g3d.NewPerspectiveCamera(75, 800.0/600.0, 0.1, 1000)
	camera.CameraNode().SetPosition(g3d.Vec3{X: 0, Y: 0.5, Z: 3})

	return scene, camera, cube
}
