//author:xunj
//createDate:2020/3/26 下午9:23
//desc:$END$
package component

import (
	"context"
	"encoding/json"
	"fmt"
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
}

func (n *News) Run() {
	ticker := time.NewTicker(time.Minute)
	var cancelFun context.CancelFunc
	var ctx context.Context
	ctx, cancelFun = context.WithCancel(context.Background())
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := n.getNews()
				if err != nil {
					//todo push error
					continue
				}
				cancelFun()
				ctx, cancelFun = context.WithCancel(context.Background())
				if wbImg, ok := n.img.(*wbimage.WB); ok {
					go RollingBanner(ctx, n.String(), wbImg, 3, 12, n.f, 1, n.notify)
				}
			}
		}
	}()
}

func (n News) String() string {

}

func (n *News) Render() image.Image {
	panic("implement me")
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
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("API Code:%d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	jdResp := wayJd{}
	log.Printf(string(body))
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
	ticker := time.NewTicker(d)
	x := 0
	for {
		defer ticker.Stop()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tools.StringSrcPic(wbImg, content, sizePx, f, x, y)
			notify <- struct{}{}
		}
		x = x + step
	}
}
