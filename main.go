package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Declares text editor widgets and its basic functionalities.
// Used for preserving and sharing state between child widgets.
// Layout can be specified freely.
type Tabs struct {
	tabBar *container.AppTabs

	createEditor   func() *widget.Entry
	editors        map[int]*widget.Entry
	editorCallback func()

	wordsLabel         *widget.Label
	sentencesLabel     *widget.Label
	paragraphsLabel    *widget.Label
	mostCommonWord     string
	calcWords          func(string) int
	calcSentences      func(string) int
	calcParagraphs     func(string) int
	calcMostCommonWord func(string) string
	updateStatistics   func(string)

	appendButton *widget.Button
	closeButton  *widget.Button
	addNewTab    func()
	closeTab     func()
}

// Implements Tabs widgets and its functionalities.
func (t *Tabs) Init() {
	t.editors = make(map[int]*widget.Entry)

	// Counts words
	t.calcWords = func(s string) int {
		words := strings.Fields(s)
		return len(words)
	}

	// Counts sentences
	t.calcSentences = func(s string) int {
		var sentences int
		if len(s) > 0 {
			sentences = 1
		} else {
			sentences = 0
		}
		insideWord := false
		newSentence := false
		for _, Rune := range s {
			switch Rune {
			case '.', '?', '!':
				if insideWord {
					newSentence = true
				}
				insideWord = false
			default:
				if unicode.IsLetter(Rune) {
					insideWord = true
					if newSentence {
						sentences++
						newSentence = false
					}
				}
			}
		}
		return sentences
	}

	// Counts paragraphs
	t.calcParagraphs = func(s string) int {
		var paragraphs int
		if len(s) > 0 {
			paragraphs = 1
		} else {
			paragraphs = 0
		}
		prevNewLine := false
		inSentence := false
		newParagraph := false
		for _, Rune := range s {
			switch Rune {
			case '\n':
				if prevNewLine && inSentence {
					newParagraph = true
					prevNewLine = false
					inSentence = false
					continue
				}
				prevNewLine = true
			default:
				prevNewLine = false
				if unicode.IsLetter(Rune) {
					inSentence = true
					if newParagraph {
						paragraphs++
						newParagraph = false
					}
				}
			}
		}
		return paragraphs
	}

	t.calcMostCommonWord = func(s string) string {
		wordsMap := make(map[string]int)
		words := strings.Fields(s)
		var maxOccurence int
		mostCommon := ""
		for _, word := range words {
			wordsMap[word]++
			if wordsMap[word] > maxOccurence {
				maxOccurence = wordsMap[word]
				mostCommon = word
			}
		}
		return mostCommon
	}

	// Root widget
	t.tabBar = container.NewAppTabs()

	// Statistics widgets
	t.wordsLabel = widget.NewLabel("")
	t.sentencesLabel = widget.NewLabel("")
	t.paragraphsLabel = widget.NewLabel("")

	// Extends editor callback for new events
	t.editorCallback = func() {
		idx := t.tabBar.CurrentTabIndex()
		if editor, ok := t.editors[idx]; ok {
			editor.OnChanged(editor.Text)
		} else if idx == -1 {
			t.updateStatistics("")
		}
	}

	// Updates statistics widgets value
	t.updateStatistics = func(text string) {
		t.wordsLabel.Text = fmt.Sprintf("Words: %v", t.calcWords(text))
		t.sentencesLabel.Text = fmt.Sprintf("Sentences: %v", t.calcSentences(text))
		t.paragraphsLabel.Text = fmt.Sprintf("Paragraphs: %v", t.calcParagraphs(text))
		t.mostCommonWord = t.calcMostCommonWord(text)
	}

	// Updates statistics for just focused tab
	t.tabBar.OnChanged = func(tab *container.TabItem) {
		t.editorCallback()
	}

	// Instantiates new text field
	t.createEditor = func() *widget.Entry {
		editor := widget.NewMultiLineEntry()
		editor.SetPlaceHolder("Start typing here...")
		editor.OnChanged = func(text string) {
			t.updateStatistics(text)
		}
		return editor
	}

	// Adds new tab and its associated text field
	t.addNewTab = func() {
		newEditor := t.createEditor()
		t.tabBar.Append(
			container.NewTabItemWithIcon(
				"New File",
				nil,
				newEditor,
			),
		)
		t.tabBar.SelectTabIndex(len(t.tabBar.Items) - 1)
		t.editors[t.tabBar.CurrentTabIndex()] = newEditor
		t.editorCallback()
	}

	// Closes currently active tab
	t.closeTab = func() {
		currentIdx := t.tabBar.CurrentTabIndex()
		if currentIdx >= 0 {
			t.tabBar.RemoveIndex(currentIdx)
			delete(t.editors, currentIdx)
			t.editorCallback()
		}
	}

	// Add/Close tab buttons
	t.appendButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), t.addNewTab)
	t.closeButton = widget.NewButtonWithIcon("", theme.CancelIcon(), t.closeTab)
}

type MainMenu struct {
	menu     *fyne.MainMenu
	fileMenu *fyne.Menu
	loadFile func()
	saveFile func()
}

func main() {
	// Main window initialization
	app := app.New()
	mainWindow := app.NewWindow("Text Editor")

	tabs := Tabs{}
	tabs.Init()

	mainMenu := MainMenu{}

	// File Menu
	mainMenu.loadFile = func() {
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
			tabs.tabBar.CurrentTab().Text = reader.URI().Name()
			tabs.tabBar.Refresh()
		}, mainWindow)
		fileDialog.Show()
	}

	mainMenu.saveFile = func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}
			if writer == nil {
				log.Println("Cancelled")
				return
			}
			defer writer.Close()
			var text string
			if editor, ok := tabs.editors[tabs.tabBar.CurrentTabIndex()]; ok {
				text = editor.Text
			} else {
				text = ""
			}
			_, err = writer.Write([]byte(text))
			if err != nil {
				dialog.ShowError(err, mainWindow)
			}
			tabs.tabBar.CurrentTab().Text = writer.URI().Name()
			tabs.tabBar.Refresh()
		}, mainWindow)
	}

	mainMenu.fileMenu = fyne.NewMenu(
		"File",
		fyne.NewMenuItem("Open", mainMenu.loadFile),
		fyne.NewMenuItem("Save", mainMenu.saveFile),
	)
	mainMenu.menu = fyne.NewMainMenu(mainMenu.fileMenu)
	mainWindow.SetMainMenu(mainMenu.menu)

	// Most commond word button and popup
	popWidget := widget.NewLabel("")
	pop := widget.NewPopUp(popWidget, mainWindow.Canvas())
	commonWordButton := widget.NewButtonWithIcon(
		"",
		theme.HelpIcon(),
		func() {
			popWidget.Text = fmt.Sprintf("Most common word:\n%v", tabs.mostCommonWord)
			pop.ShowAtPosition(
				fyne.NewPos(
					mainWindow.Canvas().Size().Width/2-pop.MinSize().Width/2,
					mainWindow.Canvas().Size().Height/2-pop.MinSize().Height/2,
				),
			)
		},
	)

	// Main window layout
	mainWindow.SetContent(
		container.NewBorder(
			container.NewHBox(
				tabs.appendButton,
				tabs.closeButton,
			),
			container.NewHBox(
				layout.NewSpacer(),
				tabs.wordsLabel,
				tabs.sentencesLabel,
				tabs.paragraphsLabel,
				commonWordButton,
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
	mainWindow.Resize(fyne.NewSize(640, 480))
	mainWindow.ShowAndRun()
}
