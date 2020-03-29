package screen

import (
	"github.com/disintegration/imaging"
	"github.com/jjjabc/clock/component"
	"github.com/jjjabc/clock/layout"
	"github.com/jjjabc/lcd"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"image/png"
	"os"
)

type Screen struct {
	main layout.Com
	bg   *wbimage.WB
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
	return &Screen{
		main:
		layout.NewContainer(layout.Vertical,
			layout.NewContainer(layout.Horizontal,
				component.NewWeather(), component.NewClock()),
			component.NewWeatherForecast(),
			component.NewNews(),
		),
		bg: bg,
	}
}
func (s *Screen) Render() image.Image {
	mainImg := s.main.Render()
	//screenImg := imaging.Crop(mainImg, image.Rect(0, 0, s.bg.Bounds().Dx(), s.bg.Bounds().Dy()))
	screenImg := imaging.Paste(s.bg, mainImg, image.Pt(1, 1))
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
			/*err := saveImg(s.Render())
			log.Printf("save")
			if err != nil {
				panic(err)
			}*/
		}
	}
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
