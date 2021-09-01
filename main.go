package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeromelesaux/martine/common"
	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/export/file"
	"github.com/jeromelesaux/martine/export/net"
	"github.com/jeromelesaux/martine/gfx"
	"github.com/jeromelesaux/martine/gfx/animate"
	cgfx "github.com/jeromelesaux/martine/gfx/common"
	"github.com/jeromelesaux/martine/gfx/effect"
	"github.com/jeromelesaux/martine/gfx/errors"
	"github.com/jeromelesaux/martine/gfx/filter"
	"github.com/jeromelesaux/martine/gfx/transformation"
)

type stringSlice []string

func (f *stringSlice) String() string {
	return ""
}

func (f *stringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var deltaFiles stringSlice
var (
	byteStatement       = flag.String("statement", "", "Byte statement to replace in ascii export (default is db), you can replace or instance by defb or byte")
	picturePath         = flag.String("in", "", "Picture path of the input file.")
	width               = flag.Int("width", -1, "Custom output width in pixels. (Will produce a sprite file .win)")
	height              = flag.Int("height", -1, "Custom output height in pixels. (Will produce a sprite file .win)")
	mode                = flag.Int("mode", -1, "Output mode to use :\n\t0 for mode0\n\t1 for mode1\n\t2 for mode2\n\tand add -f option for overscan export.\n\t")
	output              = flag.String("out", "", "Output directory")
	overscan            = flag.Bool("fullscreen", false, "Overscan mode (default no overscan)")
	resizeAlgorithm     = flag.Int("algo", 1, "Algorithm to resize the image (available : \n\t1: NearestNeighbor (default)\n\t2: CatmullRom\n\t3: Lanczos\n\t4: Linear\n\t5: Box\n\t6: Hermite\n\t7: BSpline\n\t8: Hamming\n\t9: Hann\n\t10: Gaussian\n\t11: Blackman\n\t12: Bartlett\n\t13: Welch\n\t14: Cosine\n\t15: MitchellNetravali\n\t")
	help                = flag.Bool("help", false, "Display help message")
	noAmsdosHeader      = flag.Bool("noheader", false, "No amsdos header for all files (default amsdos header added).")
	plusMode            = flag.Bool("plus", false, "Plus mode (means generate an image for CPC Plus Screen)")
	rollMode            = flag.Bool("roll", false, "Roll mode allow to walk and walk into the input file, associated with rla,rra,sra,sla, keephigh, keeplow, losthigh or lostlow options.")
	iterations          = flag.Int("iter", -1, "Iterations number to walk in roll mode, or number of images to generate in rotation mode.")
	rra                 = flag.Int("rra", -1, "Bit rotation on the right and keep pixels")
	rla                 = flag.Int("rla", -1, "Bit rotation on the left and keep pixels")
	sra                 = flag.Int("sra", -1, "Bit rotation on the right and lost pixels")
	sla                 = flag.Int("sla", -1, "Bit rotation on the left and lost pixels")
	losthigh            = flag.Int("losthigh", -1, "Bit rotation on the top and lost pixels")
	lostlow             = flag.Int("lostlow", -1, "Bit rotation on the bottom and lost pixels")
	keephigh            = flag.Int("keephigh", -1, "Bit rotation on the top and keep pixels")
	keeplow             = flag.Int("keeplow", -1, "Bit rotation on the bottom and keep pixels")
	palettePath         = flag.String("pal", "", "Apply the input palette to the image")
	info                = flag.Bool("info", false, "Return the information of the file, associated with -pal and -win options")
	winPath             = flag.String("win", "", "Filepath of the ocp win file")
	dsk                 = flag.Bool("dsk", false, "Copy files in a new CPC image Dsk.")
	tileMode            = flag.Bool("tile", false, "Tile mode to create multiples sprites from a same image.")
	tileIterationX      = flag.Int("iterx", 1, "Number of tiles on a row in the input image.")
	tileIterationY      = flag.Int("itery", 1, "Number of tiles on a column in the input image.")
	compress            = flag.Int("z", -1, "Compression algorithm : \n\t1: rle (default)\n\t2: rle 16bits\n\t3: Lz4 Classic\n\t4: Lz4 Raw\n\t5: zx0 crunch\n")
	kitPath             = flag.String("kit", "", "Path of the palette Cpc plus Kit file. (Apply the input kit palette on the image)")
	inkPath             = flag.String("ink", "", "Path of the palette Cpc ink file. (Apply the input ink palette on the image)")
	rotateMode          = flag.Bool("rotate", false, "Allow rotation on the input image, the input image must be a square (width equals height)")
	m4Host              = flag.String("host", "", "Set the ip of your M4.")
	m4RemotePath        = flag.String("remotepath", "", "Remote path on your M4 where you want to copy your files.")
	m4Autoexec          = flag.Bool("autoexec", false, "Execute on your remote CPC the screen file or basic file.")
	rotate3dMode        = flag.Bool("rotate3d", false, "Allow 3d rotation on the input image, the input image must be a square (width equals height)")
	rotate3dType        = flag.Int("rotate3dtype", 0, "Rotation type :\n\t1 rotate on X axis\n\t2 rotate on Y axis\n\t3 rotate reverse X axis\n\t4 rotate left to right on Y axis\n\t5 diagonal rotation on X axis\n\t6 diagonal rotation on Y axis\n")
	rotate3dX0          = flag.Int("rotate3dx0", -1, "X0 coordinate to apply in 3d rotation (default width of the image/2)")
	rotate3dY0          = flag.Int("rotate3dy0", -1, "Y0 coordinate to apply in 3d rotation (default height of the image/2)")
	initProcess         = flag.String("initprocess", "", "Create a new empty process file.")
	processFile         = flag.String("processfile", "", "Process file path to apply.")
	deltaMode           = flag.Bool("delta", false, "Delta mode: compute delta between two files (prefixed by the argument -df)\n\t(ex: -delta -df file1.SCR -df file2.SCR -df file3.SCR).\n\t(ex with wildcard: -delta -df file\\?.SCR or -delta file\\*.SCR")
	ditheringAlgo       = flag.Int("dithering", -1, "Dithering algorithm to apply on input image\nAlgorithms available:\n\t0: FloydSteinberg\n\t1: JarvisJudiceNinke\n\t2: Stucki\n\t3: Atkinson\n\t4: Sierra\n\t5: SierraLite\n\t6: Sierra3\n\t7: Bayer2\n\t8: Bayer3\n\t9: Bayer4\n\t10: Bayer8\n")
	ditheringMultiplier = flag.Float64("multiplier", 1.18, "Error dithering multiplier.")
	withQuantization    = flag.Bool("quantization", false, "Use additionnal quantization for dithering.")
	extendedDsk         = flag.Bool("extendeddsk", false, "Export in a Extended DSK 80 tracks, 10 sectors 400 ko per face")
	reverse             = flag.Bool("reverse", false, "Transform .scr (overscan or not) file with palette (pal or kit file) into png file")
	flash               = flag.Bool("flash", false, "generate flash animation with two ocp screens.\n\t(ex: -m 1 -flash -i input.png -o test -dsk)\n\tor\n\t(ex: -m 1 -flash -i input1.scr -pal input1.pal -m2 0 -i2 input2.scr -pal2 input2.pal -o test -dsk )")
	picturePath2        = flag.String("in2", "", "Picture path of the second input file (flash mode)")
	mode2               = flag.Int("mode2", -1, "Output mode to use :\n\t0 for mode0\n\t1 for mode1\n\t2 for mode2\n\tmode of the second input file (flash mode)")
	palettePath2        = flag.String("pal2", "", "Apply the input palette to the second image (flash mode)")
	egx1                = flag.Bool("egx1", false, "Create egx 1 output cpc image overscan (option -f) or classical (mix mode 0 / 1).\n\t(ex before generate two images one in mode 1 et one in mode 0\n\tfor instance : martine -i myimage.jpg -m 0 and martine -i myimage.jpg -m 1\n\t: -egx1 -i 1.SCR -m 0 -pal 1.PAL -i2 2.SCR -o test -m2 1 -dsk)\n\tor\n\t(ex automatic egx from image file : -egx1 -i input.png -m 0 -o test -dsk)")
	egx2                = flag.Bool("egx2", false, "Create egx 2 output cpc image overscan (option -f) or classical (mix mode 1 / 2).\n\t(ex before generate two images one in mode 1 et one in mode 2\n\tfor instance : martine -i myimage.jpg -m 0 and martine -i myimage.jpg -m 1\n\t: -egx2 -i 1.SCR -m 0 -pal 1.PAL -i2 2.SCR -o test -m2 1 -dsk)\n\tor\n\t(ex automatic egx from image file : -egx2 -i input.png -m 0 -o test -dsk)")
	sna                 = flag.Bool("sna", false, "Copy files in a new CPC image Sna.")
	spriteHard          = flag.Bool("spritehard", false, "Generate sprite hard for cpc plus.")
	splitRasters        = flag.Bool("splitrasters", false, "Create Split rastered image. (Will produce Overscan output file and .SPL with split rasters file)")
	scanlineSequence    = flag.String("scanlinesequence", "", "Scanline sequence to apply on sprite. for instance : \n\tmartine -i myimage.jpg -w 4 -h 4 -scanlinesequence 0,2,1,3 \n\twill generate a sprite stored with lines order 0 2 1 and 3.\n")
	maskSprite          = flag.String("mask", "", "Mask to apply on each bit of the sprite (to apply an and operation on each pixel with the value #AA [in hexdecimal: #AA or 0xAA, in decimal: 170] ex: martine -i myimage.png -w 40 -h 80 -mask #AA -m 0 -maskand)")
	maskOrOperation     = flag.Bool("maskor", false, "Will apply an OR operation on each byte with the mask")
	maskAdOperation     = flag.Bool("maskand", false, "Will apply an AND operation on each byte with the mask")
	zigzag              = flag.Bool("zigzag", false, "generate data in zigzag order (inc first line and dec next line for tiles)")
	tileMap             = flag.Bool("tilemap", false, "Analyse the input image and generate the tiles, the tile map and global schema.")
	initialAddress      = flag.String("address", "0xC000", "Starting address to display sprite in delta packing")
	doAnimation         = flag.Bool("animate", false, "Will produce an full screen with all sprite on the same image (add -i image.gif or -i *.png)")
	reducer             = flag.Int("reducer", -1, "Reducer mask will reduce original image colors. Available : \n\t1 : lower\n\t2 : medium\n\t3 : strong\n")
	jsonOutput          = flag.Bool("json", false, "Generate json format output.")
	txtOutput           = flag.Bool("txt", false, "Generate text format output.")
	oneLine             = flag.Bool("oneline", false, "Display every other line.")
	oneRow              = flag.Bool("onerow", false, "Display  every other row.")
	impCatcher          = flag.Bool("imp", false, "Will generate sprites as IMP-Catcher format (Impdraw V2).")
	inkSwap             = flag.String("inkswap", "", "Swap ink:\n\tfor instance mode 4 (4 inks) : 0=3,1=0,2=1,3=2\n\twill swap in output image index 0 by 3 and 1 by 0 and so on.")
	lineWidth           = flag.String("linewidth", "#50", "Line width in hexadecimal to compute the screen address in delta mode.")
	deltaPacking        = flag.Bool("deltapacking", false, "Will generate all the animation code from the followed gif file.")
	filloutGif          = flag.Bool("fillout", false, "Fill out the gif frames needed some case with deltapacking")
	saturationPal       = flag.Int("contrast", 0, "apply contrast on the color of the palette on amstrad plus screen. (max value 100 and only on CPC PLUS).")
	brightnessPal       = flag.Int("brightness", 0, "apply brightness on the color of the palette on amstrad plus screen. (max value 100 and only on CPC PLUS).")
	appVersion          = "0.29"
	version             = flag.Bool("version", false, "print martine's version")
)

func usage() {
	fmt.Fprintf(os.Stdout, "martine convert (jpeg, png format) image to Amstrad cpc screen (even overscan)\n")
	fmt.Fprintf(os.Stdout, "By Impact Sid (Version:%s)\n", appVersion)
	fmt.Fprintf(os.Stdout, "Special thanks to @Ast (for his support), @Siko and @Tronic for ideas\n")
	fmt.Fprintf(os.Stdout, "usage :\n\n")
	flag.PrintDefaults()
	os.Exit(-1)
}

func printVersion() {
	fmt.Fprintf(os.Stdout, "%s\n", appVersion)
	os.Exit(-1)
}

/*
@Todo : add zigzag on sprite and sprite hard.
*/

func main() {

	var filename, extension string
	var screenMode uint8
	var in image.Image

	flag.Var(&deltaFiles, "df", "scr file path to add in delta mode comparison. (wildcard accepted such as ? or * file filename.) ")

	flag.Parse()
	if len(flag.Args()) > 0 {
		firstArg := flag.Args()[0]
		if firstArg[0] != '-' {
			flag.Set("i", firstArg)
			for i := 1; i < len(flag.Args()); i += 2 {
				name := strings.Replace(flag.Arg(i), "-", "", 1)
				var value string
				if len(flag.Args()) > i+1 {
					if flag.Arg(i + 1)[0] == '-' {
						value = "true"
						i--
					} else {
						value = flag.Arg(i + 1)
					}
				} else {
					value = "true"
				}
				flag.Set(name, value)
			}
			flag.Parse()
		}
	}
	if *help {
		usage()
	}
	if *version {
		printVersion()
	}

	if *initProcess != "" {
		_, err := InitProcess(*initProcess)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while creating (%s) process file error :%v\n", *initProcess, err)
			os.Exit(-1)
		}
		os.Exit(0)
	}

	if *processFile != "" {
		proc, err := LoadProcessFile(*processFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while loading (%s) process file error :%v\n", *initProcess, err)
			os.Exit(-1)
		}
		proc.Apply()
		if proc.PicturePath == "" && !proc.Delta {
			err = proc.GenerateRawFile()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while loading (%s) process file error :%v\n", *initProcess, err)
				os.Exit(-1)
			}
		}
	}

	if *info {
		if *palettePath != "" {
			file.PalInformation(*palettePath)
		}
		if *winPath != "" {
			file.WinInformation(*winPath)
		}
		if *kitPath != "" {
			file.KitInformation(*kitPath)
		}
		if *inkPath != "" {
			file.InkInformation(*inkPath)
		}
		os.Exit(0)
	}

	// picture path to convert
	if *picturePath == "" && !*deltaMode {
		fmt.Fprintf(os.Stderr, "No picture to compute (option -picturepath or -delta)\n")
		usage()
	}
	filename = filepath.Base(*picturePath)
	extension = filepath.Ext(*picturePath)

	// output directory to store results
	if *output != "" {
		if err := common.CheckOutput(*output); err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting directory informations :%v, Quiting\n", err)
			os.Exit(-2)
		}
	} else {
		*output = "./"
	}

	if *mode == -1 && !*deltaMode && !*reverse {
		fmt.Fprintf(os.Stderr, "No output mode defined can not choose. Quiting\n")
		usage()
	}

	exportType, size := ExportHandler()
	screenMode = uint8(*mode)

	if *byteStatement != "" {
		file.ByteToken = *byteStatement
	}

	if *deltaPacking {
		screenAddress, err := common.ParseHexadecimal16(*initialAddress)
		exportType.Size = size
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while parsing (%s) use the starting address #C000, err : %v\n", *initialAddress, err)
			screenAddress = 0xC000
		}
		if err := animate.DeltaPacking(exportType.InputPath, exportType, screenAddress, screenMode); err != nil {
			fmt.Fprintf(os.Stderr, "Error while deltapacking error: %v\n", err)
		}
		os.Exit(0)
	}

	if !*reverse {
		fmt.Fprintf(os.Stdout, "Informations :\n%s", size.ToString())
	}
	if !*impCatcher && !exportType.DeltaMode && !*reverse && !*doAnimation && strings.ToUpper(extension) != ".SCR" {
		f, err := os.Open(*picturePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while opening file %s, error %v\n", *picturePath, err)
			os.Exit(-2)
		}
		defer f.Close()
		in, _, err = image.Decode(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot decode the image %s error %v", *picturePath, err)
			os.Exit(-2)
		}
	}

	// gestion de la taille de l'image en sortie
	if !exportType.CustomDimension && *rotateMode && !exportType.SpriteHard {
		size.Width = in.Bounds().Max.X
		size.Height = in.Bounds().Max.Y
	}
	if *spriteHard {
		size.Width = 16
		size.Height = 16
	}
	exportType.Size = size

	if !*deltaMode {
		fmt.Fprintf(os.Stdout, "Filename :%s, extension:%s\n", filename, extension)
	}

	if *ditheringAlgo != -1 {
		switch *ditheringAlgo {
		case 0:
			exportType.DitheringMatrix = filter.FloydSteinberg
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:FloydSteinberg, Type:ErrorDiffusionDither\n")
		case 1:
			exportType.DitheringMatrix = filter.JarvisJudiceNinke
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:JarvisJudiceNinke, Type:ErrorDiffusionDither\n")
		case 2:
			exportType.DitheringMatrix = filter.Stucki
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:Stucki, Type:ErrorDiffusionDither\n")
		case 3:
			exportType.DitheringMatrix = filter.Atkinson
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:Atkinson, Type:ErrorDiffusionDither\n")
		case 4:
			exportType.DitheringMatrix = filter.Sierra
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:Sierra, Type:ErrorDiffusionDither\n")
		case 5:
			exportType.DitheringMatrix = filter.SierraLite
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:SierraLite, Type:ErrorDiffusionDither\n")
		case 6:
			exportType.DitheringMatrix = filter.Sierra3
			exportType.DitheringType = constants.ErrorDiffusionDither
			fmt.Fprintf(os.Stdout, "Dither:Sierra3, Type:ErrorDiffusionDither\n")
		case 7:
			exportType.DitheringMatrix = filter.Bayer2
			exportType.DitheringType = constants.OrderedDither
			fmt.Fprintf(os.Stdout, "Dither:Bayer2, Type:OrderedDither\n")
		case 8:
			exportType.DitheringMatrix = filter.Bayer3
			exportType.DitheringType = constants.OrderedDither
			fmt.Fprintf(os.Stdout, "Dither:Bayer3, Type:OrderedDither\n")
		case 9:
			exportType.DitheringMatrix = filter.Bayer4
			exportType.DitheringType = constants.OrderedDither
			fmt.Fprintf(os.Stdout, "Dither:Bayer4, Type:OrderedDither\n")
		case 10:
			exportType.DitheringMatrix = filter.Bayer8
			exportType.DitheringType = constants.OrderedDither
			fmt.Fprintf(os.Stdout, "Dither:Bayer8, Type:OrderedDither\n")
		default:
			fmt.Fprintf(os.Stderr, "Dithering matrix not available.")
			os.Exit(-1)
		}
	}
	if *impCatcher {
		if !exportType.CustomDimension {
			fmt.Fprintf(os.Stderr, "You must set custom width and height.")
			os.Exit(-1)
		}
		sprites := make([]byte, 0)
		fmt.Fprintf(os.Stdout, "[%s]\n", *picturePath)
		spritesPaths, err := common.WilcardedFiles([]string{*picturePath})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while getting wildcard files %s error : %v\n", *picturePath, err)
		}
		for _, v := range spritesPaths {
			f, err := os.Open(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while opening file %s, error %v\n", *picturePath, err)
				os.Exit(-2)
			}
			defer f.Close()
			in, _, err = image.Decode(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot decode the image %s error %v", *picturePath, err)
				os.Exit(-2)
			}
			gfx.ApplyOneImageAndExport(in,
				exportType,
				filepath.Base(v),
				v,
				*mode,
				screenMode)

			spritePath := exportType.AmsdosFullPath(v, ".WIN")
			data, err := file.RawWin(spritePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while extracting raw content, err:%s\n", err)
			}
			sprites = append(sprites, data...)
		}
		finalFile := strings.ReplaceAll(filename, "?", "")
		if err = file.Imp(sprites, uint(len(spritesPaths)), uint(exportType.Size.Width), uint(exportType.Size.Height), uint(screenMode), finalFile, exportType); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot export to Imp-Catcher the image %s error %v", *picturePath, err)
		}
		os.Exit(0)
	} else if *reverse {

		outpath := filepath.Join(*output, strings.Replace(strings.ToLower(filename), ".scr", ".png", 1))
		if exportType.Overscan {
			p, mode, err := file.OverscanPalette(*picturePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot get the palette from file (%s) error %v\n", *picturePath, err)
				os.Exit(-1)
			}

			if err := cgfx.OverscanToPng(*picturePath, outpath, mode, p); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot convert to PNG file (%s) error %v\n", *picturePath, err)
				os.Exit(-1)
			}
			os.Exit(1)
		}
		if *mode == -1 {
			fmt.Fprintf(os.Stderr, "Mode is mandatory to convert to PNG")
			os.Exit(-1)
		}
		var p color.Palette
		var err error
		if *palettePath != "" && !*plusMode {
			p, _, err = file.OpenPal(*palettePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot open palette file (%s) error %v\n", *palettePath, err)
				os.Exit(-1)
			}
		} else {
			if *kitPath != "" && *plusMode {
				p, _, err = file.OpenKit(*kitPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot open kit file (%s) error %v\n", *kitPath, err)
					os.Exit(-1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "For screen or window image, pal or kit file palette is mandatory. (kit file must be associated with -p option)\n")
				os.Exit(-1)
			}
		}
		switch strings.ToUpper(filepath.Ext(filename)) {
		case ".WIN":
			if err := cgfx.SpriteToPng(*picturePath, outpath, uint8(*mode), p); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot convert to PNG file (%s) error %v\n", *picturePath, err)
				os.Exit(-1)
			}
		case ".SCR":
			if err := cgfx.ScrToPng(*picturePath, outpath, uint8(*mode), p); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot convert to PNG file (%s) error %v\n", *picturePath, err)
				os.Exit(-1)
			}
		}
		os.Exit(1)
	}
	if exportType.Animate {
		if !exportType.CustomDimension {
			fmt.Fprintf(os.Stderr, "You must set sprite dimensions with option -w and -h (mandatory)\n")
			os.Exit(-1)
		}
		fmt.Fprintf(os.Stdout, "animation output.\n")
		files := []string{*picturePath}
		files, err := common.WilcardedFiles(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot parse wildcard in argument (%s) error %v\n", *picturePath, err)
			os.Exit(-1)
		}
		if err := animate.Animation(files, screenMode, exportType); err != nil {
			fmt.Fprintf(os.Stderr, "Error while proceeding to animate export error : %v\n", err)
			os.Exit(-1)
		}
	} else {
		if exportType.DeltaMode {
			fmt.Fprintf(os.Stdout, "delta files to proceed.\n")
			for i, v := range deltaFiles {
				fmt.Fprintf(os.Stdout, "[%d]:%s\n", i, v)
			}
			screenAddress, err := common.ParseHexadecimal16(*initialAddress)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while parsing (%s) use the starting address #C000, err : %v\n", *initialAddress, err)
				screenAddress = 0xC000
			}
			if *mode == -1 {
				fmt.Fprintf(os.Stderr, "You must set the mode for this feature. (option -m)\n")
				os.Exit(-1)
			}
			if err := transformation.ProceedDelta(deltaFiles, screenAddress, exportType, uint8(*mode)); err != nil {
				fmt.Fprintf(os.Stderr, "error while proceeding delta mode %v\n", err)
				os.Exit(-1)
			}
		} else {
			if *tileMap {
				/*
					8x8 : 40x25
					16x8 : 20x25
					16x16 : 20x24
				*/
				if exportType.Size.Width != 4 && exportType.Size.Width != 8 {
					fmt.Fprintf(os.Stderr, "Width accepted 4 or 8 pixels")
					os.Exit(-1)
				}
				if exportType.Size.Height != 16 && exportType.Size.Height != 8 {
					fmt.Fprintf(os.Stderr, "Height accepted 16 or 8 pixels")
					os.Exit(-1)
				}
				nbTileLarge := 20
				nbTileHigh := 25
				maxTiles := 255
				switch exportType.Size.Width {
				case 8:
					nbTileLarge = 20
					if exportType.Size.Height == 16 {
						maxTiles = 240
					}
				case 4:
					nbTileLarge = 40
				}

				if !exportType.CustomDimension {
					fmt.Fprintf(os.Stderr, "You must set height and width to define the tile dimensions (options -h and -w)\n")
					os.Exit(-1)
				}
				analyze := transformation.AnalyzeTilesBoard(in, exportType.Size)
				if err := analyze.SaveSchema(filepath.Join(exportType.OutputPath, "tilesmap_schema.png")); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot save tilemap schema error :%v\n", err)
					os.Exit(-1)
				}
				if err := analyze.SaveTilemap(filepath.Join(exportType.OutputPath, "tilesmap.map")); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot save tilemap csv file error :%v\n", err)
					os.Exit(-1)
				}

				// applyOneImage
				// sort tiles
				// check < 256 tiles
				// finally export
				// 20 tiles large 25 tiles height
				tiles := analyze.Sort()
				/*	for i, v := range tiles {
					if v.Occurence > 0 {
						tile := v.Tile.Image()
						tileFilepath := filepath.Join(exportType.OutputPath, fmt.Sprintf("%.2d.png", i))
						f, err := os.Create(tileFilepath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Cannot create tiles %.2d error %v\n", i, err)
							os.Exit(-1)
						}
						defer f.Close()
						if err := png.Encode(f, tile); err != nil {
							fmt.Fprintf(os.Stderr, "Cannot encode in png tile %.2d error %v\n", i, err)
							os.Exit(-1)
						}

						var palette color.Palette
						if exportType.CpcPlus {
							palette = constants.CpcPlusPalette
						} else {
							palette = constants.CpcOldPalette
						}

						out, _ := gfx.DoDithering(tile, palette, exportType)
						palette, out, err = convert.DowngradingPalette(out, exportType.Size, exportType.CpcPlus)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Cannot downgrade colors palette for this image %s\n", tileFilepath)
						}

						palette = constants.SortColorsByDistance(palette)

						fmt.Fprintf(os.Stdout, "Saving downgraded image into (%s)\n", filename+"_down.png")
						if err := file.Png(tileFilepath+"_down.png", out); err != nil {
							os.Exit(-2)
						}
						if err := cgfx.ToSpriteAndExport(tile, palette, exportType.Size, screenMode, tileFilepath, false, exportType); err != nil {
							fmt.Fprintf(os.Stderr, "Cannot create tile from image %s, error :%v\n", tileFilepath, err)
						}
					}
				}*/
				data := make([]byte, 0)

				palette := analyze.Palette()
				finalFile := strings.ReplaceAll(filename, "?", "")
				if err := file.Kit(finalFile, palette, screenMode, false, exportType); err != nil {
					fmt.Fprintf(os.Stderr, "Error while saving file %s error :%v", finalFile, err)
					os.Exit(-1)
				}
				nbFrames := 0
				for i, v := range tiles {
					if v.Occurence > 0 {
						tile := v.Tile.Image()
						d, _, _, err := gfx.ApplyOneImage(tile,
							exportType,
							*mode,
							palette,
							screenMode)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error while transforming sprite error : %v\n", err)
						}
						data = append(data, d...)
						scenePath := filepath.Join(exportType.OutputPath, fmt.Sprintf("tile-%.2d.png", i))
						f, err := os.Create(scenePath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Cannot create scene tile-%.2d error %v\n", i, err)
							os.Exit(-1)
						}

						if err := png.Encode(f, tile); err != nil {
							fmt.Fprintf(os.Stderr, "Cannot encode in png scene tile-%.2d error %v\n", i, err)
							os.Exit(-1)
						}
						f.Close()
						nbFrames++
						if i >= maxTiles {
							fmt.Fprintf(os.Stderr, "Maximum of %d tiles accepted, skipping...\n", maxTiles)
							break
						}
					}
				}
				// save the file sprites
				finalFile = strings.ReplaceAll(filename, "?", "")
				if err := file.Imp(data, uint(nbFrames), uint(analyze.TileSize.Width), uint(analyze.TileSize.Height), uint(screenMode), finalFile, exportType); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot export to Imp-Catcher the image %s error %v", *picturePath, err)
				}

				// save the tilemap
				maps := make([]*image.RGBA, 0)
				index := 0
				for y := 0; y < in.Bounds().Max.Y; y += (nbTileHigh * analyze.TileSize.Height) {
					for x := 0; x < in.Bounds().Max.X; x += (nbTileLarge * analyze.TileSize.Width) {
						m := image.NewRGBA(image.Rect(0, 0, nbTileLarge*analyze.TileSize.Width, nbTileHigh*analyze.TileSize.Height))
						// copy of the map
						for i := 0; i < nbTileLarge*analyze.TileSize.Width; i++ {
							for j := 0; j < nbTileHigh*analyze.TileSize.Height; j++ {
								var c color.Color = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
								if x+i < in.Bounds().Max.X && y+j < in.Bounds().Max.Y {
									c = in.At(x+i, y+j)
								}
								m.Set(i, j, c)
							}
						}
						// store the map in the slice
						maps = append(maps, m)
						scenePath := filepath.Join(exportType.OutputPath, fmt.Sprintf("scene-%.2d.png", index))
						f, err := os.Create(scenePath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Cannot create scene scence-%.2d error %v\n", index, err)
							os.Exit(-1)
						}

						if err := png.Encode(f, m); err != nil {
							fmt.Fprintf(os.Stderr, "Cannot encode in png scene scene-%.2d error %v\n", index, err)
							os.Exit(-1)
						}
						f.Close()
						index++
					}
				}

				// now thread all maps images
				tileMaps := make([]byte, 0)
				for _, v := range maps {
					for y := 0; y < v.Bounds().Max.Y; y += analyze.TileSize.Height {
						for x := 0; x < v.Bounds().Max.X; x += analyze.TileSize.Width {
							sprt, err := transformation.ExtractTile(v, analyze.TileSize, x, y)
							if err != nil {
								fmt.Fprintf(os.Stderr, "Error while extracting tile size(%d,%d) at position (%d,%d) error :%v\n", size.Width, size.Height, x, y, err)
								break
							}
							index := analyze.TileIndex(sprt, tiles)
							tileMaps = append(tileMaps, byte(index))
						}
					}
				}

				if err := file.TileMap(tileMaps, finalFile, exportType); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot export to Imp-TileMap the image %s error %v", *picturePath, err)
				}
				os.Exit(0)
			} else {
				if exportType.TileMode {
					if exportType.TileIterationX == -1 || exportType.TileIterationY == -1 {
						fmt.Fprintf(os.Stderr, "missing arguments iterx and itery to use with tile mode.\n")
						usage()
						os.Exit(-1)
					}
					err := transformation.TileMode(exportType, uint8(*mode), exportType.TileIterationX, exportType.TileIterationY)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Tile mode on error : error :%v\n", err)
						os.Exit(-1)
					}
				} else {
					if *flash {
						if err := effect.Flash(*picturePath, *picturePath2,
							*palettePath, *palettePath2,
							*mode,
							*mode2,
							exportType); err != nil {
							fmt.Fprintf(os.Stderr, "Error while applying on one image :%v\n", err)
							os.Exit(-1)
						}
					} else {
						var p color.Palette
						var err error
						if exportType.CpcPlus {
							if *kitPath != "" {
								p, _, err = file.OpenKit(*kitPath)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Error while reading kit file (%s) :%v\n", *kitPath, err)
									os.Exit(-1)
								}
							}
						} else {
							if *palettePath != "" {
								p, _, err = file.OpenPal(*palettePath)
								if err != nil {
									fmt.Fprintf(os.Stderr, "Error while reading palette file (%s) :%v\n", *palettePath, err)
									os.Exit(-1)
								}
							}
						}

						if exportType.EgxFormat > 0 {
							if len(p) == 0 {
								fmt.Fprintf(os.Stderr, "Now colors found in palette, give up treatment.\n")
								os.Exit(-1)
							}
							if err := effect.Egx(*picturePath, *picturePath2,
								p,
								*mode,
								*mode2,
								exportType); err != nil {
								fmt.Fprintf(os.Stderr, "Error while applying on one image :%v\n", err)
								os.Exit(-1)
							}
						} else {
							if exportType.SplitRaster {
								if exportType.Overscan {
									if err := effect.DoSpliteRaster(in, screenMode, filename, exportType); err != nil {
										fmt.Fprintf(os.Stderr, "Error while applying splitraster on one image :%v\n", err)
										os.Exit(-1)
									}
								} else {
									fmt.Fprintf(os.Stderr, "Only overscan mode implemented for this feature, %v", errors.ErrorNotYetImplemented)
								}
							} else {
								if strings.ToUpper(extension) != ".SCR" {
									if err := gfx.ApplyOneImageAndExport(in,
										exportType,
										filename, *picturePath,
										*mode,
										screenMode); err != nil {
										fmt.Fprintf(os.Stderr, "Error while applying on one image :%v\n", err)
										os.Exit(-1)
									}
								} else {
									fmt.Fprintf(os.Stderr, "Error while applying on one image : SCR format not used for this treatment\n")
									os.Exit(-1)
								}
							}
						}
					}
				}
			}
		}
	}
	// export into bundle DSK or SNA
	if exportType.Dsk {
		if err := file.ImportInDsk(*picturePath, exportType); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot create or write into dsk file error :%v\n", err)
		}
	}
	if exportType.Sna {
		if exportType.Overscan {
			var gfxFile string
			for _, v := range exportType.DskFiles {
				if filepath.Ext(v) == ".SCR" {
					gfxFile = v
					break
				}
			}
			exportType.SnaPath = filepath.Join(*output, "test.sna")
			if err := file.ImportInSna(gfxFile, exportType.SnaPath, screenMode); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot create or write into sna file error :%v\n", err)
			}
			fmt.Fprintf(os.Stdout, "Sna saved in file %s\n", exportType.SnaPath)
		} else {
			fmt.Fprintf(os.Stderr, "Feature not implemented for this file.")
			os.Exit(-1)
		}
	}
	if exportType.M4 {
		if err := net.ImportInM4(exportType); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot send to M4 error :%v\n", err)
		}
	}
	os.Exit(0)
}
