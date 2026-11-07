// Package workspace extends the notebook through its public contracts.
package workspace

import (
	"errors"
	"fmt"
	"unicode/utf8"

	contract "github.com/xleotranx/wada-kit/examples/extensions/internal/contract/notebook"
	"github.com/xleotranx/wada-kit/kernel"
)

type plugin struct {
	book                contract.Notebook
	committed, observed int
}

func New() kernel.Plugin { return &plugin{} }

func (*plugin) Manifest() kernel.Manifest {
	return kernel.Manifest{ID: "workspace", Title: "Workspace", Requires: []kernel.Named{contract.NotebookKey}}
}

func (p *plugin) Setup(c *kernel.Context) error {
	p.book = kernel.Require(c, contract.NotebookKey)
	kernel.Transform(c, contract.Preparing, 20, func(title string) (string, func(), error) {
		if utf8.RuneCountInString(title) > 32 {
			return "", nil, errors.New("workspace titles are limited to 32 runes")
		}
		// Prepare leaves plugin state alone; Add calls this only after appending.
		return title, func() { p.committed++ }, nil
	})
	kernel.On(c, contract.Added, func(string) { p.observed++ })
	return kernel.Contribute(c, contract.Actions, "workspace.add", 10, contract.Action{
		Title: "Add welcome note", Run: func() (string, error) {
			title, err := p.book.Add("  Hello plugins  ")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s (committed=%d, observed=%d)", title, p.committed, p.observed), nil
		},
	})
}
