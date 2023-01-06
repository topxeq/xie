//go:build !linux && !darwin
// +build !linux,!darwin

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/topxeq/dlgs"
	"github.com/topxeq/xie"

	// "github.com/topxeq/go-sciter"
	// "github.com/topxeq/go-sciter/window"
	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
	"github.com/topxeq/tk"

	"github.com/kbinani/screenshot"

	"github.com/jchv/go-webview2"
)

// ...
// "github.com/jchv/go-webview2"
// "github.com/jchv/go-webview2/pkg/edge"
// )

// func main() {
// dataPath, _ := filepath.Abs("./userdata")
// w := webview2.NewWithOptions(webview2.WebViewOptions{
// 	Debug:     true,
// 	AutoFocus: true,
// 	DataPath:  dataPath,
// 	WindowOptions: webview2.WindowOptions{
// 		Title: "go-webview2 Example",
// 	},
// })
// if w == nil {
// 	log.Fatalln("Failed to load webview.")
// }
// defer w.Destroy()

// // update window icon
// w32.SendMessage(w.Window(), 0x0080, 1, w32.ExtractIcon(os.Args[0], 0))

// w.SetSize(800, 600, webview2.HintNone)

// chromium := getChromium(w)

// folderPath, _ := filepath.Abs("./public")
// webview := chromium.GetICoreWebView2_3()
// webview.SetVirtualHostNameToFolderMapping(
// 	"app.assets", folderPath,
// 	edge.COREWEBVIEW2_HOST_RESOURCE_ACCESS_KIND_DENY_CORS,
// )
// w.Navigate("http://app.assets/index.html")

// w.Run()
// }

// func getChromium(w webview2.WebView) *edge.Chromium {
// browser := reflect.ValueOf(w).Elem().FieldByName("browser")
// browser = reflect.NewAt(browser.Type(), unsafe.Pointer(browser.UnsafeAddr())).Elem()
// return browser.Interface().(*edge.Chromium)
// }

func guiHandler(actionA string, objA interface{}, dataA interface{}, paramsA ...interface{}) interface{} {
	switch actionA {
	case "init":
		rs := initGUI()
		return rs
	case "lockOSThread":
		runtime.LockOSThread()
		return nil
	case "showInfo":
		if len(paramsA) < 2 {
			return fmt.Errorf("参数不够")
		}
		return showInfoGUI(tk.ToStr(paramsA[0]), tk.ToStr(paramsA[1]), paramsA[2:]...)

	case "showError":
		if len(paramsA) < 2 {
			return fmt.Errorf("参数不够")
		}
		return showErrorGUI(tk.ToStr(paramsA[0]), tk.ToStr(paramsA[1]), paramsA[2:]...)

	case "getConfirm":
		if len(paramsA) < 2 {
			return fmt.Errorf("参数不够")
		}
		return getConfirmGUI(tk.ToStr(paramsA[0]), tk.ToStr(paramsA[1]), paramsA[2:]...)
	case "getActiveDisplayCount":
		return screenshot.NumActiveDisplays()
	case "getScreenResolution":
		var paraArgsT []string = []string{}

		for i := 0; i < len(paramsA); i++ {
			paraArgsT = append(paraArgsT, tk.ToStr(paramsA[i]))
		}

		pT := objA.(*xie.XieVM)

		formatT := pT.GetSwitchVarValue(paraArgsT, "-format=", "")

		idxStrT := pT.GetSwitchVarValue(paraArgsT, "-index=", "0")

		idxT := tk.StrToInt(idxStrT, 0)

		rectT := screenshot.GetDisplayBounds(idxT)

		if formatT == "" {
			return []interface{}{rectT.Max.X, rectT.Max.Y}
		} else if formatT == "raw" || formatT == "rect" {
			return rectT
		} else if formatT == "json" {
			return tk.ToJSONX(rectT, "-sort")
		}

		return []interface{}{rectT.Max.X, rectT.Max.Y}
	case "newWindowSciter":
		if len(paramsA) < 3 {
			return fmt.Errorf("参数不够")
		}

		var paraArgsT []string = []string{}

		for i := 3; i < len(paramsA); i++ {
			paraArgsT = append(paraArgsT, tk.ToStr(paramsA[i]))
		}

		fromFileT := tk.IfSwitchExistsWhole(paraArgsT, "-fromFile")

		titleT := tk.ToStr(paramsA[0])

		rectStrT := tk.ToStr(paramsA[1]) //tk.GetSwitchI(paramsA, "-rect=", "")

		var rectT *sciter.Rect

		if rectStrT == "" {
			rectT = sciter.DefaultRect
		} else {
			objT, errT := tk.FromJSON(rectStrT)

			if errT != nil {
				return fmt.Errorf("窗口矩阵位置大小解析错误：%v", errT)
			}

			var aryT []int

			switch nv := objT.(type) {
			case []int:
				aryT = nv
			case []float64:
				if len(nv) < 4 {
					return fmt.Errorf("窗口矩阵位置大小解析错误：%v", "数据个数错误")
				}
				aryT = []int{tk.ToInt(nv[0]), tk.ToInt(nv[1]), tk.ToInt(nv[2]), tk.ToInt(nv[3])}
			case []interface{}:
				if len(nv) < 4 {
					return fmt.Errorf("窗口矩阵位置大小解析错误：%v", "数据个数错误")
				}
				aryT = []int{tk.ToInt(nv[0]), tk.ToInt(nv[1]), tk.ToInt(nv[2]), tk.ToInt(nv[3])}
			}

			rectT = &sciter.Rect{Left: int32(aryT[0]), Top: int32(aryT[1]), Right: int32(aryT[0] + aryT[2]), Bottom: int32(aryT[1] + aryT[3])}

		}

		w, errT := window.New(sciter.DefaultWindowCreateFlag, rectT)

		if errT != nil {
			return fmt.Errorf("创建窗口失败：%v", errT)
		}

		w.SetOption(sciter.SCITER_SET_SCRIPT_RUNTIME_FEATURES, sciter.ALLOW_EVAL|sciter.ALLOW_SYSINFO|sciter.ALLOW_FILE_IO|sciter.ALLOW_SOCKET_IO)

		w.SetTitle(titleT)

		htmlT := tk.ToStr(paramsA[2])

		baseUrlT := tk.GetSwitch(paraArgsT, "-baseUrl=", "")

		// tk.Pln(fromFileT, htmlT, baseUrlT, tk.PathToURI("."))

		if fromFileT {
			htmlNewT, errT := filepath.Abs(htmlT)
			if errT == nil {
				htmlT = htmlNewT
			}

			errT = w.LoadFile(htmlT)

			if tk.IsErrX(errT) {
				return fmt.Errorf("从文件（%v）创建窗口失败：%v", htmlT, errT)
			}
		} else {
			htmlT := tk.ToStr(paramsA[2])

			if baseUrlT == "." {
				baseUrlT = tk.PathToURI(".") + "/basic.html"
			}

			w.LoadHtml(htmlT, baseUrlT)
		}

		var handlerT tk.TXDelegate

		handlerT = func(actionA string, objA interface{}, dataA interface{}, paramsA ...interface{}) interface{} {
			switch actionA {
			case "show":
				w.Show()
				w.Run()
				return nil
			case "setDelegate":
				var deleT tk.QuickDelegate = paramsA[0].(tk.QuickDelegate)

				w.DefineFunction("delegateDo", func(args ...*sciter.Value) *sciter.Value {
					// args是SciterJS中调用谢语言函数时传入的参数
					// 可以是多个，谢语言中按位置索引进行访问
					strT := args[0].String()

					rsT := deleT(strT)

					// 最后一定要返回一个值，空字符串也可以
					return sciter.NewValue(rsT)
				})

				return nil
			case "call":
				len1T := len(paramsA)
				if len1T < 1 {
					return fmt.Errorf("参数不够")
				}

				if len1T > 1 {
					aryT := make([]*sciter.Value, 0, 10)

					for i := 1; i < len1T; i++ {
						aryT = append(aryT, sciter.NewValue(paramsA[i]))
					}

					rsT, errT := w.Call(tk.ToStr(paramsA[0]), aryT...)

					if errT != nil {
						return fmt.Errorf("调用方法时发生错误：%v", errT)
					}

					return rsT.String()
				}

				rsT, errT := w.Call(tk.ToStr(paramsA[0]))

				if errT != nil {
					return fmt.Errorf("调用方法时发生错误：%v", errT)
				}

				return rsT.String()
			default:
				return fmt.Errorf("未知操作：%v", actionA)
			}

			return nil
		}

		// w.Show()
		// w.Run()

		return handlerT

	case "newWindow":
		// if len(paramsA) < 3 {
		// 	return fmt.Errorf("参数不够：%v", paramsA)
		// }

		var paraArgsT []string = []string{}

		for i := 0; i < len(paramsA); i++ {
			paraArgsT = append(paraArgsT, tk.ToStr(paramsA[i]))
		}

		titleT := tk.GetSwitch(paraArgsT, "-title=", "dialog")
		widthT := tk.GetSwitch(paraArgsT, "-width=", "800")
		heightT := tk.GetSwitch(paraArgsT, "-height=", "600")
		iconT := tk.GetSwitch(paraArgsT, "-icon=", "2")
		debugT := tk.IfSwitchExistsWhole(paraArgsT, "-debug")
		centerT := tk.IfSwitchExistsWhole(paraArgsT, "-center")
		fixT := tk.IfSwitchExistsWhole(paraArgsT, "-fix")
		maxT := tk.IfSwitchExistsWhole(paraArgsT, "-max")
		minT := tk.IfSwitchExistsWhole(paraArgsT, "-min")

		w := webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     debugT,
			AutoFocus: true,
			WindowOptions: webview2.WindowOptions{
				Title:  titleT,
				Width:  uint(tk.ToInt(widthT, 800)),
				Height: uint(tk.ToInt(heightT, 600)),
				IconId: uint(tk.ToInt(iconT, 2)), // icon resource id
				Center: centerT,
			},
		})

		if w == nil {
			return fmt.Errorf("创建窗口失败：%v", "N/A")
		}

		windowStyleT := webview2.HintNone

		if fixT {
			windowStyleT := webview2.HintFixed
		}

		if maxT {
			windowStyleT := webview2.HintMax
		}

		if minT {
			windowStyleT := webview2.HintMin
		}

		w.SetSize(tk.ToInt(widthT, 800), tk.ToInt(heightT, 600), windowStyleT)

		var handlerT tk.TXDelegate

		handlerT = func(actionA string, objA interface{}, dataA interface{}, paramsA ...interface{}) interface{} {
			switch actionA {
			case "show":
				w.Run()
				return nil
			case "navigate":
				len1T := len(paramsA)
				if len1T < 1 {
					return fmt.Errorf("参数不够")
				}

				if len1T > 0 {
					w.Navigate(tk.ToStr(paramsA[0]))
				}

				return nil
			case "setHtml":
				len1T := len(paramsA)
				if len1T < 1 {
					return fmt.Errorf("参数不够")
				}

				if len1T > 0 {
					w.SetHtml(tk.ToStr(paramsA[0]))
				}

				return nil
			case "call", "eval":
				len1T := len(paramsA)
				if len1T < 1 {
					return fmt.Errorf("参数不够")
				}

				if len1T > 0 {
					w.Dispatch(func() {
						w.Eval(tk.ToStr(paramsA[0]))
					})
				}

				return nil
			case "close":
				w.Destroy()
				return nil
			case "setDelegate":
				var deleT tk.QuickVarDelegate = paramsA[0].(tk.QuickVarDelegate)

				w.Bind("delegateDo", func(args ...interface{}) interface{} {
					// args是WebView2中调用谢语言函数时传入的参数
					// 可以是多个，谢语言中按位置索引进行访问
					// strT := args[0].String()

					rsT := deleT(args...)

					// 最后一定要返回一个值，空字符串也可以
					return rsT
				})

				return nil
			// case "call":
			// 	len1T := len(paramsA)
			// 	if len1T < 1 {
			// 		return fmt.Errorf("参数不够")
			// 	}

			// 	if len1T > 1 {
			// 		aryT := make([]*sciter.Value, 0, 10)

			// 		for i := 1; i < len1T; i++ {
			// 			aryT = append(aryT, sciter.NewValue(paramsA[i]))
			// 		}

			// 		rsT, errT := w.Call(tk.ToStr(paramsA[0]), aryT...)

			// 		if errT != nil {
			// 			return fmt.Errorf("调用方法时发生错误：%v", errT)
			// 		}

			// 		return rsT.String()
			// 	}

			// 	rsT, errT := w.Call(tk.ToStr(paramsA[0]))

			// 	if errT != nil {
			// 		return fmt.Errorf("调用方法时发生错误：%v", errT)
			// 	}

			// 	return rsT.String()
			default:
				return fmt.Errorf("未知操作：%v", actionA)
			}

			return nil
		}

		// w.Show()
		// w.Run()

		return handlerT

	default:
		return fmt.Errorf("未知方法")
	}

	return ""
}

func initGUI() error {
	applicationPathT := tk.GetApplicationPath()

	osT := tk.GetOSName()

	if tk.Contains(osT, "inux") {
	} else if tk.Contains(osT, "arwin") {
	} else {
		_, errT := exec.LookPath("sciter.dll")

		// tk.Pln("LookPath", errT)

		if errors.Is(errT, exec.ErrDot) {
			errT = nil
		}

		if errT != nil {
			if tk.IfFileExists("sciter.dll") || tk.IfFileExists(filepath.Join(applicationPathT, "sciter.dll")) {

			} else {
				tk.Pl("初始化WEB图形界面环境……")
				rs := tk.DownloadFile("http://xie.topget.org/pub/sciter.dll", applicationPathT, "sciter.dll")

				if tk.IsErrorString(rs) {
					return fmt.Errorf("初始化图形界面编程环境失败")
				}
			}
		}
	}

	// dialog.Do_init()
	// window.Do_init()

	return nil
}

func showInfoGUI(titleA string, formatA string, messageA ...interface{}) interface{} {
	rs, errT := dlgs.Info(titleA, fmt.Sprintf(formatA, messageA...))

	if errT != nil {
		return errT
	}

	return rs
}

func getConfirmGUI(titleA string, formatA string, messageA ...interface{}) interface{} {
	flagT, errT := dlgs.Question(titleA, fmt.Sprintf(formatA, messageA...), true)
	if errT != nil {
		return errT
	}

	return flagT
}

func showErrorGUI(titleA string, formatA string, messageA ...interface{}) interface{} {
	rs, errT := dlgs.Error(titleA, fmt.Sprintf(formatA, messageA...))
	if errT != nil {
		return errT
	}

	return rs
}
