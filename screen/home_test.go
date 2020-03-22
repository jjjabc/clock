package screen

import "testing"

func TestScreen_Run(t *testing.T) {
	s:=New12864ClockScreen()
	s.Run()
}