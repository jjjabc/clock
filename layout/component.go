package layout

import "image"

type Com interface {
	Render() image.Image
	Bounds() image.Rectangle
	Notify() <-chan struct{}
}

type Runner interface {
	Run()
}