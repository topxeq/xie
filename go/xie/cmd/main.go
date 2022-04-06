package main

import (
	"os"

	"github.com/topxeq/tk"
	"github.com/topxeq/xie/go/xie"
)

func main() {
	if len(os.Args) < 2 {
		tk.Pl("no input")
		return
	}

	// tk.Pln(os.Args[1])

	fcT := tk.LoadStringFromFile(os.Args[1])

	rs := xie.RunCode(fcT, nil, os.Args...)
	if rs != "TXERROR:no result" {
		tk.Pl("%v", rs)
	}
}
