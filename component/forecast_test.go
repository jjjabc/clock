package component

import (
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestWeatherForecast_Render(t *testing.T) {
	bg := wbimage.NewWB(image.Rect(0, 0, 38, 17))
	for i := range bg.Pix {
		bg.Pix[i] = true
	}
	fontBytes, err := ioutil.ReadFile("./resource/5_D75.ttf")
	if err != nil {
		panic(err)
	}
	tmpFont, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}

	fontBytes, err = ioutil.ReadFile("./resource/04.ttf")
	if err != nil {
		panic(err)
	}
	dateFont, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	pngFile, err := os.Open(iconFolder + "max_tmp.png")
	if err != nil {
		panic(err)
	}
	maxIcon, err := png.Decode(pngFile)
	if err != nil {
		panic(err)
	}
	pngFile, err = os.Open(iconFolder + "min_tmp.png")
	if err != nil {
		panic(err)
	}
	minIcon, err := png.Decode(pngFile)
	if err != nil {
		panic(err)
	}
	pngFile, err = os.Open(iconFolder + "negative_sign.png")
	if err != nil {
		panic(err)
	}
	negativeSignIcon, err := png.Decode(pngFile)
	if err != nil {
		panic(err)
	}
	area:=wbimage.NewWB(image.Rect(0,0,20,5))
	for i:=range area.Pix{
		area.Pix[i]=true
	}
	r := &forecastItemRender{bg: bg,
		tmpFont: tmpFont,
		dateFont: dateFont,
		maxIcon: maxIcon,
		minIcon: minIcon,
		negativeSignIcon: negativeSignIcon,
		tmpFontArea:area}
	img, err := r.renderItem(weatherForecastStatus{
		Code:   100,
		Date:   time.Now(),
		TmpMax: 0,
		TmpMin: 2200,
		Des:    "test",
		Eng:    "eng",
		Icon:   "sunny.png",
	})
	if err != nil {
		panic(err)
	}
	imaging.Save(img, "test.png")

}
