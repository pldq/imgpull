package display

import (
	"github.com/morikuni/aec"
)

type colorFunc func(string) string

var (
	nocolor colorFunc = func(s string) string {
		return s
	}

	TimerColor   colorFunc = aec.BlueF.Apply
	CountColor   colorFunc = aec.YellowF.Apply
	WarningColor colorFunc = aec.YellowF.With(aec.Bold).Apply
	SuccessColor colorFunc = aec.GreenF.Apply
	ErrorColor   colorFunc = aec.RedF.With(aec.Bold).Apply
)
