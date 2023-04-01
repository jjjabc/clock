package component

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/jjjabc/clock/tools"
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

	"github.com/disintegration/imaging"

	"github.com/jjjabc/lcd/wbimage"
)

const heweatherUrl = "https://devapi.qweather.com/v7/weather/now"
const iconFolder = "." + string(os.PathSeparator) + "resource" + string(os.PathSeparator) + "weather" + string(os.PathSeparator)
const key = "b765a0cac14c42adb9cc517db0e6fc3a"
const location = "101270108"

type Weather struct {
	code      int
	tmp       int
	img       image.Image
	bg        image.Image
	webClient *http.Client
	notify    chan struct{}
	font      *truetype.Font
	c         image.Image
}

func NewWeather() *Weather {
	img := wbimage.NewWB(image.Rect(0, 0, 47, 19))
	for i := range img.Pix {
		img.Pix[i] = true
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//state := c.Query("state")
	client := &http.Client{Transport: tr}
	fontBytes, err := ioutil.ReadFile("./resource/12.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	bg := wbimage.Clone(img)
	pngFile, err := os.Open(iconFolder + "c.png")
	if err != nil {
		panic(err)
	}
	c, err := png.Decode(pngFile)
	if err != nil {
		panic(err)
	}
	return &Weather{img: img, bg: bg, notify: make(chan struct{}), webClient: client, font: f, c: c}
}

func (w *Weather) Render() image.Image {
	return w.img
}

func (w *Weather) Bounds() image.Rectangle {
	return w.img.Bounds()
}

func (w *Weather) Notify() <-chan struct{} {
	return w.notify
}
func (w *Weather) renderErr(err error) error {
	w.img, _ = tools.StringSrcPic(wbimage.NewWB(image.Rect(0, 0, 47, 19)), err.Error(), 12, w.font, 0, 0)
	return nil
}
func (w *Weather) render(ws weaterStatus) error {
	f, err := os.Open(iconFolder + ws.Icon)
	if err != nil {
		return err
	}
	icon, err := png.Decode(f)
	if err != nil {
		return err
	}
	defer f.Close()
	w.img = imaging.Paste(w.bg, icon, image.Point{X: 1, Y: 1})
	w.img = imaging.Paste(w.img, w.c, image.Pt(38, 9))
	wbImg := wbimage.Convert(w.img)
	tmp := int(math.Floor(float64(w.tmp)/100.0 + 0.5))
	tmpX := 19
	switch {
	case tmp < 10:
		tmpX = 31
	case tmp >= 0:
		tmpX = 25
	case tmp < 0:
		tmpX = 19
	}
	w.img, _ = tools.StringSrcPic(wbImg, strconv.Itoa(tmp), 12, w.font, tmpX, 6)
	return nil
}
func (w *Weather) Run() {
	ticker := time.NewTicker(5 * time.Minute)
	draw := func() (ok bool, err error) {
		ws, err := w.getWeatherStatus()
		if err != nil {
			return
		}
		if w.code != ws.Code || w.tmp != ws.Tmp {
			w.code = ws.Code
			w.tmp = ws.Tmp
			err = w.render(ws)
			if err != nil {
				return
			}
		}
		ok = true
		return
	}
	_, err := draw()
	if err != nil {
		w.renderErr(err)
		return
	}
	go func() {
		defer ticker.Stop()
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

type heWeatherResp struct {
	HeWeather6 []struct {
		Basic struct {
		} `json:"basic"`
		Now struct {
			CondCode string `json:"cond_code"`
			Tmp      string `json:"tmp"`
		} `json:"now"`
		DailyForecast []struct {
			CondCodeD string `json:"cond_code_d"`
			TmpMax    string `json:"tmp_max"`
			TmpMin    string `json:"tmp_min"`
			Date      string `json:"date"`
		} `json:"daily_forecast"`
	}
}

type heWeatherForecastRespV7 struct {
	DailyForecast []struct {
		CondCodeD string `json:"iconDay"`
		TmpMax    string `json:"tempMax"`
		TmpMin    string `json:"tempMin"`
		Date      string `json:"fxDate"`
	} `json:"daily"`
}
type heWeatherNowRespV7 struct {
	Now struct {
		Tmp       string `json:"temp"`
		CondCodeD string `json:"icon"`
	} `json:"now"`
}

func (w *Weather) getWeatherStatus() (ws weaterStatus, err error) {
	resp, err := w.webClient.Get(heweatherUrl + "?" + "location=" + location + "&" + "key=" + key)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("API Code:%d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	wsResp := heWeatherNowRespV7{}
	log.Printf("get weather ok")
	err = json.Unmarshal(body, &wsResp)
	if err != nil {
		return
	}
	/*	if len(wsResp.HeWeather6) != 1 {
		err = fmt.Errorf("Resp error")
		return
	}*/
	heWeather7 := wsResp
	code, err := strconv.Atoi(heWeather7.Now.CondCodeD)
	if err != nil {
		return
	}
	var ok bool
	if ws, ok = WSMap[code]; !ok {
		ws = WSMap[999]
	} else {
		var tmp float64
		tmp, err = strconv.ParseFloat(heWeather7.Now.Tmp, 10)
		ws.Tmp = int(math.Floor(tmp*100 + 0.5))
	}
	return
}

type weaterStatus struct {
	Code int
	Tmp  int
	Des  string
	Eng  string
	Icon string
}

var WSMap map[int]weaterStatus = map[int]weaterStatus{
	100: {Code: 100, Des: "晴", Eng: "Sunny", Icon: "sunny.png"},
	101: {Code: 101, Des: "多云", Eng: "Cloudy", Icon: "cloudy.png"},
	102: {Code: 102, Des: "少云", Eng: "Few Clouds", Icon: "few_clouds.png"},
	103: {Code: 103, Des: "晴间多云", Eng: "Partly Cloudy", Icon: "cloudy.png"},
	104: {Code: 104, Des: "阴", Eng: "Overcast", Icon: "overcast.png"},
	200: {Code: 200, Des: "有风", Eng: "Windy", Icon: "windy.png"},
	201: {Code: 201, Des: "平静", Eng: "Calm", Icon: "windy.png"},
	202: {Code: 202, Des: "微风", Eng: "Light Breeze", Icon: ""},
	203: {Code: 203, Des: "和风", Eng: "Moderate/Gentle Breeze", Icon: "windy.png"},
	204: {Code: 204, Des: "清风", Eng: "Fresh Breeze", Icon: "windy.png"},
	205: {Code: 205, Des: "强风/劲风", Eng: "Strong Breeze", Icon: "windy.png"},
	206: {Code: 206, Des: "疾风", Eng: "High Wind, Near Gale", Icon: "windy.png"},
	207: {Code: 207, Des: "大风", Eng: "Gale", Icon: "windy.png"},
	208: {Code: 208, Des: "烈风", Eng: "Strong Gale", Icon: "windy.png"},
	209: {Code: 209, Des: "风暴", Eng: "Storm", Icon: "windy.png"},
	210: {Code: 210, Des: "狂爆风", Eng: "Violent Storm", Icon: "windy.png"},
	211: {Code: 211, Des: "飓风", Eng: "Hurricane", Icon: "windy.png"},
	212: {Code: 212, Des: "龙卷风", Eng: "Tornado", Icon: "windy.png"},
	213: {Code: 213, Des: "热带风暴", Eng: "Tropical Storm", Icon: "windy.png"},
	300: {Code: 300, Des: "阵雨", Eng: "Shower Rain", Icon: "shower_rain.png"},
	301: {Code: 301, Des: "强阵雨", Eng: "Heavy Shower Rain", Icon: "shower_rain.png"},
	302: {Code: 302, Des: "雷阵雨", Eng: "Thundershower", Icon: "thundershower.png"},
	303: {Code: 303, Des: "强雷阵雨", Eng: "Heavy Thunderstorm", Icon: "thundershower.png"},
	304: {Code: 304, Des: "雷阵雨伴有冰雹", Eng: "Thundershower with hail", Icon: "thundershower.png"},
	305: {Code: 305, Des: "小雨", Eng: "Light Rain", Icon: "light_rain.png"},
	306: {Code: 306, Des: "中雨", Eng: "Moderate Rain", Icon: "moderate_rain.png"},
	307: {Code: 307, Des: "大雨", Eng: "Heavy Rain", Icon: "heavy_rain.png"},
	308: {Code: 308, Des: "极端降雨", Eng: "Extreme Rain", Icon: "extreme_rain.png"},
	309: {Code: 309, Des: "毛毛雨/细雨", Eng: "Drizzle Rain", Icon: "light_rain.png"},
	310: {Code: 310, Des: "暴雨", Eng: "Storm", Icon: "heavy_storm.png"},
	311: {Code: 311, Des: "大暴雨", Eng: "Heavy Storm", Icon: "severe_storm.png"},
	312: {Code: 312, Des: "特大暴雨", Eng: "Severe Storm", Icon: "extreme_rain.png"},
	313: {Code: 313, Des: "冻雨", Eng: "Freezing Rain", Icon: "moderate_rain.png"},
	314: {Code: 314, Des: "小到中雨", Eng: "Light to moderate rain", Icon: "moderate_rain.png"},
	315: {Code: 315, Des: "中到大雨", Eng: "Moderate to heavy rain", Icon: "heavy_rain.png"},
	316: {Code: 316, Des: "大到暴雨", Eng: "Heavy rain to storm", Icon: "heavy_storm.png"},
	317: {Code: 317, Des: "暴雨到大暴雨", Eng: "Storm to heavy storm", Icon: "severe_storm.png"},
	318: {Code: 318, Des: "大暴雨到特大暴雨", Eng: "Heavy to severe storm", Icon: "extreme_rain.png"},
	399: {Code: 399, Des: "雨", Eng: "Rain", Icon: "moderate_rain.png"},
	400: {Code: 400, Des: "小雪", Eng: "Light Snow", Icon: "light_snow.png"},
	401: {Code: 401, Des: "中雪", Eng: "Moderate Snow", Icon: "moderate_snow.png"},
	402: {Code: 402, Des: "大雪", Eng: "Heavy Snow", Icon: "heavy_snow.png"},
	403: {Code: 403, Des: "暴雪", Eng: "Snowstorm", Icon: "heavy_snow.png"},
	404: {Code: 404, Des: "雨夹雪", Eng: "Sleet", Icon: "light_snow.png"},
	405: {Code: 405, Des: "雨雪天气", Eng: "Rain And Snow", Icon: "light_snow.png"},
	406: {Code: 406, Des: "阵雨夹雪", Eng: "Shower Snow", Icon: "light_snow.png"},
	407: {Code: 407, Des: "阵雪", Eng: "Snow Flurry", Icon: "light_snow.png"},
	408: {Code: 408, Des: "小到中雪", Eng: "Light to moderate snow", Icon: "light_snow.png"},
	409: {Code: 409, Des: "中到大雪", Eng: "Moderate to heavy snow", Icon: "moderate_snow.png"},
	410: {Code: 410, Des: "大到暴雪", Eng: "Heavy snow to snowstorm", Icon: "heavy_snow.png"},
	499: {Code: 499, Des: "雪", Eng: "Snow", Icon: "moderate_snow.png"},
	500: {Code: 500, Des: "薄雾", Eng: "Mist", Icon: "foggy.png"},
	501: {Code: 501, Des: "雾", Eng: "Foggy", Icon: "foggy.png"},
	502: {Code: 502, Des: "霾", Eng: "Haze", Icon: "haze.png"},
	503: {Code: 503, Des: "扬沙", Eng: "Sand", Icon: "sand.png"},
	504: {Code: 504, Des: "浮尘", Eng: "Dust", Icon: "sand.png"},
	507: {Code: 507, Des: "沙尘暴", Eng: "Duststorm", Icon: "dust_storm.png"},
	508: {Code: 508, Des: "强沙尘暴", Eng: "Sandstorm", Icon: "dust_storm.png"},
	509: {Code: 509, Des: "浓雾", Eng: "Dense fog", Icon: "dense_fog.png"},
	510: {Code: 510, Des: "强浓雾", Eng: "Strong fog", Icon: "dense_fog.png"},
	511: {Code: 511, Des: "中度霾", Eng: "Moderate haze", Icon: "haze.png"},
	512: {Code: 512, Des: "重度霾", Eng: "Heavy haze", Icon: "haze.png"},
	513: {Code: 513, Des: "严重霾", Eng: "Severe haze", Icon: "haze.png"},
	514: {Code: 514, Des: "大雾", Eng: "Heavy fog", Icon: "foggy.png"},
	515: {Code: 515, Des: "特强浓雾", Eng: "Extra heavy fog", Icon: "dense_fog.png"},
	900: {Code: 900, Des: "热", Eng: "Hot", Icon: "hot.png"},
	901: {Code: 901, Des: "冷", Eng: "Cold", Icon: "clod.png"},
	999: {Code: 999, Des: "未知", Eng: "Unknown", Icon: "unknown.png"},
}
