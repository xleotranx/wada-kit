package main

import (
	"errors"
	"slices"
	"strings"
	"testing"

	contract "github.com/xleotranx/wada-kit/examples/extensions/internal/contract/notebook"
	"github.com/xleotranx/wada-kit/examples/extensions/internal/plugins/notebook"
	"github.com/xleotranx/wada-kit/examples/extensions/internal/plugins/workspace"
	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestRun(t *testing.T) {
	var out strings.Builder
	if err := run(&out); err != nil {
		t.Fatal(err)
	}
	want := "Add welcome note: Hello plugins (committed=1, observed=1)\nList notes: Hello plugins\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestTwoPluginFlow(t *testing.T) {
	h := kerneltest.New(t, workspace.New, notebook.New) // consumer installed first
	book := notebookService(t, h)
	actions := kernel.Items(h, contract.Actions)
	if len(actions) != 2 || actions[0].ID != "workspace.add" || actions[1].ID != "notebook.list" {
		t.Fatalf("actions not ordered by their declared Order: %+v", actions)
	}
	prepared, commit, err := kernel.Run(h, contract.Preparing, "  discarded  ")
	if err != nil || prepared != "discarded" || commit == nil {
		t.Fatalf("prepare = %q, %v", prepared, err)
	}
	// Drop the prepared result: no notebook append, counter increment or event.
	if len(book.Titles()) != 0 {
		t.Fatal("preparation changed the notebook")
	}
	_, err = book.Add(strings.Repeat("x", 33))
	var pluginErr *kernel.PluginError
	if !errors.As(err, &pluginErr) || pluginErr.Plugin != "workspace" {
		t.Fatalf("length policy error = %v", err)
	}
	if got := invokeAction(t, h, "workspace.add"); got != "Hello plugins (committed=1, observed=1)" {
		t.Fatalf("dropped or rejected preparation changed counters: %s", got)
	}
	if got := invokeAction(t, h, "notebook.list"); got != "Hello plugins" {
		t.Fatalf("list = %q", got)
	}
	for _, st := range h.Plugins() {
		if st.State != kernel.Active {
			t.Fatalf("ordinary rejection failed a plugin: %+v", st)
		}
	}
}

func TestDownstreamFailureDoesNotCommitOrPublish(t *testing.T) {
	h := kerneltest.New(t, workspace.New, notebook.New)
	book := notebookService(t, h)
	kernel.Transform(h, contract.Preparing, 30, func(title string) (string, func(), error) {
		if title == "blocked" {
			return "", nil, errors.New("later stage rejected title")
		}
		return title, nil, nil
	})
	events := 0
	kernel.On(h, contract.Added, func(string) { events++ })
	if _, err := book.Add("  blocked  "); err == nil {
		t.Fatal("later-stage rejection ignored")
	}
	if len(book.Titles()) != 0 || events != 0 {
		t.Fatal("failed chain appended a title or published an event")
	}
	if got := invokeAction(t, h, "workspace.add"); got != "Hello plugins (committed=1, observed=1)" {
		t.Fatalf("failed chain committed an earlier stage: %s", got)
	}
}

func TestLengthPolicyUsesPreparedRunes(t *testing.T) {
	h := kerneltest.New(t, workspace.New, notebook.New)
	book := notebookService(t, h)
	title := strings.Repeat("é", 32)
	if got, err := book.Add("  " + title + "  "); err != nil || got != title {
		t.Fatalf("policy should count trimmed runes: title = %q, error = %v", got, err)
	}
	if _, err := book.Add(title + "é"); err == nil {
		t.Fatal("policy accepted 33 runes")
	}
}

func TestWorkspaceDisposalAndRecovery(t *testing.T) {
	h := kerneltest.New(t, workspace.New, notebook.New)
	book := notebookService(t, h)
	invokeAction(t, h, "workspace.add")
	stale := kernel.Items(h, contract.Actions)[0]
	if err := h.Disable("workspace"); err != nil {
		t.Fatal(err)
	}
	if actions := kernel.Items(h, contract.Actions); len(actions) != 1 || actions[0].ID != "notebook.list" {
		t.Fatalf("workspace action survived disposal: %+v", actions)
	}
	called := false
	kernel.Invoke(h, stale, func(contract.Action) { called = true })
	if called {
		t.Fatal("disposed action was invoked")
	}
	long := strings.Repeat("x", 33)
	if _, err := book.Add(long); err != nil {
		t.Fatalf("optional workspace policy survived disposal: %v", err)
	}
	if err := h.Enable("workspace"); err != nil {
		t.Fatal(err)
	}
	if got := invokeAction(t, h, "workspace.add"); got != "Hello plugins (committed=1, observed=1)" {
		t.Fatalf("workspace state was not fresh: %s", got)
	}
	if !slices.Equal(book.Titles(), []string{"Hello plugins", long, "Hello plugins"}) {
		t.Fatal("workspace reactivation changed the notebook's existing titles")
	}
}

func TestProviderDisposalAndRecovery(t *testing.T) {
	h := kerneltest.New(t, workspace.New, notebook.New)
	old := notebookService(t, h)
	invokeAction(t, h, "workspace.add")
	if err := h.Disable("notebook"); err != nil {
		t.Fatal(err)
	}
	if _, ok := kernel.Use(h, contract.NotebookKey); ok || len(kernel.Items(h, contract.Actions)) != 0 {
		t.Fatal("provider disposal retained its service or dependent contributions")
	}
	if st := h.Plugins()[0]; st.State != kernel.Blocked {
		t.Fatalf("required consumer not blocked: %+v", st)
	}
	if _, err := old.Add("stale call"); err == nil {
		t.Fatal("old service activation accepted an addition")
	}
	if err := h.Enable("notebook"); err != nil {
		t.Fatal(err)
	}
	if len(notebookService(t, h).Titles()) != 0 {
		t.Fatal("new notebook activation retained old in-memory state")
	}
	if got := invokeAction(t, h, "workspace.add"); got != "Hello plugins (committed=1, observed=1)" {
		t.Fatalf("dependent did not reactivate against the new service: %s", got)
	}
}

func notebookService(t *testing.T, h *kerneltest.Harness) contract.Notebook {
	t.Helper()
	book, ok := kernel.Use(h, contract.NotebookKey)
	if !ok {
		t.Fatal("notebook service unavailable")
	}
	return book
}

func invokeAction(t *testing.T, h *kerneltest.Harness, id string) string {
	t.Helper()
	for _, entry := range kernel.Items(h, contract.Actions) {
		if entry.ID == id {
			var result string
			var err error
			kernel.Invoke(h, entry, func(action contract.Action) { result, err = action.Run() })
			if err != nil {
				t.Fatal(err)
			}
			return result
		}
	}
	t.Fatalf("action %q unavailable", id)
	return ""
}
