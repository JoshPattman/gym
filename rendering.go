package gym

import (
	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/pixelgl"
)

// Create a window, and start rendering an environment in a loop.
// This function will not update the environment, you should do this inside stepFunc.
//
// Each frame, the stepFunc is called (uses vsync so will proobably be 60/s).
// If stepFunc returns true, the render loop will exit.
// stepFunc also takes a refrence to the window as a parameter, to allow for keypress checking.
func BeginRenderLoop(e Env, stepFunc func(win *pixelgl.Window) bool) {
	pixelgl.Run(func() {
		renderLoop(e, stepFunc)
	})
}

// Helper function for BeginRenderLoop.
func renderLoop(e Env, stepFunc func(win *pixelgl.Window) bool) {
	dx, dy := e.RenderSize()
	cfg := pixelgl.WindowConfig{
		Title:  "Gym: " + e.Name(),
		Bounds: pixel.R(0, 0, dx, dy),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		if stepFunc(win) {
			break
		}
		e.Render(win)
		win.Update()
	}
}
