package kernel_test

import (
	"context"
	"testing"
	"time"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestPostAsksForAFrameAndPostQuietDoesNot(t *testing.T) {
	p := plug("a", nil)
	h := kerneltest.New(t, p.New)
	h.Flush()
	before := h.Frames()

	ran := 0
	p.ctx.Post(func() { ran++ })
	h.RunPosts()
	if ran != 1 || h.Frames() != before+1 {
		t.Fatalf("Post: ran %d, frames %d → %d; want a frame after it", ran, before, h.Frames())
	}

	p.ctx.PostQuiet(func() { ran++ })
	h.RunPosts()
	if ran != 2 || h.Frames() != before+1 {
		t.Fatalf("PostQuiet: ran %d, frames %d → %d; want no frame", ran, before+1, h.Frames())
	}

	// A listener of what the quiet closure emits asks for the frame.
	kernel.On(p.ctx, ping, func(string) { p.ctx.Invalidate() })
	p.ctx.PostQuiet(func() { kernel.Emit(p.ctx, ping, "on show") })
	h.RunPosts()
	if h.Frames() != before+2 {
		t.Fatalf("frames %d after an Invalidate in a quiet closure, want %d", h.Frames(), before+2)
	}
}

func TestTaskResultsAskForAFrame(t *testing.T) {
	p := plug("a", nil)
	h := kerneltest.New(t, p.New)
	h.Flush()
	before := h.Frames()
	var task kernel.Task
	got := ""
	kernel.Go(p.ctx, &task, func(context.Context) (string, error) { return "done", nil },
		func(v string, _ error) { got = v })
	h.Flush()
	if got != "done" || h.Frames() <= before {
		t.Fatalf("result %q, frames %d → %d; want the result applied and a frame", got, before, h.Frames())
	}
}

func TestQuietRedrawBoundsTheLag(t *testing.T) {
	p := plug("a", nil)
	h := kerneltest.Start(t, kerneltest.Config{QuietRedraw: 5 * time.Millisecond}, p.New)
	h.Flush()
	before := h.Frames()
	for range 3 {
		p.ctx.PostQuiet(func() {})
	}
	h.RunPosts()
	if h.Frames() != before {
		t.Fatalf("frames %d → %d at once after quiet posts, want none yet", before, h.Frames())
	}
	deadline := time.Now().Add(2 * time.Second)
	for h.Frames() == before {
		if time.Now().After(deadline) {
			t.Fatal("no frame came after quiet posts that asked for none")
		}
		time.Sleep(time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond)
	if h.Frames() != before+1 {
		t.Fatalf("frames %d → %d after the lag, want one for the three posts", before, h.Frames())
	}
}

func TestQuietPostWithoutAFrameHookStillRuns(t *testing.T) {
	var queue []func()
	k := kernel.New(kernel.Options{Post: func(fn func()) { queue = append(queue, fn) }, Store: kerneltest.NewStore(), Strict: true})
	p := plug("a", nil)
	if err := k.Install(p.New); err != nil {
		t.Fatal(err)
	}
	if err := k.Start(); err != nil {
		t.Fatal(err)
	}
	ran := false
	p.ctx.PostQuiet(func() { ran = true; p.ctx.Invalidate() })
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]
		fn()
	}
	if !ran {
		t.Fatal("the quiet closure did not run")
	}
	if err := k.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
