package tools

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/jjjabc/lcd/wbimage"
	"image"
	"log"
)

func StringSrcPic(src *wbimage.WB, str string, sizePx int, f *truetype.Font, x, y int)(dst *wbimage.WB) {
	dst=wbimage.Clone(src)
	fg := image.Black
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(float64(sizePx))
	c.SetClip(src.Bounds())
	c.SetDst(dst)
	c.SetSrc(fg)
	pt := freetype.Pt(x, y+int(c.PointToFixed(float64(sizePx))>>6)-1)
	var err error
	pt, err = c.DrawString(str, pt)
	if err != nil {
		log.Println(err)
		return
	}
	return
}
