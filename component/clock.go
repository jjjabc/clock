package component

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/jjjabc/clock/tools"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"io/ioutil"
	"time"
)

type Clock struct {
	width, height int
	font          *truetype.Font
	imgBackGround *wbimage.WB
	notify        chan struct{}
	ticker        *time.Ticker
}

func (c *Clock) Notify() <-chan struct{} {
	return c.notify
}

func (c *Clock) Run() {
	c.ticker = time.NewTicker(time.Minute)
	go func() {
		defer c.ticker.Stop()
		for {
			select {
			case <-c.ticker.C:
				c.notify <- struct{}{}
			}
		}
	}()
}

func (c *Clock) Destroy() {
	close(c.notify)
}

func NewClock() *Clock {
	fontBytes, err := ioutil.ReadFile("./resource/clock.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	bg := wbimage.NewWB(image.Rect(0, 0, 79, 19))
	for i := range bg.Pix {
		bg.Pix[i] = true
	}
	clock := &Clock{width: 63,
		height:        19,
		font:          f,
		imgBackGround: bg,
		notify:        make(chan struct{}),
	}
	return clock
}
func (c *Clock) Render() (img image.Image) {
	img = wbimage.Clone(c.imgBackGround)
	img = tools.StringSrcPic(img.(*wbimage.WB), time.Now().Format("03:04"), 20, c.font, 1, -1)
	img = tools.StringSrcPic(img.(*wbimage.WB), time.Now().Format("PM"), 10, c.font, 62, 9)
	return
}

func (c *Clock) Bounds() image.Rectangle {
	return c.imgBackGround.Rect
}
