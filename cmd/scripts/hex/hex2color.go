package main

import (
	"fmt"
	"github.com/ebitenui/ebitenui/utilities/colorutil"
	"image/color"
)

var input = [][]string{
	{"85daeb", "c3ecf4", "82ffff"},
	{"5fc9e7", "9ce0ee", "41f0ff"},
	{"5fa1e7", "78c7c7", "4188ff"},
	{"5f6ee7", "a7acff", "4160ff"},
	{"4c60aa", "7b82d6", "1a4389"},
	{"444774", "545171", "0a106c"},
	{"32313b", "000000", "000000"},
	{"463c5e", "794961", "0e003f"},
	{"5d4776", "6868ad", "3d1070"},
	{"855395", "a870ac", "822885"},
	{"ab58a8", "cc80bf", "8a3389"},
	{"ca60ae", "ee9ce0", "f5438a"},
	{"f3a787", "edf5c5", "ff8982"},
	{"f5daa7", "d6f9e9", "ffff89"},
	{"8dd894", "cbf6e2", "84ff85"},
	{"5dc190", "99ede6", "3dca84"},
	{"4ab9a3", "86d9ea", "16a388"},
	{"4593a5", "7fc7e8", "0c8588"},
	{"5efdf7", "9adfed", "3fffff"},
	{"ff5dcc", "dd8fd0", "ff3dff"},
	{"fdfe89", "eef5c8", "ffff83"},
}

func main() {
	for idx, _ := range input {
		//noon := h2c(line[0]).(color.NRGBA)
		//morn := h2c(line[1]).(color.NRGBA)
		//night := h2c(line[2]).(color.NRGBA)
		//fmt.Printf("c%d := [3]vec3{%s, %s, %s}\n", idx, vec4(noon), vec4(morn), vec4(night))

		fmt.Printf("} else if c == c%d[0] { if Dt < 0.5 { return mix(c%d[0], c%d[1], t) } else { return mix(c%d[0], c%d[2], t) }\n", idx, idx, idx, idx, idx)
	}
}

func vec4(c color.NRGBA) string {
	return fmt.Sprintf("vec3(float(%d)/255, float(%d)/255, float(%d)/255)",
		c.R,
		c.B,
		c.G)
}

// dT 0 - 0.1   -> 0 - 1
// dT 0.1 - 0.9 -> 1
// dt 0.9 - 1.0 -> 1 - 0

func h2c(hex string) color.Color {
	c, _ := colorutil.HexToColor(hex)
	return c
}
