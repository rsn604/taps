package tui

import (
	"fmt"
	"time"
	"regexp"
	"strings"
	"github.com/gdamore/tcell/v2"
	"github.com/rsn604/taps"
)

const (
	NO_ERROR    = -1
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
		{"errmsg", "red", "default"},
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
Name = "L"
Data = "<List>"
X = 42
Y = 4
Style = "select, select_focus"
FieldType = "select"

# ---------------------------------------------
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

# ---------------------------------------------
[[Field]]
Name = "L03"
Data = "E03　Date(10)"
X = 2
Y = 8
Style = "label"
FieldType = "label"

[[Field]]
Name = "E03"
X = 20
Y = 8
FieldLen = 10
Style = "edit, edit_focus"
FieldType = "edit"

[[Field]]	
Name = "D"
Data = "<Date>"
X = 42
Y = 8
Style = "select, select_focus"
FieldType = "select"

# ---------------------------------------------
[[Field]]	
X = 19
Y = 10
FieldLen=32
Rows = 5
Rect=true
Style = "linerect"
FieldType = "label"

[[Field]]	
Data = "E04 Multi-lines"
X = 2
Y = 9
Style = "label"
FieldType = "label"

[[Field]]	
Data = "    (30x4)"
X = 2
Y = 11
Style = "label"
FieldType = "label"

[[Field]]	
Name = "E04"
X = 20
Y = 11
FieldLen=30
Rows = 4
Style = "note, note_focus"
FieldType = "edit"

# ---------------------------------------------
[[Field]]	
Name = "I"
Data = "Input "
X = 15
Y = 9996
Style = "select, select_focus"
FieldType = "select"

[[Field]]	
Name = "Q"
Data = "Quit "
X = 31
Y = 9996
Style = "select, select_focus"
FieldType = "select"

# ---------------------------------------------
[[Field]]
Name = "ERR_MSG"
X = 15
Y = 9998
Style = "errmsg"
FieldType = "label"

# ---------------------------------------------
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

var ptnYMD = regexp.MustCompile(`^[0-9]{4}/(0[1-9]|1[0-2])/(0[1-9]|[12][0-9]|3[01])$`)

func checkYMD(s string) bool {
	return ptnYMD.MatchString(s)
}

type Input struct {
	panel *taps.Panel
}

func (m *Input) errCheck() (string, int) {
	errMsg := ""
	e01 := m.panel.Get("E01")
	e03 := m.panel.Get("E03")
	
	if !strings.HasPrefix(e01, "Data") {
		errMsg = "ERROR: E01 must start with 'Data' ."
		return errMsg, m.panel.GetFieldNumber("E01")
	}

	if !checkYMD(e03) {
		errMsg = "ERROR: E03 Date format error ."
		return errMsg, m.panel.GetFieldNumber("E03")
	}
	return "OK", NO_ERROR

}

func  (m *Input) Run() {
	if m.panel == nil {
		m.panel = InputPanel()
	}
	godate := &GoDate{}
	list := &List{}

	m.panel.Store("Test Data", "E01")
	m.panel.StoreList([]string{""}, "E04")

	for {
		m.panel.Say()
		k, n := m.panel.Read()
		if k == tcell.KeyEscape || n == "Q"{
			break
		}

		if n == "L"{
			rs := list.Run()
			m.panel.Store(rs, "E01")
		}
		if n == "D"{
			rs := godate.Run(time.Now())
			m.panel.Store(rs, "E03")
		}
		if n == "I" {
			msg, num := m.errCheck()
			if num > NO_ERROR {
				m.panel.Store(msg, "ERR_MSG")
				m.panel.SelectFocus = num
			}else{
				m.panel.Store(msg, "ERR_MSG")
			}
		}
	}
}

// ---------------------------------------
// Main
// ---------------------------------------
func App() {
	m := &Input{}
	m.Run()
	taps.Quit()
	
	fmt.Println("E01:", m.panel.Get("E01"))
	fmt.Println("E02:", m.panel.Get("E02"))
	fmt.Println("E03:", m.panel.Get("E03"))
	fmt.Println("E04:", m.panel.GetList("E04"))
}

