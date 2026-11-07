package kernel_test

import (
	"testing"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestManifestDependenciesAreSnapshots(t *testing.T) {
	provided := kernel.NewKey[string]("manifest.provided")
	other := kernel.NewKey[string]("manifest.other")
	for _, source := range []string{"input", "plugins", "diagnostics", "event"} {
		t.Run(source, func(t *testing.T) {
			p := plug("consumer", func(c *kernel.Context) error {
				return kernel.Provide(c, provided, kernel.Require(c, agentKey).Name())
			})
			p.m.Provides = []kernel.Named{provided}
			p.m.Requires = []kernel.Named{agentKey}
			h := kerneltest.New(t, backend("provider").New, p.New, plug("unrelated", nil).New)
			var snapshot kernel.Manifest
			switch source {
			case "input":
				snapshot = p.m
			case "plugins":
				snapshot = h.Plugins()[1].Manifest
			case "diagnostics":
				snapshot = h.Inspect()[1].Status.Manifest
			case "event":
				kernel.On(h, kernel.PluginsChanged, func(statuses []kernel.Status) {
					snapshot = statuses[1].Manifest
				})
				if err := h.Disable("unrelated"); err != nil {
					t.Fatal(err)
				}
			}
			snapshot.Provides[0], snapshot.Requires[0] = other, other
			got := state(t, h, "consumer").Manifest
			if got.Provides[0].Name() != provided.Name() || got.Requires[0].Name() != agentKey.Name() {
				t.Fatalf("%s mutation changed installed dependencies: %+v", source, got)
			}
			if err := h.Disable("provider"); err != nil {
				t.Fatal(err)
			}
			if state(t, h, "consumer").State != kernel.Blocked || p.ctx.Alive() {
				t.Fatal("snapshot mutation broke dependent disposal")
			}
			if err := h.Enable("provider"); err != nil {
				t.Fatal(err)
			}
			if value, ok := kernel.Use(h, provided); !ok || value != "provider" || state(t, h, "consumer").State != kernel.Active {
				t.Fatal("snapshot mutation broke dependent reactivation")
			}
			closeClean(t, h)
		})
	}
}
