package component

import (
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/jjjabc/clock/tools"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

const heWeatherForecastUrl = "https://devapi.qweather.com/v7/weather/3d"

type WeatherForecast struct {
	img        image.Image
	bg         *wbimage.WB
	webClient  *http.Client
	notify     chan struct{}
	itemRender *forecastItemRender
	f          *truetype.Font
}

func NewWeatherForecast(client *http.Client) *WeatherForecast {
	img := wbimage.NewWB(image.Rect(0, 0, 126, 19))
	for i := range img.Pix {
		img.Pix[i] = true
	}
	bg := wbimage.Clone(img)

	fontBytes, err := ioutil.ReadFile("./resource/04.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	return &WeatherForecast{
		img:        img,
		bg:         bg,
		webClient:  client,
		notify:     make(chan struct{}),
		itemRender: NewForecastItemRender(),
		f:          f,
	}
}

func (w *WeatherForecast) Render() image.Image {
	return w.img
}

func (w *WeatherForecast) Bounds() image.Rectangle {
	return w.img.Bounds()
}

func (w *WeatherForecast) Notify() <-chan struct{} {
	return w.notify
}

func (w *WeatherForecast) Run() {
	ticker := time.NewTicker(5 * time.Minute)
	var wfs []weatherForecastStatus
	draw := func() (ok bool, err error) {
		newWfs, err := w.getForecast()
		if err != nil {
			return
		}
		var changed bool
		if len(newWfs) == len(wfs) {
			for i := range wfs {
				if newWfs[i] != wfs[i] {
					changed = true
				}
			}
		} else {
			changed = true
		}
		if changed {
			wfs = newWfs
			err = w.render(wfs)
			if err != nil {
				return
			}
		}
		ok = true
		return
	}
	w.renderErr(fmt.Errorf("Loading..."))
	imaging.Save(w.img, "wf.png")
	go func() {
		defer ticker.Stop()
		time.Sleep(10 * time.Second)
		_, err := draw()
		if err != nil {
			w.renderErr(err)
			log.Println(err)
			//return
		}
		for {
			select {
			case <-ticker.C:
				ok, err := draw()
				if err != nil {
					//todo push error
					continue
				}
				if ok {
					w.notify <- struct{}{}
				}
			}
		}
	}()
}
func (w *WeatherForecast) renderErr(err error) error {
	var img image.Image = wbimage.Clone(w.bg)
	w.img, _ = tools.StringSrcPic(wbimage.Convert(img), err.Error(), 8, w.f, 0, 0)
	return nil
}
func (w *WeatherForecast) render(wfs []weatherForecastStatus) (err error) {
	var img image.Image = wbimage.Clone(w.bg)
	for i := range wfs {
		itemImg, err := w.itemRender.renderItem(wfs[i])
		if err != nil {
			return err
		}
		img = imaging.Paste(img, itemImg, image.Pt(i*(itemImg.Bounds().Dx()+5)+1, 1))
	}
	w.img = img
	imaging.Save(img, "img.png")
	return nil
}

type forecastItemRender struct {
	bg                image.Image
	tmpFont, dateFont *truetype.Font
	maxIcon, minIcon  image.Image
	negativeSignIcon  image.Image
	tmpFontArea       *wbimage.WB
}

func NewForecastItemRender() *forecastItemRender {
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
	area := wbimage.NewWB(image.Rect(0, 0, 20, 5))
	for i := range area.Pix {
		area.Pix[i] = true
	}
	return &forecastItemRender{bg: bg,
		tmpFont:          tmpFont,
		dateFont:         dateFont,
		maxIcon:          maxIcon,
		minIcon:          minIcon,
		negativeSignIcon: negativeSignIcon,
		tmpFontArea:      area}
}

func (r *forecastItemRender) renderItem(wfs weatherForecastStatus) (img image.Image, err error) {
	f, err := os.Open(iconFolder + wfs.Icon)
	if err != nil {
		return
	}
	icon, err := png.Decode(f)
	if err != nil {
		return
	}
	defer f.Close()
	img = imaging.Paste(r.bg, icon, image.Pt(21, 0))
	img = imaging.Paste(img, r.maxIcon, image.Pt(15, 0))
	img = imaging.Paste(img, r.minIcon, image.Pt(15, 6))

	maxTmp := int(math.Floor(float64(wfs.TmpMax)/100.0 + 0.5))
	var maxTmpNegativeSignIcon, minTmpNegativeSignIcon image.Image
	maxTmpNegativeSignIcon = image.NewNRGBA(image.Rect(0, 0, 0, 0))
	minTmpNegativeSignIcon = image.NewNRGBA(image.Rect(0, 0, 0, 0))
	if maxTmp < 0 {
		maxTmp = 0 - maxTmp
		maxTmpNegativeSignIcon = r.negativeSignIcon
	}
	minTmp := int(math.Floor(float64(wfs.TmpMin)/100.0 + 0.5))

	if minTmp < 0 {
		minTmp = 0 - minTmp
		minTmpNegativeSignIcon = r.negativeSignIcon
	}

	maxWbImg := wbimage.Convert(r.tmpFontArea)
	minWbImg := wbimage.Convert(r.tmpFontArea)
	maxWbImg, maxEndPos := tools.StringSrcPic(maxWbImg, strconv.Itoa(maxTmp), 10, r.tmpFont, 0, -4)
	minWbImg, minEndPos := tools.StringSrcPic(minWbImg, strconv.Itoa(minTmp), 10, r.tmpFont, 0, -4)
	//会多渲染1个px的空像素
	maxTmpImg := imaging.Crop(maxWbImg, image.Rect(0, 0, maxEndPos.X-1, maxEndPos.Y))
	minTmpImg := imaging.Crop(minWbImg, image.Rect(0, 0, minEndPos.X-1, minEndPos.Y))
	img = imaging.Paste(img, maxTmpImg, image.Pt(14-maxTmpImg.Bounds().Dx(), 0))
	img = imaging.Paste(img, minTmpImg, image.Pt(14-minTmpImg.Bounds().Dx(), 6))
	//渲染负号
	img = imaging.Paste(img, maxTmpNegativeSignIcon, image.Pt(14-maxTmpImg.Bounds().Dx()-3, 2))
	img = imaging.Paste(img, minTmpNegativeSignIcon, image.Pt(14-minTmpImg.Bounds().Dx()-3, 8))

	dateWbImg := wbimage.Convert(r.tmpFontArea)
	dateWbImg, dateEndPos := tools.StringSrcPic(dateWbImg, wfs.Date.Format("01-02"), 8, r.dateFont, 0, -2)
	dateTmpImg := imaging.Crop(dateWbImg, image.Rect(0, 0, dateEndPos.X-1, dateEndPos.Y))
	img = imaging.Paste(img, dateTmpImg, image.Pt(20-dateTmpImg.Bounds().Dx(), 12))

	return img, nil
}

type weatherForecastStatus struct {
	Code   int
	Date   time.Time
	TmpMax int
	TmpMin int
	Des    string
	Eng    string
	Icon   string
}

func (w *WeatherForecast) getForecast() (wfs []weatherForecastStatus, err error) {
	resp, err := w.webClient.Get(heWeatherForecastUrl + "?" + "location=" + location + "&" + "key=" + key)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("API Code:%d", resp.StatusCode)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	wsResp := heWeatherForecastRespV7{}
	log.Printf("get forecast ok")
	err = json.Unmarshal(body, &wsResp)
	if err != nil {
		return
	}
	/*	if len(wsResp.HeWeather6) != 1 {
		err = fmt.Errorf("Resp error")
		return
	}*/
	heWeather7 := wsResp
	df := heWeather7.DailyForecast
	if len(df) < 3 {
		err = fmt.Errorf("Resp errpr")
		return
	}
	wfs = make([]weatherForecastStatus, 0)
	for i := range df {
		var code int
		var tmp float64
		code, err = strconv.Atoi(df[i].CondCodeD)
		if err != nil {
			return
		}
		var ws weaterStatus
		var ok bool
		if ws, ok = WSMap[code]; !ok {
			ws = WSMap[999]
		}
		tmp, err = strconv.ParseFloat(df[i].TmpMax, 10)
		if err != nil {
			return
		}
		tmpMax := int(math.Floor(tmp*100 + 0.5))
		tmp, err = strconv.ParseFloat(df[i].TmpMin, 10)
		if err != nil {
			return
		}
		tmpMin := int(math.Floor(tmp*100 + 0.5))
		wfsDate, err := time.Parse("2006-01-02", df[i].Date)
		if err != nil {
			return wfs, err
		}
		s := weatherForecastStatus{
			Code:   code,
			Date:   wfsDate,
			TmpMax: tmpMax,
			TmpMin: tmpMin,
			Des:    ws.Des,
			Eng:    ws.Eng,
			Icon:   ws.Icon,
		}
		wfs = append(wfs, s)
	}
	return
}
