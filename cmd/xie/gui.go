// -build nogui
package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/topxeq/dlgs"
	"github.com/topxeq/go-sciter"
	"github.com/topxeq/go-sciter/window"
	"github.com/topxeq/tk"
)

func guiHandler(actionA string, dataA interface{}, paramsA ...interface{}) interface{} {
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

	case "newWindow":
		if len(paramsA) < 2 {
			return fmt.Errorf("参数不够")
		}

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

		w.LoadHtml(htmlT, "")

		var handlerT tk.TXDelegate

		handlerT = func(actionA string, dataA interface{}, paramsA ...interface{}) interface{} {
			switch actionA {
			case "show":
				w.Show()
				w.Run()
				return nil
			case "setDelegate":
				var deleT tk.QuickDelegate = paramsA[0].(tk.QuickDelegate)

				w.DefineFunction("delegateDo", func(args ...*sciter.Value) *sciter.Value {
					// args是TIScript中调用setResult函数时传入的参数
					// 可以是多个，Gox中按位置索引进行访问
					strT := args[0].String()

					rsT := deleT(strT)

					// 最后一定要返回一个值，空字符串也可以
					return sciter.NewValue(rsT)
				})

				return nil
			case "call":
				if len(paramsA) < 1 {
					return fmt.Errorf("参数不够")
				}

				if len(paramsA) > 1 {
					rsT, errT := w.Call(tk.ToStr(paramsA[0]), sciter.NewValue(tk.ToStr(paramsA[1])))

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

		if errT != nil {
			tk.Pl("Initialzing GUI environment...")
			rs := tk.DownloadFile("http://xie.topget.org/pub/sciter.dll", applicationPathT, "sciter.dll")

			if tk.IsErrorString(rs) {
				return fmt.Errorf("Failed to initialze GUI environment.")
			}
		}
	}

	// dialog.Do_init()
	window.Do_init()

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
