package component

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
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

const heweatherUrl = "https://free-api.heweather.net/s6/weather/now?location=dayi&key=b765a0cac14c42adb9cc517db0e6fc3a"
const iconFolder = "." + string(os.PathSeparator) + "resource" + string(os.PathSeparator)

type Weather struct {
	code      int
	tmp       int
	img       image.Image
	webClient *http.Client
	notify    chan struct{}
}

func NewWeather() *Weather {
	img := wbimage.NewWB(image.Rect(0, 0, 47, 19))
	for i := range img.Pix {
		img.Pix[i] = true
	}
	img.Set(0, 0, wbimage.WBColor(false))
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//state := c.Query("state")
	client := &http.Client{Transport: tr}
	return &Weather{img: img, notify: make(chan struct{}), webClient: client}
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
func (w *Weather) render(ws weaterStatus) error {
	f, err := os.Open(iconFolder + ws.Icon)
	if err != nil {
		return err
	}
	icon, err := png.Decode(f)
	if err != nil {
		return err
	}
	w.img = imaging.Paste(w.img, icon, image.Point{X: 0, Y: 0})
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
			err = w.render(ws)
			if err != nil {
				return
			}
			w.code = ws.Code
			w.tmp = ws.Tmp
		}
		ok = true
		return
	}
	_, err := draw()
	if err != nil {
		panic(err)
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
	}
}

func (w *Weather) getWeatherStatus() (ws weaterStatus, err error) {
	resp, err := w.webClient.Get(heweatherUrl)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("API Code:%d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	wsResp := heWeatherResp{}
	log.Printf(string(body))
	err = json.Unmarshal(body, &wsResp)
	if err != nil {
		return
	}
	if len(wsResp.HeWeather6) != 1 {
		err = fmt.Errorf("Resp error")
		return
	}
	heWeather6 := wsResp.HeWeather6[0]
	code, err := strconv.Atoi(heWeather6.Now.CondCode)
	if err != nil {
		return
	}
	var ok bool
	if ws, ok = WSMap[code]; !ok {
		ws = WSMap[999]
	} else {
		var tmp float64
		tmp, err = strconv.ParseFloat(heWeather6.Now.Tmp, 10)
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
	100: {Code: 100, Des: "晴", Eng: "Sunny", Icon: "clear.png"},
	101: {Code: 101, Des: "多云", Eng: "Cloudy", Icon: ""},
	102: {Code: 102, Des: "少云", Eng: "Few Clouds", Icon: ""},
	103: {Code: 103, Des: "晴间多云", Eng: "Partly Cloudy", Icon: ""},
	104: {Code: 104, Des: "阴", Eng: "Overcast", Icon: ""},
	200: {Code: 200, Des: "有风", Eng: "Windy", Icon: ""},
	201: {Code: 201, Des: "平静", Eng: "Calm", Icon: ""},
	202: {Code: 202, Des: "微风", Eng: "Light Breeze", Icon: ""},
	203: {Code: 203, Des: "和风", Eng: "Moderate/Gentle Breeze", Icon: ""},
	204: {Code: 204, Des: "清风", Eng: "Fresh Breeze", Icon: ""},
	205: {Code: 205, Des: "强风/劲风", Eng: "Strong Breeze", Icon: ""},
	206: {Code: 206, Des: "疾风", Eng: "High Wind, Near Gale", Icon: ""},
	207: {Code: 207, Des: "大风", Eng: "Gale", Icon: ""},
	208: {Code: 208, Des: "烈风", Eng: "Strong Gale", Icon: ""},
	209: {Code: 209, Des: "风暴", Eng: "Storm", Icon: ""},
	210: {Code: 210, Des: "狂爆风", Eng: "Violent Storm", Icon: ""},
	211: {Code: 211, Des: "飓风", Eng: "Hurricane", Icon: ""},
	212: {Code: 212, Des: "龙卷风", Eng: "Tornado", Icon: ""},
	213: {Code: 213, Des: "热带风暴", Eng: "Tropical Storm", Icon: ""},
	300: {Code: 300, Des: "阵雨", Eng: "Shower Rain", Icon: ""},
	301: {Code: 301, Des: "强阵雨", Eng: "Heavy Shower Rain", Icon: ""},
	302: {Code: 302, Des: "雷阵雨", Eng: "Thundershower", Icon: ""},
	303: {Code: 303, Des: "强雷阵雨", Eng: "Heavy Thunderstorm", Icon: ""},
	304: {Code: 304, Des: "雷阵雨伴有冰雹", Eng: "Thundershower with hail", Icon: ""},
	305: {Code: 305, Des: "小雨", Eng: "Light Rain", Icon: ""},
	306: {Code: 306, Des: "中雨", Eng: "Moderate Rain", Icon: ""},
	307: {Code: 307, Des: "大雨", Eng: "Heavy Rain", Icon: ""},
	308: {Code: 308, Des: "极端降雨", Eng: "Extreme Rain", Icon: ""},
	309: {Code: 309, Des: "毛毛雨/细雨", Eng: "Drizzle Rain", Icon: ""},
	310: {Code: 310, Des: "暴雨", Eng: "Storm", Icon: ""},
	311: {Code: 311, Des: "大暴雨", Eng: "Heavy Storm", Icon: ""},
	312: {Code: 312, Des: "特大暴雨", Eng: "Severe Storm", Icon: ""},
	313: {Code: 313, Des: "冻雨", Eng: "Freezing Rain", Icon: ""},
	314: {Code: 314, Des: "小到中雨", Eng: "Light to moderate rain", Icon: ""},
	315: {Code: 315, Des: "中到大雨", Eng: "Moderate to heavy rain", Icon: ""},
	316: {Code: 316, Des: "大到暴雨", Eng: "Heavy rain to storm", Icon: ""},
	317: {Code: 317, Des: "暴雨到大暴雨", Eng: "Storm to heavy storm", Icon: ""},
	318: {Code: 318, Des: "大暴雨到特大暴雨", Eng: "Heavy to severe storm", Icon: ""},
	399: {Code: 399, Des: "雨", Eng: "Rain", Icon: ""},
	400: {Code: 400, Des: "小雪", Eng: "Light Snow", Icon: ""},
	401: {Code: 401, Des: "中雪", Eng: "Moderate Snow", Icon: ""},
	402: {Code: 402, Des: "大雪", Eng: "Heavy Snow", Icon: ""},
	403: {Code: 403, Des: "暴雪", Eng: "Snowstorm", Icon: ""},
	404: {Code: 404, Des: "雨夹雪", Eng: "Sleet", Icon: ""},
	405: {Code: 405, Des: "雨雪天气", Eng: "Rain And Snow", Icon: ""},
	406: {Code: 406, Des: "阵雨夹雪", Eng: "Shower Snow", Icon: ""},
	407: {Code: 407, Des: "阵雪", Eng: "Snow Flurry", Icon: ""},
	408: {Code: 408, Des: "小到中雪", Eng: "Light to moderate snow", Icon: ""},
	409: {Code: 409, Des: "中到大雪", Eng: "Moderate to heavy snow", Icon: ""},
	410: {Code: 410, Des: "大到暴雪", Eng: "Heavy snow to snowstorm", Icon: ""},
	499: {Code: 499, Des: "雪", Eng: "Snow", Icon: ""},
	500: {Code: 500, Des: "薄雾", Eng: "Mist", Icon: ""},
	501: {Code: 501, Des: "雾", Eng: "Foggy", Icon: ""},
	502: {Code: 502, Des: "霾", Eng: "Haze", Icon: ""},
	503: {Code: 503, Des: "扬沙", Eng: "Sand", Icon: ""},
	504: {Code: 504, Des: "浮尘", Eng: "Dust", Icon: ""},
	507: {Code: 507, Des: "沙尘暴", Eng: "Duststorm", Icon: ""},
	508: {Code: 508, Des: "强沙尘暴", Eng: "Sandstorm", Icon: ""},
	509: {Code: 509, Des: "浓雾", Eng: "Dense fog", Icon: ""},
	510: {Code: 510, Des: "强浓雾", Eng: "Strong fog", Icon: ""},
	511: {Code: 511, Des: "中度霾", Eng: "Moderate haze", Icon: ""},
	512: {Code: 512, Des: "重度霾", Eng: "Heavy haze", Icon: ""},
	513: {Code: 513, Des: "严重霾", Eng: "Severe haze", Icon: ""},
	514: {Code: 514, Des: "大雾", Eng: "Heavy fog", Icon: ""},
	515: {Code: 515, Des: "特强浓雾", Eng: "Extra heavy fog", Icon: ""},
	900: {Code: 900, Des: "热", Eng: "Hot", Icon: ""},
	901: {Code: 901, Des: "冷", Eng: "Cold", Icon: ""},
	999: {Code: 999, Des: "未知", Eng: "Unknown", Icon: ""},
}