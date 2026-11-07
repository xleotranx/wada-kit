package notebook_test

import (
	"slices"
	"testing"

	contract "github.com/xleotranx/wada-kit/examples/extensions/internal/contract/notebook"
	"github.com/xleotranx/wada-kit/examples/extensions/internal/plugins/notebook"
	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestAddPreparesCommitsAndPublishes(t *testing.T) {
	h := kerneltest.New(t, notebook.New)
	book, ok := kernel.Use(h, contract.NotebookKey)
	if !ok {
		t.Fatal("notebook service unavailable")
	}
	committed := false
	kernel.Transform(h, contract.Preparing, 30, func(title string) (string, func(), error) {
		return title, func() { committed = true }, nil
	})
	var events []string
	kernel.On(h, contract.Added, func(title string) {
		if !committed || !slices.Equal(book.Titles(), []string{title}) {
			t.Error("event arrived before append and commit")
		}
		events = append(events, title)
	})
	title, err := book.Add("  Draft  ")
	if err != nil || title != "Draft" || !slices.Equal(events, []string{"Draft"}) {
		t.Fatalf("add = %q, %v; events = %v", title, err, events)
	}
	snapshot := book.Titles()
	snapshot[0] = "edited snapshot"
	if !slices.Equal(book.Titles(), []string{"Draft"}) {
		t.Fatal("snapshot mutation changed notebook state")
	}
}

func TestBlankTitleHasNoEffects(t *testing.T) {
	h := kerneltest.New(t, notebook.New)
	book, _ := kernel.Use(h, contract.NotebookKey)
	events := 0
	kernel.On(h, contract.Added, func(string) { events++ })
	if _, err := book.Add(" \t "); err == nil {
		t.Fatal("blank title accepted")
	}
	if len(book.Titles()) != 0 || events != 0 {
		t.Fatal("rejected title changed notebook or emitted an event")
	}
}
