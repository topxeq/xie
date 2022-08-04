//go:build nogui
// +build nogui

package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/topxeq/dlgs"
	"github.com/topxeq/go-sciter/window"
	"github.com/topxeq/tk"
)

func guiHandler(actionA string, dataA interface{}, paramsA ...interface{}) interface{} {
	return fmt.Errorf("未设置GUI引擎")
}
