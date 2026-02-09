package cmd

import (
	"hash/fnv"

	"github.com/fatih/color"
)

var colorPalette = []color.Attribute{
	color.FgRed,
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
	color.FgHiRed,
	color.FgHiGreen,
	color.FgHiYellow,
	color.FgHiBlue,
	color.FgHiMagenta,
	color.FgHiCyan,
}

type Colorizer struct {
	enabled bool
	cache   map[string]*color.Color
}

func NewColorizer(enabled bool) *Colorizer {
	return &Colorizer{
		enabled: enabled,
		cache:   make(map[string]*color.Color),
	}
}

func (c *Colorizer) Sprint(managerName, s string) string {
	if c == nil || !c.enabled {
		return s
	}
	clr, ok := c.cache[managerName]
	if !ok {
		h := fnv.New32a()
		h.Write([]byte(managerName))
		idx := int(h.Sum32()) % len(colorPalette)
		clr = color.New(colorPalette[idx])
		c.cache[managerName] = clr
	}
	return clr.Sprint(s)
}
