　**taps** を利用するプログラムは、入力パネルを **TOML** 形式で指定します。具体的なプログラムを例に説明しましょう。TOMLは主に設定ファイルに使用されるフォーマットの1つです。詳細は、<a href="https://ja.wikipedia.org/wiki/TOML" target="_blank">こちら</a>などを参照してください。

---
## [1] 入力プログラムの例 (**examples/01_input/input.go**)

　下記のような入力画面があるとします。

![INPUT](/imsges/input.png)

### (1) パネルの定義
　最初にフィールド属性定義がありますが、これは後述するとして、"**var doc**" 以下を見てください。
まずは、パネルの定義をします。**StartX**、**StartY** が始点、**EndX**、**EndY** が終点を示します。
"**9999**" には特別な意味があり、ターミナルサイズに置き換わります。例えば、サイズ **80x24** のターミナルならば、**EndX** は "79"、**EndY** は "23" となります。これにより、ターミナルのサイズによる違いを吸収するというわけです。 

```
	var doc = `
StartX = 0
StartY = 0
EndX = 9999
EndY = 9999
```

### (2) フィールドの定義
　続けてフィールドの設定を行います。下記のように、**[[Field]]** タグを使用して、各フィールドを定義していきます。

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
　
この例では、最初の表示フィールドと入力フィールドを設定しています。フィールドタイプは、**FieldType** で指定します。tapsでサポートしているフィールドタイプは、"**label**"、"**edit**"、"**select**" の3種類のみです。

また、**E02**フィールドのように、**Attr = "N" **を指定すると、数字入力のみ("+"、"-"、"."を含む)が可能なフィールドとなります。この例では、**DataLen = 6** でデータの桁数も制限しています。

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

### (3) フィールドのスタイル定義
　フィールドスタイルとは表示される際のアトリビュートのことで、単に表示されている場合と、フォーカスが当たっている場合の2つについて指定することができます。

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
1項目目に、Styleで指定した文字列、2番めがForeGroundの属性、3番めにBackGroundの属性を指定します。色属性は、tcellが定数定義しているものなので、<a href="https://github.com/gdamore/tcell/blob/main/color.go" target="_blank">こちら</a>を参照ください。

### (4) パネルの登録
　すべてのフィールドを定義して、下記のように **NewPanel**関数にパラメータとして渡します。

```
func InputPanel() *taps.Panel {
	 :
	 :
	return taps.NewPanel(doc, styleMatrix, "")
}
```

### (5) 入力フィールドの編集
　入力フィールドでは下記のキーを使用することができます。

|Key           |function|
|--------------|---------|
|ESC           |パネルの終了|
|CTRL+B(ArrowR)|カーソルを右に移動|
|CTRL+F(ArrowL)|カーソルを右に移動|
|TAB           |フィールド間の移動|
|CTRL+A        |カーソルを行頭に移動|
|CTRL+E        |カーソルを行末に移動|
|CTRL+K        |カーソル以下を削除|
|CTRL+D(DEL)   |カーソル位置を削除|
|CTRL+H(BS)    |カーソル前を削除。マルチラインの行頭では行連結|
|ENTER         |"edit"の場合、次のフィールドへ。マルチラインでは行分割|

フィールドを **Disbale/Enable**、**Browse**モードに変更するには、関数を使用します。

```
m.panel.SetDisabled("E01")
m.panel.SetEnabled("E01")

m.panel.SetBrowseMode("E03", true)
m.panel.SetBrowseMode("E03", false)
```

### (6) プログラミングパターン
　通常のアプリケーションでは、複数のコードを使用することになると想定されるので、まずは **struct** を定義し、そこに使用するパネルを格納しておきます。


```
type Input struct {
	panel *taps.Panel
}

func  (m *Input) Run() {
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

　次に、データをパネル上のフィールドに設定しておきます。上記では、**E01** フィールドに、定数 "Test Data" を格納しています。フィールドにデータをセットする命令は **Store** となります。
その後に **Say** で画面に表示、**Read** で入力データを受け取るという手順が一般的でしょう。**Read** は、入力されたキーコードとフィールド名を返すので、その内容によってロジックを分岐させます。
　上の例では、"**Escape**"キーが押されたか、フィールド "Q" がセレクトされた場合、このループからブレイクすることになります。
また、フィールド "D" がセレクトされた場合には、**Get** によって、その内容を読み取っています。このようなパターンで、プログラムを構成することになります。

---
## [2] リスト形式の例 (**examples/02_list/list.go**)
　リスト形式の構造について説明します。一般にListBoxと呼ばれる形式で、リスト表示した一覧からデータを選択する際に利用されることになるでしょう。

![LIST](/imsges/list.png)

　パネルは、ターミナルの一部を指定します。また選択項目には、**Rows = 12** という行数を指定すれば、上記のような画面を定義できます。 

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
この "**LIST**"フィールドには、**StoreList** で、stringの配列を渡すことで、実際のデータを設定できます。

```
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
        :
```
**example**では、適当なデータをセットしておきました。

---
### (3) GoDate (**examples/03_godate/godate.go**)
　次は少し複雑な構成のプログラムを見てみましょう。いわゆる **Calendar** プログラムです。

![GODATE](/imsges/godate.png)

パネル定義に "**ExitKey**" という設定を追加しています。

```
StartX = 10
StartY = 2
EndX = 48
EndY = 16
Rect = true
ExitKey = ["F2", "F3", "F4", "F5", "F6", "F7", "F8", "F10", "F12"]
```

　**Read関数**は、通常では "**ESC**" と、 **select**フィールドで "**ENTER**" が叩かれた場合のみ制御を戻します。他のキーは無視されるので、**Read** から抜け出したいキーをここで指定します。
なお、キーは、tcellが定数定義しているものなので、<a href="https://github.com/gdamore/tcell/blob/main/key.go" target="_blank">こちら</a>を参照ください。

　次に、**Calendar** 本体の定義は、下記になります。

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
**Cols =7**、 **Rows = 6** **7 x 6** のselectフィールドとして定義されています。

操作については、画面を見ていただければ解ると思いますが、日、月、年の増減が下記の操作で可能となっています。

カーソルを合わせて、"Enter"
キーを直接入力　("d"、"D"、"T"、"m"、"M"、"y"、"Y")
PFキー(F2からF8)
マウスで選択

---
### (4) TestApp　(examples/04_testapp/testapp.go)
　最後に、ここまでのプログラム3つを組み合わせたアプリケーションを示します。

![TESTAPP](/imsges/testapp.png)

このプログラムでは、下記の処理を行います。

・ 入力データのエラーチェック

```
   if n == "I" {
    	msg, num := m.errCheck()
		if num > NO_ERROR {
			m.panel.Store(msg, "ERR_MSG")
			m.panel.SelectFocus = num
                :
```
この例のように、エラー時にはメッセージを表示し、該当フィールドに制御を移すには、**SelectFocus** にフィールドナンバーを代入します。

・ ＜List＞フィールドからの選択入力

・ ＜Date＞フィールドからの選択入力

サブプログラムの起動は、下記のように実行します。

```
godate := &GoDate{}
   :

rs := godate.Run(time.Now())

```
これにより、プログラムがポップアップし、必要なデータを取り出すことができます。
