package ui

import (
	"fmt"
	"image/gif"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jeromelesaux/fyne-io/custom_widget"
	"github.com/jeromelesaux/martine/config"
	"github.com/jeromelesaux/martine/export/amsdos"
	impPalette "github.com/jeromelesaux/martine/export/impdraw/palette"
	"github.com/jeromelesaux/martine/export/ocpartstudio"
	"github.com/jeromelesaux/martine/export/png"
	"github.com/jeromelesaux/martine/gfx/animate"
	"github.com/jeromelesaux/martine/ui/martine-ui/menu"
	w2 "github.com/jeromelesaux/martine/ui/martine-ui/widget"
)

func (m *MartineUI) exportAnimationDialog(a *menu.AnimateMenu, w fyne.Window) {
	cont := container.NewVBox(
		container.NewHBox(
			widget.NewButtonWithIcon("Export into folder", theme.DocumentSaveIcon(), func() {
				fo := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, m.window)
						return
					}
					if lu == nil {
						// cancel button
						return
					}
					cfg := m.NewConfig(a.ImageMenu, false)
					if cfg == nil {
						return
					}
					cfg.Compression = m.animateExport.ExportCompression
					if a.ExportVersion == 0 {
						a.ExportVersion = animate.DeltaExportV1
					}
					address, err := strconv.ParseUint(a.InitialAddress.Text, 16, 64)
					if err != nil {
						dialog.ShowError(err, m.window)
						return
					}
					m.animateExport.ExportFolderPath = lu.Path()
					fmt.Println(m.animateExport.ExportFolderPath)
					pi := custom_widget.NewProgressInfinite("Exporting, please wait.", m.window)
					pi.Show()
					code, err := animate.ExportDeltaAnimate(
						a.RawImages[0],
						a.DeltaCollection,
						a.Palette(),
						a.IsSprite,
						cfg,
						uint16(address),
						uint8(a.Mode),
						a.ExportVersion,
					)
					pi.Hide()
					if err != nil {
						dialog.ShowError(err, m.window)
						return
					}
					err = amsdos.SaveOSFile(m.animateExport.ExportFolderPath+string(filepath.Separator)+"code.asm", []byte(code))
					if err != nil {
						dialog.ShowError(err, m.window)
						return
					}
					dialog.ShowInformation("Save", "Your files are save in folder \n"+m.animateExport.ExportFolderPath, m.window)
				}, m.window)
				fo.Resize(savingDialogSize)
				fo.Show()
			}),
			widget.NewLabel("Export version (V1 not optimized, V2 optimized)"),
			widget.NewSelect([]string{"Version 1", "Version 2"}, func(v string) {
				switch v {
				case "Version 1":
					a.ExportVersion = animate.DeltaExportV1
				case "Version 2":
					a.ExportVersion = animate.DeltaExportV2
				default:
					a.ExportVersion = animate.DeltaExportV1
				}
			}),
		),
	)

	d := dialog.NewCustom("Export  animation", "Ok", cont, w)
	d.Resize(w.Canvas().Size())
	d.Show()
}

func (m *MartineUI) refreshAnimatePalette() {
	m.animate.SetPaletteImage(png.PalToImage(m.animate.Palette()))
}

func CheckWidthSize(width, mode int) bool {
	var colorPerPixel int

	switch mode {
	case 0:
		colorPerPixel = 2
	case 1:
		colorPerPixel = 4
	case 2:
		colorPerPixel = 8
	}
	remain := width % colorPerPixel
	return remain == 0
}

func (m *MartineUI) AnimateApply(a *menu.AnimateMenu) {
	cfg := m.NewConfig(a.ImageMenu, false)
	if cfg == nil {
		return
	}
	cfg.Compression = m.animateExport.ExportCompression
	pi := custom_widget.NewProgressInfinite("Computing, Please wait.", m.window)
	pi.Show()
	address, err := strconv.ParseUint(a.InitialAddress.Text, 16, 64)
	if err != nil {
		pi.Hide()
		dialog.ShowError(err, m.window)
		return
	}
	// controle de de la taille de la largeur en fonction du mode
	width := cfg.Size.Width
	mode := a.Mode
	// get all images from widget imagetable
	if !CheckWidthSize(width, mode) {
		pi.Hide()
		dialog.ShowError(fmt.Errorf("the width in not a multiple of color per pixel, increase the width"), m.window)
		return
	}
	imgs := a.AnimateImages.Images()[0]
	deltaCollection, rawImages, palette, err := animate.DeltaPackingMemory(imgs, cfg, uint16(address), uint8(a.Mode))
	pi.Hide()
	if err != nil {
		dialog.NewError(err, m.window).Show()
		return
	}
	a.DeltaCollection = deltaCollection
	a.SetPalette(palette)
	a.RawImages = rawImages
	a.SetPaletteImage(png.PalToImage(a.Palette()))
}

func (m *MartineUI) ImageIndexToRemove(row, col int) {
	m.animate.ImageToRemoveIndex = col
}

func (m *MartineUI) newAnimateTab(a *menu.AnimateMenu) fyne.CanvasObject {
	importOpen := NewImportButton(m, a.ImageMenu)

	paletteOpen := NewOpenPaletteButton(a.ImageMenu, m.window)

	forcePalette := widget.NewCheck("use palette", func(b bool) {
		a.UsePalette = b
	})

	openFileWidget := widget.NewButton("Add image", func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			if reader == nil {
				return
			}
			pi := custom_widget.NewProgressInfinite("Opening file, Please wait.", m.window)
			pi.Show()
			path := reader.URI()
			if strings.ToUpper(filepath.Ext(path.Path())) != ".GIF" {
				img, err := openImage(path.Path())
				if err != nil {
					pi.Hide()
					dialog.ShowError(err, m.window)
					return
				}
				if a.IsEmpty {
					a.AnimateImages.SubstitueImage(0, 0, canvas.NewImageFromImage(img))
				} else {
					a.AnimateImages.AppendImage(0, canvas.NewImageFromImage(img))
				}
				a.IsEmpty = false
				pi.Hide()
			} else {
				fr, err := os.Open(path.Path())
				if err != nil {
					pi.Hide()
					dialog.ShowError(err, m.window)
					return
				}
				defer fr.Close()
				gifCfg, err := gif.DecodeConfig(fr)
				if err != nil {
					pi.Hide()
					dialog.ShowError(err, m.window)
					return
				}
				fmt.Println(gifCfg.Height)
				_, err = fr.Seek(0, io.SeekStart)
				if err != nil {
					pi.Hide()
					dialog.ShowError(err, m.window)
					return
				}
				gifImages, err := gif.DecodeAll(fr)
				if err != nil {
					pi.Hide()
					dialog.ShowError(err, m.window)
					return
				}
				imgs := animate.ConvertToImage(*gifImages)
				for index, img := range imgs {
					if index == 0 {
						a.AnimateImages.SubstitueImage(0, 0, canvas.NewImageFromImage(img))
					} else {
						a.AnimateImages.AppendImage(0, canvas.NewImageFromImage(img))
					}
					a.IsEmpty = false
				}
				pi.Hide()
			}
			m.window.Resize(m.window.Content().Size())
		}, m.window)
		d.SetFilter(imagesFilesFilter)
		d.Resize(dialogSize)
		d.Show()
	})

	resetButton := widget.NewButtonWithIcon("Reset", theme.CancelIcon(), func() {
		a.AnimateImages.Reset()
		a.IsEmpty = true
	})

	exportButton := widget.NewButtonWithIcon("Export", theme.DocumentSaveIcon(), func() {
		m.exportAnimationDialog(a, m.window)
	})

	applyButton := widget.NewButtonWithIcon("Compute", theme.VisibilityIcon(), func() {
		fmt.Println("compute.")
		m.AnimateApply(a)
	})

	removeButton := widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {
		fmt.Printf("image index to remove %d\n", a.ImageToRemoveIndex)
		images := a.AnimateImages.Images()
		if len(images[0]) <= a.ImageToRemoveIndex {
			return
		}
		images[0] = append(images[0][:a.ImageToRemoveIndex], images[0][a.ImageToRemoveIndex+1:]...)
		canvasImages := custom_widget.NewImageTableCache(len(images), len(images[0]), fyne.NewSize(50, 50))
		for x := 0; x < len(images); x++ {
			for y := 0; y < len(images[x]); y++ {
				canvasImages.Set(x, y, canvas.NewImageFromImage(images[x][y]))
			}
		}
		a.AnimateImages.Update(canvasImages, canvasImages.ImagesPerRow, canvasImages.ImagesPerColumn)
		a.AnimateImages.Refresh()
	})

	openFileWidget.Icon = theme.FileImageIcon()

	isPlus := widget.NewCheck("CPC Plus", func(b bool) {
		a.IsCpcPlus = b
	})

	modes := widget.NewSelect([]string{"0", "1", "2"}, func(s string) {
		mode, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error %s cannot be cast in int\n", s)
		}
		a.Mode = mode
	})
	modes.SetSelected("0")
	modeSelection = modes
	modeLabel := widget.NewLabel("Mode:")

	widthLabel := widget.NewLabel("Width")
	a.Width().Validator = validation.NewRegexp("\\d+", "Must contain a number")

	heightLabel := widget.NewLabel("Height")
	a.Height().Validator = validation.NewRegexp("\\d+", "Must contain a number")

	a.AnimateImages = custom_widget.NewEmptyImageTable(fyne.NewSize(menu.AnimateSize, menu.AnimateSize))
	a.AnimateImages.IndexCallbackFunc = m.ImageIndexToRemove

	initalAddressLabel := widget.NewLabel("initial address")
	a.InitialAddress = widget.NewEntry()
	a.InitialAddress.SetText("c000")

	isSprite := widget.NewCheck("Is sprite", func(b bool) {
		a.IsSprite = b
	})
	m.animateExport.ExportCompression = -1
	compressData := widget.NewCheck("Compress data", func(b bool) {
		if b {
			m.animateExport.ExportCompression = 0
		} else {
			m.animateExport.ExportCompression = -1
		}
	})

	oneLine := widget.NewCheck("Every other line", func(b bool) {
		a.ImageMenu.OneLine = b
	})
	oneRow := widget.NewCheck("Every other row", func(b bool) {
		a.ImageMenu.OneRow = b
	})

	return container.New(
		layout.NewGridLayout(1),
		container.New(
			layout.NewGridLayoutWithRows(1),
			container.NewScroll(
				a.AnimateImages),
		),
		container.New(
			layout.NewVBoxLayout(),
			container.New(
				layout.NewHBoxLayout(),
				openFileWidget,
				resetButton,
				removeButton,
				paletteOpen,
				applyButton,
				exportButton,
				importOpen,
			),
			container.New(
				layout.NewGridLayoutWithColumns(2),
				container.New(
					layout.NewVBoxLayout(),
					isPlus,
					container.New(
						layout.NewVBoxLayout(),
						initalAddressLabel,
						a.InitialAddress,
					),
				),
				container.New(
					layout.NewGridLayoutWithColumns(2),
					container.New(
						layout.NewVBoxLayout(),
						isSprite,
						compressData,
					),
					container.New(
						layout.NewVBoxLayout(),
						container.New(
							layout.NewHBoxLayout(),
							modeLabel,
							modes,
						),
						container.New(
							layout.NewHBoxLayout(),
							widthLabel,
							a.Width(),
						),
						container.New(
							layout.NewHBoxLayout(),
							heightLabel,
							a.Height(),
						),
					),
				),
			),
			container.New(
				layout.NewGridLayoutWithRows(3),
				container.New(
					layout.NewVBoxLayout(),
					oneLine,
					oneRow,
				),
				container.New(
					layout.NewGridLayoutWithColumns(2),
					a.PaletteImage(),
					container.New(
						layout.NewHBoxLayout(),
						forcePalette,
						widget.NewButtonWithIcon("Swap", theme.ColorChromaticIcon(), func() {
							w2.SwapColor(m.SetPalette, a.Palette(), m.window, m.refreshAnimatePalette)
						}),
						widget.NewButtonWithIcon("export", theme.DocumentSaveIcon(), func() {
							d := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
								if err != nil {
									dialog.ShowError(err, m.window)
									return
								}
								if uc == nil {
									return
								}

								paletteExportPath := uc.URI().Path()
								uc.Close()
								os.Remove(uc.URI().Path())
								cfg := config.NewMartineConfig(filepath.Base(paletteExportPath), paletteExportPath)
								cfg.NoAmsdosHeader = false
								if err := impPalette.SaveKit(paletteExportPath+".kit", a.Palette(), false); err != nil {
									dialog.ShowError(err, m.window)
								}
								if err := ocpartstudio.SavePal(paletteExportPath+".pal", a.Palette(), uint8(a.Mode), false); err != nil {
									dialog.ShowError(err, m.window)
								}
							}, m.window)
							d.Show()
						}),
					),
				),

				container.New(
					layout.NewVBoxLayout(),
					widget.NewButton("show cmd", func() {
						e := widget.NewMultiLineEntry()
						e.SetText(a.CmdLine())

						d := dialog.NewCustom("Command line generated",
							"Ok",
							e,
							m.window)
						fmt.Printf("%s\n", a.CmdLine())
						size := m.window.Content().Size()
						size = fyne.Size{Width: size.Width / 2, Height: size.Height / 2}
						d.Resize(size)
						d.Show()
					}),
				),
			),
		),
	)
}
