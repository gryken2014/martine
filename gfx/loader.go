package gfx

import (
	"encoding/binary"
	"fmt"
	"github.com/jeromelesaux/m4client/cpc"
	"github.com/jeromelesaux/martine/constants"
	"image/color"
	"os"
)

// CPC plus loader nb colors *2 offset 0x1d

var (
	BasicLoader = []byte{
		0x36, 0x00, 0x05, 0x00, 0x8c, 0x20, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30,
		0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30,
		0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c, 0x30, 0x30, 0x2c,
		0x30, 0x30, 0x2c, 0x30, 0x30, 0x00, 0x0e, 0x00, 0x0a, 0x00, 0xaa, 0x20, 0x1c, 0x00, 0x40, 0x20,
		0xf5, 0x20, 0x0f, 0x00, 0x18, 0x00, 0x14, 0x00, 0xa8, 0x22, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
		0x20, 0x20, 0x2e, 0x70, 0x61, 0x6c, 0x22, 0x2c, 0x1c, 0x00, 0x40, 0x00, 0x0e, 0x00, 0x1e, 0x00,
		0xad, 0x20, 0xff, 0x12, 0x28, 0x1c, 0x00, 0x40, 0x29, 0x00, 0x13, 0x00, 0x28, 0x00, 0x9e, 0x20,
		0x0d, 0x00, 0x00, 0xf0, 0xef, 0x0e, 0x20, 0xec, 0x20, 0x19, 0x0f, 0x20, 0x00, 0x0b, 0x00, 0x32,
		0x00, 0xc3, 0x20, 0x0d, 0x00, 0x00, 0xe3, 0x00, 0x10, 0x00, 0x46, 0x00, 0xa2, 0x20, 0x0d, 0x00,
		0x00, 0xf0, 0x2c, 0x0d, 0x00, 0x00, 0xe3, 0x00, 0x07, 0x00, 0x50, 0x00, 0xb0, 0x20, 0x00, 0x18,
		0x00, 0x5a, 0x00, 0xa8, 0x22, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x2e, 0x73, 0x63,
		0x72, 0x22, 0x2c, 0x1c, 0x00, 0xc0, 0x00, 0x00, 0x00, 0x0a}
	startPaletteValues   = 6
	startPaletteName     = 58 + 16
	startScreenName      = 149 + 16
	PaletteCPCPlusLoader = []byte{
		0xf3, 0x01, 0x00, 0xbc, 0x21, 0x2d, 0x30, 0x1e, 0x11, 0x7e, 0xed, 0x79, 0x23, 0x1d, 0x20, 0xf9,
		0xfb, 0x01, 0xb8, 0x7f, 0xed, 0x49, 0x21, 0x3e, 0x30, 0x11, 0x00, 0x64, 0x01, 0x20, 0x00, 0xed,
		0xb0, 0x21, 0xf9, 0xb7, 0xc3, 0xdd, 0xbc, 0x01, 0xa0, 0x7f, 0xed, 0x49, 0xc9, 0xff, 0x00, 0xff,
		0x77, 0xb3, 0x51, 0xa8, 0xd4, 0x62, 0x39, 0x9c, 0x46, 0x2b, 0x15, 0x8a, 0xcd, 0xee, 0x66, 0x06,
		0x63, 0x06, 0x00, 0x00, 0x96, 0x06, 0x33, 0x03, 0x63, 0x03, 0x93, 0x06, 0x96, 0x06, 0x96, 0x09,
		0xc9, 0x0c, 0x63, 0x06, 0x96, 0x06, 0xc6, 0x09, 0xc9, 0x09, 0x63, 0x03, 0x99, 0x09}
	// offset file name 24
	BasicCPCPlusLoader = []byte{
		0x1a, 0x38, 0x00, 0x0a, 0x00, 0xaa, 0x20, 0x1c, 0xff, 0x2f, 0x01, 0xad, 0x20, 0x0e, 0x01, 0xa8,
		0x22, 0x70, 0x61, 0x6c, 0x70, 0x6c, 0x75, 0x73, 0x2e, 0x62, 0x69, 0x6e, 0x22, 0x2c, 0x1c, 0x00,
		0x30, 0x01, 0xa8, 0x22, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x2e, 0x73, 0x63, 0x72, 
		0x22, 0x2c, 0x1c, 0x00, 0xc0, 0x01, 0x83, 0x20, 0x1c, 0x00, 0x30, 0x00, 0x00, 0x00, 0x1a}
)

func Loader(filePath string, p color.Palette, exportType *ExportType) error {
	var out string
	for i := 0; i < len(p); i++ {
		v, err := constants.FirmwareNumber(p[i])
		if err == nil {
			out += fmt.Sprintf("%0.2d", v)
		} else {
			fmt.Fprintf(os.Stderr, "Error while getting the hardware values for color %v, error :%v\n", p[0], err)
		}
		if i+1 < len(p) {
			out += ","
		}

	}

	var loader []byte
	loader = BasicLoader
	copy(loader[startPaletteValues:], out[0:len(out)])
	filename := exportType.AmsdosFilename()
	copy(loader[startPaletteName:], filename[:])
	copy(loader[startScreenName:], filename[:])
	fmt.Println(loader)
	header := cpc.CpcHead{Type: 0, User: 0, Address: 0x170, Exec: 0x0,
		Size:        uint16(binary.Size(loader)),
		Size2:       uint16(binary.Size(loader)),
		LogicalSize: uint16(binary.Size(loader))}
	file := string(filename) + ".BAS"
	copy(header.Filename[:], file)
	header.Checksum = uint16(header.ComputedChecksum16())
	osFilepath := exportType.AmsdosFullPath(filePath, ".BAS")
	fw, err := os.Create(osFilepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating file (%s) error :%s\n", osFilepath, err)
		return err
	}
	if !exportType.NoAmsdosHeader {
		binary.Write(fw, binary.LittleEndian, header)
	}
	binary.Write(fw, binary.LittleEndian, loader)
	fw.Close()

	exportType.AddFile(osFilepath)
	return nil
}
