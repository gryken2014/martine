package sprite

import (
	"image"
	"image/color"

	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/export"
	"github.com/jeromelesaux/martine/gfx"
)

func SplitBoardToSprite(
	im image.Image,
	p color.Palette,
	nbSpritePerRow, nbSpritePerColmun int,
	mode uint8,
	isSpriteHard bool) ([][]byte, []*image.NRGBA, error) {

	results := make([][]byte, 0)
	spriteWidth := im.Bounds().Max.X / nbSpritePerColmun
	spriteHeight := im.Bounds().Max.Y / nbSpritePerRow
	sprites := make([]*image.NRGBA, 0)
	for x := 0; x < im.Bounds().Max.X; x += spriteWidth {
		for y := 0; y < im.Bounds().Max.Y; y += spriteHeight {
			i := image.NewNRGBA(image.Rect(0, 0, spriteWidth, spriteHeight))
			for x0 := 0; x0 < spriteWidth; x0++ {
				for y0 := 0; y0 < spriteHeight; y0++ {
					i.Set(x0, y0, im.At(x+x0, y+y0))
				}
			}
			sprites = append(sprites, i)
		}
	}
	cont := export.NewMartineContext("", "")
	rawSprites := make([]*image.NRGBA, 0)
	cont.Size = constants.Size{Width: spriteWidth, Height: spriteHeight}
	for _, v := range sprites {
		r, sp, _, _, err := gfx.ApplyOneImage(v, cont, int(mode), p, mode)
		if err != nil {
			return results, sprites, err
		}

		results = append(results, r)
		rawSprites = append(rawSprites, sp)
	}
	return results, rawSprites, nil
}
