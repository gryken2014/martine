package overscan

import (
	"image/color"

	"github.com/jeromelesaux/martine/export"
	"github.com/jeromelesaux/martine/export/amsdos"
)

func SaveGo(filePath string, data []byte, p color.Palette, screenMode uint8, cont *export.MartineContext) error {
	data1 := make([]byte, 0x4000)
	data2 := make([]byte, 0x4000)
	copy(data1, data[0x0:0x4040])
	copy(data2, data[0x4040:0x8000])
	go1Filename := cont.AmsdosFullPath(filePath, ".GO1")
	go2Filename := cont.AmsdosFullPath(filePath, ".GO2")
	if !cont.NoAmsdosHeader {
		if err := amsdos.SaveAmsdosFile(go1Filename, ".GO1", data1, 0, 0, 0x20, 0); err != nil {
			return err
		}
		if err := amsdos.SaveAmsdosFile(go2Filename, ".GO2", data2, 0, 0, 0x4000, 0); err != nil {
			return err
		}
	} else {
		if err := amsdos.SaveOSFile(go1Filename, data1); err != nil {
			return err
		}
		if err := amsdos.SaveOSFile(go2Filename, data2); err != nil {
			return err
		}
	}

	cont.AddFile(go1Filename)
	cont.AddFile(go2Filename)

	return nil
}