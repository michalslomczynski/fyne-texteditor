package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Main initialization
	app := app.New()
	mainWindow := app.NewWindow("Text Editor")

	// TabBar
	tabBar := container.NewAppTabs()
	addNewTab := func() {
		tabBar.Append(
			container.NewTabItemWithIcon(
				"New File",
				nil,
				func() *widget.Entry {
					editor := widget.NewMultiLineEntry()
					editor.SetPlaceHolder("Start typing here...")
					return editor
				}(),
			),
		)
	}
	appendTabButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), addNewTab)

	closeTab := func() {
		currentIdx := tabBar.CurrentTabIndex()
		fmt.Println(currentIdx)
		if currentIdx >= 0 {
			tabBar.RemoveIndex(currentIdx)
		}
	}
	closeTabButton := widget.NewButtonWithIcon("", theme.CancelIcon(), closeTab)

	toolBar := container.NewHBox(
		appendTabButton,
		closeTabButton,
		layout.NewSpacer(),
	)

	// Menu
	loadFile := func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}
			if reader == nil {
				log.Println("Cancelled")
				return
			}
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				log.Fatal(err)
			}
			tabBar.Append(
				container.NewTabItemWithIcon(
					"New File",
					nil,
					func() *widget.Entry {
						editor := widget.NewMultiLineEntry()
						editor.Text = string(data)
						return editor
					}(),
				),
			)
			tabBar.SelectTabIndex(len(tabBar.Items) - 1)
		}, mainWindow)
		fileDialog.Show()
	}
	fileMenu := fyne.NewMenu(
		"File",
		fyne.NewMenuItem("Open", loadFile),
		fyne.NewMenuItem("Save", nil),
		fyne.NewMenuItem("Save As...", nil),
	)
	mainMenu := fyne.NewMainMenu(fileMenu)
	mainWindow.SetMainMenu(mainMenu)

	// Main window layout
	mainWindow.SetContent(
		container.NewBorder(
			toolBar,
			nil,
			nil,
			nil,
			container.NewHScroll(
				tabBar,
			),
		),
	)

	// Main window parameters and launch
	mainWindow.SetMaster()
	mainWindow.Resize(fyne.NewSize(640, 460))
	mainWindow.ShowAndRun()
}
