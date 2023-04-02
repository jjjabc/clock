package screen

import (
	"crypto/tls"
	"github.com/disintegration/imaging"
	"github.com/jjjabc/clock/component"
	"github.com/jjjabc/clock/layout"
	"github.com/jjjabc/lcd"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"image/png"
	"net/http"
	"os"
	"sync"
)

var (
	panda1 = mustOpenImg("." + string(os.PathSeparator) + "resource" + string(os.PathSeparator) + "panda1.png")
	panda2 = mustOpenImg("." + string(os.PathSeparator) + "resource" + string(os.PathSeparator) + "panda2.png")
)

func mustOpenImg(file string) image.Image {
	img, err := imaging.Open(file)
	if err != nil {
		panic(err)
	}
	return img
}

type Screen struct {
	main  layout.Com
	bg    *wbimage.WB
	alert *wbimage.WB
	mutex sync.Mutex
	n     chan struct{}
}

func (s *Screen) Bounds() image.Rectangle {
	return s.main.Bounds()
}

func (s *Screen) Notify() <-chan struct{} {
	panic("implement me")
}

func New12864ClockScreen() *Screen {
	bg := wbimage.NewWB(image.Rect(0, 0, 128, 64))
	for i := range bg.Pix {
		bg.Pix[i] = true
	}
	for x := bg.Bounds().Min.X; x <= bg.Bounds().Max.X; x++ {
		bg.Set(x, bg.Bounds().Min.Y, wbimage.WBColor(false))
		bg.Set(x, bg.Bounds().Max.Y-1, wbimage.WBColor(false))
	}
	for y := bg.Bounds().Min.Y; y <= bg.Bounds().Max.Y; y++ {
		bg.Set(bg.Bounds().Min.X, y, wbimage.WBColor(false))
		bg.Set(bg.Bounds().Max.X-1, y, wbimage.WBColor(false))
	}
	/*	return &Screen{
		main: layout.NewContainer(layout.Vertical,
			layout.NewContainer(layout.Horizontal,
				&component.Weather{}, component.NewClock())),
		bg: bg,
	}*/
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//state := c.Query("state")
	client := &http.Client{Transport: tr}
	return &Screen{
		main: layout.NewContainer(layout.Vertical,
			layout.NewContainer(layout.Horizontal,
				component.NewWeather(client), component.NewClock()),
			component.NewWeatherForecast(client),
			component.NewNews(),
			component.NewStatusBar(),
		),
		bg: bg,
		n:  make(chan struct{}),
	}
}
func (s *Screen) Render() image.Image {
	mainImg := s.main.Render()
	//screenImg := imaging.Crop(mainImg, image.Rect(0, 0, s.bg.Bounds().Dx(), s.bg.Bounds().Dy()))
	screenImg := imaging.Paste(s.bg, mainImg, image.Pt(1, 1))
	s.mutex.Lock()
	if s.alert != nil {
		screenImg = imaging.Paste(screenImg, s.alert, image.Pt((screenImg.Bounds().Dx()-s.alert.Bounds().Dx())/2, (screenImg.Bounds().Dy()-s.alert.Bounds().Dy())/2))
	}
	s.mutex.Unlock()
	return screenImg
}
func (s *Screen) Run() {
	if runner, ok := s.main.(layout.Runner); ok {
		runner.Run()
	}
	n := s.main.Notify()
	lcd.Picture(s.Render())
	/*err := saveImg(s.Render())
	if err != nil {
		panic(err)
	}*/
	for {
		select {
		case _, isOpen := <-n:
			if !isOpen {
				return
			}
			lcd.Picture(s.Render())
		case <-s.n:
			lcd.Picture(s.Render())
		}
	}
}
func (s *Screen) ShowAlert(text string) {
	func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		alertImg := wbimage.NewWB(image.Rect(0, 0, 100, 50))
		content := wbimage.NewWB(image.Rect(0, 0, 98, 48))
		for i := range content.Pix {
			content.Pix[i] = true
		}
		img := imaging.Paste(content, panda1, image.Pt(13, 13))
		img = imaging.Paste(img, panda2, image.Pt(63, 13))
		img = imaging.Paste(alertImg, img, image.Pt(1, 1))
		s.alert = wbimage.Convert(img)
	}()
	s.n <- struct{}{}
}
func (s *Screen) HideAlert() {
	func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.alert = nil
	}()
	s.n <- struct{}{}
}
func saveImg(image image.Image) (err error) {
	f, err := os.Create("out.png")
	if err != nil {
		return
	}
	err = png.Encode(f, image)
	defer f.Close()
	return
}
