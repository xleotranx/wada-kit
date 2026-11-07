// Package notebook owns the example's in-memory notebook capability.
package notebook

import (
	"errors"
	"slices"
	"strings"

	contract "github.com/xleotranx/wada-kit/examples/extensions/internal/contract/notebook"
	"github.com/xleotranx/wada-kit/kernel"
)

type plugin struct {
	ctx    *kernel.Context
	titles []string
}

func New() kernel.Plugin { return &plugin{} }

func (*plugin) Manifest() kernel.Manifest {
	return kernel.Manifest{ID: "notebook", Title: "Notebook", Provides: []kernel.Named{contract.NotebookKey}}
}

func (p *plugin) Setup(c *kernel.Context) error {
	p.ctx = c
	kernel.Transform(c, contract.Preparing, 10, func(title string) (string, func(), error) {
		title = strings.TrimSpace(title)
		if title == "" {
			return "", nil, errors.New("title is empty")
		}
		return title, nil, nil
	})
	if err := kernel.Provide[contract.Notebook](c, contract.NotebookKey, p); err != nil {
		return err
	}
	return kernel.Contribute(c, contract.Actions, "notebook.list", 20, contract.Action{
		Title: "List notes", Run: func() (string, error) { return strings.Join(p.titles, ", "), nil },
	})
}

func (p *plugin) Add(title string) (string, error) {
	title, commit, err := kernel.Run(p.ctx, contract.Preparing, title)
	if err != nil {
		return "", err
	}
	p.titles = append(p.titles, title)
	commit()
	kernel.Emit(p.ctx, contract.Added, title)
	return title, nil
}

func (p *plugin) Titles() []string { return slices.Clone(p.titles) }
