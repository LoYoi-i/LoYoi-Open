package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/bitly/go-simplejson"
	"github.com/flopp/go-findfont"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"log"
	"os"
	"strings"
)

// 定义全局变量
var (
	num         int
	isHandle    int
	noHandle    int
	baiduHandle int
	dictPath    string = "D:\\Administrator\\Desktop\\测试\\翻译\\词典"
	appID              = 1 //百度翻译api的id
	password           = ""//百度翻译api的key  需要自己申请
	Url                = "http://api.fanyi.baidu.com/api/trans/vip/translate"
)

type TranslateModel struct {
	Q     string
	From  string
	To    string
	Appid int
	Salt  int
	Sign  string
}
type DictRequest struct {
	TransType string   `json:"trans_type"`
	Source    []string `json:"source"`
	RequestId string   `json:"request_id"`
	Detect    bool     `json:"detect"`
}

// MainShow 主界面函数
func MainShow(w fyne.Window) {
	//var ctrl *beep.Ctrl
	//title := widget.NewLabel("")
	fyFile := widget.NewLabel("路径:")
	dictFile := widget.NewLabel("词典:")
	entry1 := widget.NewEntry() //文本输入框
	//entry1.SetText("E:\\rename_temp2\\123.txt")
	list := widget.NewSelect(DictList(), func(s string) {})
	dia1 := widget.NewButton("打开", func() { //回调函数：打开选择文件对话框
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				log.Println("Cancelled")
				return
			}
			entry1.SetText(reader.URI().Path()) //把读取到的路径显示到输入框中
		}, w)
		//如果不要就是显示所有文件
		//fd.SetFilter(storage.NewMimeTypeFileFilter([]string{"text/*","image/*","application/*","audio/*"})) //打开的文件格式类型
		fd.Show() //控制是否弹出选择文件目录对话框
	}) //打开文件选择器

	text := widget.NewMultiLineEntry() //多行输入组件
	text.SetMinRowsVisible(6)
	//text.Disable()                     //禁用输入框，不能更改数据
	list.PlaceHolder = "请选择一个词典（默认则没有词典）" //默认的提示
	//list.Alignment = 0
	list.SetSelectedIndex(0)
	bt1 := widget.NewButton("test", func() { //  测试按钮
		//text.SetText(strings.Replace(dictPath, "\\", "/", -1) + "/" + list.Selected)
		path := "D:\\Administrator\\Desktop\\测试\\翻译\\词典\\src.json"
		anyMap := ReadDict(strings.Replace(path, "\\", "/", -1))
		fmt.Println(len(anyMap))
	})
	bt2 := widget.NewButton("整合json", func() {
		var s []string // 新建一个切片
		s, _ = GetAllFile(strings.Replace(entry1.Text, "\\", "/", -1), s, "dict")
		text.SetText("检测到是目录，共有" + strconv.Itoa(len(s)) + "个.jdict文件,正在整合中...\n")
		// 返回了一个整合好的文件信息
		text.Text += "翻译已完成，路径：" + Conformity(s, strings.Replace(entry1.Text, "\\", "/", -1))
		text.Refresh()
	})
	//开始更名按钮
	bt3 := widget.NewButton("开始翻译", func() {
		dF := strings.Replace(dictPath, "\\", "/", -1) + "/" + list.Selected // 生成路径
		anyMap := make(map[string]interface{})                               // 定义变量
		if list.Selected != "不需要词典" {
			anyMap = ReadDict(dF) // 如果选择了词典，就读取这个词典
		}
		fmt.Println(len(anyMap))
		if IsDir(entry1.Text) { //如果是文件夹
			var s []string                                                          // 定义一个保存，所有需要翻译文件的变量
			s, _ = GetAllFile(strings.Replace(entry1.Text, "\\", "/", -1), s, "go") //获取文件夹中所有的变量 这个变量保存了，这个文件夹内所有go文件

			for i := 0; i < len(s); i++ {
				text.SetText("检测到是目录，共有" + strconv.Itoa(len(s)) + "个.go文件\n")
				text.Text += "正在翻译第" + strconv.Itoa(i) + "个文件...\n"
				text.Refresh()
				text.Text += "翻译结果:" + StartFy(strings.Replace(s[i], "\\", "/", -1), anyMap) + "\n" // 开始翻译 翻译是传一个需要处理的文件路径，再传一个字典
				text.Refresh()
				//StartFy(strings.Replace(s[i], "\\", "/", -1))
			}
			//text.Text += "翻译已完成！\n"
			text.Refresh()
			dialog.ShowInformation("提示", "你的翻译好了...", w)
		} else { //否则判断文件是否存在
			exists, _ := PathExists(entry1.Text) //如果只是文件，就先判断这个晚间在不在，
			if exists {
				text.SetText("文件存在，正在翻译")
				text.SetText(StartFy(strings.Replace(entry1.Text, "\\", "/", -1), anyMap)) //翻译是传一个需要处理的文件路径，再传一个字典
			} else {
				text.SetText("文件不存在，请重新选择\n")
			}
			text.Text += "翻译已完成！\n"
			text.Refresh()
			dialog.ShowInformation("提示", "你的翻译好了...", w)
		}
	})
	//大概意思就是，在上面创建，各种元素，在中间，来安排布局，自定义布局，最后才然会显示顺序，创建的时候，可以不按顺序创建，最后显示的时候，来定义顺序
	//标题，定义了一个容器来承载
	//head := container.NewCenter()
	//主体，
	v1 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), fyFile, dia1, entry1)
	v4 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), dictFile, bt1, list)
	v5 := container.NewHBox(bt2, bt3)
	v5Center := container.NewCenter(v5)

	ctnt := container.NewVBox(v1, v4, v5Center, text) //控制显示位置顺序
	w.SetContent(ctnt)
}

// ReadDict 读取词典 传字典文件的路径 返回
func ReadDict(dictFile string) map[string]interface{} {
	anyMap := make(map[string]interface{}) //定义保存词典的变量
	//dictFile := PathRoute(ConfJson(strings.Replace(entry1.Text, "\\", "/", -1)), "json") //定义词典文件。传入后缀名，返回和处理文件名字一样的json文件
	exists, _ := PathExists(dictFile) //判断文件是否存在
	if !exists {                      //不存在就直接{},目前，基本上不可能出现不在的情况了
		jsondata := []byte("{}")
		if err := json.Unmarshal(jsondata, &anyMap); err != nil {
			anyMap = make(map[string]interface{})
			fmt.Printf("读取json错误1：%s\n", err)
		}
	} else { //存在就读取词典的数据
		//jsondata := []byte(`{"1":"1"}`)
		filePrt, err := os.Open(dictFile)
		if err != nil {
			fmt.Println("file open failed", err.Error())
		}
		defer func(filePrt *os.File) {
			err := filePrt.Close()
			if err != nil {

			}
		}(filePrt)
		//create json encoder
		docoder := json.NewDecoder(filePrt)
		err = docoder.Decode(&anyMap)
		if err != nil {
			fmt.Println("字典读取错误" + err.Error())
		}
	}
	return anyMap
}

// DictList 词典列表
func DictList() []string {
	var arr []string
	//path := "D:\\Administrator\\Desktop\\测试\\翻译\\词典"
	rd, err := ioutil.ReadDir(dictPath)
	if err != nil {
		fmt.Println("读取目录失败:", err)
		//return s, err
	}
	arr = append(arr, "不需要词典")
	for _, fi := range rd {
		arr = append(arr, fi.Name())
	}
	return arr
}

// Conformity 整合翻译数据
func Conformity(arr []string, fold string) string {
	var data = make(map[string]interface{}) // 定义一个map
	for i := 0; i < len(arr); i++ {
		anyMap := make(map[string]string) //读取需要翻译文件 的内容
		jsondata, err := readAll(arr[i])
		if err != nil {
			jsondata = []byte("{}") //如果出错就返回{}
		}
		if err := json.Unmarshal(jsondata, &anyMap); err == nil { //然后吧数据序列化
			//anyMap = make(map[string]interface{}) //出错就等于{}
			for k, v := range anyMap { //遍历map
				data[k] = v
			}
			err3 := os.Remove(arr[i])
			if err3 != nil {
				fmt.Println("删除失败")
			}
		}
	}
	if len(data) == 0 {
		fmt.Println("没有数据")
		return ConfJson(fold)
	}
	str, err := json.Marshal(data)
	if err != nil {
		fmt.Println("JSON ERR:", err)
	}
	e1 := writeAll(ConfJson(fold), str)
	if e1 != nil {
		fmt.Printf("写入错误：%s\n", e1)
	}
	return ConfJson(fold) // 返回这个文件路径
}

// ConfJson 传入一个文件夹，吧这个文件夹名字的json文件
func ConfJson(fold string) string {
	arr := strings.Split(fold, "/")
	arr = append(arr, arr[len(arr)-1]+".json")
	return strings.Join(arr, "/")
}

// GetAllFile 获取文件内所有文件
func GetAllFile(pathname string, s []string, suffixName string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname) // 读取磁盘的内容
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return s, err
	}
	for _, fi := range rd { // 开始遍历
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s, err = GetAllFile(fullDir, s, suffixName)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s, err
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			if !strings.Contains(fullName, `_test`) { // 如果没有_test 就吧这个路径添加进去
				if suffix(fullName) == suffixName {
					s = append(s, fullName) //向这个数组，添加内容
				}
			}

		}
	}
	return s, nil
}

// 获取文件后缀，文件名
func suffix(s string) string {
	arr := strings.Split(s, ".")
	return arr[len(arr)-1]
}

// IsDir 判断路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// PathExists 判断一个文件或文件夹是否存在
// 输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// PathRoute 路径处理，吧这个路径的后缀改了
func PathRoute(path string, name string) string {
	arr := strings.Split(path, ".") //  去掉最后一个一个数组，或者把最后一个数组替换了
	arr[len(arr)-1] = name          //  字符串数组转字符串返回回去
	return strings.Join(arr, ".")
}

// 读取文件
func readAll(filename string) (data []byte, e error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 写入文件
func writeAll(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

// 字符串转数组
func arrayToString(arr []string) string {
	var result string
	for _, i := range arr { //遍历数组中所有元素追加成string
		result += i + "\n"
	}
	return result
}

// IsChineseChar 判断字符串内是否有中文汉字
func IsChineseChar(str string) bool {
	// 如果包含 \\t 就返回真
	if len(str) < 6 {
		return true
	} // 小于5个字符，就不翻译
	s := make(map[string]string, 0)
	s[""] = str
	t, err := json.Marshal(s) // 需要用json转一下
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
	strJson := strconv.Quote(string(t))
	if strings.Contains(strJson, `\\t`) {
		return true
	}
	//Symbol:= strings.Contains(strJson, `\\t`)
	//"(",")","=","[","]","{","}","<",">","+","-","*","/","^","[0-9]+"
	//如果，字符串包含链接，包含var 包含type 就不翻译
	if strings.Index(str, `https://`) != -1 || strings.Index(str, `http://`) != -1 || strings.Index(str, `var `) != -1 || strings.Index(str, `0x`) != -1 || strings.Index(str, `type `) != -1 || str[0:1] == "|" || str[0:1] == "±" || str[0:1] == "~" || strings.Contains(str, ":=") {
		return true
	}

	//包含中文就直接返回真
	x := 0
	y := 0
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return true
		}
		if unicode.IsLetter(r) == true {
			x++
		}
		if string(r) == `[` || string(r) == `]` || string(r) == `{` || string(r) == `}` || string(r) == `+` || string(r) == `^` || string(r) == `*` || string(r) == `=` || string(r) == `-` || string(r) == `+` || string(r) == `/` {
			y++
		}
	}

	if x == 0 || y > 2 { //没有就是不翻译
		return true
	}

	if strings.Contains(str, `,`) || strings.Contains(str, `.`) || strings.Contains(str, `:`) || strings.Contains(str, `;`) || strings.Contains(str, `'`) {
		return false
	}

	return false
}
func Sy(str1 string, str2 string) bool {
	return strings.Contains(str1, str2)
}

// DictMach 字符串和词典匹配  返回翻译后的值， 和翻以前，和翻译后的数组
func DictMach(key string, dict map[string]interface{}) (value string, arr [2]string) {
	//var s []string
	mach := string([]byte(key)[strings.Index(key, "// ")+3:]) //处理字符串，去掉"// "
	if IsChineseChar(mach) {                                  //如果字符串包含中文，直接返回成功，传进来是啥，返回就是啥
		isHandle++
		//fmt.Println("跳过翻译："+mach)
		return key, [2]string{"null", "null"}
	}
	data, ok := dict[mach].(string) //匹配json词典
	if ok {                         //这后面就是词典匹配成功的
		isHandle++ // 字典有这个值
		fmt.Println("本地翻译：" + data)
		return strings.Replace(key, mach, data, -1), [2]string{"null", "null"}
	} else { //如果没有匹配上，就直接返回key，否则就把字符串匹配后的处理了返回
		dst, e1 := QueryCaiYun([]string{mach})
		//dst, e1 := Bdidu(mach)             //发送到百度进行翻译
		if e1 == nil { //如果翻译成功
			baiduHandle++ //百度处理数量+1
			fmt.Println("网络翻译：" + dst)
			time.Sleep(time.Millisecond * 100) //事件限制，100毫秒
			//翻译成功，就吧原来地替换了，返回回去，然后，再生成一个翻以前和翻译后的数组，返回回去
			return strings.Replace(key, mach, dst, -1), [2]string{mach, dst} //然后返回翻译后的字符串，和翻译的数据
		} else { //没有成功，就是处理失败，这里可能是百度没有正确返回值
			noHandle++
			fmt.Println("网络翻译失败：" + mach)
			return key, [2]string{"null", "null"}
		}
	}
}

// Bdidu 调用百度翻译api
func Bdidu(words string) (x string, err error) {
	//fmt.Printf(words)
	tran := NewTranslateModeler(words, "en", "zh")
	values := tran.ToValues()
	resp, err := http.PostForm(Url, values)
	if err != nil {
		//fmt.Println(err)
		return words, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//fmt.Println(err)
		return words, err
	}
	//txt := string(body)
	//fmt.Println(txt)
	js, err := simplejson.NewJson(body)
	if err != nil {
		//fmt.Println(err)
		return words, err
	}
	dst := js.Get("trans_result").GetIndex(0).Get("dst").MustString()
	if dst == "" {
		fmt.Println("没找到值")
		return words, errors.New("没有值")
	} else {
		//fmt.Println(dst)
		return dst, nil //只有成功了，才会返回空值
	}
}

// QueryCaiYun 彩云翻译
func QueryCaiYun(word []string) (s string, err error) {
	client := &http.Client{} //定义客户
	request := DictRequest{TransType: "en2zh", Source: word, RequestId: "LoYoi", Detect: true}
	buf, err := json.Marshal(request)
	if err != nil {
		fmt.Println("错误：4")
		return word[0], errors.New("没有值")
		//log.Fatal(err)
	}
	req, err := http.NewRequest("POST", "https://api.interpreter.caiyunai.com/v1/translator", bytes.NewReader(buf))
	if err != nil {
		fmt.Println("错误：3")
		//log.Fatal(err) //打印日志，退出程序
		return word[0], errors.New("没有值")
	}
	//设置请求头

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Authorization", "token:") //这里申请自己的token
	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err) //打印日志，退出程序
		fmt.Println(err.Error() + "2")
		return word[0], errors.New("没有值")
	}
	defer resp.Body.Close() //defer 会在函数结束后从后往前触发，Close() 手动关闭 Body流，防止内存资源泄露
	// 读取响应
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Fatal(err) //打印日志，退出程序
		fmt.Println(err.Error() + "1")
		return word[0], errors.New("没有值")
	}
	if resp.StatusCode != 200 { // 防御式编程，判断状态码是否正确
		fmt.Println("错误码：" + strconv.Itoa(resp.StatusCode) + "\n错误信息：" + string(bodyText))
		return word[0], errors.New("没有值")
		//log.Fatal("bad StatusCode:", resp.StatusCode, "body", string(bodyText)) //打印日志，退出程序
	}
	js, err := simplejson.NewJson(bodyText)
	dst := js.Get("target").GetIndex(0).MustString() //.Get("dst").
	//fmt.Println(dst)
	if dst == "" {
		fmt.Println("没找到值")
		return word[0], errors.New("没有值")
	} else {
		return dst, nil //只有成功了，才会返回空值
	}
}

// ToValues 处理url
func (tran TranslateModel) ToValues() url.Values {
	values := url.Values{
		"q":     {tran.Q},
		"from":  {tran.From},
		"to":    {tran.To},
		"appid": {strconv.Itoa(tran.Appid)},
		"salt":  {strconv.Itoa(tran.Salt)},
		"sign":  {tran.Sign},
	}
	return values
}

// NewTranslateModeler 百度翻译请求函数
func NewTranslateModeler(q, from, to string) TranslateModel {
	tran := TranslateModel{
		Q:    q,
		From: from,
		To:   to,
	}
	tran.Appid = appID
	tran.Salt = time.Now().Second()
	content := strconv.Itoa(appID) + q + strconv.Itoa(tran.Salt) + password
	sign := SumString(content) //计算sign值
	tran.Sign = sign
	return tran
}

// SumString 字符串转换函数
func SumString(content string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(content))
	bys := md5Ctx.Sum(nil)
	//bys := md5.Sum([]byte(content))//这个md5.Sum返回的是数组,不是切片哦
	value := hex.EncodeToString(bys)
	return value
}

// CreateDict 调用百度翻译后，创建词典，方便下次使用
func CreateDict(anyMap map[string]string, dictFile string) {
	str, err := json.Marshal(anyMap) // 序列化
	if err != nil {
		fmt.Println("JSON ERR:", err)
	}

	e1 := writeAll(dictFile, str)
	if e1 != nil {
		fmt.Printf("写入错误：%s\n", e1)
	}
}

// StartFy 开始翻译 翻译文件的路径，字典数据
func StartFy(Fyfile string, dict map[string]interface{}) string {
	//读取需要翻译的文件
	num = 0
	isHandle = 0
	noHandle = 0
	baiduHandle = 0
	var arr []string //储存处理完成的数据

	data, err := readAll(Fyfile) //读取需要处理的文件
	if err != nil {
		//fmt.Println("文件读取错误")
		return "读取翻译文件是失败，错误内容：" + err.Error()
	}
	srt := strings.Split(string(data), "\n") //对字符串数组化处理

	anyMap := make(map[string]string) // 这个是用来保存联网翻译后的数据
	//分给为数组后，开始翻译
	for i := 0; i < len(srt); i++ {
		if len(srt[i]) > 2 {
			//if string(srt[i][0:2]) == "//" {
			if strings.Contains(srt[i], "// ") { //这个字符串中，是否有//
				num++
				value, Fy := DictMach(srt[i], dict)     //和词典做匹配
				if Fy[0] != "null" && Fy[1] != "null" { // 如果不是空，就翻译成功，吧内容添加到map
					anyMap[Fy[0]] = Fy[1] //给翻译后的数组添加内容
				}
				arr = append(arr, value) //添加翻译后的字符串
			} else {
				arr = append(arr, srt[i]) //添加成原来的
			}
		} else {
			arr = append(arr, srt[i]) //否则也是把原来的的添加进去
		}
	}

	if len(anyMap) != 0 { // 保存这个文件翻译后的新字典
		CreateDict(anyMap, PathRoute(Fyfile, "dict")) // 这里还是保存为和翻译文件一样的dict文件
	}
	Chmoderr := os.Chmod(Fyfile, 0777) //设置文件权限为可读可写
	if Chmoderr != nil {
		return Chmoderr.Error()
		//fmt.Println(Chmoderr)
	}

	err2 := writeAll(Fyfile, []byte(arrayToString(arr))) // 吧翻译后的内容重新写入
	if err2 != nil {
		return err2.Error()
		//fmt.Printf("写入错误：%s\n", err2)
	}

	err3 := os.Chmod(Fyfile, 0444) //重新设置文件权限为只读
	if err3 != nil {
		return err3.Error()
	}
	//return ("处理完成,共%d行注释，本地已处理%d行，百度翻译%d行，%d行未处理", num, isHandle, baiduHandle, noHandle)
	//最后返回翻译情况，然后就可以开始下个循环了
	return "共" + strconv.Itoa(num) + "行,本地已处理" + strconv.Itoa(isHandle) + "行,百度翻译" + strconv.Itoa(baiduHandle) + "行," + strconv.Itoa(noHandle) + "行未处理"
	//return "处理完成,共" + strconv.Itoa(num) + "行\n本地已处理" + strconv.Itoa(isHandle) + "行\n百度翻译" + strconv.Itoa(baiduHandle) + "行\n" + strconv.Itoa(noHandle) + "行未处理"
}

// 设置字体
func init() {
	fontPaths := findfont.List()
	for _, fontPath := range fontPaths {
		//fmt.Println(fontPath)
		//楷体:simkai.ttf
		//黑体:simhei.ttf
		//微软雅黑：msyh.ttc
		if strings.Contains(fontPath, "AlibabaPuHuiTi-2-75-SemiBold.ttf") {
			err := os.Setenv("FYNE_FONT", fontPath)
			if err != nil {
				return
			}
			break
		}
	}
}

func main() {
	//新建一个app
	a := app.New()
	//设置窗口栏，任务栏图标

	//新建一个窗口
	w := a.NewWindow("自动翻译注释v1.0_test")
	//主界面框架布局
	MainShow(w)
	//尺寸
	w.Resize(fyne.Size{Width: 500, Height: 100})
	//w居中显示
	w.CenterOnScreen()
	//循环运行
	w.ShowAndRun()
	//支持中文
	err := os.Unsetenv("FYNE_FONT")
	if err != nil {
		return
	}
}
