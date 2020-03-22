package component

import (
	"github.com/jjjabc/lcd/wbimage"
	"image"
)

type Weather struct {
}

func (w *Weather) Render() image.Image {
	img:= wbimage.NewWB(image.Rect(0, 0, 47, 19))
	for i:=range img.Pix{
		img.Pix[i]=true
	}
	img.Set(0,0,wbimage.WBColor(false))
	return img
}

func (w *Weather) Bounds() image.Rectangle {
	return image.Rect(0, 0, 47, 19)
}

func (w *Weather) Notify() <-chan struct{} {
	return make(chan struct{})
}
