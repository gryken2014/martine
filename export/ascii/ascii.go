package ascii

import (
	"fmt"
	"image/color"
	"os"
	"runtime"

	"github.com/jeromelesaux/martine/constants"
	x "github.com/jeromelesaux/martine/export"
	"github.com/jeromelesaux/martine/export/amsdos"
	"github.com/jeromelesaux/martine/export/compression"
)

// ByteToken is the token by default
var ByteToken = "db" // "BYTE"

func Ascii(filePath string, data []byte, p color.Palette, dontImportDsk bool, cont *x.MartineConfig) error {
	eol := "\n"
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}

	var out string

	data, _ = compression.Compress(data, cont.Compression)

	cpcFilename := string(cont.AmsdosFilename()) + ".TXT"
	osFilepath := cont.AmsdosFullPath(filePath, ".TXT")
	fmt.Fprintf(os.Stdout, "Writing ascii file (%s) data length (%d)\n", osFilepath, len(data))
	sizeInfos := fmt.Sprintf("; width %d height %d %s", cont.Size.Width, cont.Size.Height, eol)
	out += "; Screen " + cpcFilename + eol + ".screen:" + eol + sizeInfos
	out += FormatAssemblyDatabyte(data, eol)
	out += "; Palette " + cpcFilename + eol + ".palette:" + eol + ByteToken + " "

	if cont.CpcPlus {
		out += FormatAssemblyCPCPlusPalette(p, eol)
	} else {
		out += FormatAssemblyCPCPalette(p, eol)
		out += eol + "; Basic Palette " + cpcFilename + eol + ".basic_palette:" + eol + ByteToken + " "
		out += FormatAssemblyBasicPalette(p, eol)
		out += eol
	}
	if !cont.NoAmsdosHeader {

		if err := amsdos.SaveAmsdosFile(osFilepath, ".TXT", []byte(out), 0, 0, 0, 0); err != nil {
			return err
		}
	} else {
		if err := amsdos.SaveOSFile(osFilepath, []byte(out)); err != nil {
			return err
		}
	}

	if !dontImportDsk {
		cont.AddFile(osFilepath)
	}
	if cont.Json {
		palette := make([]string, len(p))
		for i := 0; i < len(p); i++ {
			v, err := constants.FirmwareNumber(p[i])
			if err == nil {
				palette[i] = fmt.Sprintf("%.2d", v)
			} else {
				fmt.Fprintf(os.Stderr, "Error while getting the hardware values for color %v, error :%v\n", p[0], err)
			}
		}
		hardwarepalette := make([]string, len(p))
		for i := 0; i < len(p); i++ {
			fcolor, _ := constants.FirmwareNumber(p[i])
			hardwarepalette[i] = fmt.Sprintf("0x%.2x", fcolor)
		}
		screen := make([]string, len(data))
		for i := 0; i < len(data); i++ {
			screen[i] = fmt.Sprintf("0x%.2x", data[i])
		}
		j := x.NewJson(cont.Filename(), cont.Size.Width, cont.Size.Height, screen, palette, hardwarepalette)
		fmt.Fprintf(os.Stdout, "Filepath:%s\n", filePath)
		if cont.TileMode {
			cont.Tiles.Sprites = append(cont.Tiles.Sprites, j)
			return nil
		}
		return j.Save(cont.OsFullPath(filePath, ".json"))
	}
	return nil
}

func AsciiByColumn(filePath string, data []byte, p color.Palette, dontImportDsk bool, mode uint8, cont *x.MartineConfig) error {
	eol := "\n"
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}

	var out string
	var i int
	var jsonData []string

	cpcFilename := string(cont.AmsdosFilename()) + "C.TXT"
	osFilepath := cont.AmsdosFullPath(filePath, "C.TXT")
	fmt.Fprintf(os.Stdout, "Writing ascii file (%s) values by columns data length (%d)\n", osFilepath, len(data))
	sizeInfos := fmt.Sprintf("; width %d height %d %s", cont.Size.Width, cont.Size.Height, eol)
	out += "; Screen by column " + cpcFilename + eol + ".screen:" + eol + sizeInfos
	var adjustMode int
	switch mode {
	case 0:
		adjustMode = 2
	case 1:
		adjustMode = 4
	case 2:
		adjustMode = 8
	}
	pas := cont.Size.Width / adjustMode
	h := 0
	nbValues := 1
	octetsRead := 0
	end := 17
	if (cont.Size.Width + 1) < end {
		end = (cont.Size.Width + 1)
	}
	for {

		if nbValues == 1 {
			out += fmt.Sprintf("%s ", ByteToken)
		}
		out += fmt.Sprintf("#%0.2x", data[i])
		jsonData = append(jsonData, fmt.Sprintf("0x%.2x", data[i]))
		nbValues++

		i += pas
		octetsRead++
		if nbValues < end && octetsRead != len(data) {
			out += " ,"
		}
		if octetsRead == len(data) {
			break
		}

		if i >= len(data) {
			h++
			i = h
		}

		if nbValues == end {
			out += eol
			nbValues = 1
		}
	}
	out += eol
	out += "; Palette " + cpcFilename + eol + ".palette:" + eol + ByteToken + " "

	if cont.CpcPlus {
		out += FormatAssemblyCPCPlusPalette(p, eol)
	} else {
		out += FormatAssemblyCPCPalette(p, eol)
		out += eol + "; Basic Palette " + cpcFilename + eol + ".basic_palette:" + eol + ByteToken + " "
		out += FormatAssemblyBasicPalette(p, eol)
		out += eol
	}

	if !cont.NoAmsdosHeader {
		if err := amsdos.SaveAmsdosFile(osFilepath, ".TXT", []byte(out), 0, 0, 0, 0); err != nil {
			return err
		}
		//binary.Write(fw, binary.LittleEndian, header)
	} else {
		if err := amsdos.SaveOSFile(osFilepath, []byte(out)); err != nil {
			return err
		}
	}

	if !dontImportDsk {
		cont.AddFile(osFilepath)
	}

	if cont.Json {
		palette := make([]string, len(p))
		for i := 0; i < len(p); i++ {
			v, err := constants.FirmwareNumber(p[i])
			if err == nil {
				palette[i] = fmt.Sprintf("%.2d", v)
			} else {
				fmt.Fprintf(os.Stderr, "Error while getting the hardware values for color %v, error :%v\n", p[0], err)
			}
		}
		hardwarepalette := make([]string, len(p))
		for i := 0; i < len(p); i++ {
			fcolor, _ := constants.FirmwareNumber(p[i])
			hardwarepalette[i] = fmt.Sprintf("0x%.2x", fcolor)
		}

		j := x.NewJson(cont.Filename(), cont.Size.Width, cont.Size.Height, jsonData, palette, hardwarepalette)
		fmt.Fprintf(os.Stdout, "Filepath:%s\n", filePath)
		if cont.TileMode {
			cont.Tiles.Sprites = append(cont.Tiles.Sprites, j)
			return nil
		}
		return j.Save(cont.OsFullPath(filePath, "_column.json"))
	}
	return nil
}

func FormatAssemblyString(data []string, eol string) string {
	var out string
	for i := 0; i < len(data); i += 8 {
		out += fmt.Sprintf("%s ", ByteToken)
		if i < len(data) {
			out += data[i]
		}
		if i+1 < len(data) {
			out += fmt.Sprintf(", %s", data[i+1])
		}
		if i+2 < len(data) {
			out += fmt.Sprintf(", %s", data[i+2])
		}
		if i+3 < len(data) {
			out += fmt.Sprintf(", %s", data[i+3])
		}
		if i+4 < len(data) {
			out += fmt.Sprintf(", %s", data[i+4])
		}
		if i+5 < len(data) {
			out += fmt.Sprintf(", %s", data[i+5])
		}
		if i+6 < len(data) {
			out += fmt.Sprintf(", %s", data[i+6])
		}
		if i+7 < len(data) {
			out += fmt.Sprintf(", %s", data[i+7])
		}
		out += eol
	}
	return out
}

func FormatAssemblyDatabyte(data []byte, eol string) string {
	var out string
	for i := 0; i < len(data); i += 8 {
		out += fmt.Sprintf("%s ", ByteToken)
		if i < len(data) {
			out += fmt.Sprintf("#%0.2x", data[i])
		}
		if i+1 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+1])
		}
		if i+2 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+2])
		}
		if i+3 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+3])
		}
		if i+4 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+4])
		}
		if i+5 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+5])
		}
		if i+6 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+6])
		}
		if i+7 < len(data) {
			out += fmt.Sprintf(", #%0.2x", data[i+7])
		}
		out += eol
	}
	return out
}

func FormatAssemblyCPCPalette(p color.Palette, eol string) string {
	var out string
	for i := 0; i < len(p); i++ {
		v, err := constants.HardwareValues(p[i])
		if err == nil {
			out += fmt.Sprintf("#%0.2x", v[0])
			if (i+1)%8 == 0 && i+1 < len(p) {
				out += eol + ByteToken + " "
			} else {
				if i+1 < len(p) {
					out += ", "
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error while getting the hardware values for color %v, error :%v\n", p[0], err)
		}
	}
	return out
}

func FormatAssemblyBasicPalette(p color.Palette, eol string) string {
	var out string
	for i := 0; i < len(p); i++ {
		v, err := constants.FirmwareNumber(p[i])
		if err == nil {
			out += fmt.Sprintf("%0.2d", v)
			if (i+1)%8 == 0 && i+1 < len(p) {
				out += eol + ByteToken + " "
			} else {
				if i+1 < len(p) {
					out += ", "
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error while getting the hardware values for color %v, error :%v\n", p[0], err)
		}
	}
	return out
}

func FormatAssemblyCPCPlusPalette(p color.Palette, eol string) string {
	var out string
	for i := 0; i < len(p); i++ {
		cp := constants.NewCpcPlusColor(p[i])
		v := cp.Value()
		out += fmt.Sprintf("#%.2x, #%.2x", byte(v), byte(v>>8))
		if (i+1)%8 == 0 && i+1 < len(p) {
			out += eol + ByteToken + " "
		} else {
			if i+1 < len(p) {
				out += ", "
			}
		}
	}
	return out
}
