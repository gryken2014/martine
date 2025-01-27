package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (m *MartineUI) newGreedings() fyne.CanvasObject {
	return container.New(
		layout.NewHBoxLayout(),
		widget.NewLabel(`Some greedings.
		Thanks a lot to all the Impact members.
		Ast, CMP, Demoniak, Kris and Drill
		Specials thanks for support to : 
		***      Tronic        ***
		***        Siko          ***
		*** Roudoudou ***
		and thanks a lot to all users^^
		for more informations about this tool, go to https://github.com/jeromelesaux/martine
		for more informations about my tool go to https://github.com/jeromelesaux
		to follow me on my old website https://http://koaks.amstrad.free.fr/amstrad/
		to chat with us got to https://amstradplus.forumforever.com/index.php  or
		https://discord.com/channels/453480213032992768/454619697485447169 on discord
		`),
		layout.NewSpacer(),
	)
}
