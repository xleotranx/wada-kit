package main

import (
	"github.com/xleotranx/wada-kit/examples/starter/internal/documents"
	documentplugin "github.com/xleotranx/wada-kit/examples/starter/internal/plugins/documents"
	"github.com/xleotranx/wada-kit/plugins/inspector"
	"github.com/xleotranx/wada-kit/plugins/settings"
	"github.com/xleotranx/wada-kit/shell"
)

func main() {
	shell.Run(shell.App{
		Name: "Starter", Folder: "wada-kit-starter", Env: "STARTER",
		Width: 900, Height: 640, MinWidth: 600, MinHeight: 400,
	}, documentplugin.New(documents.Samples()), inspector.New, settings.New)
}
