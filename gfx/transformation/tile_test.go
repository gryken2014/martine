package transformation

import (
	"fmt"
	"image/png"
	"os"
	"testing"

	"github.com/jeromelesaux/martine/constants"
)

func TestBoardSprite(t *testing.T) {
	f, err := os.Open("../images/ak.png")
	if err != nil {
		t.Fatalf("Cannot open file error %v\n", err)
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatalf("Cannot decode png file error :%v\n", err)
	}
	a := AnalyzeTilesBoard(im, constants.Size{Width: 16, Height: 16})
	t.Log(a.String())
	fmt.Println(a.String())
	a.SaveSchema("alexkidd_board.png")
	a.SaveTilemap("alexkidd.map")
}
