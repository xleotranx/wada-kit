package documents

import (
	"context"
	"errors"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/xleotranx/wada-kit/contract/shell"
	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
	"github.com/xleotranx/wada-kit/kittest"
)

type repository struct {
	fail      bool
	cancelled int
}

func (r *repository) Load(ctx context.Context, id string) (string, error) {
	if ctx.Err() != nil {
		r.cancelled++
	}
	if r.fail {
		return "", errors.New("offline")
	}
	return "Body of " + id, nil // deliberately ignores cancellation to test stale results
}

func start(t *testing.T, store *kerneltest.Store, repo *repository) (*kerneltest.Harness, *Plugin, *kittest.Host) {
	t.Helper()
	host := &kittest.Host{}
	var p *Plugin
	h := kerneltest.Start(t, kerneltest.Config{Store: store}, host.New, kerneltest.Keep(&p, func() *Plugin { return &Plugin{repo: repo} }))
	h.RunPosts() // finish startup before controlling document work
	return h, p, host
}

func TestSwitchDiscardsOldDocumentAndLeavingCancelsLoad(t *testing.T) {
	repo := &repository{}
	h, p, host := start(t, nil, repo)
	tt := ui.NewTester(p.view, 900, 640)
	tt.Frame()
	if !tt.HasText("Loading…") {
		t.Fatal(tt.Texts())
	}
	p.open("workflow.txt")
	h.RunWork(1) // second document finishes first
	h.RunPosts()
	h.RunWork(0) // old repository returns despite cancellation
	h.RunPosts()
	tt.Frame()
	if !tt.HasText("Body of workflow.txt") || tt.HasText("Body of welcome.txt") {
		t.Fatal(tt.Texts())
	}
	h.Flush()
	p.open("welcome.txt")
	host.ShowPage("settings.page")
	h.Flush()
	if repo.cancelled != 2 {
		t.Fatalf("cancelled loads = %d", repo.cancelled)
	}
	host.ShowPage(pageID)
	h.Flush()
	tt.Frame()
	if !tt.HasText("Body of workflow.txt") {
		t.Fatal(tt.Texts())
	}
}

func TestRetryAndRememberDocumentAcrossActivations(t *testing.T) {
	store := kerneltest.NewStore()
	repo := &repository{fail: true}
	h, p, _ := start(t, store, repo)
	h.Flush()
	tt := ui.NewTester(p.view, 900, 640)
	tt.Frame()
	if !tt.HasText("offline") {
		t.Fatal(tt.Texts())
	}
	repo.fail = false
	if err := tt.Click("Retry"); err != nil {
		t.Fatal(err)
	}
	h.Flush()
	p.open("workflow.txt")
	h.Flush()
	h.Close()
	h2, p2, _ := start(t, store, repo)
	h2.Flush()
	tt2 := ui.NewTester(p2.view, 900, 640)
	tt2.Frame()
	if !tt2.HasText("Body of workflow.txt") {
		t.Fatal(tt2.Texts())
	}
}

func TestDisableDropsPendingDocument(t *testing.T) {
	repo := &repository{}
	h, _, _ := start(t, nil, repo)
	if err := h.Disable("documents"); err != nil {
		t.Fatal(err)
	}
	h.Flush()
	if repo.cancelled != 1 {
		t.Fatalf("cancelled loads = %d", repo.cancelled)
	}
	if len(kernel.Items(h, shell.Pages)) != 0 {
		t.Fatal("page survived disable")
	}
}
