package taps

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/pelletier/go-toml/v2"
	"strconv"
	"strings"
	//"log"
)

const (
	MAXRC       = 9999
	NORMAL_MODE = 0x00
	LIST_MODE   = 0x01
	BROWSE_MODE = 0x02
	DISABLED    = 0x80
	INVALID_KEY = -1
	EDIT        = "EDIT"
	SELECT      = "SELECT"
	LABEL       = "LABEL"
	LIST_SEP    = "_$$"
	GRID_SEP    = "_$#"
)

var taps = &Taps{}

type Taps struct {
	screen tcell.Screen
	style  tcell.Style
	err    error
}

type Panel struct {
	Field          []*DataField
	StartX, StartY int
	EndX, EndY     int
	SelectFocus    int
	Rect           bool
	ExitKey        []string
	styleMatrix    [][]string
	doc            string
	help           string
}

type ListField struct {
	data          string
	hDataPos      int
	hStartDataPos int
	hCursorX      int
	hCursorY      int
}

type DataMain struct {
	FieldType      string
	Name           string
	RData          []rune
	Data           string
	X, Y, FieldLen int
	Style          string
	Attr           string
	DataLen        int
	Picture        string
	Rect           bool
	ExitKey        []string
}

type DataField struct {
	DataMain
	GridFields []*DataField
	Cols       int
	Rows       int
	ColSpaces     int
	RowSpaces     int
	currentStyle       tcell.Style
	// ------
	normalStyle  tcell.Style
	focusedStyle tcell.Style
	listStart    int
	listData      []ListField
	hMode         byte
	hDataPos      int
	hStartDataPos int
	hCursorX      int
	hCursorY      int
}

// ---------------------------------------------
// Taps
// ---------------------------------------------
func Init() error {
	if taps.screen == nil {
		if taps.screen, taps.err = tcell.NewScreen(); taps.err != nil {
			return taps.err
		}
		if taps.err = taps.screen.Init(); taps.err != nil {
			return taps.err
		}

		//taps.screen.SetStyle(taps.style)
		taps.screen.EnableMouse(tcell.MouseButtonEvents)
	}
	return nil
}

func Quit() {
	if taps.screen != nil {
		taps.screen.Fini()
	}
}

func GetWindowSize() (int, int) {
	x, y := taps.screen.Size()
	return x - 1, y - 1
}

func checkXY(x, y int) bool {
	mx, my := GetWindowSize()
	if y > my || x > mx {
		return false
	}
	return true
}

func SetContent(x int, y int, r rune, bc []rune, style tcell.Style) {
	if checkXY(x, y) {
		taps.screen.SetContent(x, y, r, bc, style)
	}
}

func Show() {
	taps.screen.Show()
}

func ShowCursor(x, y int) {
	if checkXY(x, y) {
		taps.screen.ShowCursor(x, y)
		Show()
	}
}

func EraseCursor() {
	taps.screen.ShowCursor(-1, -1)
}

func ClrEol(j int) {
	mx, _ := GetWindowSize()
	for i := 0; i < mx; i++ {
		SetContent(i, j, ' ', nil, taps.style)
	}
	Show()
}

func Clear() {
	taps.screen.Clear()
}

func Fill(r rune, style tcell.Style) {
	taps.screen.Fill(r, style)
}

func ConsoleOut(ss string, x, y int, style tcell.Style) {
	p := x
	s := []rune(ss)
	for i := 0; i < len(s); i++ {
		SetContent(p, y, s[i], nil, style)
		p += runewidth.RuneWidth(s[i])
	}
	Show()
}

// ----------------------------------------------------------
func ClearRect(sx, sy int, ex, ey int, style tcell.Style) {
	for i := sx; i < ex; i++ {
		for j := sy; j < ey; j++ {
			SetContent(i, j, ' ', nil, style)
		}
	}
	Show()
}
func LineRect(sx, sy, ex, ey int, style tcell.Style) {
	//if ex == sx+1 {
	if ex == sx {
		SetContent(sx, sy, '│', nil, style)
		SetContent(ex, sy, '│', nil, style)
		SetContent(sx, ey, '│', nil, style)
		SetContent(ex, ey, '│', nil, style)

		//} else if ey == sy+1 {
	} else if ey == sy {
		SetContent(sx, sy, '─', nil, style)
		SetContent(ex, sy, '─', nil, style)
		SetContent(sx, ey, '─', nil, style)
		SetContent(ex, ey, '─', nil, style)
	} else {
		SetContent(sx, sy, '┌', nil, style)
		SetContent(ex, sy, '┐', nil, style)
		SetContent(sx, ey, '└', nil, style)
		SetContent(ex, ey, '┘', nil, style)
	}

	for i := sx + 1; i < ex; i++ {
		SetContent(i, sy, '─', nil, style)
	}
	for i := sx + 1; i < ex; i++ {
		SetContent(i, ey, '─', nil, style)
	}

	for j := sy + 1; j < ey; j++ {
		SetContent(sx, j, '│', nil, style)
	}
	for j := sy + 1; j < ey; j++ {
		SetContent(ex, j, '│', nil, style)
	}

	Show()
}

// ============================================
// ---------------------------------------------
// Panel
// ---------------------------------------------
func NewPanel(doc string, styleMatrix [][]string, help string) *Panel {
	var p Panel
	err := toml.Unmarshal([]byte(doc), &p)
	if err != nil {
		panic(err)
	}
	p.setFieldStyle(doc, styleMatrix, help)
	return &p
}

func ModifyPanelPosition(base *Panel, startX, startY int) *Panel {
	var p Panel

	//log.Printf("ModifyPanel:%s\n", base.doc)
	err := toml.Unmarshal([]byte(base.doc), &p)
	if err != nil {
		panic(err)
	}
	p.EndX = p.EndX - (p.StartX - startX)
	p.EndY = p.EndY - (p.StartY - startY)
	p.StartX = startX
	p.StartY = startY
	//log.Printf("ModifyPanel p.StartX:%d p.StartY:%d p.EndX:%d p.EndY:%d\n",p.StartX, p.StartY, p.EndX, p.EndY)

	p.setFieldStyle(base.doc, base.styleMatrix, base.help)
	return &p
}

// ---------------------------------------------
// Style
// ---------------------------------------------
func getDetailStyle(style tcell.Style, s string, foreground bool) tcell.Style {
	ss := strings.Split(s, ",")
	if foreground {
		style = style.Foreground(tcell.ColorNames[strings.TrimSpace(ss[0])])
	} else {
		style = style.Background(tcell.ColorNames[strings.TrimSpace(ss[0])])
	}
	/*
		style = style.Underline(false)
		style = style.Bold(false)
		style = style.Reverse(false)
		style = style.Italic(false)
		style = style.Blink(false)
		style = style.StrikeThrough(false)
		style = style.Dim(false)
	*/
	for i := 1; i < len(ss); i++ {
		switch strings.TrimSpace(ss[i]) {
		case "underline":
			style = style.Underline(true)
		case "bold":
			style = style.Bold(true)
		case "reverse":
			style = style.Reverse(true)
		case "italic":
			style = style.Italic(true)
		case "blink":
			style = style.Blink(true)
		case "strikethrough":
			style = style.StrikeThrough(true)
		case "dim":
			style = style.Dim(true)
		}
	}
	return style
}

func getStyle(styleString string, m [][]string) (tcell.Style, tcell.Style) {
	ss := strings.Split(styleString, ",")
	style := tcell.StyleDefault
	focused := tcell.StyleDefault
	for i := 0; i < len(m); i++ {
		if strings.TrimSpace(ss[0]) == m[i][0] {
			style = getDetailStyle(getDetailStyle(style, m[i][1], true), m[i][2], false)
		}
		if len(ss) > 1 && strings.TrimSpace(ss[1]) == m[i][0] {
			focused = getDetailStyle(getDetailStyle(style, m[i][1], true), m[i][2], false)
		}
	}
	return style, focused
}

// ---------------------------------------------------------
func (p *Panel) setGridFieldStyle(i int) int {
	pos := i
	gridFields := p.Field[pos].GridFields
	
	gridFieldLen := GetFieldX(p.Field[pos].FieldLen)
	maxLen := 0
	for k := 0; k < len(gridFields); k++ {
		if maxLen < GetFieldX(gridFields[k].FieldLen){
			maxLen = GetFieldX(gridFields[k].FieldLen)
		}
	}
	if gridFieldLen == 0{
		if maxLen > 0{
			gridFieldLen = maxLen
		}else{
			return i
		}
	}
	
	gridRows := GetFieldY(p.Field[pos].Rows)
	gridCols := GetFieldY(p.Field[pos].Cols)
	if gridRows == 0{
		gridRows = 1
	}
	if gridCols == 0{
		gridCols = 1
	}
	colSpaces := GetFieldY(p.Field[pos].ColSpaces)
	rowSpaces := GetFieldY(p.Field[pos].RowSpaces)

	rowWidth := 1
	minY := 9999
	maxY := 0
	for k := 0; k < len(gridFields); k++ {
		if minY > GetFieldY(gridFields[k].Y){
			minY = GetFieldY(gridFields[k].Y)
		}
		if maxY < GetFieldY(gridFields[k].Y){
			maxY = GetFieldY(gridFields[k].Y)
		}
		if rowWidth < GetFieldY(gridFields[k].Rows){
			rowWidth = GetFieldY(gridFields[k].Rows)
		}
	}
	if rowWidth < maxY - minY + 1{
		rowWidth = maxY - minY + 1
	}
	
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			for k := 0; k < len(gridFields); k++ {
				xpos := GetFieldX(gridFields[k].X) + GetFieldX(p.StartX)
				ypos := GetFieldY(gridFields[k].Y) + GetFieldY(p.StartY)
				s0, s1 := getStyle(gridFields[k].Style, p.styleMatrix)
				fieldRows := GetFieldY(gridFields[k].Rows)
				if fieldRows == 0{
					fieldRows = 1
				}
				
				for fr:=0; fr<fieldRows; fr++{
					s := new(DataField)
					s.currentStyle = s0
					s.normalStyle = s0
					s.focusedStyle = s1
					s.FieldType = gridFields[k].FieldType
					s.Attr = gridFields[k].Attr
					s.DataLen = gridFields[k].DataLen
					s.Picture = gridFields[k].Picture
					s.ExitKey = gridFields[k].ExitKey
					s.FieldLen = gridFields[k].FieldLen

					s.X = xpos + (gridFieldLen + colSpaces)*col
					s.Y = ypos + fr +(rowWidth + rowSpaces)*row

					s.Name = gridFields[k].Name + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row)

					if GetFieldY(gridFields[k].Rows) > 0{
						s.hMode = LIST_MODE
						s.Name = s.Name + LIST_SEP + fmt.Sprintf("%03d", fr)

					}else{
						s.hMode = NORMAL_MODE
					}
				
					if col == 0 && row == 0 && k == 0 && fr == 0{
						p.Field[i] = s
						p.Field[i].GridFields = gridFields
					} else {
						if i >= len(p.Field) {
							p.Field = append(p.Field, s)
						} else {
							p.Field = append(p.Field[:i], p.Field[i-1:]...)
							p.Field[i] = s
						}
					}
					p.Store(gridFields[k].Data, p.Field[i].Name)
					i++
				}
			}

		}
	}

	return i
}
func (p *Panel) setListFieldStyle(i int, s0 tcell.Style, s1 tcell.Style) int {
	pos := i
	x := GetFieldX(p.Field[pos].Cols)
	y := GetFieldY(p.Field[pos].Rows)
	name := p.Field[pos].Name
	fieldLen := 0
	xpos := GetFieldX(p.Field[pos].X) + GetFieldX(p.StartX)
	ypos := GetFieldY(p.Field[pos].Y) + GetFieldY(p.StartY)
	endy := GetFieldY(p.Field[pos].Rows)
	if p.Field[pos].FieldLen == 0 {
		fieldLen = GetFieldX(p.EndX) - p.Field[pos].X - GetFieldX(p.StartX)
		if p.Rect {
			fieldLen = fieldLen - 2
		}
	} else {
		fieldLen = p.Field[pos].FieldLen
	}

	k := 0
	j := 0
	fnum := 0
	for {
		s := new(DataField)
		s.hMode = LIST_MODE
		s.currentStyle = s0
		s.normalStyle = s0
		s.focusedStyle = s1
		s.FieldType = p.Field[pos].FieldType
		s.Attr = p.Field[pos].Attr
		s.DataLen = p.Field[pos].DataLen
		s.Picture = p.Field[pos].Picture
		s.ExitKey = p.Field[pos].ExitKey
		s.FieldLen = fieldLen

		s.Name = name + LIST_SEP + fmt.Sprintf("%03d", fnum)
		s.Rows = endy - j - 1

		//@@@@
		s.X = xpos + (fieldLen + p.Field[pos].ColSpaces)*k
		s.Y = ypos + j

		if j == 0 && k == 0 {
			p.Field[i] = s
		} else {
			if i >= len(p.Field) {
				p.Field = append(p.Field, s)
			} else {
				p.Field = append(p.Field[:i], p.Field[i-1:]...)
				p.Field[i] = s
			}
		}
		i++
		k++
		fnum++
		if k >= x {
			k = 0
			if j < y-1 {
				j++
			} else {
				break
			}
		}

	}
	return i
}


func (p *Panel) setFieldStyle(doc string, styleMatrix [][]string, help string) {
	i := 0
	p.doc = doc
	p.styleMatrix = styleMatrix
	p.help = help

	for {
		if i >= len(p.Field) {
			break
		}
		s0, s1 := getStyle(p.Field[i].Style, styleMatrix)

		if (GetFieldY(p.Field[i].Rows) == 0 && GetFieldY(p.Field[i].Cols) == 0) || p.Field[i].Rect {
			p.Field[i].currentStyle = s0
			p.Field[i].normalStyle = s0
			p.Field[i].focusedStyle = s1
			//@@@@
			p.Field[i].X = p.Field[i].X + GetFieldX(p.StartX)
			p.Field[i].Y = p.Field[i].Y + GetFieldY(p.StartY)
			p.Field[i].RData = []rune(p.Field[i].Data)
			p.Field[i].hMode = NORMAL_MODE
			i++
			continue
		}

		if GetFieldY(p.Field[i].Rows) > 0 || GetFieldY(p.Field[i].Cols) > 0 {
			if len(p.Field[i].GridFields) == 0 {
				i = p.setListFieldStyle(i, s0, s1)

			} else {
				i = p.setGridFieldStyle(i)
			}
		}

	}
}

func (p *Panel) GetFieldStyle(style string) (tcell.Style, tcell.Style){
	return getStyle(style, p.styleMatrix)
}

func (p *Panel) ResetFieldStyle(n, style string){
	s0, s1 := getStyle(style, p.styleMatrix)
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, n) {
			f.currentStyle = s0
			f.normalStyle = s0
			f.focusedStyle = s1
		}
	}

}

func (p *Panel) GetHelp() string{
	return p.help
}

// ---------------------------------------------
// Check field attribute
// ---------------------------------------------
func isEdit(f *DataField) bool {
	if strings.ToUpper(f.FieldType) == EDIT {
		return true
	}
	return false
}

func isSelect(f *DataField) bool {
	if strings.ToUpper(f.FieldType) == SELECT {
		return true
	}
	return false
}

func isLabel(f *DataField) bool {
	if strings.ToUpper(f.FieldType) == LABEL {
		return true
	}
	return false
}

func (f *DataField) Disabled() {
	if f != nil {
		f.hMode = f.hMode | DISABLED
	}
}

func (f *DataField) Enabled() {
	if f != nil {
		f.hMode = f.hMode & (0xff ^ DISABLED)
	}
}

func (p *Panel) SetEnabled(n string) {
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, n) {
			f.Enabled()
		}
	}
}

func (p *Panel) SetDisabled(n string) {
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, n) {
			f.Disabled()
		}
	}
}

func isDisabled(f *DataField) bool {
	if f.hMode & DISABLED != 0x00 {
		return true
	}
	return false
}

func isListMode(f *DataField) bool {
	if f.hMode & LIST_MODE != 0x00 {
		return true
	}
	return false
}

func (f *DataField) BrowseMode() {
	if f != nil {
		f.hMode = f.hMode | BROWSE_MODE
		//@@@
		/*
		f.hDataPos = 0
		f.hStartDataPos = 0
		f.hCursorX = 0
		f.hCursorY = 0
		*/
	}
}

func (f *DataField) EditMode() {
	if f != nil {
		f.hMode = f.hMode & (0xff ^ BROWSE_MODE)
	}
}

func isBrowseMode(f *DataField) bool {
	if f.hMode & BROWSE_MODE != 0x00 {
		return true
	}
	return false
}

func (p *Panel) SetBrowseMode(n string, flag bool) {
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, n) {
			if flag{
				f.BrowseMode()
			}else{
				f.EditMode()
			}

		}
	}
}

func (p *Panel) AddExitKey(n string, key string) {
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, n) {
			//log.Printf(fmt.Sprintf("f.Name:%s n:%s key:%s\n", f.Name, n, key))
			f.ExitKey = append(f.ExitKey, key)
		}
	}
}

// ---------------------------------------------
// Field size 
// ---------------------------------------------
func GetFieldSize(n, mn int) int {
	n2 := MAXRC - n
	if n2 <= mn {
		return mn - n2
	}
	return n
}

func GetFieldY(y int) int {
	_, my := GetWindowSize()
	return GetFieldSize(y, my)
}

func GetFieldX(x int) int {
	mx, _ := GetWindowSize()
	return GetFieldSize(x, mx)
}

func (f *DataField) GetFieldLen() int {
	mx, _ := GetWindowSize()
	if GetFieldX(f.X)+f.FieldLen >= mx {
		return mx - GetFieldX(f.X)
	} else {
		return f.FieldLen
	}
}

// ---------------------------------------------
// Write Field
// ---------------------------------------------
func (f *DataField) clearField() {
	mx, my := GetWindowSize()
	if GetFieldY(f.Y) > my || GetFieldX(f.X) >= mx {
		return
	}
	x := 0
	y := 0
	for {
		if x+GetFieldX(f.X) >= mx || x >= f.GetFieldLen() {
			break
		}
		SetContent(x+GetFieldX(f.X), y+GetFieldY(f.Y), ' ', nil, f.normalStyle)
		x++
	}
	EraseCursor()

}

func (f *DataField) writeField() {
	y := GetFieldY(f.Y)
	mx, my := GetWindowSize()
	if y > my || GetFieldX(f.X) >= mx {
		return
	}

	if f.Rect && isLabel(f) {
		if f.Rows > 0 {
			//LineRect(GetFieldX(f.X), GetFieldY(f.Y), GetFieldX(f.X)+GetFieldX(f.FieldLen), GetFieldY(f.Y)+GetFieldY(f.Rows)-1, f.normalStyle)
			LineRect(GetFieldX(f.X), GetFieldY(f.Y), GetFieldX(f.X)+GetFieldX(f.FieldLen), GetFieldY(f.Y)+GetFieldY(f.Rows), f.normalStyle)
		} else if f.Cols > 0 {
			LineRect(GetFieldX(f.X), GetFieldY(f.Y), GetFieldX(f.X)+GetFieldX(f.Cols), GetFieldY(f.Y), f.normalStyle)
		}
		return
	}

	x := 0
	for i := 0; i < len(f.RData); i++ {
		if (x+GetFieldX(f.X) >= mx) || (f.FieldLen > 0 && x >= f.GetFieldLen()) {
			//@@@@@
			//if isListMode(f) && isLabel(f) && y < GetFieldY(f.Y)+GetFieldY(f.Rows) {
			if isListMode(f) && isEdit(f)  && y < GetFieldY(f.Y)+GetFieldY(f.Rows) {
				x = 0
				y++
			} else {
				break
			}
		}

		SetContent(x+GetFieldX(f.X), y, f.RData[i], nil, f.currentStyle)
		x += runewidth.RuneWidth(f.RData[i])
	}
	Show()
}

func (f *DataField) writeEdit() {

	y := GetFieldY(f.Y)
	mx, my := GetWindowSize()
	if y > my || GetFieldX(f.X) > mx {
		return
	}

	x := 0
	for i := f.hStartDataPos; i < len(f.RData); i++ {
		//@@@@ Zenkaku/Hankaku
		if (x+GetFieldX(f.X) >= mx) || ((f.FieldLen > 0) && (x >= f.GetFieldLen())) {
			if isListMode(f) && y < GetFieldY(f.Y)+GetFieldY(f.Rows) {
				x = 0
				y++
			} else {
				break
			}
		}

		SetContent(x+GetFieldX(f.X), y, f.RData[i], nil, f.currentStyle)
		x += runewidth.RuneWidth(f.RData[i])
	}
}

// ---------------------------------------------
// Say Data
// ---------------------------------------------
func (f *DataField) Say() {
	if isDisabled(f) {
		return
	}

	f.clearField()
	if isEdit(f) {
		f.writeEdit()
		ShowCursor(f.getCursorPosX(), f.getCursorPosY())
	} else {
		f.writeField()
	}
}

func (p *Panel) additionalLines(f *DataField, ss string) int {
	x := 0
	lines := 0
	s := []rune(ss)

	for i := f.hStartDataPos; i < len(s); i++ {
		if (f.FieldLen > 0) && (x >= f.GetFieldLen()) {
			lines++
			x = 0
		}
		x += runewidth.RuneWidth(s[i])
	}
	return lines
}

func (p *Panel) ClearGridData(n string) {
	name := strings.Split(n, GRID_SEP)[0] + GRID_SEP
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, name){
			p.Store("", f.Name)
			//@@@@
			f.clearField()
			SetNormalStyle(f)
		}
	}
}

func (p *Panel) ClearGridList(n string) {
	_, col, row := p.getLastGrid(n)
	for j := 0; j <= row; j++ {
		for i := 0; i <= col; i++ {
			p.StoreGridList([]string{""}, n, i, j)
		}
	}
}

func (p *Panel) ClearList(n string) {
	name := strings.Split(n, LIST_SEP)[0] + LIST_SEP
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, name) {
			p.Store("", f.Name)
			f.clearField()
			SetNormalStyle(f)
		}
	}
}

func (p *Panel) ModifyFieldLen(name string, fieldLen int) {
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, name) {
			f.FieldLen = fieldLen
		}
	}
}

func (p *Panel) SayListData(n string) {
	s := p.getFirstList(n)
	pos := p.GetFieldNumber(s.Name)
	start := p.Field[pos].listStart
	i := pos
	lines := 0
	dataPos := 0
	listLen := p.getListLen(p.Field[pos].Name)

	p.ClearList(p.Field[pos].Name)
	//@@@@
	if isDisabled(p.Field[pos]){
		return
	}
	if len(s.listData) == 0{
		return
	}
	//

	for i < pos+listLen {
		if lines == 0 {
			p.Field[i].Enabled()

			if dataPos < len(s.listData)-start {
				p.Store(s.listData[dataPos+start].data, p.Field[i].Name)
				p.Field[i].hDataPos = s.listData[dataPos+start].hDataPos
				p.Field[i].hStartDataPos = s.listData[dataPos+start].hStartDataPos
				p.Field[i].hCursorX = s.listData[dataPos+start].hCursorX
				p.Field[i].hCursorY = s.listData[dataPos+start].hCursorY

				//@@@@
				//if !isSelect(p.Field[i]) {
				if isEdit(p.Field[i]) {
					lines = p.additionalLines(p.Field[i], s.listData[dataPos+start].data)
				}
			} else {
				p.Store("", p.Field[i].Name)
				p.Field[i].Disabled()
			}
			dataPos++
		} else {
			p.Field[i].Disabled()
			p.Store("", p.Field[i].Name)
			lines--
		}
		p.Field[i].Say()
		i++
	}
	Show()
}

func (p *Panel) Say() {
	//@@@ Style
	if GetFieldX(p.StartX) == 0 && GetFieldY(p.StartY) == 0 && p.EndX == 9999 && p.EndY == 9999 {
		Clear()
	} else {
		//@@@@
		ClearRect(GetFieldX(p.StartX)-1, GetFieldY(p.StartY), GetFieldX(p.EndX+1), GetFieldY(p.EndY), tcell.StyleDefault)
	}

	if p.Rect {
		LineRect(GetFieldX(p.StartX), GetFieldY(p.StartY), GetFieldX(p.EndX), GetFieldY(p.EndY), tcell.StyleDefault)
	}

	i := 0
	next := 0
	for i < len(p.Field) {
		if isListMode(p.Field[i]) {
			next = p.getListLen(p.Field[i].Name)
			p.SayListData(p.Field[i].Name)
			i += next
		} else {
			if !isDisabled(p.Field[i]) {
				p.Field[i].hDataPos = 0
				p.Field[i].hStartDataPos = 0
				p.Field[i].hCursorX = 0
				p.Field[i].hCursorY = 0
				p.Field[i].Say()
			}
			i++
		}

	}
	Show()
}

// ============================================
// Position
// ============================================
func resetCursorPos(startDataPos, dataPos, fieldLen int, data []rune, listMode bool) (int, int) {
	posx := 0
	posy := 0
	i := startDataPos
	for {
		if i >= dataPos {
			break
		}
		if listMode && i < len(data)-1 && posx >= fieldLen-runewidth.RuneWidth(data[i+1]) {
			posx = 0
			posy++
		} else if posx >= fieldLen {
			break
		} else {
			posx += runewidth.RuneWidth(data[i])
		}
		i++
	}
	return posx, posy
}

func (f *DataField) setCursorPos() {
	f.hCursorX, f.hCursorY = resetCursorPos(f.hStartDataPos, f.hDataPos, f.GetFieldLen(), f.RData, isListMode(f))
}

func (f *DataField) resetStartDataPos() {
	startPos := f.hStartDataPos
	posx := 0
	i := 0
	for {
		if i >= f.hDataPos {
			break
		}

		if i < len(f.RData)-1 && posx >= f.GetFieldLen()-runewidth.RuneWidth(f.RData[i+1]) {
			startPos = i + 1
			posx = 0
		} else {
			posx += runewidth.RuneWidth(f.RData[i])
		}
		i++
	}

	f.hStartDataPos = startPos
}

func (f *DataField) setStartDataPos() {
	if isListMode(f) {
		return
	}
	f.resetStartDataPos()
}

// ------------------------------------------

func (f *DataField) resetDataPos(x, y int) {
	posx := 0
	posy := 0
	i := f.hStartDataPos
	for {
		if i >= len(f.RData) {
			//posx = x
			break
		}

		if posy == y && posx >= x {
			break
		}

		if isListMode(f) && i < len(f.RData)-1 && posx >= f.GetFieldLen()-runewidth.RuneWidth(f.RData[i+1]) {
			posx = 0
			posy++
		} else {
			posx += runewidth.RuneWidth(f.RData[i])
		}
		i++
	}
	f.hDataPos = i
	f.hCursorX = posx
	f.hCursorY = posy
}

func (f *DataField) goFirstLinePos() {
	f.hStartDataPos = 0
	f.resetDataPos(0, 0)
}

func (f *DataField) getCursorPosX() int {
	return f.hCursorX + GetFieldX(f.X)
}

func (f *DataField) getCursorPosY() int {
	return f.hCursorY + GetFieldX(f.Y)
}

// ============================================
// Edit
// ============================================
func (p *Panel) input_del(i int) {
	if p.Field[i].hDataPos < len(p.Field[i].RData) {
		p.Field[i].RData = append(p.Field[i].RData[:p.Field[i].hDataPos], p.Field[i].RData[p.Field[i].hDataPos+1:]...)
		if isListMode(p.Field[i]) {
			p.updateList(p.Field[i].Name)
			p.SayListData(p.Field[i].Name)
			SetFocusedStyle(p.Field[i])
		}
		p.Field[i].Say()
		ShowCursor(p.Field[i].getCursorPosX(), p.Field[i].getCursorPosY())
	}
}

func (p *Panel) input_bs(i int) {
	if p.Field[i].hDataPos > 0 {
		p.Field[i].RData = append(p.Field[i].RData[:p.Field[i].hDataPos-1], p.Field[i].RData[p.Field[i].hDataPos:]...)
		p.Field[i].hDataPos--
		p.Field[i].setStartDataPos()
		p.Field[i].setCursorPos()

		if isListMode(p.Field[i]) {
			p.updateList(p.Field[i].Name)
			p.SayListData(p.Field[i].Name)
			SetFocusedStyle(p.Field[i])
		}

		p.Field[i].Say()

		ShowCursor(p.Field[i].getCursorPosX(), p.Field[i].getCursorPosY())
	}
}

func (p *Panel) input_rt(i int) {
	if isListMode(p.Field[i]) {
		if p.Field[i].hDataPos >= len(p.Field[i].RData) {
			return
		}
		s := p.getFirstList(p.Field[i].Name)
		pos := p.GetFieldNumber(s.Name)
		listLen := p.getListLen(p.Field[pos].Name)
		if p.Field[i].hCursorX >= p.Field[i].FieldLen-runewidth.RuneWidth(p.Field[i].RData[p.Field[i].hDataPos]) && p.Field[i].hCursorY >= pos+listLen-i-1 {
			return
		}
	}

	if p.Field[i].hDataPos < len(p.Field[i].RData) {
		p.Field[i].hDataPos++
	}
	p.Field[i].setStartDataPos()
	p.Field[i].setCursorPos()
	p.Field[i].Say()
	ShowCursor(p.Field[i].getCursorPosX(), p.Field[i].getCursorPosY())
}

func (p *Panel) input_lt(i int) {
	if p.Field[i].hCursorX == 0 {
		if p.Field[i].hCursorY > 0 {
			p.Field[i].hDataPos--
			p.Field[i].setCursorPos()
		} else {
			if p.Field[i].hStartDataPos > 0 {
				p.Field[i].hStartDataPos--
			}
			p.Field[i].hDataPos = p.Field[i].hStartDataPos
		}
	} else if p.Field[i].hDataPos > 0 {
		p.Field[i].hDataPos--
		//@@@@ Error
		p.Field[i].hCursorX -= runewidth.RuneWidth(p.Field[i].RData[p.Field[i].hDataPos])
	}

	p.Field[i].Say()
	ShowCursor(p.Field[i].getCursorPosX(), p.Field[i].getCursorPosY())

}

func (f *DataField) isNumeric(r rune) bool {
	if '0' <= r && r <= '9' {
		return true
	}

	if '+' == r || '-' == r {
		if f.hDataPos > 0 {
			return false
		}
		return true
	}

	// @@@ Check decimal point
	if '.' == r {
		for i := 0; i <= f.hDataPos; i++ {
			if i >= len(f.RData) {
				break
			}
			if f.RData[i] == '.' {
				return false
			}
		}
		return true
	}
	return false
}

func (f *DataField) checkPicture() bool {
	hLen := len(f.Picture)
	if hLen == 0 {
		return true
	}
	if f.DataLen < hLen {
		if f.Picture[f.hDataPos] == '9' {
			return f.isNumeric(f.RData[f.hDataPos])
		}
	}
	return true
}

func (p *Panel) input_data(i int, r rune) {
	// Check data length
	if p.Field[i].DataLen > 0 && len(p.Field[i].RData) == p.Field[i].DataLen {
		return
	}

	// Numeric check
	if p.Field[i].Attr == "N" || p.Field[i].Attr == "n" {
		if p.Field[i].isNumeric(r) == false {
			return
		}
	}

	// Check picture string
	p.Field[i].checkPicture()

	if p.Field[i].hDataPos < len(p.Field[i].RData) {
		p.Field[i].RData = append(p.Field[i].RData[:p.Field[i].hDataPos+1], p.Field[i].RData[p.Field[i].hDataPos:]...)
		p.Field[i].RData[p.Field[i].hDataPos] = r
	} else {
		p.Field[i].RData = append(p.Field[i].RData, r)
	}

	if p.Field[i].hCursorX < p.Field[i].GetFieldLen() {
		p.Field[i].hCursorX += runewidth.RuneWidth(p.Field[i].RData[p.Field[i].hDataPos])
	}

	p.Field[i].hDataPos++

	p.Field[i].setStartDataPos()
	p.Field[i].setCursorPos()

	if isListMode(p.Field[i]) {
		p.updateList(p.Field[i].Name)
		p.SayListData(p.Field[i].Name)
		SetFocusedStyle(p.Field[i])
	}
	p.Field[i].Say()

	ShowCursor(p.Field[i].getCursorPosX(), p.Field[i].getCursorPosY())
}

// ---------------------------------------------
// Getter
// ---------------------------------------------
func (p *Panel) GetFieldNumber(name string) int {
	pos := 0
	for i, f := range p.Field {
		if f.Name == name {
			pos = i
			break
		}
	}
	return pos
}

func (p *Panel) GetGridFieldNumber(n string, col, row int) int {
	return p.GetFieldNumber(n + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row))
}

func (p *Panel) GetFieldName(pos int) string {
	name := ""
	for i, f := range p.Field {
		if i == pos {
			name = f.Name
			break
		}
	}
	return name
}

func (p *Panel) GetGridFieldName(n string, col, row int) string {
	return n + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row)
}

func (p *Panel) GetDataFieldWithNumber(name string) (*DataField, int) {
	i := 0
	for _, f := range p.Field {
		if f.Name == name {
			return f, i
		}
		i++
	}
	return nil, -1
}

func (p *Panel) GetDataField(name string) *DataField {
	f, _ := p.GetDataFieldWithNumber(name)
	return f
}

// ---------------------------------------------
// Get
// ---------------------------------------------
func (p *Panel) Get(n string) string {
	f := p.GetDataField(n)
	if f != nil {
		f.Data = string(f.RData)
		return f.Data
	}
	return ""
}

func (p *Panel) GetGridData(n string, col, row int) string {
	return p.Get(n + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row))

}

func (p *Panel) GetList(n string) []string {
	f := p.getFirstList(n)
	if f == nil {
		return nil
	}
	return f.getListData()
}

func (p *Panel) GetGridList(n string, col, row int) []string {
	return p.GetList(n + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row))
}

// ---------------------------------------------
// Store
// ---------------------------------------------
func (p *Panel) Store(sData string, n string) {
	f := p.GetDataField(n)
	if f != nil {
		f.Data = sData
		f.RData = []rune(sData)
	}
}

func (p *Panel) StoreGridData(sData string, n string, col, row int) {
	p.Store(sData, n + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row))
}

func (p *Panel) StoreList(listData []string, n string) {
	f := p.getFirstList(n)
	if f != nil {
		f.setListData(listData)
		f.listStart = 0
		return
	}
}

func (p *Panel) StoreGridList(listData []string, n string, col, row int) {
	p.StoreList(listData, n + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row))
}

// ============================================
// List
// ============================================-
func (p *Panel) SetListStart(n string, listStart int) {
	f := p.getFirstList(n)
	if f != nil {
		f.listStart = listStart
	}
}

func (p *Panel) GetListFocus(name string) int {
	n := strings.Split(name, LIST_SEP)
	if len(n) == 2 {
		n2, _ := strconv.Atoi(n[1])
		return n2
	} else {
		return -1
	}
}

func (p *Panel) GetListCount(n string) (string, int) {
	lastname, cnt := p.getListCountX(n, false)
	return lastname, cnt
}

// -------------------------------
func (f *DataField) getListData() []string {
	var listData []string
	for i := 0; i < len(f.listData); i++ {
		listData = append(listData, f.listData[i].data)
	}
	return listData
}

func (f *DataField) setListData(listData []string) {
	f.listData = nil
	var s ListField
	for _, data := range listData {
		s.data = data
		f.listData = append(f.listData, s)
	}
}

func (p *Panel) GetNthListName(name string, i int) string {
	return strings.Split(name, LIST_SEP)[0] + LIST_SEP + fmt.Sprintf("%03d", i)
}

func (p *Panel) GetListFieldName(n string, i int) string {
	return n + LIST_SEP + fmt.Sprintf("%03d", i)
}

func (p *Panel) GetFirstListName(name string) string {
	return p.GetNthListName(name, 0)
}

func (p *Panel) GetNthGridName(name string, col, row int) string {
	return strings.Split(name, GRID_SEP)[0] + GRID_SEP + fmt.Sprintf("%03d:%03d", col, row)
}

func (p *Panel) GetFirstGridName(name string) string {
	return p.GetNthGridName(name, 0, 0)
}

func (p *Panel) getFirstList(name string) *DataField {
	f := p.GetDataField(p.GetNthListName(name, 0))
	return f
}

func (p *Panel) getLastList(n string) (string, int) {
	cnt := 0
	name := strings.Split(n, LIST_SEP)[0] + LIST_SEP
	lastname := ""
	for i := len(p.Field) - 1; i >= 0; i-- {
		if strings.HasPrefix(p.Field[i].Name, name) {
			if !(isDisabled(p.Field[i])) {
				cnt++
				if lastname == "" {
					lastname = p.Field[i].Name
				}
			}
		}
	}
	return lastname, cnt
}

func (p *Panel) getLastGrid(n string) (string, int, int) {
	name := strings.Split(n, GRID_SEP)[0] + GRID_SEP
	lastname := ""
	for i := len(p.Field) - 1; i >= 0; i-- {
		if strings.HasPrefix(p.Field[i].Name, name) {
			if !(isDisabled(p.Field[i])) {
				//if lastname == "" {
					lastname = p.Field[i].Name
					break
				//}
			}
		}
	}
	col, _ := strconv.Atoi(lastname[len(name):len(name)+3])
	row, _ := strconv.Atoi(lastname[len(name)+3+1:len(name)+3+1+3])
	return lastname, col, row
}

func (p *Panel) getListCountX(n string, isBreakName bool) (string, int) {
	cnt := 0
	name := strings.Split(n, LIST_SEP)[0] + LIST_SEP
	lastname := ""
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, name) {
			if f.Name == n && isBreakName {
				break
			}
			if !(isDisabled(f)) {
				cnt++
				lastname = f.Name
			}
		}

	}
	return lastname, cnt
}

func (p *Panel) getListCountUntil(n string) (string, int) {
	lastname, cnt := p.getListCountX(n, true)
	return lastname, cnt
}

func getListDataLen(f *DataField) int {
	return len(f.listData)
}

func (p *Panel) getListLen(n string) int {
	cnt := 0
	name := strings.Split(n, LIST_SEP)[0] + LIST_SEP
	for _, f := range p.Field {
		if strings.HasPrefix(f.Name, name) {
			cnt++
		}
	}
	return cnt
}

// -----------------------------------------------------
func (p *Panel) expandList(i int) {
	s := p.getFirstList(p.Field[i].Name)
	if s == nil {
		return
	}
	start := s.listStart
	_, curNum := p.getListCountUntil(p.Field[i].Name)
	s.listData[start+curNum].data = string(p.Field[i].RData[:p.Field[i].hDataPos])
	if len(s.listData) > start+curNum {
		s.listData = append(s.listData[:start+curNum+1], s.listData[start+curNum:]...)
		s.listData[start+curNum+1].data = string(p.Field[i].RData[p.Field[i].hDataPos:])
		s.listData[start+curNum+1].hDataPos = 0
		s.listData[start+curNum+1].hStartDataPos = 0
		s.listData[start+curNum+1].hCursorX = 0
		s.listData[start+curNum+1].hCursorY = 0

	} else {
		var t ListField
		t.data = string(p.Field[i].RData[p.Field[i].hDataPos:])
		s.listData = append(s.listData, t)
	}
}

func (p *Panel) concateList(i int) {
	p.concateList2(i, true)
}

func (p *Panel) concateList2(i int, isConcate bool) {
	_, curNum := p.getListCountUntil(p.Field[i].Name)

	if curNum == 0 && p.Field[i].listStart == 0 {
		return
	}

	s := p.getFirstList(p.Field[i].Name)
	if s == nil {
		return
	}
	start := s.listStart
	curPos := start + curNum - 1
	fieldLen := p.Field[i].GetFieldLen()
	s.listData[curPos].hDataPos = len([]rune(s.listData[curPos].data))
	s.listData[curPos].hCursorX, s.listData[curPos].hCursorY = resetCursorPos(s.listData[curPos].hStartDataPos, s.listData[curPos].hDataPos, fieldLen, []rune(s.listData[curPos].data), true)

	if isConcate {
		s.listData[curPos].data = s.listData[curPos].data + p.Get(p.Field[i].Name)
	} else {
		s.listData[curPos].data = p.Get(p.Field[i].Name)
	}

	if len(s.listData) > start+curNum {
		s.listData = append(s.listData[:start+curNum], s.listData[start+curNum+1:]...)
	}

}

func (p *Panel) killList(i int) {
	s := p.getFirstList(p.Field[i].Name)
	if s == nil {
		return
	}
	start := s.listStart
	_, curNum := p.getListCountUntil(p.Field[i].Name)
	s.listData[start+curNum].data = string(p.Field[i].RData[:p.Field[i].hDataPos])
	s.listData[start+curNum].hDataPos = p.Field[i].hDataPos
	s.listData[start+curNum].hStartDataPos = p.Field[i].hStartDataPos
	s.listData[start+curNum].hCursorX = p.Field[i].hCursorX
	s.listData[start+curNum].hCursorY = p.Field[i].hCursorY
}

func (p *Panel) insertList(i int) {
	s := p.getFirstList(p.Field[i].Name)
	if s == nil {
		return
	}

	start := s.listStart
	_, curNum := p.getListCountUntil(p.Field[i].Name)

	if len(s.listData) > start+curNum {
		s.listData = append(s.listData[:start+curNum+1], s.listData[start+curNum:]...)
		s.listData[start+curNum].data = ""
		s.listData[start+curNum].hDataPos = 0
		s.listData[start+curNum].hStartDataPos = 0
		s.listData[start+curNum].hCursorX = 0
		s.listData[start+curNum].hCursorY = 0
	} else {
		var t ListField
		t.data = ""
		s.listData = append(s.listData, t)
	}
}

func (p *Panel) updateList(n string) {
	s := p.getFirstList(n)
	if s == nil {
		return
	}

	start := s.listStart
	i := 0
	cnt := 0
	for {
		name := strings.Split(n, LIST_SEP)[0] + LIST_SEP + fmt.Sprintf("%03d", i)
		f := p.GetDataField(name)
		if f == nil {
			break
		}
		if !(isDisabled(f)) && isEdit(f) {
			if len(s.listData) > start+cnt {
				s.listData[start+cnt].data = p.Get(name)
				s.listData[start+cnt].hDataPos = f.hDataPos
				s.listData[start+cnt].hStartDataPos = f.hStartDataPos
				s.listData[start+cnt].hCursorX = f.hCursorX
				s.listData[start+cnt].hCursorY = f.hCursorY
				cnt++
			} else {
				var t ListField
				t.data = p.Get(name)
				s.listData = append(s.listData, t)
			}
		}
		i++
	}
}

// ============================================
// Read
// ============================================-
func getClickedField(sf []*DataField, e *tcell.EventMouse) (*DataField, int) {
	x, y := e.Position()
	editFlag := false

	for i := len(sf) - 1; i >= 0; i-- {
		w := GetFieldX(sf[i].FieldLen)
		if sf[i].FieldLen == 0 {
			w = len(sf[i].RData)
		}
		if x >= GetFieldX(sf[i].X) && x < GetFieldX(sf[i].X)+w && y == GetFieldY(sf[i].Y) && !editFlag {
			if isSelect(sf[i]) && len(sf[i].RData) > 0 {
				return sf[i], i
			}
			if isEdit(sf[i]) {
				editFlag = true
			}
		}
		if editFlag && !isDisabled(sf[i]) {
			sf[i].resetDataPos(x-GetFieldX(sf[i].X), y-GetFieldY(sf[i].Y))
			return sf[i], i
		}
	}
	return nil, -1
}
func checkExitKey(sf []*DataField, r rune) (*DataField, int) {
	s := string(r)
	var fsave *DataField
	isave := -1
	for i, f := range sf {

		if isSelect(f) {
			if s == f.Name {
				return f, i
			}
			if strings.ToUpper(s) == strings.ToUpper(f.Name) {
				fsave = f
				isave = i
			}
		}
	}
	if isave != -1 {
		return fsave, isave
	}
	return nil, -1
}

func SetNormalStyle(f *DataField) {
	f.currentStyle = f.normalStyle
}

func SetFocusedStyle(f *DataField) {
	f.currentStyle = f.focusedStyle
}

// ---------------------------------------------
func (p *Panel) priorSelect(i int, cKey tcell.Key) int {
	hCurSel := i
	hPriorSel := i - 1
	editField := isEdit(p.Field[hCurSel]) && (cKey == tcell.KeyUp)
	hSaveSel := hCurSel

	for {
		if hPriorSel < 0 {
			if hSaveSel != hCurSel {
				hPriorSel = hSaveSel
			} else {
				hPriorSel = hCurSel
			}
			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

		if isDisabled(p.Field[hPriorSel]) {
			hPriorSel--
			continue
		}

		if isLabel(p.Field[hPriorSel]) {
			hPriorSel--
			continue
		}

		if isEdit(p.Field[hPriorSel]) {
			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

		if len(p.Field[hPriorSel].RData) == 0 {
			hPriorSel--
			continue
		}

		if cKey == tcell.KeyLeft {
			if GetFieldY(p.Field[hCurSel].Y) != GetFieldY(p.Field[hPriorSel].Y) {
				if hSaveSel == hCurSel {
					hSaveSel = hPriorSel
				}
				hPriorSel--
				continue
			}

			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

		if cKey == tcell.KeyUp {
			if GetFieldY(p.Field[hCurSel].Y) <= GetFieldY(p.Field[hPriorSel].Y) {
				hPriorSel--
				continue
			}
			if hSaveSel == hCurSel {
				hSaveSel = hPriorSel
			}
			if GetFieldX(p.Field[hCurSel].X) == GetFieldX(p.Field[hPriorSel].X) {
				SetNormalStyle(p.Field[hCurSel])
				p.Field[hCurSel].Say()
				break
			}
			hPriorSel--
			continue
		}

		if editField {
			if hSaveSel == hCurSel {
				hSaveSel = hPriorSel
			}
			hPriorSel--
			continue
		} else {
			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}
	}
	return hPriorSel
}

func (p *Panel) nextSelect(i int, cKey tcell.Key) int {
	hCurSel := i
	hNextSel := i + 1
	editField := isEdit(p.Field[hCurSel]) && (cKey == tcell.KeyDown || cKey == tcell.KeyEnter)
	hSaveSel := hCurSel
	for {
		if hNextSel == len(p.Field) {
			if hSaveSel != hCurSel {
				hNextSel = hSaveSel
			} else {
				hNextSel = hCurSel
			}
			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

		if isDisabled(p.Field[hNextSel]) {
			hNextSel++
			continue
		}

		if isLabel(p.Field[hNextSel]) {
			hNextSel++
			continue
		}

		if isEdit(p.Field[hNextSel]) {
			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

		if len(p.Field[hNextSel].RData) == 0 {
			hNextSel++
			continue
		}

		if cKey == tcell.KeyRight {
			if GetFieldY(p.Field[hCurSel].Y) != GetFieldY(p.Field[hNextSel].Y) {
				if hSaveSel == hCurSel {
					hSaveSel = hNextSel
				}
				hNextSel++
				continue
			}

			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

		if cKey == tcell.KeyDown {
			if GetFieldY(p.Field[hCurSel].Y) >= GetFieldY(p.Field[hNextSel].Y) {
				hNextSel++
				continue
			}
			if hSaveSel == hCurSel {
				hSaveSel = hNextSel
			}
			if GetFieldX(p.Field[hCurSel].X) == GetFieldX(p.Field[hNextSel].X) {
				SetNormalStyle(p.Field[hCurSel])
				p.Field[hCurSel].Say()
				break
			}
			hNextSel++
			continue
		}

		//log.Printf("nextSelect i:%d hCurSel:%d hNextSel:%d hSaveSel:%d\n", i, hCurSel, hNextSel, hSaveSel)
		if editField {
			if hSaveSel == hCurSel {
				hSaveSel = hNextSel
			}
			hNextSel++
			continue
		} else {
			SetNormalStyle(p.Field[hCurSel])
			p.Field[hCurSel].Say()
			break
		}

	}
	return hNextSel
}

// ---------------------------------------------
func (p *Panel) doEdit(i int, cKey tcell.Key, rKey rune) (bool, int) {

	if isBrowseMode(p.Field[i]) {
		return false, i
	}
	
	if cKey == tcell.KeyRight || cKey == tcell.KeyCtrlF {
		p.input_rt(i)
	}

	if cKey == tcell.KeyLeft || cKey == tcell.KeyCtrlB {
		p.input_lt(i)
	}

	if cKey == tcell.KeyRune {
		p.input_data(i, rKey)
	}

	if cKey == tcell.KeyDelete || cKey == tcell.KeyCtrlD {
		p.input_del(i)
	}

	if cKey == tcell.KeyBackspace2 || cKey == tcell.KeyCtrlH {
		p.input_bs(i)
	}

	if cKey == tcell.KeyCtrlA {
		p.Field[i].goFirstLinePos()
		p.Field[i].Say()
	}

	if isListMode(p.Field[i]) {
		return false, i
	}

	// ---------------------------------------
	if cKey == tcell.KeyEnter {
		i = p.nextSelect(i, cKey)
		SetFocusedStyle(p.Field[i])
		p.Field[i].Say()
		return true, i
	}

	if cKey == tcell.KeyCtrlK {
		p.Store(string(p.Field[i].RData[:p.Field[i].hDataPos]), p.Field[i].Name)
		p.Field[i].Say()
		return true, i
	}

	if cKey == tcell.KeyCtrlE {
		p.Field[i].hDataPos = len(p.Field[i].RData)
		p.Field[i].resetStartDataPos()
		p.Field[i].resetDataPos(p.Field[i].GetFieldLen(), p.Field[i].hCursorY)

		p.Field[i].Say()
	}
	return false, i
}

// -----------------------------------------
func (p *Panel) goLineUp(i int) {
	startPos := p.Field[i].hStartDataPos
	posx := p.Field[i].GetFieldLen()

	for {
		if startPos == 0 || posx < 0 {
			break
		}
		posx -= runewidth.RuneWidth(p.Field[i].RData[startPos])
		startPos--
	}
	p.Field[i].hStartDataPos = startPos
}

func (p *Panel) scrollUp(i, pos, lines int, lastListName string) (bool, int) {
	if !isEdit(p.Field[i]) {
		return false, i
	}

	if lines > 0 && p.Field[i].hCursorY > 0 {
		p.Field[i].hCursorY--
		p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
		return true, i
	}

	if i == pos {
		if p.Field[i].hCursorY == 0 {
			if p.Field[i].hStartDataPos == 0 && p.Field[i].listStart == 0 {
				return false, i
			}
		}

		if p.Field[i].hStartDataPos > 0 {
			p.goLineUp(i)
			p.updateList(lastListName)

		} else {
			p.Field[i].listStart--
		}

		p.SayListData(lastListName)
		p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
		return true, i

	}
	return false, i

}

// -----------------------------------------
func (p *Panel) goLineDown2(i, sposition, lineCount int) int {
	startPos := sposition
	posx := 0
	posy := 1
	saveStartPos := sposition
	for {
		if startPos >= len(p.Field[i].RData) {
			startPos = saveStartPos
			break
		}
		if posx >= p.Field[i].GetFieldLen()-runewidth.RuneWidth(p.Field[i].RData[startPos]) {
			if posy > lineCount {
				break
			} else {
				posy++
				saveStartPos = startPos
			}
		}
		posx += runewidth.RuneWidth(p.Field[i].RData[startPos])
		startPos++
	}

	return startPos
}

func (p *Panel) goLineDown(i int) {
	p.Field[i].hStartDataPos = p.goLineDown2(i, p.Field[i].hStartDataPos, 1)
}

func (p *Panel) scrollDown(i, pos, lines, cnt, listLen int, lastListName string) (bool, int) {

	if !isEdit(p.Field[i]) {
		return false, i
	}

	if p.Field[i].Name != lastListName {
		if p.Field[i].hCursorY < lines {
			p.Field[i].hCursorY++
			p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
			return true, i
		}
		return false, i
	}

	// Last List && last field
	if p.Field[pos].listStart+cnt >= getListDataLen(p.Field[pos]) {
		if p.Field[i].hCursorY >= lines {
			return false, i
		} else if p.Field[i].hCursorY < pos+listLen-i-1 {
			p.Field[i].hCursorY++
			p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
			return true, i
		}
	}

	// Last List && not last field
	if p.Field[i].hCursorY < lines && p.Field[i].hCursorY < pos+listLen-i-1 {
		p.Field[i].hCursorY++
		p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
		return true, i
	}

	if i == pos || isDisabled(p.Field[pos+1]) {
		p.goLineDown(pos)
		p.updateList(lastListName)
	} else {
		p.Field[pos].listStart++
	}
	p.SayListData(lastListName)

	n, _ := p.getLastList(lastListName)
	i = p.GetFieldNumber(n)
	linex := p.additionalLines(p.Field[i], string(p.Field[i].RData))
	if p.Field[i].hCursorY < linex && p.Field[i].hCursorY < pos+listLen-i-1 {
		p.Field[i].hCursorY++
	}
	p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
	return true, i
}

// ---------------------------------------------
func (p *Panel) doList(i int, cKey tcell.Key, rKey rune) (bool, int) {
	var isContinue bool
	s := p.getFirstList(p.Field[i].Name)
	pos := p.GetFieldNumber(s.Name)
	lastListName, cnt := p.getLastList(p.Field[pos].Name)
	lines := p.additionalLines(p.Field[i], p.Field[i].Data)
	listLen := p.getListLen(p.Field[pos].Name)

	if cKey == tcell.KeyUp {
		isContinue, i = p.scrollUp(i, pos, lines, lastListName)
		if isContinue {
			SetFocusedStyle(p.Field[pos])
			p.Field[i].Say()
			p.updateList(p.Field[pos].Name)
			return true, i
		}

		if p.Field[pos].listStart > 0 && p.Field[i].Name == p.Field[pos].Name && isSelect(p.Field[i]) {
			p.Field[pos].listStart--
			p.SayListData(p.Field[pos].Name)
			SetFocusedStyle(p.Field[pos])
			p.Field[pos].Say()
			return true, i
		}
	}

 	if cKey == tcell.KeyDown {
		if p.Field[i].Name == lastListName && isSelect(p.Field[i]) {
			if getListDataLen(p.Field[pos]) > p.Field[pos].listStart+cnt {
				p.Field[pos].listStart++
				p.SayListData(p.Field[i].Name)
				SetFocusedStyle(p.Field[i])
				p.Field[i].Say()
				return true, i
			}
		}

		isContinue, i = p.scrollDown(i, pos, lines, cnt, listLen, lastListName)
		if isContinue {
			SetFocusedStyle(p.Field[i])
			p.Field[i].Say()
			p.updateList(lastListName)
			return true, i
		}
	}

	if !isEdit(p.Field[i]) || isBrowseMode(p.Field[i]) {
		return false, i
	}

	// ------------------------------------------------
	if cKey == tcell.KeyEnter {
		p.expandList(i)
		p.SayListData(p.Field[i].Name)
		lastListName, _ := p.getLastList(p.Field[i].Name)
		if p.Field[i].Name != lastListName {
			i = p.nextSelect(i, cKey)
			SetFocusedStyle(p.Field[i])
			p.Field[i].Say()
			return true, i
		}
	}

	if cKey == tcell.KeyBackspace2 || cKey == tcell.KeyCtrlH {
		if p.Field[i].hDataPos == 0 {
			p.concateList(i)
			p.SayListData(p.Field[i].Name)
			i = p.priorSelect(i, cKey)
			SetFocusedStyle(p.Field[i])
			p.Field[i].Say()
			return true, i
		}
	}

	if cKey == tcell.KeyCtrlK {
		p.killList(i)
		p.SayListData(p.Field[i].Name)
		SetFocusedStyle(p.Field[i])
		p.Field[i].Say()
		return true, i
	}

	if cKey == tcell.KeyCtrlO {
		p.insertList(i)
		p.SayListData(p.Field[i].Name)
		p.Field[i].Say()
		return true, i
	}

	if cKey == tcell.KeyCtrlE {
		count := lines - p.Field[i].hCursorY
		for {
			if count == 0 {
				p.Field[i].resetDataPos(p.Field[i].GetFieldLen(), p.Field[i].hCursorY)
				SetFocusedStyle(p.Field[i])
				p.Field[i].Say()
				return true, i
			}
			isContinue, i = p.scrollDown(i, pos, lines, cnt, listLen, lastListName)
			if isContinue {
				p.updateList(lastListName)
			}
			s = p.getFirstList(p.Field[i].Name)
			pos = p.GetFieldNumber(s.Name)
			lastListName, cnt = p.getLastList(p.Field[pos].Name)
			//lines = p.additionalLines(p.Field[i], p.Field[i].Data)

			count--
		}
	}

	return false, i
}

// ---------------------------------------------
func isExitKey(exitKey []string, cKey tcell.Key) bool {
	for _, x := range exitKey {
		for k, v := range tcell.KeyNames {
			if v == x && k == cKey {
				//log.Printf("isExitKet v:%s x:%s\n", v, x)

				return true
			}
		}
	}
	return false
}

// ---------------------------------------------
func (p *Panel) checkBreak(i int, cKey tcell.Key, rKey rune) (bool, string) {
	if isExitKey(p.ExitKey, cKey) || isExitKey(p.Field[i].ExitKey, cKey) {
		SetNormalStyle(p.Field[i])
		p.Field[i].Say()
		p.SelectFocus = i
		if isExitKey(p.ExitKey, cKey) {
			return true, ""
		} else {
			return true, p.Field[i].Name
		}
	}

	if cKey == tcell.KeyEscape {
		SetNormalStyle(p.Field[i])
		p.Field[i].Say()
		p.SelectFocus = i
		return true, p.Field[i].Name
	}

	if isSelect(p.Field[i]) {
		if cKey == tcell.KeyEnter {
			SetNormalStyle(p.Field[i])
			p.Field[i].Say()
			p.SelectFocus = i
			return true, p.Field[i].Name
		}
		if cKey == tcell.KeyRune {
			fld, focus := checkExitKey(p.Field, rKey)
			if fld != nil {
				SetNormalStyle(p.Field[i])

				p.Field[i].Say()
				p.SelectFocus = focus
				return true, fld.Name
			}
		}
	}
	return false, p.Field[i].Name
}

func (p *Panel) locateField(i int) int {
	if p.Field == nil {
		return INVALID_KEY
	}
	startField := i
	for {
		if i >= len(p.Field) {
			if startField == 0 {
				return INVALID_KEY
			} else {
				startField = 0
				i = 0
				continue
			}
		}

		if isDisabled(p.Field[i]) {
			i++
			continue
		}

		//if isLabel(p.Field[i]) && !(isListMode(p.Field[i])) {
		if isLabel(p.Field[i]) {
			i++
			continue
		}
		//@@@@
		if isEdit(p.Field[i]) || len(p.Field[i].RData) > 0 || getListDataLen(p.Field[i]) > 0 {
			SetFocusedStyle(p.Field[i])
			p.Field[i].Say()
			break
		}
		i++
	}
	return i
}

// ---------------------------------------------
func (p *Panel) read2(i int) (tcell.Key, string) {
	var isContinue bool

	i = p.locateField(i)
	if i == INVALID_KEY {
		//return INVALID_KEY, ""
		return tcell.KeyEscape, ""
	}

	for {
		if isDisabled(p.Field[i]) && !(isListMode(p.Field[i])) {
			i++
			continue
		}

		if isLabel(p.Field[i]) && !(isListMode(p.Field[i])) {
			i++
			continue
		}

		ev := taps.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			cKey := ev.Key()
			rKey := ev.Rune()
			isBreak, n := p.checkBreak(i, cKey, rKey)
			if isBreak {
				return cKey, n
			}

			if isEdit(p.Field[i]) && !isDisabled(p.Field[i]) {
				isContinue, i = p.doEdit(i, cKey, rKey)
				if isContinue {
					continue
				}
			}

			if isListMode(p.Field[i]) {
				isContinue, i = p.doList(i, cKey, rKey)
				if isContinue {
					continue
				}
			}

			if (isSelect(p.Field[i]) && cKey == tcell.KeyLeft) || cKey == tcell.KeyUp || cKey == tcell.KeyBacktab {
				i = p.priorSelect(i, cKey)
				SetFocusedStyle(p.Field[i])
				p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
				p.Field[i].Say()
			}

			if (isSelect(p.Field[i]) && cKey == tcell.KeyRight) || cKey == tcell.KeyDown || cKey == tcell.KeyTab {
				i = p.nextSelect(i, cKey)
				SetFocusedStyle(p.Field[i])
				p.Field[i].resetDataPos(p.Field[i].hCursorX, p.Field[i].hCursorY)
				p.Field[i].Say()
			}

		case *tcell.EventMouse:
			/*
				if ev.Buttons()&tcell.Button5 != 0 {

					Say(p.Field[i])
					p.SelectFocus = i
					return tcell.KeyEscape, p.Field[i].Name
				}
			*/
			if ev.Buttons()&tcell.Button1 != 0 {
				f, num := getClickedField(p.Field, ev)
				if f != nil {
					SetNormalStyle(p.Field[i])
					p.Field[i].Say()
					p.SelectFocus = num
					if isSelect(f) {
						return tcell.KeyEnter, f.Name
					}
					if isEdit(f) {
						i = num
						SetFocusedStyle(p.Field[i])

						p.Field[i].Say()
					}
				}
			}

			if ev.Buttons()&tcell.WheelUp != 0 || ev.Buttons()&tcell.WheelDown != 0 {
				f, _ := getClickedField(p.Field, ev)
				if f != nil {
					if isListMode(f) {
						s := p.getFirstList(f.Name)
						if s != nil {
							if ev.Buttons()&tcell.WheelUp != 0 {
								if s.listStart > 0 {
									s.listStart--
									p.Say()
									continue
								}
							}
							if ev.Buttons()&tcell.WheelDown != 0 {
								_, cnt := p.getLastList(s.Name)
								if getListDataLen(s) > s.listStart+cnt {
									s.listStart++
									p.Say()
									continue
								}
							}
						}
					}
				}
			}
		}
	}
}

func (p *Panel) Read() (tcell.Key, string) {
	cKey, n := p.read2(p.SelectFocus)
	return cKey, n
}

// ---------------------------------------
/*
func (taps *Taps) SetStyle(style tcell.Style) {
	taps.style = style
	if taps.screen != nil {
		taps.screen.SetStyle(style)
	}
}
*/
func Main(app func()) {
	if err := Init(); err != nil {
		return
	}
	defer Quit()
	app()
}
