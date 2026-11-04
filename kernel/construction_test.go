package kernel_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

type manifestPanic struct{ kernel.Plugin }

func (manifestPanic) Manifest() kernel.Manifest { panic("manifest failed") }

func caughtPanic(fn func()) (value any) {
	defer func() { value = recover() }()
	fn()
	return nil
}

func TestInstallRollsBackFailedBatch(t *testing.T) {
	a, b := backend("a"), plug("b", nil)
	other := kernel.NewKey[agent]("construction.other")
	a.m.Requires = []kernel.Named{other}
	b.m.Provides, b.m.Requires = []kernel.Named{other}, []kernel.Named{agentKey}
	for name, batch := range map[string][]kernel.Factory{
		"invalid":   {plug("new", nil).New, plug("Bad_ID", nil).New},
		"duplicate": {plug("new", nil).New, plug("existing", nil).New},
		"cycle":     {a.New, b.New},
	} {
		t.Run(name, func(t *testing.T) {
			k := kernel.New(kernel.Options{Post: func(func()) {}, Store: kerneltest.NewStore()})
			t.Cleanup(func() {
				if err := k.Close(context.Background()); err != nil {
					t.Error(err)
				}
			})
			if err := k.Install(plug("existing", nil).New); err != nil {
				t.Fatal(err)
			}
			if err := k.Install(batch...); err == nil {
				t.Fatal("accepted invalid batch")
			}
			if got := k.Plugins(); len(got) != 1 || got[0].Manifest.ID != "existing" {
				t.Fatalf("failed batch changed registry: %+v", got)
			}
			// IDs from the failed batch must be available on a corrected retry.
			if err := k.Install(plug("new", nil).New, backend("a").New, plug("b", nil).New); err != nil {
				t.Fatal(err)
			}
			if err := k.Start(); err != nil {
				t.Fatal(err)
			}
			for _, st := range k.Plugins() {
				if st.State != kernel.Active {
					t.Fatalf("retry did not activate: %+v", st)
				}
			}
		})
	}
}

func TestInstallConstructionPanicPolicy(t *testing.T) {
	for _, strict := range []bool{false, true} {
		for name, factory := range map[string]kernel.Factory{
			"factory":  func() kernel.Plugin { panic("factory failed") },
			"manifest": func() kernel.Plugin { return manifestPanic{} },
		} {
			t.Run(fmt.Sprintf("%s/strict=%t", name, strict), func(t *testing.T) {
				k := kernel.New(kernel.Options{Post: func(func()) {}, Store: kerneltest.NewStore(), Strict: strict})
				var err error
				value := caughtPanic(func() { err = k.Install(plug("new", nil).New, factory) })
				if strict {
					if value != name+" failed" {
						t.Fatalf("strict panic = %v", value)
					}
				} else {
					var pe *kernel.PanicError
					if value != nil || !errors.As(err, &pe) || pe.Value != name+" failed" || len(pe.Stack) == 0 {
						t.Fatalf("panic = %v, error = %v", value, err)
					}
				}
				if got := k.Plugins(); len(got) != 0 {
					t.Fatalf("panic left partial installation: %+v", got)
				}
				if err := k.Install(plug("new", nil).New); err != nil {
					t.Fatal(err)
				}
				if err := k.Close(context.Background()); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestReactivationConstructionFailure(t *testing.T) {
	for _, mode := range []string{"factory", "manifest", "nil"} {
		t.Run(mode, func(t *testing.T) {
			p := backend("provider")
			fail := false
			factory := func() kernel.Plugin {
				if fail {
					switch mode {
					case "factory":
						panic("factory failed")
					case "manifest":
						return manifestPanic{}
					case "nil":
						return nil
					}
				}
				return p.New()
			}
			var used agent
			h := kerneltest.Start(t, kerneltest.Config{Lenient: true,
				Restart: kernel.Restart{Max: 3, Within: time.Minute}},
				factory, consumer("consumer", &used).New, plug("unrelated", nil).New)
			if err := h.Disable("provider"); err != nil {
				t.Fatal(err)
			}
			fail = true
			var err error
			if value := caughtPanic(func() { err = h.Enable("provider") }); value != nil {
				t.Fatalf("construction panic escaped: %v", value)
			}
			st := state(t, h, "provider")
			if err == nil || st.State != kernel.Failed || st.Restarting {
				t.Fatalf("enable = %v, status = %+v", err, st)
			}
			if mode != "nil" {
				var pe *kernel.PanicError
				if !errors.As(err, &pe) || !strings.Contains(h.Inspect()[0].Stack, "TestReactivationConstructionFailure") {
					t.Fatalf("missing panic diagnostics: %v", err)
				}
			}
			if d := h.Inspect()[0]; d.Scopes != 0 || len(d.Registrations) != 0 {
				t.Fatalf("failed construction left resources: %+v", d)
			}
			h.Flush()
			if state(t, h, "provider").State != kernel.Failed || state(t, h, "consumer").State != kernel.Blocked || state(t, h, "unrelated").State != kernel.Active {
				t.Fatalf("construction failure affected other lifetimes: %+v", h.Plugins())
			}
			fail = false
			if err := h.Enable("provider"); err != nil {
				t.Fatal(err)
			}
			if state(t, h, "provider").State != kernel.Active || state(t, h, "consumer").State != kernel.Active {
				t.Fatalf("explicit retry did not restore dependents: %+v", h.Plugins())
			}
			closeClean(t, h)
		})
	}
}

func TestStrictReactivationConstructionPanics(t *testing.T) {
	for _, mode := range []string{"factory", "manifest"} {
		t.Run(mode, func(t *testing.T) {
			fail := false
			h := kerneltest.New(t, func() kernel.Plugin {
				if fail {
					if mode == "factory" {
						panic("factory failed")
					}
					return manifestPanic{}
				}
				return plug("p", nil).New()
			})
			if err := h.Disable("p"); err != nil {
				t.Fatal(err)
			}
			fail = true
			if value := caughtPanic(func() { _ = h.Enable("p") }); value != mode+" failed" {
				t.Fatalf("strict panic = %v", value)
			}
		})
	}
}

func TestRestartConstructionPanicStaysFailed(t *testing.T) {
	setups := 0
	p, user := crasher(&setups)
	fail := false
	h := kerneltest.Start(t, kerneltest.Config{Lenient: true,
		Restart: kernel.Restart{Max: 3, Within: time.Minute}}, func() kernel.Plugin {
		if fail {
			panic("restart factory failed")
		}
		return p.New()
	}, user.New)
	fail = true
	kernel.Emit(h, ping, "crash")
	if !state(t, h, "p").Restarting {
		t.Fatal("runtime failure did not schedule restart")
	}
	h.Flush()
	st := state(t, h, "p")
	var pe *kernel.PanicError
	if st.State != kernel.Failed || st.Restarting || !errors.As(st.Err, &pe) || pe.Value != "restart factory failed" {
		t.Fatalf("restart construction failure: %+v", st)
	}
	if setups != 1 || state(t, h, "user").State != kernel.Blocked {
		t.Fatalf("failed restart ran Setup or restored dependents: %+v", h.Plugins())
	}
	fail = false
	if err := h.Enable("p"); err != nil {
		t.Fatal(err)
	}
	if setups != 2 || state(t, h, "user").State != kernel.Active {
		t.Fatal("explicit retry did not restore dependent")
	}
}
