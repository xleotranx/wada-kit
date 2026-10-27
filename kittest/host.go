// Package kittest helps an app's tests: Host fakes what the shell provides
// plugins, ColorsStayInThePalette guards the app's colors and
// PluginsSpawnTheirGoroutines its plugins' goroutines.
package kittest

import (
	"context"

	"github.com/xleotranx/wada-kit/contract/shell"
	"github.com/xleotranx/wada-kit/kernel"

	"github.com/egoist/mygo/ui"
)

// Host provides shell.HostKey and records what plugins ask of the window.
type Host struct {
	ctx        *kernel.Context
	PageID     string
	Toasts     []string
	Copied     string
	Opened     []string
	Behind     string // what Backdrop reports
	Mode       string
	PluginList []kernel.Status
	Toggled    []string
	Notices    []shell.Notice
	Clip       shell.Clipboard // what Clipboard returns
	Background bool            // the window is not in front
	Hidden     bool            // what SetHidden set last
	Data       string          // what DataDir returns
	// Web is what NewWebView gives a plugin; nil returns shell.ErrNoWebViews.
	Web *WebView
}

// New installs the fake: it is its own factory, so a test reads what it
// recorded across restarts. A fake may; a plugin of the app may not.
func (h *Host) New() kernel.Plugin { return h }

func (*Host) Manifest() kernel.Manifest {
	return kernel.Manifest{ID: "fake-host", Provides: []kernel.Named{shell.HostKey}}
}

func (h *Host) Setup(c *kernel.Context) error {
	h.ctx = c
	return kernel.Provide[shell.Host](c, shell.HostKey, h)
}

// ShowPage records the page and emits shell.PageChanged when it changed, as
// the shell does.
func (h *Host) ShowPage(id string) {
	if id == h.PageID {
		return
	}
	h.PageID = id
	if h.ctx != nil {
		kernel.Emit(h.ctx, shell.PageChanged, id)
	}
}

func (h *Host) Page() string               { return h.PageID }
func (h *Host) Toast(t ui.Toast)           { h.Toasts = append(h.Toasts, t.Title) }
func (h *Host) Copy(text string) error     { h.Copied = text; return nil }
func (h *Host) Clipboard() shell.Clipboard { return h.Clip }
func (h *Host) OpenURL(url string)         { h.Opened = append(h.Opened, url) }
func (h *Host) Appearance() string         { return h.Mode }
func (h *Host) SetAppearance(mode string)  { h.Mode = mode }
func (h *Host) Backdrop() string           { return h.Behind }
func (h *Host) SetBackdrop(mode string)    { h.Behind = mode }
func (h *Host) Plugins() []kernel.Status   { return h.PluginList }
func (h *Host) Enable(id string) error     { h.Toggled = append(h.Toggled, "+"+id); return nil }
func (h *Host) Disable(id string) error    { h.Toggled = append(h.Toggled, "-"+id); return nil }
func (h *Host) Notify(n shell.Notice)      { h.Notices = append(h.Notices, n) }
func (h *Host) Foreground() bool           { return !h.Background }
func (h *Host) SetHidden(hidden bool)      { h.Hidden = hidden }
func (h *Host) DataDir() string            { return h.Data }

func (h *Host) NewWebView() (shell.WebView, error) {
	if h.Web == nil {
		return nil, shell.ErrNoWebViews
	}
	return h.Web, nil
}

// WebView fakes a page: it records what the plugin loads, and its Eval and
// Capture answer what a test sets before the plugin calls them. A test reads
// its fields on the UI thread, after the plugin's work has come back.
type WebView struct {
	Loaded  []string // the files LoadFile was given, in order
	Reloads int
	Closed  bool
	// EvalFunc answers Eval, from the plugin's worker; nil returns nothing.
	EvalFunc func(ctx context.Context, code string) (any, error)
	PNG      []byte // what Capture returns
}

var _ shell.WebView = (*WebView)(nil)

// Show draws a plain box where the page would show.
func (v *WebView) Show(c *ui.Context) ui.Element { return ui.Box(c).Label("Web page") }

func (v *WebView) LoadFile(path string) error { v.Loaded = append(v.Loaded, path); return nil }
func (v *WebView) Reload()                    { v.Reloads++ }
func (v *WebView) Capture() ([]byte, error)   { return v.PNG, nil }
func (v *WebView) Close()                     { v.Closed = true }

func (v *WebView) Eval(ctx context.Context, code string) (any, error) {
	if v.EvalFunc == nil {
		return nil, nil
	}
	return v.EvalFunc(ctx, code)
}
