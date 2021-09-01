package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Tabs struct {
	tabBar          *container.AppTabs
	addNewTab       func()
	closeTab        func()
	newEditor       func() *widget.Entry
	currentEditor   *widget.Entry
	calculateWords  func(string) int
	wordCountLabel  *widget.Label
	appendTabButton *widget.Button
	closeTabButton  *widget.Button
}

func (t *Tabs) Init() {
	t.calculateWords = func(s string) int {
		words := strings.Fields(s)
		return len(words)
	}
	t.wordCountLabel = widget.NewLabel("")
	t.tabBar = container.NewAppTabs()

	t.newEditor = func() *widget.Entry {
		editor := widget.NewMultiLineEntry()
		editor.SetPlaceHolder("Start typing here...")
		editor.OnChanged = func(s string) {
			t.wordCountLabel.Text = fmt.Sprintf("Words count: %v", t.calculateWords(s))
		}
		return editor
	}

	t.addNewTab = func() {
		t.currentEditor = t.newEditor()
		t.tabBar.Append(
			container.NewTabItemWithIcon(
				"New File",
				nil,
				t.currentEditor,
			),
		)
		t.tabBar.SelectTabIndex(len(t.tabBar.Items) - 1)
	}

	t.closeTab = func() {
		currentIdx := t.tabBar.CurrentTabIndex()
		if currentIdx >= 0 {
			t.tabBar.RemoveIndex(currentIdx)
		}
	}

	t.appendTabButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), t.addNewTab)
	t.closeTabButton = widget.NewButtonWithIcon("", theme.CancelIcon(), t.closeTab)
}

func main() {
	tabs := Tabs{}
	tabs.Init()

	// Main initialization
	app := app.New()
	mainWindow := app.NewWindow("Text Editor")

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
			tabs.tabBar.Append(
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
			tabs.tabBar.SelectTabIndex(len(tabs.tabBar.Items) - 1)
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
			container.NewHBox(
				tabs.appendTabButton,
				tabs.closeTabButton,
			),
			container.NewHBox(
				layout.NewSpacer(),
				tabs.wordCountLabel,
			),
			nil,
			nil,
			container.NewHScroll(
				tabs.tabBar,
			),
		),
	)

	// Main window parameters and launch
	mainWindow.SetMaster()
	mainWindow.Resize(fyne.NewSize(640, 460))
	mainWindow.ShowAndRun()
}
