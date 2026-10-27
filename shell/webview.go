package shell

import (
	"context"

	"github.com/xleotranx/wada-kit/contract/shell"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// webPage is a mygo.WebView of the shell's window, as plugins see it.
type webPage struct{ v *mygo.WebView }

var _ shell.WebView = webPage{}

func newWebPage(win *mygo.Window) (shell.WebView, error) {
	v, err := win.NewWebView(mygo.WebViewOptions{})
	if err != nil {
		return nil, err
	}
	return webPage{v}, nil
}

func (p webPage) Show(c *ui.Context) ui.Element { return ui.WebView(c, p.v) }
func (p webPage) LoadFile(path string) error    { return p.v.Page().LoadFile(path) }
func (p webPage) Reload()                       { p.v.Page().ReloadIgnoringCache() }
func (p webPage) Capture() ([]byte, error)      { return p.v.CapturePage() }
func (p webPage) Close()                        { p.v.Destroy() }

func (p webPage) Eval(ctx context.Context, code string) (any, error) {
	return p.v.Page().EvalContext(ctx, code)
}
