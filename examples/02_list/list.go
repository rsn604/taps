package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rsn604/taps"
)

// ---------------------------------------
// Panel
// ---------------------------------------
func ListPanel() *taps.Panel {
	var styleMatrix = [][]string{
		{"list", "white", "default"},
		{"list_focus", "white", "aqua"},
	}

	var doc = `
StartX = 20
StartY = 4
EndX = 50
EndY = 17
Rect = true

[[Field]]	
Name = "LIST"
X = 2
Y = 1
Rows = 12
Style = "list, list_focus"
FieldType = "select"
`
	return taps.NewPanel(doc, styleMatrix, "")
}

// -------------------------------------------------
type List struct {
	panel *taps.Panel
}

func  (m *List) getList() []string {
	var listData []string
	for i := 0; i < 30; i++ {
		listData = append(listData, "Data" + fmt.Sprintf("%d", i))
	}
	return listData
}

func (m *List) Run() string {
	if m.panel == nil{
		m.panel = ListPanel()
	}
	m.panel.StoreList(m.getList(), "LIST")

	for {
		m.panel.Say()
		k, n := m.panel.Read()
		if k == tcell.KeyEscape{
			break
		}

		if k == tcell.KeyEnter {
			return m.panel.Get(n)
		}
	}
	return "Not selected"
}
// ---------------------------------------
// Main
// ---------------------------------------
func app() {
	m := &List{}
	s := m.Run()
	taps.Quit()
	fmt.Println("Data:", s)
}

func main() {
	taps.Main(app)
}
