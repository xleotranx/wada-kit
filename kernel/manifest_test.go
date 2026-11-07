package kernel_test

import (
	"context"
	"errors"
	"strings"
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

func TestInstallValidatesServiceKeyDeclarations(t *testing.T) {
	for _, field := range []string{"Provides", "Requires"} {
		for _, tc := range []struct {
			name string
			key  kernel.Named
			want string
		}{
			{"topic", ping, "service key"},
			{"pipeline", outgoing, "service key"},
			{"point", sidebar, "service key"},
			{"name-only", named(agentKey.Name()), "service key"},
			{"nil", nil, "unnamed key"},
			{"nil-key-pointer", (*kernel.Key[string])(nil), "unnamed key"},
			{"zero-key", kernel.Key[string]{}, "unnamed key"},
		} {
			t.Run(field+"/"+tc.name, func(t *testing.T) {
				k := kernel.New(kernel.Options{Post: func(func()) {}, Store: kerneltest.NewStore()})
				t.Cleanup(func() {
					if err := k.Close(context.Background()); err != nil {
						t.Error(err)
					}
				})
				if err := k.Install(plug("existing", nil).New); err != nil {
					t.Fatal(err)
				}
				invalid := plug("invalid", nil)
				if field == "Provides" {
					invalid.m.Provides = []kernel.Named{tc.key}
				} else {
					invalid.m.Requires = []kernel.Named{tc.key}
				}
				err := k.Install(plug("staged", nil).New, invalid.New)
				var panicErr *kernel.PanicError
				if err == nil || errors.As(err, &panicErr) || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "invalid") {
					t.Fatalf("invalid declaration returned %v, want validation error containing %q and plugin ID", err, tc.want)
				}
				if got := k.Plugins(); len(got) != 1 || got[0].Manifest.ID != "existing" {
					t.Fatalf("failed batch changed registry: %+v", got)
				}
				if err := k.Install(plug("staged", nil).New, plug("invalid", nil).New); err != nil {
					t.Fatal(err)
				}
				if err := k.Start(); err != nil {
					t.Fatal(err)
				}
				for _, st := range k.Plugins() {
					if st.State != kernel.Active {
						t.Fatalf("corrected batch did not activate: %+v", st)
					}
				}
			})
		}
	}
}

func TestInstallSnapshotsServiceKeyPointers(t *testing.T) {
	provided := kernel.NewKey[string]("manifest.pointer")
	required := agentKey
	original := provided
	p := plug("consumer", func(c *kernel.Context) error {
		return kernel.Provide(c, original, kernel.Require(c, agentKey).Name())
	})
	p.m.Provides = []kernel.Named{&provided}
	p.m.Requires = []kernel.Named{&required}
	h := kerneltest.New(t, backend("provider").New, p.New)
	provided = kernel.NewKey[string]("manifest.pointer-other")
	required = kernel.NewKey[agent]("manifest.pointer-missing")
	got := state(t, h, "consumer").Manifest
	if got.Provides[0].Name() != original.Name() || got.Requires[0].Name() != agentKey.Name() {
		t.Fatalf("key pointer mutation changed installed dependencies: %+v", got)
	}
	if err := h.Disable("provider"); err != nil {
		t.Fatal(err)
	}
	if state(t, h, "consumer").State != kernel.Blocked {
		t.Fatal("key pointer mutation broke dependent disposal")
	}
	if err := h.Enable("provider"); err != nil {
		t.Fatal(err)
	}
	if value, ok := kernel.Use(h, original); !ok || value != "provider" {
		t.Fatal("key pointer mutation broke dependent reactivation")
	}
	closeClean(t, h)
}
