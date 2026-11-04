package kernel_test

import (
	"context"
	"testing"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestStaleContextCannotRegisterOrFailNewSetup(t *testing.T) {
	for _, mode := range []string{"provide", "contribute"} {
		t.Run(mode, func(t *testing.T) {
			var stale *kernel.Context
			p := backend("provider")
			p.setup = func(c *kernel.Context) error {
				if stale != nil {
					var err error
					if mode == "provide" {
						err = kernel.Provide[agent](stale, agentKey, named("stale"))
					} else {
						err = kernel.Contribute(stale, sidebar, "provider.stale", 0, "stale")
					}
					if err == nil {
						t.Error("registration through stale context succeeded")
					}
				}
				return kernel.Provide[agent](c, agentKey, named("fresh"))
			}
			h := kerneltest.New(t, p.New)
			stale = p.ctx
			if err := h.Disable("provider"); err != nil {
				t.Fatal(err)
			}
			if err := h.Enable("provider"); err != nil {
				t.Fatalf("stale registration poisoned new Setup: %v", err)
			}
			if v, ok := kernel.Use(h, agentKey); !ok || v.Name() != "fresh" {
				t.Fatalf("wrong provider: %v, %v", v, ok)
			}
			if err := h.Disable("provider"); err != nil {
				t.Fatal(err)
			}
			if _, ok := kernel.Use(h, agentKey); ok {
				t.Fatal("stale service survived disposal")
			}
			closeClean(t, h)
		})
	}
}

func TestDisposedSetupCannotProvideService(t *testing.T) {
	p := backend("provider")
	p.setup = func(c *kernel.Context) error {
		c.Dispose()
		return kernel.Provide[agent](c, agentKey, named("dead"))
	}
	h := kerneltest.New(t, p.New)
	if state(t, h, "provider").State != kernel.Failed {
		t.Fatal("disposed Setup became active")
	}
	if _, ok := kernel.Use(h, agentKey); ok {
		t.Fatal("disposed Setup leaked a service")
	}
	closeClean(t, h)
}

func TestSetupCannotActivateDisposedContext(t *testing.T) {
	p := plug("p", func(c *kernel.Context) error { c.Dispose(); return nil })
	h := kerneltest.New(t, p.New)
	if st := state(t, h, "p"); st.State != kernel.Failed || st.Err == nil {
		t.Fatalf("disposed context became active: %+v", st)
	}
	closeClean(t, h)
}

func TestFailedRegistrationOnCurrentChildStillFailsSetup(t *testing.T) {
	p := plug("p", func(c *kernel.Context) error {
		child := c.NewScope()
		child.Dispose()
		_ = kernel.Contribute(child, sidebar, "p.dead", 0, "dead")
		return nil
	})
	h := kerneltest.New(t, p.New)
	if st := state(t, h, "p"); st.State != kernel.Failed || st.Err == nil {
		t.Fatalf("ignored current-activation registration error: %+v", st)
	}
	closeClean(t, h)
}

func TestImmediatePointCallbackCannotLeakContribution(t *testing.T) {
	k := kernel.New(kernel.Options{Post: func(fn func()) { fn() }, Background: func(fn func()) { fn() }, Store: kerneltest.NewStore(), Strict: true})
	t.Cleanup(func() {
		if err := k.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	p := plug("p", nil)
	if err := k.Install(p.New); err != nil {
		t.Fatal(err)
	}
	if err := k.Start(); err != nil {
		t.Fatal(err)
	}
	armed := true
	kernel.On(k, kernel.PointChanged, func(point string) {
		if armed && point == sidebar.Name() {
			armed = false
			if err := k.Disable("p"); err != nil {
				t.Error(err)
			}
		}
	})
	if err := kernel.Contribute(p.ctx, sidebar, "p.item", 0, "item"); err != nil {
		t.Fatal(err)
	}
	if len(kernel.Items(k, sidebar)) != 0 {
		t.Fatal("immediate disposal left contribution registered")
	}
	if err := k.Enable("p"); err != nil {
		t.Fatal(err)
	}
	if err := kernel.Contribute(p.ctx, sidebar, "p.item", 0, "replacement"); err != nil {
		t.Fatalf("disposed contribution ID could not be reused: %v", err)
	}
}

func TestDisposedCallerCannotEmitOrRunPipeline(t *testing.T) {
	p := plug("p", nil)
	h := kerneltest.New(t, p.New)
	heard, transformed := 0, 0
	kernel.On(h, ping, func(string) { heard++ })
	kernel.Transform(h, outgoing, 0, func(v string) (string, func(), error) {
		transformed++
		return v, nil, nil
	})
	s := p.ctx.NewScope()
	s.Dispose()
	kernel.Emit(s, ping, "stale")
	if _, commit, err := kernel.Run(s, outgoing, "stale"); err == nil || commit != nil {
		t.Fatalf("disposed caller ran pipeline: error=%v", err)
	}
	if heard != 0 || transformed != 0 {
		t.Fatalf("disposed scope delivered callbacks: heard=%d transformed=%d", heard, transformed)
	}
}

func TestPipelineCommitExpiresWithCaller(t *testing.T) {
	p := plug("p", nil)
	h := kerneltest.New(t, p.New)
	committed := 0
	kernel.Transform(h, outgoing, 0, func(v string) (string, func(), error) {
		return v, func() { committed++ }, nil
	})
	s := p.ctx.NewScope()
	_, commit, err := kernel.Run(s, outgoing, "x")
	if err != nil {
		t.Fatal(err)
	}
	s.Dispose()
	commit()
	commit()
	if committed != 0 {
		t.Fatal("pipeline committed after caller disposal")
	}
}

func TestPipelineStopsWhenCallerDisposedDuringRunOrCommit(t *testing.T) {
	for _, phase := range []string{"run", "commit"} {
		t.Run(phase, func(t *testing.T) {
			p := plug("p", nil)
			h := kerneltest.New(t, p.New)
			caller := p.ctx.NewScope()
			laterRuns, laterCommits := 0, 0
			kernel.Transform(h, outgoing, 0, func(v string) (string, func(), error) {
				if phase == "run" {
					caller.Dispose()
				}
				return v, caller.Dispose, nil
			})
			kernel.Transform(h, outgoing, 1, func(v string) (string, func(), error) {
				laterRuns++
				return v, func() { laterCommits++ }, nil
			})
			_, commit, err := kernel.Run(caller, outgoing, "x")
			if phase == "run" {
				if err == nil || commit != nil || laterRuns != 0 {
					t.Fatalf("continued after caller disposal: error=%v laterRuns=%d", err, laterRuns)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				commit()
				commit()
				if laterCommits != 0 {
					t.Fatal("continued committing after caller disposal")
				}
			}
		})
	}
}
