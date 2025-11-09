## TOML Screen definittion 

|Key	         |Attribute |Definition|
|:---------------|:---------|:---------|
|StartX          |int|Panel start col |
|StartY          |int|Panel start row|
|EndX            |int|Panel end col|
|EndY            |int|Panel end row|
|Rect            |bool|"true"; surrunding panel by line |
|ExitKey         |[]string|Key to exit "READ" function|
|[[Field]]       ||Field definition
|Name            |string|Field name|
|X               |int|Field start col, relative in Panel.|
|Y               |int|Field start row, relative in Panel.|
|FieldLen        |int|Field length|
|Style           |string|Field style|
|FieldType       |string|"label", "edit", "select"|
|Data            |string|Initial data|
|Attr            |string|"N" means numeric field|
|DataLen         |int|Data length|
|Rect            |bool|bool|"true"; surrunding by line|
|ExitKey         |[]string|Key to exit "READ" function|
|Cols            |int|Number of repetitions for col|
|Rows            |int|Number of repetitions for row|
|ColSpaces       |int|Space within col|
|RowSpaces       |int|Space within row|

---
## TAPS functions

### (1) Define Panel
func NewPanel(doc string, styleMatrix [][]string, help string)(*Panel)

### (2) Panel Read/Write 
func (p *Panel)Say()
func (p *Panel)Read()(tcell.Key, string)

### (2) Store data to Field 
func (p *Panel)Store(sData string, n string)
func (p *Panel)StoreList(listData []string, n string)

### (3) Get data from Field 
func (p *Panel)Get(n string)(string)
func (p *Panel)GetList(n string)([]string)

### (4) Change attribute of Field
func (p *Panel)SetEnabled(n string)
func (p *Panel)SetDisabled(n string)
func (p *Panel)SetBrowseMode(n string, flag bool)

### (5) Add Exitkey by code
func (p *Panel)AddExitKey(n string, key string)

### (6) Get Field number or name
func (p *Panel)GetFieldNumber(name string)(int)
func (p *Panel)GetFieldName(pos int)(string)

### (7) Get Field structure
func (p *Panel)GetDataField(name string)(*DataField)
func (p *Panel)GetDataFieldWithNumber(name string)(*DataField, int)

### (8) List
func (p *Panel)GetListFocus(name string)(int)
func (p *Panel)GetListFieldName(n string, i int)(string)






