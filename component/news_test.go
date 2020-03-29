package component

import (
	"context"
	"github.com/golang/freetype"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"io/ioutil"
	"testing"
	"time"
)

func TestRollingBanner(t *testing.T) {
	ctx,cancelFun:=context.WithTimeout(context.Background(),5*time.Second)
	defer cancelFun()
	fontBytes, err := ioutil.ReadFile("./resource/12.ttf")
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	RollingBanner(ctx,"hello world,hahaha",wbimage.NewWB(image.Rect(0,0,20,20)),5,12,f,0,make(chan struct{},500))
}
