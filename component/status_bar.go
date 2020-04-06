package component

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/jjjabc/clock/tools"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"io/ioutil"
	"log"
	"time"
)

type Status struct {
	img *wbimage.WB
	f   *truetype.Font
	notify    chan struct{}
}

func (s *Status) Run() {
	ip, err := tools.ExternalIP()
	if err != nil {
		log.Printf(err.Error())
		return
	}
	log.Printf(ip.String())
	img, _ := tools.StringSrcPic(s.img, "IP:"+ip.String(), 8, s.f, 63, -1)
	s.img = img

	go func(){
		now:=time.Now()
		s.img, _ = tools.StringSrcPic(s.img, now.Format("Mon Jan 2"), 8, s.f, 13, -1)
		timer:=time.NewTimer(time.Until(time.Date(now.Year(),now.Month(),now.Day()+1,0,0,1,0,now.Location())))
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				now:=time.Now()
				img, _ = tools.StringSrcPic(s.img, now.Format("Mon Jan 2"), 8, s.f, 13, -1)
				s.img = img
				s.notify<- struct{}{}
				timer.Reset(time.Until(time.Date(now.Year(),now.Month(),now.Day()+1,0,0,1,0,now.Location())))
			}

		}


	}()
}

func (s *Status) Render() image.Image {
	return s.img
}

func NewStatusBar() *Status {
	bg := wbimage.NewWB(image.Rect(0, 0, 126, 7))
	for i := range bg.Pix {
		bg.Pix[i] = true
	}
	fontBytes, err := ioutil.ReadFile("./resource/04.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	return &Status{img: bg, f: f}
}
func (s *Status) Bounds() image.Rectangle {
	return s.img.Bounds()
}

func (s *Status) Notify() <-chan struct{} {
	return s.notify
}
