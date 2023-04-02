// createDate:2020/3/26 下午9:23
// desc:$END$
//
//author:xunj
package component

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/jjjabc/clock/tools"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type News struct {
	notify    chan struct{}
	img       image.Image
	webClient *http.Client
	news      []jdNewsItem
	f         *truetype.Font
	disString string
}

func NewNews() *News {
	bg := wbimage.NewWB(image.Rect(0, 0, 126, 14))
	for i := range bg.Pix {
		bg.Pix[i] = true
	}
	fontBytes, err := ioutil.ReadFile("./resource/12.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//state := c.Query("state")
	client := &http.Client{Transport: tr}
	return &News{notify: make(chan struct{}), img: bg, webClient: client, f: f}
}
func (n *News) Run() {
	ticker := time.NewTicker(5 * time.Minute)
	var cancelFun context.CancelFunc
	var ctx context.Context
	ctx, cancelFun = context.WithCancel(context.Background())
	err := n.getNews()
	if err != nil {
		n.disString = "错误:5分钟后重试（" + err.Error() + "）"
	}
	if wbImg, ok := n.img.(*wbimage.WB); ok {
		n.disString = n.String()
		go RollingBanner(ctx, n.String(), wbImg, 2, 12, n.f, 0, n.notify)
	}
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := n.getNews()
				if err != nil {
					//todo 错误处理
					//n.disString="错误:1分钟后重试（"+err.Error()+"）"
				}
				if n.disString == n.String() {
					continue
				}
				cancelFun()
				ctx, cancelFun = context.WithCancel(context.Background())
				if wbImg, ok := n.img.(*wbimage.WB); ok {
					go RollingBanner(ctx, n.String(), wbImg, 2, 12, n.f, 0, n.notify)
				}
			}
		}
	}()
}

func (n News) String() string {
	dis := ""
	for _, item := range n.news {
		dis = dis + item.Title + " " + item.Src + " " + item.Time + "    "
	}
	return dis
}

func (n *News) Render() image.Image {
	return n.img
}

func (n *News) Bounds() image.Rectangle {
	return n.img.Bounds()
}

func (n *News) Notify() <-chan struct{} {
	return n.notify
}

type wayJd struct {
	Result struct {
		Status int             `json:"status"`
		Result json.RawMessage `json:"result"`
	} `json:"result"`
}
type jdNews struct {
	Channel string       `json:"channel"`
	List    []jdNewsItem `json:"list"`
}
type jdNewsItem struct {
	Title string `json:"title"`
	Time  string `json:"time"`
	Src   string `json:"src"`
}

func (n *News) getNews() (err error) {
	resp, err := n.webClient.Get("https://way.jd.com/jisuapi/get?channel=%E6%96%B0%E9%97%BB&num=10&start=0&appkey=28cef4566c4bd850b11e77a36bd9a5ed")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("API Code:%d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	jdResp := wayJd{}
	err = json.Unmarshal(body, &jdResp)
	if err != nil {
		return
	}
	if jdResp.Result.Status != 0 {
		err = fmt.Errorf("get news error:status!=0")
		return
	}
	news := jdNews{}
	err = json.Unmarshal(jdResp.Result.Result, &news)
	if err != nil {
		return
	}
	n.news = make([]jdNewsItem, 0)
	for _, item := range news.List {
		switch item.Src {
		case "微博短视频", "黑猫投诉":
			continue
		default:
			n.news = append(n.news, item)
		}
	}
	return nil
}

func RollingBanner(ctx context.Context, content string, wbImg *wbimage.WB, speed int, sizePx int, f *truetype.Font, y int, notify chan<- struct{}) {
	d := 100 * time.Millisecond
	step := 1
	switch speed {
	case 1:
		d = d * 4
	case 2:
		d = d * 2
	case 3:

	case 4:
		step = 2
	case 5:
		step = 4
	default:

	}
	var pt image.Point

	bg := wbimage.NewWB(wbImg.Bounds())
	for i := range wbImg.Pix {
		bg.Pix[i] = true
	}
	_, pt = tools.StringSrcPic(bg, content, sizePx, f, 0, y)
	width := pt.X - 0
	contentImg := wbimage.NewWB(image.Rect(0, 0, width, wbImg.Bounds().Dy()))
	for i := range contentImg.Pix {
		contentImg.Pix[i] = true
	}
	contentImg, _ = tools.StringSrcPic(contentImg, content, sizePx, f, 0, y)
	ticker := time.NewTicker(d)
	x := 0
	dstNRGB := imaging.Paste(bg, contentImg, image.Pt(x, y))
	dst := wbimage.Convert(dstNRGB)
	for i := range dst.Pix {
		wbImg.Pix[i] = dst.Pix[i]
	}
	log.Printf("News:%s", content)
	for {
		defer ticker.Stop()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if x+width < 0 {
				x = wbImg.Bounds().Max.X
			}
			dstNRGB := imaging.Paste(bg, contentImg, image.Pt(x, y))
			dst := wbimage.Convert(dstNRGB)
			for i := range dst.Pix {
				wbImg.Pix[i] = dst.Pix[i]
			}
			notify <- struct{}{}
		}
		x = x - step
	}
}
