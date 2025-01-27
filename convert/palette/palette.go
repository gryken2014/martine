package palette

import (
	"image/color"

	"github.com/jeromelesaux/martine/gfx/errors"
)

// PalettePosition returns the position of the color c in the palette
// overwise ErrorColorNotFound error
func PalettePosition(c color.Color, p color.Palette) (int, error) {
	r, g, b, a := c.RGBA()
	for index, cp := range p {
		//fmt.Fprintf(os.Stdout,"index(%d), c:%v,cp:%v\n",index,c,cp)
		rp, gp, bp, ap := cp.RGBA()
		if r == rp && g == gp && b == bp && a == ap {
			//fmt.Fprintf(os.Stdout,"Position found")
			return index, nil
		}
	}
	return -1, errors.ErrorColorNotFound
}
