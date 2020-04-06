package button

import (
	"github.com/stianeikeland/go-rpio/v4"
	"time"
)

var (
	input  = rpio.Pin(4)
)

type Button struct {
	input    rpio.Pin
	notify   chan struct{}
	callback func()
}

func NewButton(pin rpio.Pin) *Button {
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.RiseEdge)
	return &Button{input:pin}
}

func (b *Button) DownEvent() <-chan struct{} {
	if b.notify == nil {
		b.run()
	}
	return b.notify
}
func (b *Button) Callback(callbackFun func()) {
	//log.Panicln(b.input.Read())
	if b.notify==nil{
		b.run()
	}
	b.callback = callbackFun
}

func (b *Button) run() {
	b.notify = make(chan struct{})
	go func() {
		for {
			if b.input.EdgeDetected(){
				select {
				case b.notify<- struct{}{}:
				default:
				}
				go b.callback()
			}
			time.Sleep(200*time.Millisecond)
		}
	}()
}
