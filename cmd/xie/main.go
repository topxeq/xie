package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/topxeq/tk"
	"github.com/topxeq/xie"

	_ "github.com/denisenkom/go-mssqldb"

	_ "github.com/godror/godror"
	_ "github.com/sijms/go-ora/v2"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func test() {
	// fontPaths := findfont.List()
	// for _, path := range fontPaths {
	// 	// fmt.Println(path)
	// 	//楷体:simkai.ttf
	// 	//黑体:simhei.ttf
	// 	if strings.Contains(path, "simhei.ttf") {
	// 		os.Setenv("FYNE_FONT", path)
	// 		break
	// 	}
	// }

	// a := app.New()
	// w := a.NewWindow("Hello今天")

	// hello := widget.NewLabel("Hello Fyne我们!")
	// w.SetContent(container.NewVBox(
	// 	hello,
	// 	widget.NewButton("Hi!", func() {
	// 		hello.SetText("Welcome大家 :)")
	// 	}),
	// ))

	// w.ShowAndRun()
}

func runInteractiveShell() int {
	tk.Pl(`谢语言（Xielang）版本（ver.） %v`, xie.VersionG)
	xie.ShellModeG = true
	xie.SetLeVSilent(true)

	var following bool
	var source string
	scanner := bufio.NewScanner(os.Stdin)

	vmT := xie.NewXie(nil)

	vmT.SetVar("argsG", os.Args)
	vmT.SetVar("全局参数", os.Args)

	var guiHandlerG tk.TXDelegate = guiHandler

	vmT.SetVar("guiG", guiHandlerG)

	for {
		if following {
			source += "\n"
			fmt.Print("  ")
		} else {
			fmt.Print("> ")
		}

		if !scanner.Scan() {
			break
		}
		source += scanner.Text()
		if source == "" {
			continue
		}

		if source == "quit" || source == "退出" {
			break
		} else if source == "#debug" {
			vmT.Debug()
			following = false
			source = ""
			continue
		}

		retG := ""

		lrs := vmT.Load(source)

		if tk.IsErrStr(lrs) {
			following = false
			source = ""
			fmt.Println("载入源码失败：", tk.GetErrStr(lrs))
			continue
		}

		rs := vmT.Run(tk.StrToInt(lrs))

		noResultT := (rs == "TXERROR:no result")

		if tk.IsErrStrX(rs) && !noResultT {
			fmt.Fprintln(os.Stderr, "运行失败："+tk.GetErrStr(rs))
			following = false
			source = ""
			continue
		}

		if !noResultT {
			fmt.Println(retG)
		}

		following = false
		source = ""
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			fmt.Fprintln(os.Stderr, "读取字符串失败：", err)
			return 12
		}
	}

	return 0
}

var muxG *http.ServeMux
var portG = ":80"
var sslPortG = ":443"
var basePathG = "."
var webPathG = "."
var certPathG = "."
var verboseG = false
var verbosePlusG = false
var scriptPathG = ""

var staticFS http.Handler = nil

func serveStaticDirHandler(w http.ResponseWriter, r *http.Request) {
	if staticFS == nil {
		// tk.Pl("staticFS: %#v", staticFS)
		// staticFS = http.StripPrefix("/w/", http.FileServer(http.Dir(filepath.Join(basePathG, "w"))))
		hdl := http.FileServer(http.Dir(webPathG))
		// tk.Pl("hdl: %#v", hdl)
		staticFS = hdl
	}

	old := r.URL.Path

	if verboseG {
		tk.PlNow("URL: %v", r.URL.Path)
	}

	name := filepath.Join(webPathG, path.Clean(old))

	// tk.Pl("name: %v", name)

	info, err := os.Lstat(name)
	if err == nil {
		if !info.IsDir() {
			staticFS.ServeHTTP(w, r)
			// http.ServeFile(w, r, name)
		} else {
			if tk.IfFileExists(filepath.Join(name, "index.html")) {
				staticFS.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		}
	} else {
		http.NotFound(w, r)
	}

}

func startHttpsServer(portA string) {
	if !tk.StartsWith(portA, ":") {
		portA = ":" + portA
	}

	err := http.ListenAndServeTLS(portA, filepath.Join(certPathG, "server.crt"), filepath.Join(certPathG, "server.key"), muxG)
	if err != nil {
		tk.PlNow("启动https服务失败：%v", err)
	}

}

func genFailCompact(titleA, msgA string, optsA ...string) string {
	mapT := map[string]string{
		"msgTitle":    titleA,
		"msg":         msgA,
		"subMsg":      "",
		"actionTitle": "返回",
		"actionHref":  "javascript:history.back();",
	}

	var fileNameT = "fail.html"

	if tk.IfSwitchExists(optsA, "-compact") {
		fileNameT = "failcompact.html"
	}

	tmplT := tk.LoadStringFromFile(filepath.Join(basePathG, "tmpl", fileNameT))

	if tk.IsErrStr(tmplT) {
		tmplT = `<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<meta http-equiv="content-type" content="text/html; charset=UTF-8" />
			<meta name='viewport' content='width=device-width; initial-scale=1.0; maximum-scale=4.0; user-scalable=1;' />
		</head>
		
		<body>
			<div>
				<h2>TX_msgTitle_XT</h2>
				<p>TX_msg_XT</p>
			</div>
			<div>
				<p>TX_subMsg_XT</p>
			</div>
			<div style="display: none;">
				<p>
					<a href="TX_actionHref_XT">TX_actionTitle_XT</a>
				</p>
			</div>
		</body>
		
		</html>`
	}

	tmplT = tk.ReplaceHtmlByMap(tmplT, mapT)

	return tmplT
}

func doXms(res http.ResponseWriter, req *http.Request) {
	if res != nil {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "*")
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	if req != nil {
		req.ParseForm()
		req.ParseMultipartForm(1000000000000)
	}

	reqT := tk.GetFormValueWithDefaultValue(req, "xms", "")

	if verboseG {
		tk.Pl("请求URI： %v", req.RequestURI)
	}

	if reqT == "" {
		if tk.StartsWith(req.RequestURI, "/xms") {
			reqT = req.RequestURI[4:]
		}
	}

	tmps := tk.Split(reqT, "?")
	if len(tmps) > 1 {
		reqT = tmps[0]
	}

	if tk.StartsWith(reqT, "/") {
		reqT = reqT[1:]
	}

	var paraMapT map[string]string
	var errT error

	vo := tk.GetFormValueWithDefaultValue(req, "vo", "")

	if vo == "" {
		paraMapT = tk.FormToMap(req.Form)
	} else {
		paraMapT, errT = tk.MSSFromJSON(vo)

		if errT != nil {
			res.Write([]byte(genFailCompact("操作失败", "参数格式错误", "-compact")))
			return
		}
	}

	if verboseG {
		tk.Pl("[%v] REQ: %#v (%#v)", tk.GetNowTimeStringFormal(), reqT, paraMapT)
	}

	toWriteT := ""

	fileNameT := reqT

	if !tk.EndsWith(fileNameT, ".xie") {
		fileNameT += ".xie"
	}

	// fcT := tk.LoadStringFromFile(filepath.Join(basePathG, "xms", fileNameT))
	// absT, _ := filepath.Abs(filepath.Join(basePathG, fileNameT))
	// tk.Pln("loading", absT)
	fcT := tk.LoadStringFromFile(filepath.Join(basePathG, fileNameT))
	if tk.IsErrStr(fcT) {
		res.Write([]byte(genFailCompact("操作失败", tk.GetErrStr(fcT), "-compact")))
		return
	}

	vmT := xie.NewXie(nil)

	vmT.SetVar("paraMapG", paraMapT)
	vmT.SetVar("requestG", req)
	vmT.SetVar("responseG", res)
	vmT.SetVar("reqNameG", reqT)
	vmT.SetVar("basePathG", basePathG)

	// vmT.SetVar("inputG", objA)

	lrs := vmT.Load(fcT)

	contentTypeT := res.Header().Get("Content-Type")

	if tk.IsErrStr(lrs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("操作失败：%v", tk.GetErrStr(lrs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("操作失败", tk.GetErrStr(lrs), "-compact")))
		return
	}

	rs := vmT.Run()

	contentTypeT = res.Header().Get("Content-Type")

	// tk.Pln("contentType:", contentTypeT)

	// if errT != nil {
	// 	if tk.StartsWith(contentTypeT, "text/json") {
	// 		res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("操作失败：%v", tk.GetErrStr(lrs)), req)))
	// 		return
	// 	}

	// 	res.Write([]byte(genFailCompact("操作失败", errT.Error(), "-compact")))
	// 	return
	// }

	if tk.IsErrStr(rs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("操作失败：%v", tk.GetErrStr(rs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("操作失败", tk.GetErrStr(rs), "-compact")))
		return
	}

	toWriteT = rs

	if toWriteT == "TX_END_RESPONSE_XT" {
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	res.Write([]byte(toWriteT))

}

func doXmsContent(res http.ResponseWriter, req *http.Request) {
	if res != nil {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "*")
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	if req != nil {
		req.ParseForm()
		req.ParseMultipartForm(1000000000000)
	}

	reqT := tk.GetFormValueWithDefaultValue(req, "xc", "")

	if verboseG {
		tk.Pl("请求URI： %v", req.RequestURI)
	}

	if reqT == "" {
		if tk.StartsWith(req.RequestURI, "/xc") {
			reqT = req.RequestURI[3:]
		}
	}

	tmps := tk.Split(reqT, "?")
	if len(tmps) > 1 {
		reqT = tmps[0]
	}

	if tk.StartsWith(reqT, "/") {
		reqT = reqT[1:]
	}

	var paraMapT map[string]string
	var errT error

	vo := tk.GetFormValueWithDefaultValue(req, "vo", "")

	if vo == "" {
		paraMapT = tk.FormToMap(req.Form)
	} else {
		paraMapT, errT = tk.MSSFromJSON(vo)

		if errT != nil {
			res.Write([]byte(genFailCompact("操作失败", "参数格式错误", "-compact")))
			return
		}
	}

	if verboseG {
		tk.Pl("[%v] REQ: %#v (%#v)", tk.GetNowTimeStringFormal(), reqT, paraMapT)
	}

	toWriteT := ""

	fileNameT := "doxc"

	if !tk.EndsWith(fileNameT, ".xie") {
		fileNameT += ".xie"
	}

	// fcT := tk.LoadStringFromFile(filepath.Join(basePathG, "xms", fileNameT))
	// absT, _ := filepath.Abs(filepath.Join(basePathG, fileNameT))
	// tk.Pln("loading", absT)
	fcT := tk.LoadStringFromFile(filepath.Join(basePathG, fileNameT))
	if tk.IsErrStr(fcT) {
		res.Write([]byte(genFailCompact("操作失败", tk.GetErrStr(fcT), "-compact")))
		return
	}

	vmT := xie.NewXie(nil)

	vmT.SetVar("paraMapG", paraMapT)
	vmT.SetVar("requestG", req)
	vmT.SetVar("responseG", res)
	vmT.SetVar("reqNameG", reqT)
	vmT.SetVar("basePathG", basePathG)

	// vmT.SetVar("inputG", objA)

	lrs := vmT.Load(fcT)

	contentTypeT := res.Header().Get("Content-Type")

	if tk.IsErrStr(lrs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("操作失败：%v", tk.GetErrStr(lrs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("操作失败", tk.GetErrStr(lrs), "-compact")))
		return
	}

	rs := vmT.Run()

	contentTypeT = res.Header().Get("Content-Type")

	if tk.IsErrStr(rs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("操作失败：%v", tk.GetErrStr(rs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("操作失败", tk.GetErrStr(rs), "-compact")))
		return
	}

	toWriteT = rs

	if toWriteT == "TX_END_RESPONSE_XT" {
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	res.Write([]byte(toWriteT))

}

func RunServer() {
	portG = tk.GetSwitch(os.Args, "-port=", portG)
	sslPortG = tk.GetSwitch(os.Args, "-sslPort=", sslPortG)

	if !tk.StartsWith(portG, ":") {
		portG = ":" + portG
	}

	if !tk.StartsWith(sslPortG, ":") {
		sslPortG = ":" + sslPortG
	}

	basePathG = tk.GetSwitch(os.Args, "-dir=", basePathG)
	webPathG = tk.GetSwitch(os.Args, "-webDir=", basePathG)
	certPathG = tk.GetSwitch(os.Args, "-certDir=", certPathG)

	muxG = http.NewServeMux()

	muxG.HandleFunc("/xms/", doXms)
	muxG.HandleFunc("/xms", doXms)

	muxG.HandleFunc("/xc/", doXmsContent)
	muxG.HandleFunc("/xc", doXmsContent)

	muxG.HandleFunc("/", serveStaticDirHandler)

	tk.PlNow("谢语言微服务框架 版本%v -port=%v -sslPort=%v -dir=%v -webDir=%v -certDir=%v", xie.VersionG, portG, sslPortG, basePathG, webPathG, certPathG)

	if sslPortG != "" {
		tk.PlNow("在端口%v上启动https服务...", sslPortG)
		go startHttpsServer(sslPortG)
	}

	tk.Pl("在端口%v上启动http服务 ...", portG)
	err := http.ListenAndServe(portG, muxG)

	if err != nil {
		tk.PlNow("启动服务失败：%v", err)
	}

}

func main() {

	// tk.Pln(os.Args[1])
	argsT := os.Args

	if tk.IfSwitchExistsWhole(argsT, "-test") {
		test()
		return
	}

	if tk.IfSwitchExistsWhole(argsT, "-version") {
		tk.Pl("谢语言 版本%v", xie.VersionG)
		return
	}

	verboseG = tk.IfSwitchExistsWhole(argsT, "-verbose")

	verbosePlusG = tk.IfSwitchExistsWhole(argsT, "-vv")

	if tk.IfSwitchExistsWhole(argsT, "-server") {
		RunServer()
		return
	}

	ifExampleT := tk.IfSwitchExistsWhole(argsT, "-example")
	ifExamT := tk.IfSwitchExistsWhole(argsT, "-exam")
	ifGoPathT := tk.IfSwitchExistsWhole(argsT, "-gopath")
	ifCloudT := tk.IfSwitchExistsWhole(argsT, "-cloud")
	ifRemoteT := tk.IfSwitchExistsWhole(argsT, "-remote")
	ifClipT := tk.IfSwitchExistsWhole(argsT, "-clip")
	ifLocalT := tk.IfSwitchExistsWhole(argsT, "-local")
	ifViewT := tk.IfSwitchExistsWhole(argsT, "-view")
	ifCompileT := tk.IfSwitchExistsWhole(argsT, "-compile")
	ifPipeT := tk.IfSwitchExistsWhole(argsT, "-pipe")

	ifInExeT := false
	inExeCodeT := ""

	binNameT, errT := os.Executable()
	if errT != nil {
		binNameT = ""
	}

	baseBinNameT := filepath.Base(binNameT)

	if binNameT != "" {
		if !tk.StartsWith(baseBinNameT, "xie") {
			text1T := tk.Trim(`740404`)
			text2T := tk.Trim(`690415`)
			text3T := tk.Trim(`040626`)

			buf1, errT := tk.LoadBytesFromFileE(binNameT)
			if errT == nil {
				re := regexp.MustCompile(text1T + text2T + text3T + `(.*?) *` + text3T + text2T + text1T)
				matchT := re.FindAllSubmatch(buf1, -1)

				if matchT != nil && len(matchT) > 0 {
					codeStrT := string(matchT[len(matchT)-1][1])

					decCodeT := tk.DecryptStringByTXDEF(codeStrT, "topxeq")
					if !tk.IsErrStr(decCodeT) {
						ifInExeT = true
						inExeCodeT = decCodeT
					}

				}
			}
		}
	}

	if !ifInExeT && len(tk.GetAllParameters(argsT)) < 2 {
		// if tk.IsErrX(scriptT) {
		fileListT := tk.GetFileList(".", "-pattern=auto*.xie", "-sort=asc", "-sortKey=Name")

		// tk.Pln(fileListT)
		// }

		var guiHandlerG tk.TXDelegate = guiHandler

		if len(fileListT) > 0 {
			for i, v := range fileListT {

				fcT := tk.LoadStringFromFile(v["Path"])

				if tk.IsErrX(fcT) {
					tk.Pl("载入自动脚本([%v] %v)失败：%v", i, v["Path"], tk.GetErrStrX(fcT))
					return
				}

				scriptPathG = "."

				rs := xie.RunCode(fcT, map[string]interface{}{"guiG": guiHandlerG, "scriptPathG": scriptPathG}, nil, argsT...)
				if rs != "TXERROR:no result" {
					tk.Pl("%v", rs)
				}
			}

			return
		}

		runInteractiveShell()

		// tk.Pl("no input")
		return
	}

	var scriptT string = ""

	filePathT := tk.GetParameterByIndexWithDefaultValue(argsT, 1, "")

	if ifInExeT && inExeCodeT != "" {
		scriptT = inExeCodeT
	} else if ifExampleT {
		if (!tk.EndsWith(filePathT, ".xie")) && (!tk.EndsWith(filePathT, ".谢")) {
			filePathT += ".谢"
		}

		pathT := "http://xie.topget.org/xc/t/c/xielang/example/" + tk.UrlEncode2(filePathT)
		scriptT = tk.DownloadWebPageX(pathT)
		scriptPathG = pathT

	} else if ifExamT {
		if (!tk.EndsWith(filePathT, ".xie")) && (!tk.EndsWith(filePathT, ".谢")) {
			filePathT += ".xie"
		}

		pathT := "http://xie.topget.org/xc/t/c/xielang/example/" + tk.UrlEncode2(filePathT)
		scriptT = tk.DownloadWebPageX(pathT)
		scriptPathG = pathT

	} else if ifGoPathT {
		if (!tk.EndsWith(filePathT, ".xie")) && (!tk.EndsWith(filePathT, ".谢")) {
			filePathT += ".xie"
		}

		filePathT = filepath.Join(tk.GetEnv("GOPATH"), "src", "github.com", "topxeq", "xie", "cmd", "xie", "scripts", filePathT)
		// tk.Pl(filePathT)
		scriptT = tk.LoadStringFromFile(filePathT)
		scriptPathG = filePathT

	} else if ifPipeT {
		// fmt.Println("pipe")
		bufT := bufio.NewReader(os.Stdin)

		b, err := io.ReadAll(bufT)
		if err != nil {
			log.Fatal(err)
		}

		// Prints the data in buffer
		// fmt.Println("s1T", string(b))

		filePathT = "#PIPE"

		scriptT = string(b)

	} else if ifCloudT {
		if (!tk.EndsWith(filePathT, ".xie")) && (!tk.EndsWith(filePathT, ".谢")) {
			filePathT += ".xie"
		}

		basePathT, errT := tk.EnsureBasePath("xie")

		gotT := false

		if errT == nil {
			cfgPathT := tk.JoinPath(basePathT, "cloud.cfg")

			cfgStrT := tk.Trim(tk.LoadStringFromFile(cfgPathT))

			if !tk.IsErrorString(cfgStrT) {
				scriptT = tk.DownloadPageUTF8(cfgStrT+filePathT, nil, "", 30)

				scriptPathG = cfgStrT + filePathT

				gotT = true
			}

		}

		if !gotT {
			scriptT = tk.DownloadPageUTF8(scriptT, nil, "", 30)
			scriptPathG = scriptT
		}

	} else if ifRemoteT {
		scriptPathG = scriptT
		scriptT = tk.DownloadPageUTF8(scriptT, nil, "", 30)

	} else if ifClipT {
		scriptPathG = "clip"
		scriptT = tk.GetClipText()

	} else if ifLocalT {
		if (!tk.EndsWith(filePathT, ".xie")) && (!tk.EndsWith(filePathT, ".谢")) {
			filePathT += ".xie"
		}

		basePathT, _ := tk.EnsureBasePath("xie")

		cfgPathT := tk.JoinPath(basePathT, "local.cfg")

		cfgStrT := tk.Trim(tk.LoadStringFromFile(cfgPathT))

		if tk.IsErrorString(cfgStrT) {
			tk.Pl("获取配置文件信息失败：%v", tk.GetErrorString(cfgStrT))

			return
		}

		// if tk.GetEnv("GOXVERBOSE") == "true" {
		// 	tk.Pl("Try to load script from %v", filepath.Join(localPathT, scriptT))
		// }

		scriptPathG = filepath.Join(cfgStrT, filePathT)

		scriptT = tk.LoadStringFromFile(scriptPathG)
	} else if strings.HasPrefix(filePathT, "http") {
		rsT := tk.DownloadWebPageX(filePathT)
		scriptPathG = filePathT

		if tk.IsErrStr(rsT) {
			scriptT = ""
		} else {
			scriptT = rsT
		}
	} else {
		scriptT = tk.LoadStringFromFile(filePathT)
		scriptPathG = filePathT
	}

	if ifViewT {
		tk.Pl("%v", scriptT)

		return
	}

	if tk.IfSwitchExists(argsT, "-dotest") {
		tk.Pl("codeG: %v", codeG)
		return
	}

	if ifCompileT {
		appPathT, errT := os.Executable()

		tk.CheckError(errT)

		outputT := tk.Trim(tk.GetSwitch(argsT, "-output=", "output.exe"))

		if scriptT == "" {
			tk.Fatalf("代码为空")
		}

		fcT := scriptT

		buf1, errT := tk.LoadBytesFromFileE(appPathT)
		if errT != nil {
			tk.Fatalf("读取主程序文件失败：%v", errT)
		}

		encTextT := tk.EncryptStringByTXDEF(fcT, "topxeq")

		encBytesT := []byte(encTextT)

		lenEncT := len(encBytesT)

		text1T := tk.Trim("740404")
		text2T := tk.Trim("690415")
		text3T := tk.Trim("040626")

		re := regexp.MustCompile(text1T + text2T + text3T + `(.*)` + text3T + text2T + text1T)
		matchT := re.FindSubmatchIndex(buf1)
		if matchT == nil {
			tk.Fatalf("无效的主程序文件")
		}

		bufCodeLenT := matchT[3] - matchT[2]

		var buf3 bytes.Buffer

		if bufCodeLenT < lenEncT {
			buf3.Write(buf1)
			buf3.Write([]byte("74040469" + "0415840215"))
			buf3.Write(encBytesT)
			buf3.Write([]byte("840215690" + "415740404"))
		} else {
			buf3.Write(buf1[:matchT[2]])
			buf3.Write(encBytesT)
			buf3.Write(buf1[matchT[2]+lenEncT:])
		}

		errT = tk.SaveBytesToFileE(buf3.Bytes(), outputT)
		tk.CheckError(errT)

		return

	}

	if strings.HasPrefix(scriptT, "//TXDEF#") {
		scriptT = tk.TKX.DecryptStringByTXDEF(scriptT)

		if tk.IsErrStrX(scriptT) {
			tk.Fatalf("无效的代码")
		}
	}

	if tk.IsErrX(scriptT) {
		fileListT := tk.GetFileList(".", "-pattern=auto*.xie")

		tk.Pln(fileListT)
	}

	var guiHandlerG tk.TXDelegate = guiHandler

	rs := xie.RunCode(scriptT, map[string]interface{}{"guiG": guiHandlerG, "scriptPathG": scriptPathG}, nil, argsT...)
	if rs != "TXERROR:no result" {
		tk.Pl("%v", rs)
	}
}
