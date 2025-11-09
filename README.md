# Taps

 **Taps** stands for "**Tcell and TOML based Terminal Application Script**" and is a framework for creating **TUI** applications in **Golang**. It has the following features.

**Japanese** is [here](README_JP.md) 

## (1) Screen Creation with TOML
Many **UI** tools use a method of stacking objects on the screen. While this method is versatile, it has the problem of not being able to easily check the screen image. Compilation will not pass unless you fully understand the specifications for component generation and object associations, so the initial learning curve is significant.

In **Taps**, you can create a format by using **TOML** for screen definitions. **TOML** is generally a format used for program configuration files, **Taps** uses it for screen definitions, making it easy to define screens.
For example, you can define a screen in the following format:
```
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
   :

```
---
## (2) Simple Programming Style
 Many UI apps implement behavior by inserting logic corresponding to input field and key events in various places. This format often results in code being scattered throughout, hindering readability.

 Taps aims to avoid this event-driven structure and instead aims for a simple program structure.
Each event is absorbed by the **Read** function, and only when a specific key is pressed, the key code and its field name returned to the program. This results in a simple code flowing from top to bottom.

```
	m.panel.Store("Test Data", "E01")
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

---
## (3) Tutorial
 For further explanation, please refer to the example program [Tutorial](TUTORIAL.md) in the **examples** directory.

