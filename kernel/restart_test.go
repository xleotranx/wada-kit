package kernel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestPluginRestartPolicyOverridesHostDefault(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fallback kernel.Restart
		override *kernel.Restart
		allowed  int
	}{
		{name: "default off"},
		{name: "inherit", fallback: kernel.Restart{Max: 1, Within: time.Minute}, allowed: 1},
		{name: "explicit off", fallback: kernel.Restart{Max: 2, Within: time.Minute}, override: &kernel.Restart{}},
		{name: "opt in", override: &kernel.Restart{Max: 1, Within: time.Minute}, allowed: 1},
		{name: "override budget", fallback: kernel.Restart{Max: 1, Within: time.Minute}, override: &kernel.Restart{Max: 2, Within: time.Minute}, allowed: 2},
		{name: "invalid override", fallback: kernel.Restart{Max: 1, Within: time.Minute}, override: &kernel.Restart{Max: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setups := 0
			p, _ := crasher(&setups)
			p.m.Restart = tc.override
			h := kerneltest.Start(t, kerneltest.Config{Lenient: true, Restart: tc.fallback}, p.New)
			for i := range tc.allowed {
				kernel.Emit(h, ping, "crash")
				if st := state(t, h, "p"); !st.Restarting || st.Restarts != i+1 {
					t.Fatalf("restart %d: %+v", i+1, st)
				}
				h.Flush()
				if state(t, h, "p").State != kernel.Active {
					t.Fatal("restart did not activate plugin")
				}
			}
			kernel.Emit(h, ping, "crash")
			h.Flush()
			if st := state(t, h, "p"); st.State != kernel.Failed || st.Restarting || setups != tc.allowed+1 {
				t.Fatalf("exceeded budget: %+v, setups=%d", st, setups)
			}
		})
	}
}

func TestPluginRestartPolicyIsSnapshotAtInstall(t *testing.T) {
	policy := kernel.Restart{Max: 1, Within: time.Minute}
	setups := 0
	p, _ := crasher(&setups)
	p.m.Restart = &policy
	h := kerneltest.Start(t, kerneltest.Config{Lenient: true}, p.New)
	policy.Max = 10
	state(t, h, "p").Manifest.Restart.Max = 20
	if got := state(t, h, "p").Manifest.Restart.Max; got != 1 {
		t.Fatalf("caller mutated installed policy: Max=%d", got)
	}
	for range 2 {
		kernel.Emit(h, ping, "crash")
		h.Flush()
	}
	if st := state(t, h, "p"); st.State != kernel.Failed || st.Restarts != 1 || setups != 2 {
		t.Fatalf("caller changed restart budget: %+v, setups=%d", st, setups)
	}
}

func TestPluginRestartWindowAndBudgetAreIndependent(t *testing.T) {
	now := time.Unix(0, 0)
	makePlugin := func(id string) *fake {
		return plug(id, func(c *kernel.Context) error {
			return kernel.Contribute(c, kernel.Commands, id+".crash", 0,
				kernel.Command{Run: func() { panic("crash") }})
		})
	}
	a, b := makePlugin("a"), makePlugin("b")
	a.m.Restart = &kernel.Restart{Max: 1, Within: time.Second}
	b.m.Restart = &kernel.Restart{Max: 1, Within: time.Minute}
	h := kerneltest.Start(t, kerneltest.Config{Lenient: true, Now: func() time.Time { return now }}, a.New, b.New)
	crash := func(id string) {
		for _, e := range kernel.Items(h, kernel.Commands) {
			if e.Plugin == id {
				kernel.Invoke(h, e, func(c kernel.Command) { c.Run() })
				h.Flush()
				return
			}
		}
		t.Fatalf("missing command for %s", id)
	}
	crash("a")
	crash("b")
	if state(t, h, "a").State != kernel.Active || state(t, h, "b").State != kernel.Active {
		t.Fatalf("plugins shared a restart budget: %+v", h.Plugins())
	}
	now = now.Add(time.Second)
	if state(t, h, "a").Restarts != 0 || state(t, h, "b").Restarts != 1 {
		t.Fatalf("wrong per-plugin time window: %+v", h.Plugins())
	}
	crash("a")
	crash("b")
	if state(t, h, "a").State != kernel.Active || state(t, h, "b").State != kernel.Failed {
		t.Fatalf("wrong per-plugin restart limits: %+v", h.Plugins())
	}
}

func TestRestartDoesNotReplayFailedCommand(t *testing.T) {
	setups, runs := 0, 0
	p := plug("p", func(c *kernel.Context) error {
		setups++
		return kernel.Contribute(c, kernel.Commands, "p.run", 0, kernel.Command{Run: func() {
			runs++
			panic("failed operation")
		}})
	})
	p.m.Restart = &kernel.Restart{Max: 1, Within: time.Minute}
	h := kerneltest.Start(t, kerneltest.Config{Lenient: true}, p.New)
	old := kernel.Items(h, kernel.Commands)[0]
	kernel.Invoke(h, old, func(c kernel.Command) { c.Run() })
	h.Flush()
	kernel.Invoke(h, old, func(c kernel.Command) { c.Run() })
	if setups != 2 || runs != 1 || state(t, h, "p").State != kernel.Active {
		t.Fatalf("restart replayed command: setups=%d runs=%d", setups, runs)
	}
	// A fresh user action still runs the new contribution.
	kernel.Invoke(h, kernel.Items(h, kernel.Commands)[0], func(c kernel.Command) { c.Run() })
	if runs != 2 {
		t.Fatalf("fresh command ran %d times", runs)
	}
	closeClean(t, h)
}

func TestNonpositiveRestartLimitsDisableRestart(t *testing.T) {
	for _, policy := range []kernel.Restart{
		{}, {Max: -1, Within: time.Minute}, {Max: 1}, {Max: 1, Within: -time.Second},
	} {
		t.Run(fmt.Sprint(policy), func(t *testing.T) {
			setups := 0
			p, _ := crasher(&setups)
			h := kerneltest.Start(t, kerneltest.Config{Lenient: true, Restart: policy}, p.New)
			kernel.Emit(h, ping, "crash")
			h.Flush()
			if st := state(t, h, "p"); st.State != kernel.Failed || st.Restarting || st.Restarts != 0 || setups != 1 {
				t.Fatalf("nonpositive policy restarted: %+v, setups=%d", st, setups)
			}
		})
	}
}
