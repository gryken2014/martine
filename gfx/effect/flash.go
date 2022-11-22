package effect

import (
	"image"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/convert"
	"github.com/jeromelesaux/martine/export"
	"github.com/jeromelesaux/martine/export/ocpartstudio"
	"github.com/jeromelesaux/martine/gfx"
)

func Flash(filepath1, filepath2, palpath1, palpath2 string, m1, m2 int, cont *export.MartineConfig) error {
	if filepath2 == "" && filepath1 != "" {
		filename := filepath.Base(filepath1)
		f, err := os.Open(filepath1)
		if err != nil {
			return err
		}
		defer f.Close()
		in, _, err := image.Decode(f)
		if err != nil {
			return err
		}
		return AutoFlash(in, cont, filename, filepath1, m1, uint8(m1))
	}
	filename1 := filepath.Base(filepath1)
	filename2 := filepath.Base(filepath2)
	p1, _, err := ocpartstudio.OpenPal(palpath1)
	if err != nil {
		return err
	}
	p2, _, err := ocpartstudio.OpenPal(palpath2)
	if err != nil {
		return err
	}
	cont.AddFile(filepath1)
	cont.AddFile(filepath2)
	cont.AddFile(palpath1)
	cont.AddFile(palpath2)
	return ocpartstudio.FlashLoader(filename1, filename2, p1, p2, uint8(m1), uint8(m2), cont)
}

func AutoFlash(in image.Image,
	cont *export.MartineConfig,
	filename, picturePath string,
	mode int,
	screenMode uint8) error {

	var err error
	size := constants.Size{
		Width:  cont.Size.Width * 2,
		Height: cont.Size.Height * 2}
	im := convert.Resize(in, size, cont.ResizingAlgo)
	leftIm := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{cont.Size.Width, cont.Size.Height}})
	rigthIm := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{cont.Size.Width, cont.Size.Height}})
	indexExtFilename := strings.LastIndex(filename, ".")
	indexExtPath := strings.LastIndex(picturePath, ".")

	bFilename := make([]byte, indexExtFilename)
	bPath := make([]byte, indexExtPath)
	formerExt := filename[indexExtFilename:]
	copy(bFilename, filename[0:indexExtFilename])
	copy(bPath, picturePath[0:indexExtPath])

	filenameLeft := string(bFilename) + "1" + formerExt
	filenameRigth := string(bFilename) + "2" + formerExt
	filepathLeft := string(bPath) + "1" + formerExt
	filepathRigth := string(bPath) + "2" + formerExt
	x := 0
	for i := 0; i < size.Width; i += 2 {
		y := 0
		for j := 0; j < size.Height; j += 2 {
			leftIm.Set(x, y, im.At(i, j))
			y++
		}
		x++
		y = 0
	}
	ext := path.Ext(filename)
	name := strings.Replace(filename, ext, "", 1)
	namesize := len(name)
	if namesize > 8 {
		namesize = 7
	}
	flashPaletteFilename1 := strings.ToUpper(name)[0:namesize] + "1.PAL"
	flashPalettePath1 := filepath.Join(cont.OutputPath, flashPaletteFilename1)

	err = gfx.ApplyOneImageAndExport(leftIm,
		cont,
		filenameLeft, filepathLeft,
		mode,
		screenMode)
	if err != nil {
		return err
	}
	p1, _, err := ocpartstudio.OpenPal(flashPalettePath1)
	if err != nil {
		return err
	}

	x = 0
	for i := 1; i < size.Width; i += 2 {
		y := 0
		for j := 1; j < size.Height; j += 2 {
			rigthIm.Set(x, y, im.At(i, j))
			y++
		}
		x++
		y = 0
	}

	flashMode := mode - 1
	if flashMode < 0 {
		flashMode = 2
	}

	switch flashMode {
	case 0:
		cont.Size = constants.Mode0
	case 1:
		cont.Size = constants.Mode1
	case 2:
		cont.Size = constants.Mode2
	}

	flashPaletteFilename2 := strings.ToUpper(name)[0:namesize] + "2.PAL"
	flashPalettePath2 := filepath.Join(cont.OutputPath, flashPaletteFilename2)

	cont.PalettePath = flashPalettePath1

	err = gfx.ApplyOneImageAndExport(rigthIm,
		cont,
		filenameRigth, filepathRigth,
		flashMode,
		uint8(flashMode))
	if err != nil {
		return err
	}
	p2, _, err := ocpartstudio.OpenPal(flashPalettePath2)
	if err != nil {
		return err
	}
	return ocpartstudio.FlashLoader(filenameLeft, filenameRigth, p1, p2, uint8(mode), uint8(flashMode), cont)
}
