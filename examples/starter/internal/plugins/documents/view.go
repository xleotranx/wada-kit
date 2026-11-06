package documents

import (
	"github.com/egoist/mygo/ui"
	"github.com/xleotranx/wada-kit/native"
)

func (p *Plugin) view(c *ui.Context) {
	pal := native.PaletteFor(c)
	ui.Column(c).Fill().Padding(24).Gap(16).Children(func() {
		ui.Row(c).FillWidth().Gap(8).Children(func() {
			c.Key("welcome")
			native.Choice(c, "Welcome", p.selected == "welcome.txt", func() { p.open("welcome.txt") })
			c.Key("workflow")
			native.Choice(c, "Workflow", p.selected == "workflow.txt", func() { p.open("workflow.txt") })
		})
		switch {
		case p.loading:
			ui.Text(c, "Loading…").TextColor(pal.Muted)
		case p.err != nil:
			ui.Text(c, p.err.Error()).TextColor(pal.Danger)
			ui.Button(c, "Retry").OnClick(func() { p.open(p.selected) })
		case p.body == "":
			ui.Text(c, "This document is empty.").TextColor(pal.Muted)
		default:
			ui.Text(c, p.body).TextColor(pal.Text)
		}
	})
}
