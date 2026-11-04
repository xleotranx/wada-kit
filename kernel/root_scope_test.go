package kernel_test

import (
	"slices"
	"testing"
	"time"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestRootScopeDisposalStopsDependentsAndAllowsExplicitRetry(t *testing.T) {
	for _, locked := range []bool{false, true} {
		t.Run(map[bool]string{false: "optional", true: "locked"}[locked], func(t *testing.T) {
			var order []string
			provider := backend("provider")
			provider.m.Locked = locked
			provider.m.Restart = &kernel.Restart{Max: 3, Within: time.Minute}
			setup := provider.setup
			provider.setup = func(c *kernel.Context) error {
				c.OnDispose(func() { order = append(order, "provider") })
				return setup(c)
			}
			var value agent
			dependent := consumer("dependent", &value)
			setupDependent := dependent.setup
			dependent.setup = func(c *kernel.Context) error {
				c.OnDispose(func() { order = append(order, "dependent") })
				return setupDependent(c)
			}
			optional := plug("optional", nil)
			h := kerneltest.New(t, provider.New, dependent.New, optional.New)
			oldRoot := provider.ctx
			oldRoot.Dispose()
			h.Flush()
			if st := state(t, h, "provider"); st.State != kernel.Failed || st.Err == nil || st.Restarting || oldRoot.Alive() {
				t.Fatalf("root disposal did not stop activation: %+v", st)
			}
			if state(t, h, "dependent").State != kernel.Blocked || dependent.ctx.Alive() || state(t, h, "optional").State != kernel.Active {
				t.Fatalf("wrong dependent lifetimes: %+v", h.Plugins())
			}
			if !slices.Equal(order, []string{"dependent", "provider"}) {
				t.Fatalf("root disposal order: %v", order)
			}
			if _, ok := kernel.Use(optional.ctx, agentKey); ok {
				t.Fatal("disposed root still provides service")
			}
			if err := h.Enable("provider"); err != nil {
				t.Fatal(err)
			}
			oldRoot.Dispose()
			if state(t, h, "provider").State != kernel.Active || state(t, h, "dependent").State != kernel.Active || provider.ctx == oldRoot {
				t.Fatal("stale root affected replacement activation")
			}
			child := provider.ctx.NewScope()
			child.Dispose()
			if state(t, h, "provider").State != kernel.Active || state(t, h, "dependent").State != kernel.Active {
				t.Fatal("child disposal stopped plugin activation")
			}
			closeClean(t, h)
		})
	}
}

func TestRootDisposalDuringDependentCleanupPreservesProvider(t *testing.T) {
	provider := backend("provider")
	var value agent
	dependent := consumer("dependent", &value)
	h := kerneltest.New(t, provider.New, dependent.New)
	dependent.ctx.OnDispose(func() {
		provider.ctx.Dispose() // Already being disposed through Kernel.Disable.
		if _, ok := kernel.Use(dependent.ctx, agentKey); !ok || !provider.ctx.Alive() {
			t.Error("reentrant root disposal withdrew provider before dependent cleanup finished")
		}
	})
	if err := h.Disable("provider"); err != nil {
		t.Fatal(err)
	}
	if state(t, h, "provider").State != kernel.Disabled || state(t, h, "dependent").State != kernel.Blocked {
		t.Fatalf("root disposal changed ongoing teardown: %+v", h.Plugins())
	}
	closeClean(t, h)
}
