package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type TokyoNightStormTheme struct{}

func NewTokyoNightStormTheme() *TokyoNightStormTheme {
	return &TokyoNightStormTheme{}
}

// Credits: https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
// Yes i could have written it myself
// but someone else has probably already made it :)
func parseHexColor(s string) (c color.RGBA) {
	c.A = 0xff
	switch len(s) {
	case 7:
		fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
	}
	return
}

var (
	bgColor        = parseHexColor("#24283b")
	fgColor        = parseHexColor("#a9b1d6")
	primaryColor   = parseHexColor("#bb9af7")
	secondaryColor = parseHexColor("#414868")
	highlightColor = parseHexColor("#565f89")
	borderColor    = parseHexColor("#1a1b26")
	errorColor     = parseHexColor("#f7768e")
)

func (t *TokyoNightStormTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return bgColor
	case theme.ColorNameForeground:
		return fgColor
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameInputBackground:
		return borderColor
	case theme.ColorNameFocus:
		return highlightColor
	case theme.ColorNameButton, theme.ColorNameHover:
		return secondaryColor
	case theme.ColorNameError:
		return errorColor
	case theme.ColorNameSelection:
		return secondaryColor
	default:
		return bgColor
	}
}

func (t *TokyoNightStormTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func (t *TokyoNightStormTheme) Font(name fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(name)
}

func (t *TokyoNightStormTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
