package menu

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/jeromelesaux/fyne-io/custom_widget"
)

var SpriteSize float32 = 80.

type SpriteExportFormat string

var (
	SpriteFlatExport  SpriteExportFormat = "Flat"
	SpriteFilesExport SpriteExportFormat = "Files"
	SpriteImpCatcher  SpriteExportFormat = "Impcatcher"
)

type SpriteMenu struct {
	IsHardSprite    bool
	OriginalBoard   canvas.Image
	OriginalPalette canvas.Image

	Palette                color.Palette
	SpritesData            [][][]byte
	CompileSprite          bool
	IsCpcPlus              bool
	OriginalImages         *custom_widget.ImageTable
	SpritesCollection      [][]*image.NRGBA
	SpriteNumberPerRow     int
	SpriteNumberPerColumn  int
	Mode                   int
	SpriteWidth            int
	SpriteHeight           int
	ExportFormat           SpriteExportFormat
	ExportDsk              bool
	ExportText             bool
	ExportWithAmsdosHeader bool
	ExportZigzag           bool
	ExportJson             bool
	ExportCompression      int
	ExportFolderPath       string
}

func NewSpriteMenu() *SpriteMenu {
	return &SpriteMenu{
		OriginalBoard:     canvas.Image{},
		OriginalImages:    custom_widget.NewEmptyImageTable(fyne.NewSize(SpriteSize, SpriteSize)),
		SpritesCollection: make([][]*image.NRGBA, 0),
		SpritesData:       make([][][]byte, 0),
	}
}
