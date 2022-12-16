package ui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jeromelesaux/martine/config"
	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/convert/image"
	"github.com/jeromelesaux/martine/export/amsdos"
	"github.com/jeromelesaux/martine/export/ascii"
	"github.com/jeromelesaux/martine/export/diskimage"
	impPalette "github.com/jeromelesaux/martine/export/impdraw/palette"
	"github.com/jeromelesaux/martine/export/m4"

	"github.com/jeromelesaux/martine/export/ocpartstudio"
	"github.com/jeromelesaux/martine/export/png"
	"github.com/jeromelesaux/martine/export/snapshot"
	"github.com/jeromelesaux/martine/gfx"
	"github.com/jeromelesaux/martine/ui/martine-ui/menu"
	w2 "github.com/jeromelesaux/martine/ui/martine-ui/widget"
)

func (m *MartineUI) ExportOneImage(me *menu.ImageMenu) {
	pi := dialog.NewProgressInfinite("Saving....", "Please wait.", m.window)
	pi.Show()
	cfg := m.NewConfig(me, true)
	if cfg == nil {
		return
	}
	if m.imageExport.ExportText {
		if cfg.Overscan {
			cfg.ExportAsGoFile = true
		}

		out, _, palette, _, err := gfx.ApplyOneImage(me.OriginalImage.Image, cfg, me.Mode, me.Palette, uint8(me.Mode))
		if err != nil {
			pi.Hide()
			dialog.ShowError(err, m.window)
			return
		}
		code := ascii.FormatAssemblyDatabyte(out, "\n")
		palCode := ascii.FormatAssemblyCPCPalette(palette, "\n")
		content := fmt.Sprintf("; Generated by martine\n; from file %s\nImage:\n%s\n\n; palette\npalette: \n%s\n ",
			me.OriginalImagePath.Path(),
			code,
			palCode)
		filename := filepath.Base(me.OriginalImagePath.Path())
		fileExport := m.imageExport.ExportFolderPath + string(filepath.Separator) + filename + ".asm"
		if err = amsdos.SaveStringOSFile(fileExport, content); err != nil {
			pi.Hide()
			dialog.ShowError(err, m.window)
			return
		}
	} else {
		// palette export
		defer func() {
			os.Remove("temporary_palette.kit")
		}()
		if err := impPalette.SaveKit("temporary_palette.kit", me.Palette, false); err != nil {
			pi.Hide()
			dialog.ShowError(err, m.window)
		}
		cfg.KitPath = "temporary_palette.kit"
		filename := filepath.Base(me.OriginalImagePath.Path())
		if err := gfx.ApplyOneImageAndExport(
			me.OriginalImage.Image,
			cfg,
			filename,
			m.imageExport.ExportFolderPath+string(filepath.Separator)+filename,
			me.Mode,
			uint8(me.Mode)); err != nil {
			pi.Hide()
			dialog.NewError(err, m.window).Show()
			return
		}
		if cfg.Dsk {
			if err := diskimage.ImportInDsk(me.OriginalImagePath.Path(), cfg); err != nil {
				dialog.NewError(err, m.window).Show()
				return
			}
		}
		if cfg.Sna {
			if cfg.Overscan {
				var gfxFile string
				for _, v := range cfg.DskFiles {
					if filepath.Ext(v) == ".SCR" {
						gfxFile = v
						break
					}
				}
				cfg.SnaPath = filepath.Join(m.imageExport.ExportFolderPath, "test.sna")
				if err := snapshot.ImportInSna(gfxFile, cfg.SnaPath, uint8(me.Mode)); err != nil {
					dialog.NewError(err, m.window).Show()
					return
				}
			}
		}
	}
	if m.imageExport.ExportToM2 {
		if err := m4.ImportInM4(cfg); err != nil {
			dialog.NewError(err, m.window).Show()
			fmt.Fprintf(os.Stderr, "Cannot send to M4 error :%v\n", err)
		}
	}
	pi.Hide()
	dialog.ShowInformation("Save", "Your files are save in folder \n"+m.imageExport.ExportFolderPath, m.window)

}

func (m *MartineUI) monochromeColor(c color.Color) {

	m.main.Palette = image.ColorMonochromePalette(c, m.main.Palette)
	m.main.PaletteImage.Image = png.PalToImage(m.main.Palette)
	m.main.PaletteImage.Refresh()
}

func (m *MartineUI) ApplyOneImage(me *menu.ImageMenu) {
	cfg := m.NewConfig(me, true)
	if cfg == nil {
		return
	}

	var inPalette color.Palette
	if me.UsePalette {
		inPalette = me.Palette
		maxPalette := len(inPalette)
		switch me.Mode {
		case 1:
			if maxPalette > 4 {
				maxPalette = 4
			}
			inPalette = inPalette[0:maxPalette]
		case 2:
			if maxPalette > 2 {
				maxPalette = 2
			}
			inPalette = inPalette[0:maxPalette]
		}

	}
	pi := dialog.NewProgressInfinite("Computing", "Please wait.", m.window)
	pi.Show()
	out, downgraded, palette, _, err := gfx.ApplyOneImage(me.OriginalImage.Image, cfg, me.Mode, inPalette, uint8(me.Mode))
	pi.Hide()
	if err != nil {
		dialog.NewError(err, m.window).Show()
		return
	}
	me.Data = out
	me.Downgraded = downgraded
	if !me.UsePalette {
		me.Palette = palette
	}
	if me.IsSprite || me.IsHardSprite {
		newSize := constants.Size{Width: cfg.Size.Width * 50, Height: cfg.Size.Height * 50}
		me.Downgraded = image.Resize(me.Downgraded, newSize, me.ResizeAlgo)
	}
	me.CpcImage.Image = me.Downgraded
	me.CpcImage.FillMode = canvas.ImageFillStretch
	me.CpcImage.Refresh()
	me.PaletteImage.Image = png.PalToImage(me.Palette)
	me.PaletteImage.Refresh()
}

func (m *MartineUI) newImageTransfertTab(me *menu.ImageMenu) fyne.CanvasObject {
	importOpen := NewImportButton(m, me)

	paletteOpen := NewOpenPaletteButton(me, m.window)

	forcePalette := widget.NewCheck("use palette", func(b bool) {
		me.UsePalette = b
	})

	openFileWidget := widget.NewButton("Image", func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			if reader == nil {
				return
			}

			me.OriginalImagePath = reader.URI()
			img, err := openImage(me.OriginalImagePath.Path())
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			me.OriginalImage.Image = img
			me.OriginalImage.FillMode = canvas.ImageFillContain
			me.OriginalImage.Refresh()
			// m.window.Canvas().Refresh(&me.OriginalImage)
			// m.window.Resize(m.window.Content().Size())
		}, m.window)
		d.SetFilter(imagesFilesFilter)
		d.Resize(dialogSize)
		d.Show()
	})

	exportButton := widget.NewButtonWithIcon("Export", theme.DocumentSaveIcon(), func() {
		m.exportDialog(m.imageExport, m.window)
	})

	applyButton := widget.NewButtonWithIcon("Apply", theme.VisibilityIcon(), func() {
		fmt.Println("apply.")
		m.ApplyOneImage(me)
	})

	openFileWidget.Icon = theme.FileImageIcon()

	winFormat := w2.NewWinFormatRadio(me)

	colorReducerLabel := widget.NewLabel("Color reducer")
	colorReducer := widget.NewSelect([]string{"none", "Lower", "Medium", "Strong"}, func(s string) {
		switch s {
		case "none":
			me.Reducer = 0
		case "Lower":
			me.Reducer = 1
		case "Medium":
			me.Reducer = 2
		case "Strong":
			me.Reducer = 3
		}
	})
	colorReducer.SetSelected("none")

	resize := w2.NewResizeAlgorithmSelect(me)
	resizeLabel := widget.NewLabel("Resize algorithm")

	ditheringMultiplier := widget.NewSlider(0., 2.5)
	ditheringMultiplier.Step = 0.1
	ditheringMultiplier.SetValue(1.18)
	ditheringMultiplier.OnChanged = func(f float64) {
		me.DitheringMultiplier = f
	}
	dithering := w2.NewDitheringSelect(me)

	ditheringWithQuantification := widget.NewCheck("With quantification", func(b bool) {
		me.WithQuantification = b
	})

	enableDithering := widget.NewCheck("Enable dithering", func(b bool) {
		me.ApplyDithering = b
	})
	isPlus := widget.NewCheck("CPC Plus", func(b bool) {
		me.IsCpcPlus = b
	})

	oneLine := widget.NewCheck("Every other line", func(b bool) {
		me.OneLine = b
	})
	oneRow := widget.NewCheck("Every other row", func(b bool) {
		me.OneRow = b
	})
	modes := widget.NewSelect([]string{"0", "1", "2"}, func(s string) {
		mode, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error %s cannot be cast in int\n", s)
		}
		me.Mode = mode
	})
	modes.SetSelected("0")
	modeSelection = modes
	modeLabel := widget.NewLabel("Mode:")

	widthLabel := widget.NewLabel("Width")
	me.Width = widget.NewEntry()
	me.Width.Validator = validation.NewRegexp("\\d+", "Must contain a number")

	heightLabel := widget.NewLabel("Height")
	me.Height = widget.NewEntry()
	me.Height.Validator = validation.NewRegexp("\\d+", "Must contain a number")

	brightness := widget.NewSlider(0.0, 1.0)
	brightness.SetValue(1.)
	brightness.Step = .01
	brightness.OnChanged = func(f float64) {
		me.Brightness = f
	}
	saturationLabel := widget.NewLabel("Saturation")
	saturation := widget.NewSlider(0.0, 1.0)
	saturation.SetValue(1.)
	saturation.Step = .01
	saturation.OnChanged = func(f float64) {
		me.Saturation = f
	}
	brightnessLabel := widget.NewLabel("Brightness")

	warningLabel := widget.NewLabel("Setting thoses parameters will affect your palette, you can't force palette.")
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}

	return container.New(
		layout.NewGridLayoutWithColumns(2),
		container.New(
			layout.NewGridLayoutWithRows(2),
			container.NewScroll(
				me.OriginalImage),
			container.NewScroll(
				me.CpcImage),
		),
		container.New(
			layout.NewVBoxLayout(),
			container.New(
				layout.NewHBoxLayout(),
				openFileWidget,
				paletteOpen,
				applyButton,
				exportButton,
				importOpen,
			),
			container.New(
				layout.NewHBoxLayout(),
				isPlus,
				winFormat,

				container.New(
					layout.NewVBoxLayout(),
					container.New(
						layout.NewVBoxLayout(),
						modeLabel,
						modes,
					),
					container.New(
						layout.NewHBoxLayout(),
						widthLabel,
						me.Width,
					),
					container.New(
						layout.NewHBoxLayout(),
						heightLabel,
						me.Height,
					),
				),
			),
			container.New(
				layout.NewGridLayoutWithRows(7),
				container.New(
					layout.NewGridLayoutWithRows(2),
					container.New(
						layout.NewGridLayoutWithColumns(2),
						resizeLabel,
						resize,
					),
					container.New(
						layout.NewGridLayoutWithColumns(4),
						enableDithering,
						dithering,
						ditheringMultiplier,
						ditheringWithQuantification,
					),
				),
				container.New(
					layout.NewGridLayoutWithRows(2),
					oneLine,
					oneRow,
				),
				container.New(
					layout.NewGridLayoutWithRows(2),
					me.PaletteImage,
					container.New(
						layout.NewHBoxLayout(),
						forcePalette,
						widget.NewButtonWithIcon("Swap", theme.ColorChromaticIcon(), func() {
							w2.SwapColor(m.SetPalette, me.Palette, m.window, func() {
								forcePalette.SetChecked(true)
							})
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
								if err := impPalette.SaveKit(paletteExportPath+".kit", me.Palette, false); err != nil {
									dialog.ShowError(err, m.window)
								}
								if err := ocpartstudio.SavePal(paletteExportPath+".pal", me.Palette, uint8(me.Mode), false); err != nil {
									dialog.ShowError(err, m.window)
								}
							}, m.window)
							d.Show()
						}),
						widget.NewButton("Gray", func() {
							if me.IsCpcPlus {
								me.Palette = image.MonochromePalette(me.Palette)
								me.PaletteImage.Image = png.PalToImage(me.Palette)
								me.PaletteImage.Refresh()
								forcePalette.SetChecked(true)
								forcePalette.Refresh()
							}
						}),

						widget.NewButton("Monochome", func() {
							if me.IsCpcPlus {
								w2.ColorSelector(m.monochromeColor, me.Palette, m.window, func() {
									forcePalette.SetChecked(true)
								})
							}
						}),
					),
				),
				container.New(
					layout.NewVBoxLayout(),
					warningLabel,
				),
				container.New(
					layout.NewVBoxLayout(),
					colorReducerLabel,
					colorReducer,
				),
				container.New(
					layout.NewVBoxLayout(),
					brightnessLabel,
					brightness,
				),
				container.New(
					layout.NewVBoxLayout(),
					saturationLabel,
					saturation,
					widget.NewButton("show cmd", func() {
						e := widget.NewMultiLineEntry()
						e.SetText(me.CmdLine())

						d := dialog.NewCustom("Command line generated",
							"Ok",
							e,
							m.window)
						fmt.Printf("%s\n", me.CmdLine())
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
