package main

import (
	"fmt"
	"github.com/jjjabc/clock/button"
	"github.com/jjjabc/clock/screen"
	"github.com/jjjabc/lcd"
	"github.com/stianeikeland/go-rpio/v4"
	"os"
)

// GOARCH=arm GOOS=linux go build
func main() {
	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()
	lcd.Init()
	lcd.ImageMod()
	lcd.Clear()
	s := screen.New12864ClockScreen()
	b := button.NewButton(4)
	var alertShown bool
	b.Callback(func() {
		if alertShown {
			s.HideAlert()
			alertShown=false
		} else {
			s.ShowAlert("Hello")
			alertShown=true
		}
	})
	s.Run()
}
