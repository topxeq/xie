package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/topxeq/tk"
	"github.com/topxeq/xie/go/xie"
)

func runInteractiveShell() int {
	var following bool
	var source string
	scanner := bufio.NewScanner(os.Stdin)

	vmT := xie.NewXie()

	vmT.SetVar("argsG", os.Args)

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
		}

		retG := ""

		lrs := vmT.Load(source)

		if tk.IsErrStr(lrs) {
			fmt.Println("failed to load source")
			continue
		}

		rs := vmT.Run(tk.StrToInt(lrs))

		noResultT := (rs == "TXERROR:no result")

		if tk.IsErrStr(rs) && !noResultT {
			fmt.Fprintln(os.Stderr, "failed to run: "+tk.GetErrStr(rs))
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
			fmt.Fprintln(os.Stderr, "ReadString error:", err)
			return 12
		}
	}

	return 0
}

func main() {

	if len(os.Args) < 2 {

		runInteractiveShell()

		// tk.Pl("no input")
		return
	}

	// tk.Pln(os.Args[1])

	fcT := tk.LoadStringFromFile(os.Args[1])

	rs := xie.RunCode(fcT, nil, os.Args...)
	if rs != "TXERROR:no result" {
		tk.Pl("%v", rs)
	}
}
