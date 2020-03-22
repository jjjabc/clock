package layout

import (
	"github.com/disintegration/imaging"
	"image"
	"sync"
)

const (
	Horizontal = 0
	Vertical   = 1
)

type Container struct {
	direction int
	coms      []Com
	img       image.Image
	notify    chan struct{}
	destroyCh chan struct{}
	once      sync.Once
}

func NewContainer(dir int, coms ...Com) *Container {
	return &Container{
		direction: dir,
		coms:      coms,
		img:       nil,
		notify:    make(chan struct{}),
		destroyCh: make(chan struct{}),
		once:      sync.Once{},
	}
}

func (c *Container) Notify() <-chan struct{} {
	return c.notify
}

func (c *Container) Run() {
	for i := range c.coms {
		if com,ok:=c.coms[i].(Runner);ok{
			com.Run()
		}
		n := c.coms[i].Notify()
		go func() {
			for {
				select {
				case <-c.destroyCh:
					return
				case notify, isOpen := <-n:
					if !isOpen {
						return
					}
					c.notify <- notify
				}
			}
		}()
	}
}

func (c *Container) Destroy() {
	c.once.Do(func() {
		close(c.destroyCh)
	})
}

func (c *Container) Render() (img image.Image) {
	img = image.NewRGBA(image.Rect(0, 0, 0, 0))
	for i := range c.coms {
		comImg := c.coms[i].Render()
		//splitLine是垂直分割线
		splitLine := image.NewNRGBA(image.Rect(0, 0, 0, 0))
		if i != len(c.coms)-1 {
			switch c.direction {
			case Horizontal:
				maxDy := img.Bounds().Dy()
				if maxDy < comImg.Bounds().Dy() {
					maxDy = comImg.Bounds().Dy()
				}
				splitLine = imaging.Crop(image.Black, image.Rect(0, 0, 1, maxDy))
			case Vertical:
				maxDx := img.Bounds().Dx()
				if maxDx < comImg.Bounds().Dx() {
					maxDx = comImg.Bounds().Dx()
				}
				splitLine = imaging.Crop(image.Black, image.Rect(0, 0, 1, maxDx))
			}
		}
		var isVertical bool
		switch c.direction {
		case Horizontal:
			isVertical = false
		case Vertical:
			isVertical = true

		}
		img = appendImg(isVertical, img, comImg, splitLine)
	}
	c.img = img
	return img
}

func (c *Container) Bounds() image.Rectangle {
	return c.img.Bounds()
}
func (c *Container) Lasted() image.Image {
	return c.img
}
func (c *Container) Append(elems ...Com) (img image.Image) {
	c.coms = append(c.coms, elems...)
	return c.Render()
}

func appendImg(isVertical bool, src image.Image, elems ...image.Image) (dst image.Image) {
	dst = imaging.Clone(src)
	for i := range elems {
		//0大小图片略过
		if elems[i].Bounds().Dx()+elems[i].Bounds().Dy()==0{
			continue
		}
		var dx, dy int
		if isVertical {
			dx = dst.Bounds().Dx() - 1
			//选择更宽的图片作为扩展后图片的宽
			if dst.Bounds().Dx() < elems[i].Bounds().Dx() {
				dx = elems[i].Bounds().Dx() - 1
			}
			dy = dst.Bounds().Dy() + elems[i].Bounds().Dy() - 1
		} else {
			dy = dst.Bounds().Dy()
			//选择更高的图片作为扩展后图片的高
			if dst.Bounds().Dy() < elems[i].Bounds().Dy() {
				dy = elems[i].Bounds().Dy()
			}
			dx = dst.Bounds().Dx() + elems[i].Bounds().Dx()
		}
		p := dst
		dst = image.NewNRGBA(image.Rect(0, 0, dx, dy))
		dst = imaging.Paste(dst,p,image.Pt(0,0))
		if isVertical {
			//垂直下扩
			dst = imaging.Paste(dst, elems[i], image.Pt(0, p.Bounds().Dy()-1))
		} else {
			//水平右扩
			dst = imaging.Paste(dst, elems[i], image.Pt(p.Bounds().Dx(), 0))
		}
	}
	return
}
