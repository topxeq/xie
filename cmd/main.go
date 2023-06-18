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
	"runtime"
	"strings"

	"github.com/kardianos/service"
	"github.com/topxeq/tk"
	"github.com/topxeq/xie"
	// "tinygo.org/x/bluetooth"
)

// func main() {

// 	// v := tk.Undefined

// 	// tk.Pl("is: %v", tk.IsUndefined(0))

// 	// tk.Pl("is2: %v", tk.IsUndefined(v))

// 	codeT := `
// 	= $s1 qqq

// 	pln abc $s1

// 	= $outG 10
// 	`

// 	// compiledT := xie.Compile(codeT)

// 	// if tk.IsErrX(compiledT) {
// 	// 	tk.Pl("failed to compile: %v", compiledT)
// 	// 	os.Exit(1)
// 	// }

// 	// tk.Pl("compiled: %#v", compiledT)

// 	vmT := xie.NewVM(codeT)

// 	if tk.IsError(vmT) {
// 		tk.Pl("failed to create VM: %v", vmT)
// 		tk.Exit()
// 	}

// 	// rs1 := vmT.LoadCompiled(compiledT.(*xie.CompiledCode))

// 	// tk.Pl("rs1: %v", rs1)
// 	nv := vmT.(*xie.XieVM)

// 	rs := nv.Run(0)

// 	if !tk.IsUndefined(rs) {
// 		tk.Pl("running result: %v", rs)
// 	}

// }

var serviceNameG = "xieService"
var configFileNameG = serviceNameG + ".cfg"
var serviceModeG = false
var currentOSG = ""

type program struct {
	BasePath string
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	// basePathG = p.BasePath
	// logWithTime("basePath: %v", basePathG)
	serviceModeG = true

	go p.run()

	return nil
}

func (p *program) run() {
	go doWork()
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func initSvc() *service.Service {
	if tk.GetOSName() == "windows" {
		currentOSG = "win"
		if tk.Trim(basePathG) == "." || strings.TrimSpace(basePathG) == "" {
			basePathG = "c:\\" + "xie" // serviceNameG
		}
		configFileNameG = serviceNameG + "win.cfg"
	} else {
		currentOSG = "linux"
		if tk.Trim(basePathG) == "." || strings.TrimSpace(basePathG) == "" {
			basePathG = "/" + "xie" //  + serviceNameG
		}
		configFileNameG = serviceNameG + "linux.cfg"
	}

	if !tk.IfFileExists(basePathG) {
		os.MkdirAll(basePathG, 0777)
	}

	tk.SetLogFile(filepath.Join(basePathG, serviceNameG+".log"))

	svcConfigT := &service.Config{
		Name:        serviceNameG,
		DisplayName: serviceNameG,
		Description: serviceNameG + " V" + xie.VersionG,
		Arguments:   []string{"-service"},
	}

	prgT := &program{BasePath: basePathG}
	var s, err = service.New(prgT, svcConfigT)

	if err != nil {
		tk.LogWithTimeCompact("%v unable to init servcie: %v\n", svcConfigT.DisplayName, err)
		return nil
	}

	return &s
}

func Svc() {
	if tk.GetOSName() == "windows" {
		currentOSG = "win"
		if tk.Trim(basePathG) == "." || strings.TrimSpace(basePathG) == "" {
			basePathG = "c:\\" + "xie" // serviceNameG
		}
		configFileNameG = serviceNameG + "win.cfg"
	} else {
		currentOSG = "linux"
		if tk.Trim(basePathG) == "." || strings.TrimSpace(basePathG) == "" {
			basePathG = "/" + "xie" //  + serviceNameG
		}
		configFileNameG = serviceNameG + "linux.cfg"
	}

	if !tk.IfFileExists(basePathG) {
		os.MkdirAll(basePathG, 0777)
	}

	tk.SetLogFile(filepath.Join(basePathG, serviceNameG+".log"))

	defer func() {
		if v := recover(); v != nil {
			tk.LogWithTimeCompact("panic in service: %v", v)
		}
	}()

	tk.DebugModeG = true

	tk.LogWithTimeCompact("%v V%v", serviceNameG, xie.VersionG)
	tk.LogWithTimeCompact("os: %v, basePathG: %v, configFileNameG: %v", runtime.GOOS, basePathG, configFileNameG)
	tk.LogWithTimeCompact("command-line args: %v", os.Args)

	// tk.Pl("os: %v, basePathG: %v, configFileNameG: %v", runtime.GOOS, basePathG, configFileNameG)

	cfgFileNameT := filepath.Join(basePathG, configFileNameG)
	if tk.IfFileExists(cfgFileNameT) {
		fileContentT := tk.LoadSimpleMapFromFile(cfgFileNameT)

		if fileContentT != nil {
			basePathG = fileContentT["xieBasePath"]
		}
	}

	tk.LogWithTimeCompact("Service started.")
	// tk.LogWithTimeCompact("Using config file: %v", cfgFileNameT)

	runAutoRemoveTask := func() {
		for {
			taskFileListT := tk.GetFileList(basePathG, "-pattern=autoRemoveTask*.xie", "-sort=asc", "-sortKey=Name")

			if len(taskFileListT) > 0 {
				for i, v := range taskFileListT {

					fcT := tk.LoadStringFromFile(v["Abs"])

					if tk.IsErrX(fcT) {
						tk.LogWithTimeCompact("failed to load run-then-remove task - [%v] %v: %v", i, v["Abs"], tk.GetErrStrX(fcT))
						continue
					}

					tk.LogWithTimeCompact("running run-then-remove task: %v ...", v["Abs"])

					scriptPathG = v["Abs"]

					rs := xie.RunCode(fcT, nil, map[string]interface{}{"scriptPathG": scriptPathG, "basePathG": basePathG})
					if !tk.IsUndefined(rs) {
						tk.LogWithTimeCompact("task result: %v", rs)
					}

					tk.RemoveFile(v["Abs"])
				}
			}

			tk.Sleep(5.0)

		}

	}

	go runAutoRemoveTask()

	taskFileListT := tk.GetFileList(basePathG, "-pattern=task*.xie", "-sort=asc", "-sortKey=Name")

	if len(taskFileListT) > 0 {
		for i, v := range taskFileListT {

			fcT := tk.LoadStringFromFile(v["Abs"])

			if tk.IsErrX(fcT) {
				tk.LogWithTimeCompact("failed to load auto task - [%v] %v: %v", i, v["Abs"], tk.GetErrStrX(fcT))
				continue
			}

			tk.LogWithTimeCompact("running task: %v ...", v["Abs"])

			scriptPathG = v["Abs"]

			rs := xie.RunCode(fcT, nil, map[string]interface{}{"scriptPathG": scriptPathG, "basePathG": basePathG})
			if !tk.IsUndefined(rs) {
				tk.LogWithTimeCompact("auto task result: %v", rs)
			}
		}
	}

	// c := 0
	for {
		tk.Sleep(60.0)

		// c++
		// tk.Pl("c: %v", c)
		// tk.LogWithTimeCompact("c: %v", c)
	}

}

var exitG = make(chan struct{})

func doWork() {
	serviceModeG = true

	go Svc()

	for {
		select {
		case <-exitG:
			os.Exit(0)
			return
		}
	}
}

func test() {
	iter := tk.NewCompactIterator(10, 5, 10, 2, 0)

	for {
		_, ki, item, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Println(ki, item)
	}

	iter0 := tk.NewCompactIterator(5)

	for {
		_, ki, item, ok := iter0.Next()
		if !ok {
			break
		}
		fmt.Println(ki, item)
	}

	iter1 := tk.NewCompactIterator(5.0, 5, 10, 2.3, 0)

	for {
		_, ki, item, ok := iter1.Next()
		if !ok {
			break
		}
		fmt.Println(ki, item)
	}

	iter2 := tk.NewCompactIterator("abc123", 0, -1, 1, 0)

	for {
		_, k, item, ok := iter2.Next()
		if !ok {
			break
		}
		fmt.Println(k, item)
	}

	iter3 := tk.NewCompactIterator("abc123", -1, 0, -1, 3)

	for {
		_, k, item, ok := iter3.Next()
		if !ok {
			break
		}
		fmt.Println(k, item)
	}

	iter4 := tk.NewCompactIterator([]map[string]string{map[string]string{"a": "avalue", "b": "bvalue", "c": "cvalue"}, map[string]string{"1": "1v", "2": "2v", "3": "3v"}}, 0, -1, 1, 0)

	for {
		_, k, item, ok := iter4.Next()
		if !ok {
			break
		}
		fmt.Println(k, item)
	}

	iter5 := tk.NewCompactIterator(map[string]string{"a": "avalue", "b": "bvalue", "c": "cvalue"}, 0, -1, 1, 0)

	for {
		_, k, item, ok := iter5.Next()
		if !ok {
			break
		}
		fmt.Println(k, item)
	}

	iter6 := tk.NewCompactIterator([]bool{false, true, true, false, true})

	for {
		_, k, item, ok := iter6.Next()
		if !ok {
			break
		}
		fmt.Println(k, item)
	}

	iter7 := tk.NewCompactIterator(map[string]byte{"a": byte(1), "b": byte(2)})

	if iter7 == nil {
		tk.Pl("failed to create iterator")
		return
	}

	for {
		_, k, item, ok := iter7.Next()
		if !ok {
			break
		}
		fmt.Println(k, item)
	}

}

var guiHandlerG tk.TXDelegate

func runInteractiveShell() int {
	tk.Pl(`Xielang(谢语言) V(版本)%v`, xie.VersionG)
	xie.GlobalsG.Vars["ShellModeG"] = true
	xie.GlobalsG.Vars["leSilentG"] = true

	var following bool
	var source string
	scanner := bufio.NewScanner(os.Stdin)

	vm0T := xie.NewVM()

	if tk.IsError(vm0T) {
		tk.Pl("failed to initialize VM(初始化虚拟机失败): %v", tk.GetErrStrX(vm0T))
		os.Exit(1)
	}

	vmT := vm0T.(*xie.XieVM)

	vmT.SetVar(vmT.Running, "argsG", os.Args)

	guiHandlerG = guiHandler

	vmT.SetVar(vmT.Running, "guiG", guiHandlerG)

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

		if source == "quit" {
			break
		} else if source == "#debug" {
			vmT.Debug()
			following = false
			source = ""
			continue
		}

		retG := ""

		originalCodeLenT := vmT.GetCodeLen(vmT.Running)

		lrs := vmT.Load(vmT.Running, source)

		if tk.IsError(lrs) {
			following = false
			source = ""
			fmt.Println("failed to load source code of the script(载入代码失败): ", tk.GetErrStrX(lrs))
			continue
		}

		rs := vmT.Run(originalCodeLenT)

		noResultT := tk.IsUndefined(rs) // == "TXERROR:no result")

		if tk.IsErrX(rs) {
			fmt.Fprintln(os.Stderr, "failed to run(运行失败): "+tk.GetErrStrX(rs))
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
			fmt.Fprintln(os.Stderr, "failed to read char(获取键盘输入失败):", err)
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

// var verboseG = false
// var verbosePlusG = false
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

	if xie.GlobalsG.VerboseLevel > 0 {
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
		tk.PlNow("failed to start https service: %v", err)
	}

}

func genFailCompact(titleA, msgA string, optsA ...string) string {
	mapT := map[string]string{
		"msgTitle":    titleA,
		"msg":         msgA,
		"subMsg":      "",
		"actionTitle": "back",
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

	if xie.GlobalsG.VerboseLevel > 0 {
		tk.Pl("RequestURI: %v", req.RequestURI)
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
			res.Write([]byte(genFailCompact("action failed", "invalid parameter format", "-compact")))
			return
		}
	}

	if xie.GlobalsG.VerboseLevel > 0 {
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
		res.Write([]byte(genFailCompact("action failed", tk.GetErrStr(fcT), "-compact")))
		return
	}

	vmT := xie.NewVMQuick()

	vmT.SetVar(vmT.Running, "paraMapG", paraMapT)
	vmT.SetVar(vmT.Running, "requestG", req)
	vmT.SetVar(vmT.Running, "responseG", res)
	vmT.SetVar(vmT.Running, "reqNameG", reqT)
	vmT.SetVar(vmT.Running, "basePathG", basePathG)

	// vmT.SetVar("inputG", objA)

	lrs := vmT.Load(vmT.Running, fcT)

	contentTypeT := res.Header().Get("Content-Type")

	if tk.IsError(lrs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("action failed: %v", tk.GetErrStrX(lrs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("action failed", tk.GetErrStrX(lrs), "-compact")))
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

	if tk.IsErrX(rs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("action failed: %v", tk.GetErrStrX(rs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("action failed", tk.GetErrStrX(rs), "-compact")))
		return
	}

	toWriteT = tk.ToStr(rs)

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

	if xie.GlobalsG.VerboseLevel > 0 {
		tk.Pl("RequestURI: %v", req.RequestURI)
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
			res.Write([]byte(genFailCompact("action failed", "invalid parameter format", "-compact")))
			return
		}
	}

	if xie.GlobalsG.VerboseLevel > 0 {
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
		res.Write([]byte(genFailCompact("action failed", tk.GetErrStr(fcT), "-compact")))
		return
	}

	vmT := xie.NewVMQuick(nil)

	vmT.SetVar(nil, "paraMapG", paraMapT)
	vmT.SetVar(nil, "requestG", req)
	vmT.SetVar(nil, "responseG", res)
	vmT.SetVar(nil, "reqNameG", reqT)
	vmT.SetVar(nil, "basePathG", basePathG)

	// vmT.SetVar("inputG", objA)

	lrs := vmT.Load(nil, fcT)

	contentTypeT := res.Header().Get("Content-Type")

	if tk.IsErrX(lrs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("action failed: %v", tk.GetErrStrX(lrs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("action failed", tk.GetErrStrX(lrs), "-compact")))
		return
	}

	rs := vmT.Run()

	contentTypeT = res.Header().Get("Content-Type")

	if tk.IsErrX(rs) {
		if tk.StartsWith(contentTypeT, "text/json") {
			res.Write([]byte(tk.GenerateJSONPResponse("fail", tk.Spr("action failed: %v", tk.GetErrStrX(rs)), req)))
			return
		}

		res.Write([]byte(genFailCompact("action failed", tk.GetErrStrX(rs), "-compact")))
		return
	}

	toWriteT = tk.ToStr(rs)

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

	tk.PlNow("Xie micro-service framework V%v -port=%v -sslPort=%v -dir=%v -webDir=%v -certDir=%v", xie.VersionG, portG, sslPortG, basePathG, webPathG, certPathG)

	if sslPortG != "" {
		tk.PlNow("starting https service on port %v...", sslPortG)
		go startHttpsServer(sslPortG)
	}

	tk.Pl("starting http service on port %v...", portG)
	err := http.ListenAndServe(portG, muxG)

	if err != nil {
		tk.PlNow("failed to start service: %v", err)
	}

}

func main() {
	// var bluetoothAdapter = bluetooth.DefaultAdapter

	// errT := bluetoothAdapter.Enable()

	// if errT != nil {
	// 	tk.Pl("enable Bluetooth function failed: %v", errT)
	// 	// exit()
	// }

	// tk.Pln(os.Args[1])
	argsT := os.Args

	if tk.IfSwitchExistsWhole(argsT, "-test") {
		test()
		return
	}

	if tk.IfSwitchExistsWhole(argsT, "-version") {
		tk.Pl("Xielang(谢语言) Version(版本) %v", xie.VersionG)
		return
	}

	xie.GlobalsG.VerboseLevel = 0

	verboseT := tk.IfSwitchExistsWhole(argsT, "-verbose")

	if verboseT {
		xie.GlobalsG.VerboseLevel = 1
	}

	verbosePlusT := tk.IfSwitchExistsWhole(argsT, "-vv")

	if verbosePlusT {
		xie.GlobalsG.VerboseLevel = 2
	}

	if tk.IfSwitchExistsWhole(argsT, "-server") {
		RunServer()
		return
	}

	if tk.IfSwitchExistsWhole(argsT, "-service") {
		tk.Pl("%v V%v is running in service(server) mode. Running the application with argument \"-service\" will cause it running in service mode.\n", serviceNameG, xie.VersionG, serviceNameG, xie.VersionG)
		serviceModeG = true

		s := initSvc()

		if s == nil {
			tk.LogWithTimeCompact("Failed to init service")
			return
		}

		err := (*s).Run()
		if err != nil {
			tk.LogWithTimeCompact("Service \"%s\" failed to run.", (*s).String())
		}

		return
	}

	if tk.IfSwitchExistsWhole(argsT, "-installService") {
		s := initSvc()

		if s == nil {
			tk.Pl("failed to initialize service")
			return
		}

		tk.Pl("installing service \"%v\"...", (*s).String())

		errT := (*s).Install()
		if errT != nil {
			tk.Pl("failed to install service: %v", errT)
			return
		}

		tk.Pl("service installed - \"%s\" .", (*s).String())

		// tk.Pl("启动服务（starting service） \"%v\"...", (*s).String())

		// errT = (*s).Start()
		// if errT != nil {
		// 	tk.Pl("启动服务失败（failed to start）: %v", errT)
		// 	return
		// }

		// tk.Pl("服务已启动（service started） - \"%s\" .", (*s).String())

		return

	}

	if tk.IfSwitchExistsWhole(argsT, "-startService") {
		s := initSvc()

		if s == nil {
			tk.Pl("failed to init service")
			return
		}

		tk.Pl("starting service \"%v\"...", (*s).String())

		errT := (*s).Start()
		if errT != nil {
			tk.Pl("failed to start: %v", errT)
			return
		}

		tk.Pl("service started - \"%s\" ", (*s).String())

		return

	}

	if tk.IfSwitchExistsWhole(argsT, "-stopService") {
		s := initSvc()

		if s == nil {
			tk.Pl("failed to init service")
			return
		}

		errT := (*s).Stop()
		if errT != nil {
			tk.Pl("failed to stop service: %s", errT)
		} else {
			tk.Pl("service stopped - \"%s\" ", (*s).String())
		}

		return

	}

	if tk.IfSwitchExistsWhole(argsT, "-removeService") || tk.IfSwitchExistsWhole(argsT, "-uninstallService") {
		s := initSvc()

		if s == nil {
			tk.Pl("failed to init service")
			return
		}

		errT := (*s).Stop()
		if errT != nil {
			tk.Pl("failed to stop service: %s", errT)
		} else {
			tk.Pl("service stopped - \"%s\" ", (*s).String())
		}

		errT = (*s).Uninstall()
		if errT != nil {
			tk.Pl("failed to remove service: %v", errT)
			return
		}

		tk.Pl("service removed - \"%s\" ", (*s).String())

		return

	}

	if tk.IfSwitchExistsWhole(argsT, "-reinstallService") {
		s := initSvc()

		if s == nil {
			tk.Pl("failed to init service")
			return
		}

		errT := (*s).Stop()
		if errT != nil {
			tk.Pl("failed to stop service: %s", errT)
		} else {
			tk.Pl("service stopped - \"%s\" ", (*s).String())
		}

		errT = (*s).Uninstall()
		if errT != nil {
			tk.Pl("failed to remove service: %v", errT)
		} else {
			tk.Pl("service removed - \"%s\" ", (*s).String())
		}

		tk.Pl("installing service \"%v\"...", (*s).String())

		errT = (*s).Install()
		if errT != nil {
			tk.Pl("failed to install service: %v", errT)
			return
		}

		tk.Pl("service installed - \"%s\" .", (*s).String())

		tk.Pl("starting service \"%v\"...", (*s).String())

		errT = (*s).Start()
		if errT != nil {
			tk.Pl("failed to start: %v", errT)
			return
		}

		tk.Pl("service started - \"%s\" ", (*s).String())

		return

	}

	if tk.IfSwitchExistsWhole(argsT, "-restartService") {
		s := initSvc()

		if s == nil {
			tk.Pl("failed to init service")
			return
		}

		errT := (*s).Stop()
		if errT != nil {
			tk.Pl("failed to stop service: %s", errT)
		} else {
			tk.Pl("service stopped - \"%s\" ", (*s).String())
		}

		tk.Pl("starting service \"%v\"...", (*s).String())

		errT = (*s).Start()
		if errT != nil {
			tk.Pl("failed to start: %v", errT)
			return
		}

		tk.Pl("service started - \"%s\" ", (*s).String())

		return

	}

	// main start

	ifExampleT := tk.IfSwitchExistsWhole(argsT, "-example")
	ifExamT := tk.IfSwitchExistsWhole(argsT, "-exam")
	ifGoPathT := tk.IfSwitchExistsWhole(argsT, "-gopath")
	ifCloudT := tk.IfSwitchExistsWhole(argsT, "-cloud")
	ifRemoteT := tk.IfSwitchExistsWhole(argsT, "-remote")
	ifClipT := tk.IfSwitchExistsWhole(argsT, "-clip")
	ifEditT := tk.IfSwitchExistsWhole(argsT, "-edit")
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

				if len(matchT) > 0 {
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

	if !ifInExeT && len(tk.GetAllParameters(argsT)) < 2 && !ifClipT && !ifEditT {
		// if tk.IsErrX(scriptT) {
		fileListT := tk.GetFileList(".", "-pattern=auto*.xie", "-sort=asc", "-sortKey=Name")

		// tk.Pln(fileListT)
		// }

		guiHandlerG = guiHandler

		if len(fileListT) > 0 {
			for i, v := range fileListT {

				fcT := tk.LoadStringFromFile(v["Abs"])

				if tk.IsErrX(fcT) {
					tk.Pl("failed to load auto-run script([%v] %v): %v", i, v["Abs"], tk.GetErrStrX(fcT))
					return
				}

				scriptPathG = v["Abs"]

				rs := xie.RunCode(fcT, nil, map[string]interface{}{"guiG": guiHandlerG, "scriptPathG": scriptPathG, "basePathG": basePathG}, argsT...)
				if !tk.IsUndefined(rs) {
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
		// if !tk.EndsWith(filePathT, ".xie") {
		// 	filePathT += ".xie"
		// }

		pathT := "http://xie.topget.org/xc/t/c/xielang/example/" + tk.UrlEncode2(filePathT)
		scriptT = tk.DownloadWebPageX(pathT)
		scriptPathG = pathT

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to get script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}
	} else if ifExamT {
		// if !tk.EndsWith(filePathT, ".xie") {
		// 	filePathT += ".xie"
		// }

		pathT := "http://xie.topget.org/xc/t/c/xielang/example/" + tk.UrlEncode2(filePathT)
		scriptT = tk.DownloadWebPageX(pathT)
		scriptPathG = pathT

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to get script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}

	} else if ifGoPathT {
		// if !tk.EndsWith(filePathT, ".xie") {
		// 	filePathT += ".xie"
		// }

		filePathT = filepath.Join(tk.GetEnv("GOPATH"), "src", "github.com", "topxeq", "xie", "cmd", "scripts", filePathT)
		// tk.Pl(filePathT)
		scriptT = tk.LoadStringFromFile(filePathT)
		scriptPathG = filePathT

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to load script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}
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
		// if !tk.EndsWith(filePathT, ".xie") {
		// 	filePathT += ".xie"
		// }

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

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to get script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}
	} else if ifRemoteT {
		// if !tk.EndsWith(filePathT, ".xie") {
		// 	filePathT += ".xie"
		// }

		scriptPathG = filePathT
		// tk.Pl("scriptT: %v", filePathT)
		scriptT = tk.DownloadPageUTF8(filePathT, nil, "", 30)

		if tk.IsErrStrX(scriptT) {
			tk.Pl("failed to load script: %v", tk.GetErrStrX(scriptT))

			return

		}

	} else if ifClipT {
		scriptPathG = "clip"
		scriptT = tk.GetClipText()

	} else if ifLocalT {
		// if !tk.EndsWith(filePathT, ".xie") {
		// 	filePathT += ".xie"
		// }

		basePathT, _ := tk.EnsureBasePath("xie")

		cfgPathT := tk.JoinPath(basePathT, "local.cfg")

		cfgStrT := tk.Trim(tk.LoadStringFromFile(cfgPathT))

		if tk.IsErrorString(cfgStrT) {
			tk.Pl("failed to get config file content: %v", tk.GetErrorString(cfgStrT))

			return
		}

		// if tk.GetEnv("GOXVERBOSE") == "true" {
		// 	tk.Pl("Try to load script from %v", filepath.Join(localPathT, scriptT))
		// }

		scriptPathG = filepath.Join(cfgStrT, filePathT)

		scriptT = tk.LoadStringFromFile(scriptPathG)

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to load script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}
	} else if strings.HasPrefix(filePathT, "http") {
		rsT := tk.DownloadWebPageX(filePathT)
		scriptPathG = filePathT

		if tk.IsErrStr(rsT) {
			scriptT = ""
		} else {
			scriptT = rsT
		}

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to load script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}
	} else if filePathT == "" {
		scriptT = ""
		scriptPathG = ""
	} else {
		scriptT = tk.LoadStringFromFile(filePathT)
		scriptPathG = filePathT

		if tk.IsErrX(scriptT) {
			tk.Pl("failed to load script: %v", tk.GetErrStrX(scriptT))
			tk.Exit()
		}
	}

	if ifViewT {
		tk.Pl("%v", scriptT)

		return
	}

	if ifEditT {
		guiHandlerG = guiHandler

		editCodeT := tk.DownloadWebPageX(`http://xie.topget.org/pub/script/edit.xie`)

		if tk.IsErrX(editCodeT) {
			editCodeT = tk.DecryptStringByTXDEF(editCodeG)
		}

		rs := xie.RunCode(editCodeT, scriptT, map[string]interface{}{"guiG": guiHandlerG, "scriptPathG": scriptPathG}, append(argsT, "-fromInput")...)
		if !tk.IsUndefined(rs) {
			tk.Pl("%v", rs)
		}

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

	guiHandlerG = guiHandler

	rs := xie.RunCode(scriptT, nil, map[string]interface{}{"guiG": guiHandlerG, "scriptPathG": scriptPathG}, argsT...)
	if !tk.IsUndefined(rs) {
		tk.Pl("%v", rs)
	}
}
