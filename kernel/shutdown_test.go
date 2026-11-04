package kernel_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestCloseCancelsAllLifetimesBeforeDisposal(t *testing.T) {
	provider := backend("provider")
	var used agent
	dependent := consumer("dependent", &used)
	h := kerneltest.New(t, provider.New, dependent.New)
	dependent.ctx.OnDispose(func() {
		if provider.ctx.Lifetime().Err() == nil {
			t.Error("disposer ran before provider lifetime was cancelled")
		}
	})
	closeClean(t, h)
}

func TestCloseRejectsActivationFromDisposer(t *testing.T) {
	setups := 0
	late := plug("late", func(*kernel.Context) error { setups++; return nil })
	late.m.DefaultOff = true
	p := plug("p", nil)
	h := kerneltest.New(t, p.New, late.New)
	p.ctx.OnDispose(func() {
		if err := h.Enable("late"); err == nil {
			t.Error("activated plugin during shutdown")
		}
	})
	closeClean(t, h)
	if setups != 0 {
		t.Fatalf("late plugin set up %d times", setups)
	}
}

func TestClosedKernelCannotBeInstalledOrStarted(t *testing.T) {
	k := kernel.New(kernel.Options{Post: func(func()) {}, Store: kerneltest.NewStore()})
	if err := k.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := k.Install(plug("late", nil).New); err == nil {
		t.Error("Install accepted after Close")
	}
	if err := k.Start(); err == nil {
		t.Error("Start accepted after Close")
	}
	if len(k.Plugins()) != 0 || kernel.Leaks(k) != 0 {
		t.Fatal("closed kernel acquired plugin state")
	}
}

func TestCloseReturnsDisposerFailuresAndContinuesCleanup(t *testing.T) {
	cleaned := false
	p := plug("p", func(c *kernel.Context) error {
		c.OnDispose(func() { cleaned = true })
		c.OnDispose(func() { panic("cleanup failed") })
		return kernel.Contribute(c, sidebar, "p.item", 0, "item")
	})
	other := plug("other", func(c *kernel.Context) error {
		c.OnDispose(func() { panic("other cleanup failed") })
		return nil
	})
	h := kerneltest.Start(t, kerneltest.Config{Lenient: true}, p.New, other.New)
	err := h.Kernel.Close(context.Background())
	var pe *kernel.PluginError
	var panicErr *kernel.PanicError
	if !errors.As(err, &pe) || pe.Plugin != "p" || !errors.As(err, &panicErr) || panicErr.Value != "cleanup failed" {
		t.Fatalf("missing attributed disposal error: %v", err)
	}
	if !cleaned || kernel.Leaks(h.Kernel) != 0 {
		t.Fatal("disposer panic prevented remaining cleanup")
	}
	for _, st := range h.Plugins() {
		if st.Err == nil || !errors.Is(err, st.Err) {
			t.Fatalf("Close omitted plugin error: %+v; returned %v", st, err)
		}
	}
	closeClean(t, h)
}

func TestCloseBoundsBlockingWorkerCleanup(t *testing.T) {
	cleanupStarted, release, stopped := make(chan struct{}), make(chan struct{}), make(chan struct{})
	p := plug("p", func(c *kernel.Context) error {
		c.Spawn(func(ctx context.Context) {
			defer close(stopped)
			<-ctx.Done()
			close(cleanupStarted)
			<-release // Simulate a resource Close that does not accept cancellation.
		})
		return nil
	})
	h := kerneltest.New(t, p.New)
	t.Cleanup(func() {
		close(release)
		select {
		case <-stopped:
		case <-time.After(2 * time.Second):
			t.Error("cleanup worker leaked")
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := h.Kernel.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Close did not report unfinished worker: %v", err)
	}
	select {
	case <-cleanupStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start cleanup on cancellation")
	}
	if d := h.Inspect()[0]; d.Scopes != 0 || d.Workers != 1 {
		t.Fatalf("unfinished cleanup not tracked after shutdown: %+v", d)
	}
}

func TestClosePersistsDisposerSettingsOnce(t *testing.T) {
	disposals := 0
	p := plug("p", func(c *kernel.Context) error {
		c.OnDispose(func() {
			disposals++
			if err := c.Settings().Set("closed", true); err != nil {
				t.Error(err)
			}
		})
		return nil
	})
	h := kerneltest.New(t, p.New)
	closeClean(t, h)
	closeClean(t, h)
	if disposals != 1 || string(h.Store.Raw("p", "closed")) != "true" || h.Store.Saves() != 1 {
		t.Fatalf("shutdown persistence: disposals=%d saves=%d value=%s", disposals, h.Store.Saves(), h.Store.Raw("p", "closed"))
	}
}
