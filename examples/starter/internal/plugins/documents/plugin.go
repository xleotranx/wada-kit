// Package documents connects document access to the kit's lifecycle.
package documents

import (
	"context"

	"github.com/xleotranx/wada-kit/contract/shell"
	"github.com/xleotranx/wada-kit/examples/starter/internal/documents"
	"github.com/xleotranx/wada-kit/kernel"
)

const pageID = "documents.page"

type Plugin struct {
	ctx            *kernel.Context
	host           shell.Host
	repo           documents.Repository
	last           *kernel.Setting[string]
	document       *kernel.Scope
	task           kernel.Task
	selected, body string
	err            error
	loading        bool
}

func New(repo documents.Repository) kernel.Factory {
	return func() kernel.Plugin { return &Plugin{repo: repo} }
}

func (*Plugin) Manifest() kernel.Manifest {
	return kernel.Manifest{ID: "documents", Title: "Documents", Requires: []kernel.Named{shell.HostKey}}
}

func (p *Plugin) Setup(c *kernel.Context) error {
	p.ctx = c
	p.host = kernel.Require(c, shell.HostKey)
	p.last = kernel.NewSetting(c, "last-document", "welcome.txt")
	kernel.Contribute(c, shell.Pages, pageID, 0, shell.Page{Title: func() string { return "Documents" }, Render: p.view})
	kernel.Contribute(c, kernel.Commands, "documents.open", 0, kernel.Command{
		Title: "Documents", Menu: "View", Keys: []string{"cmd+1"}, Run: func() { p.host.ShowPage(pageID) },
	})
	kernel.On(c, shell.PageChanged, p.pageChanged)
	// Post runs after host startup. On v0.2.0 the default page has an empty
	// Host.Page; this starter's order-0 page is the default in that release too.
	c.Post(func() { p.pageChanged(p.host.Page()) })
	return nil
}

func (p *Plugin) pageChanged(id string) {
	if id == pageID || id == "" {
		if p.document == nil {
			p.open(p.last.Get())
		}
		return
	}
	p.closeDocument()
}

func (p *Plugin) closeDocument() {
	if p.document != nil {
		p.document.Dispose()
		p.document = nil
	}
	p.task.Cancel()
	p.loading = false
}

// open belongs to input/navigation handlers, never to rendering. Each selection
// owns its load. A repository that ignores cancellation still cannot apply an
// old result because its scope is disposed and Task also checks generations.
func (p *Plugin) open(id string) {
	p.closeDocument()
	if id != "welcome.txt" && id != "workflow.txt" {
		id = "welcome.txt"
	}
	p.document = p.ctx.NewScope()
	p.selected, p.body, p.err, p.loading = id, "", nil, true
	repo := p.repo // capture worker-safe inputs, never p or a UI context
	kernel.Go(p.document, &p.task, func(ctx context.Context) (string, error) {
		return repo.Load(ctx, id)
	}, func(body string, err error) {
		p.body, p.err, p.loading = body, err, false
		if err == nil {
			p.err = p.last.Set(id)
		}
	})
}
