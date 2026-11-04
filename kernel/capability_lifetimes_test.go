package kernel_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestOptionalProviderRemovalAndRecovery(t *testing.T) {
	for _, mode := range []string{"disable", "failure"} {
		t.Run(mode, func(t *testing.T) {
			activations, optionalSetups := 0, 0
			provider := backend("provider")
			provider.setup = func(c *kernel.Context) error {
				activations++
				kernel.On(c, ping, func(string) { panic("provider failed") })
				return kernel.Provide[agent](c, agentKey, named(fmt.Sprint(activations)))
			}
			optional := plug("optional", func(*kernel.Context) error { optionalSetups++; return nil })
			var requiredValue agent
			required := consumer("required", &requiredValue)
			setup := required.setup
			required.setup = func(c *kernel.Context) error {
				c.OnDispose(func() {
					if _, ok := kernel.Use(c, agentKey); !ok {
						t.Error("provider withdrawn before required consumer cleanup")
					}
				})
				return setup(c)
			}
			h := kerneltest.Start(t, kerneltest.Config{Lenient: true}, required.New, optional.New, provider.New)
			oldRequired := required.ctx
			if v, ok := kernel.Use(optional.ctx, agentKey); !ok || v.Name() != "1" {
				t.Fatal("optional consumer could not find initial provider")
			}
			if mode == "disable" {
				if err := h.Disable("provider"); err != nil {
					t.Fatal(err)
				}
			} else {
				kernel.Emit(h, ping, "fail")
			}
			if _, ok := kernel.Use(optional.ctx, agentKey); ok {
				t.Fatal("optional lookup retained removed provider")
			}
			if state(t, h, "optional").State != kernel.Active || state(t, h, "required").State != kernel.Blocked || oldRequired.Alive() {
				t.Fatalf("wrong removal lifetimes: %+v", h.Plugins())
			}
			if err := h.Enable("provider"); err != nil {
				t.Fatal(err)
			}
			if v, ok := kernel.Use(optional.ctx, agentKey); !ok || v.Name() != "2" || requiredValue.Name() != "2" {
				t.Fatal("consumers did not receive replacement provider")
			}
			if optionalSetups != 1 || required.ctx == oldRequired || state(t, h, "required").State != kernel.Active {
				t.Fatal("wrong consumer recovery lifetimes")
			}
			closeClean(t, h)
		})
	}
}

func TestOldTaskCannotAffectReplacementActivation(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(fmt.Sprint("panic=", panics), func(t *testing.T) {
			p := plug("p", nil)
			h := kerneltest.Start(t, kerneltest.Config{Lenient: true}, p.New)
			var task kernel.Task
			applied := ""
			old := p.ctx
			kernel.Go(old, &task, func(context.Context) (string, error) {
				if panics {
					panic("obsolete task")
				}
				return "old", nil
			}, func(v string, _ error) { applied = v })
			old.Post(func() { panic("obsolete post") })
			old.PostQuiet(func() { panic("obsolete quiet post") })
			if err := h.Disable("p"); err != nil {
				t.Fatal(err)
			}
			if err := h.Enable("p"); err != nil {
				t.Fatal(err)
			}
			kernel.Go(p.ctx, &task, func(context.Context) (string, error) { return "new", nil },
				func(v string, _ error) { applied = v })
			h.RunWork(0) // Complete the old task after the new activation started.
			h.RunPosts()
			if applied != "" || !task.Running() || state(t, h, "p").State != kernel.Active {
				t.Fatalf("old work changed new activation: applied=%q running=%v", applied, task.Running())
			}
			h.Flush()
			if applied != "new" || task.Running() || state(t, h, "p").State != kernel.Active {
				t.Fatalf("new result did not survive old work: applied=%q", applied)
			}
			closeClean(t, h)
		})
	}
}
