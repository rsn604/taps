The programs that use **Taps**, specify the input panel using **TOML** format. Let's explain this with a concrete program example. 
(**TOML** is a file format primarily used for configuration files. For details, please refer to resources like <a href="https://ja.wikipedia.org/wiki/TOML" target="_blank">this</a> .)

**Japanese** is [here](TUTORIAL_JP.md) 

---
## [1] Input Program Example (examples/01_input/input.go)

Assume you have an input screen like this:

![INPUT](/imsges/input.png)

### (1) Panel Definition

 We'll discuss the field attribute definition later, but first, look at the part below "**var doc**".
 First, we define the panel. **StartX** and **StartY** indicate the starting point, and **EndX** and **EndY** indicate the end point.
The value "**9999**" has a special meaning: it is replaced by the terminal size. For example, in an **80x24** terminal, EndX will be "79" and EndY will be "23". This absorbs differences caused by terminal size.

```
    var doc = `
StartX = 0
StartY = 0
EndX = 9999
EndY = 9999
```

### (2) Field Definition

 Next, we configure the fields. Each field is defined using the **[[Field]]** tag, as shown below.

```
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
```

This example sets up the first display field and the input field. The field type is specified by FieldType. **Taps** supports only three field types: "**label**," "**edit**" and "**select**" .

Also, for a field like **E02**, specifying Attr = "N" makes it an input field restricted to numbers (including "+", "-", and "."). In this example, the data length is also limited by "**DataLen = 6**" .

```
[[Field]]
Name = "E02"
X = 20
Y = 6
Style = "edit, edit_focus"
FieldLen = 6
DataLen = 6
Attr = "N"
FieldType = "edit"
```

### (3) Field Style Definition

 Field style refers to the display attributes, which can be specified for two cases: when the field is simply displayed, and when it has focus.

```
    var styleMatrix = [][]string{
        {"label", "lightcyan", "default"},
        {"select", "yellow", "default"},
        {"select_focus", "red,bold", "white"},
        {"edit", "white, underline", "black"},
        {"edit_focus", "yellow", "black"},
        {"note", "white", "black"},
        {"note_focus", "yellow,underline", "black"},
    }
```

The first item is the string specified in Style, the second is the ForeGround attribute, and the third is the BackGround attribute.
(The color attributes are constants defined by tcell; please refer to <a href="https://github.com/gdamore/tcell/blob/main/color.go" target="_blank">this page</a>.)

### (4) Panel Registration

 After defining all fields, they are passed as parameters to the NewPanel function as shown below.

```
func InputPanel() *taps.Panel {
     :
     :
    return taps.NewPanel(doc, styleMatrix, "")
}
```

### (5) Editing Input Fields

The following keys can be used in input fields:

|Key	        |Function|
|:---------------|:---------|
|ESC           |Close Panel|
|CTRL+B (ArrowR)|Move cursor right|
|CTRL+F (ArrowL)|Move cursor left |
|TAB           |Move next or previous field|
|CTRL+A	        |Move cursor to the start of the line|
|CTRL+E	        |Move cursor to the end of the line|
|CTRL+K	        |Delete from the cursor to the end of the line|
|CTRL+D (DEL)   |Delete character at the cursor position|
|CTRL+H (BS)    |Delete character before the cursor. Joins lines at the start of a multi-line field.|
|ENTER	        |Moves to the next field for "edit." Splits the line for multi-line fields.|

To **Disable/Enable** or change to **Browse mode**, use the following functions:

```
m.panel.SetDisabled("E01")
m.panel.SetEnabled("E01")

m.panel.SetBrowseMode("E03", true)
m.panel.SetBrowseMode("E03", false)
```

### (6)  Programming Pattern
 In a typical application, multiple code are expected to be used, so first, we define a struct and store the panel to be used within it.

```
type Input struct {
    panel *taps.Panel
}

func (m *Input) Run() {
    if m.panel == nil {
        m.panel = InputPanel()
    }

    m.panel.Store("Test Data", "E01")
    m.panel.StoreList([]string{""}, "E03")

    for {
        m.panel.Say()
        k, n := m.panel.Read()
        if k == tcell.KeyEscape || n == "Q"{
            break
        }
        if n == "D"{
            s := m.panel.Get(n)

            :
```

 Next, set data to the fields on the panel. In the example above, the constant "Test Data" is stored in the E01 field using **Store** function.
The general procedure is then to display the screen with **Say** and receive input data with **Read**.**Read** returns the entered key code and the field name, allowing the logic to branch based on the content.

In the example above, the loop breaks if the "Escape" key is pressed or if field "Q" is selected. Also, if field "D" is selected, its content is read using **Get**. This pattern is used to construct the program.

---
## [2] List Format Example (examples/02_list/list.go)

 Let's explain the structure of the list format. This is generally called a ListBox and will be used when selecting data from a displayed list.

![LIST](/imsges/list.png)

The panel specifies a portion of the terminal. By specifying the number of *rows as **Rows = 12** for the selection item, a screen like the one above can be defined.

```
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
```

Actual data can be set to this "LIST" field by passing an array of strings using StoreList.

```
func (m *List) getList() []string {
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
        :
```

In this example, string array has been set.

---
### (3) GoDate (examples/03_godate/godate.go)

Next, let's look at a program with a slightly more complex structure: the so-called Calendar program.

![GODATE](/imsges/godate.png)

"**ExitKey**" is added to the panel definition.

```
StartX = 10
StartY = 2
EndX = 48
EndY = 16
Rect = true
ExitKey = ["F2", "F3", "F4", "F5", "F6", "F7", "F8", "F10", "F12"]
```

The **Read** function normally only returns control when "**ESC**" is pressed or **ENTER** is hit on a select field. Other keys are ignored, so keys that should cause an exit from **Read** are specified "ExitKey".
(Note that the keys are constants defined by tcell; please refer to <a href="https://github.com/gdamore/tcell/blob/main/key.go" target="_blank">this page</a>.)

Next is the definition of the Calendar body itself.

```
[[Field]]
Name = "CAL"
X = 6
Y = 3
FieldLen = 4
Cols = 7
Rows = 6
Style = "CAL, CAL_FOCUS"
FieldType = "select"
```

It is defined as a 7×6 select field by **Cols = 7 and Rows = 6**.

: the display allows for increasing/decreasing the day, month, and year.
The operations should be clear from the screen. You can increase or decrease the day, month, and year using the following operations.

Move the cursor and press "Enter."
Direct key input ("d", "D", "T", "m", "M", "y", "Y")
PF keys (F2 to F8)
Mouse selection

---
### (4) TestApp (examples/04_testapp/testapp.go)
 Finally, we present an application that combines the three programs discussed so far.

![TESTAPP](/imsges/testapp.png)

This program performs the following processes:

・ Error checking of input data

```
   if n == "I" {
    	msg, num := m.errCheck()
		if num > NO_ERROR {
			m.panel.Store(msg, "ERR_MSG")
			m.panel.SelectFocus = num
                :
```
As in this example, to display error message when an error occurs and transfer control to the relevant field, assign the field number to **SelectFocus**.

・ Selection input from the <List> field

・ Selection input from the <Date> field

Subprograms are launched as follows:

```
godate := &GoDate{}
   :

rs := godate.Run(time.Now())
```
This will pop up a program that will extract the data you need.