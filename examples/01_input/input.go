package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rsn604/taps"
)

func InputPanel() *taps.Panel {
	var styleMatrix = [][]string{
		{"label", "lightcyan", "default"},
		{"select", "yellow", "default"},
		{"select_focus", "red,bold", "white"},
		{"edit", "white, underline", "black"},
		{"edit_focus", "yellow", "black"},
		{"note", "white", "black"},
		{"note_focus", "yellow,underline", "black"},
	}

	var doc = `
StartX = 0
StartY = 0
EndX = 9999
EndY = 9999

[[Field]]
Name = "L01"
Data = "E01　String(20)"
X = 2
Y = 4
Style = "label"
FieldType = "label"

[[Field]]
Name = "E01"
X = 20
Y = 4
Style = "edit, edit_focus"
FieldLen = 20
FieldType = "edit"

[[Field]]	
Name = "D"
Data = "<Disabled>"
X = 42
Y = 4
Style = "select, select_focus"
FieldType = "select"


[[Field]]
Name = "L02"
Data = "E02　Numeric(6)"
X = 2
Y = 6
Style = "label"
FieldType = "label"

[[Field]]
Name = "E02"
#Data = "123456"
X = 20
Y = 6
Style = "edit, edit_focus"
FieldLen = 6
DataLen = 6
Attr = "N"
FieldType = "edit"

[[Field]]	
X = 19
Y = 8
FieldLen=32
Rows = 7
Rect=true
Style = "linerect"
FieldType = "label"

[[Field]]	
Data = "E03 Multi-lines"
X = 2
Y = 9
Style = "label"
FieldType = "label"

[[Field]]	
Data = "    (30x4)"
X = 2
Y = 10
Style = "label"
FieldType = "label"

[[Field]]	
Name = "E03"
X = 20
Y = 9
FieldLen=30
Rows = 6
Style = "note, note_focus"
FieldType = "edit"

[[Field]]	
Name = "B"
Data = "<Edit>"
X = 53
Y = 9
Style = "select, select_focus"
FieldType = "select"

[[Field]]	
Name = "Q"
Data = "Quit "
X = 31
Y = 9996
Style = "select, select_focus"
FieldType = "select"

[[Field]]
Name = "L99"
Data = "Last line is here . "
X = 0
Y = 9999
Style = "label"
FieldType = "label"
`
	return taps.NewPanel(doc, styleMatrix, "")
}

type Input struct {
	panel *taps.Panel
}


func  (m *Input) Run() {
	lines := []string{"AAAA", "BBBB", "CCCC", "DDDD"}

	if m.panel == nil {
		m.panel = InputPanel()
	}

	m.panel.Store("Test Data", "E01")
	m.panel.StoreList(lines, "E03")
	m.panel.SetBrowseMode("E03", true)

	for {
		m.panel.Say()
		k, n := m.panel.Read()
		if k == tcell.KeyEscape || n == "Q"{
			break
		}

		if n == "D"{
			s := m.panel.Get(n)
			if s == "<Disabled>"{
				m.panel.SetDisabled("E01")
				m.panel.Store("<Enabled>", "D")
			}else{
				m.panel.SetEnabled("E01")
				m.panel.Store("<Disabled>", "D")
			}
		}
		if n == "B"{
			s := m.panel.Get(n)
			if s == "<Browse>"{
				m.panel.SetBrowseMode("E03", true)
				m.panel.Store("<Edit>", "B")
			}else{
				m.panel.SetBrowseMode("E03", false)
				m.panel.Store("<Browse>", "B")
			}
		}
	}
}

// ---------------------------------------
// Main
// ---------------------------------------
func app() {
	m := &Input{}
	m.Run()
	taps.Quit()
	
	fmt.Println("E01:", m.panel.Get("E01"))
	fmt.Println("E02:", m.panel.Get("E02"))
	fmt.Println("E03:", m.panel.GetList("E03"))
}

func main() {
	taps.Main(app)
}

