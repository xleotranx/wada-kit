// Package notebook declares this example's plugin contracts, not kit APIs.
// Every call and callback runs on the host thread (the UI thread in a desktop app).
package notebook

import "github.com/xleotranx/wada-kit/kernel"

// Notebook owns the in-memory titles. Add prepares a title, appends it, commits
// prepared effects, then emits Added. A preparation error leaves it unchanged.
// Titles returns an independent snapshot. Required consumers may retain Notebook
// only for their activation; optional consumers re-query NotebookKey on each use.
type Notebook interface {
	Add(title string) (string, error)
	Titles() []string
}

// Action is a host-invoked operation. Run returns display text or an error;
// the host invokes it through kernel.Invoke, never directly from Entry.Value.
type Action struct {
	Title string
	Run   func() (string, error)
}

var (
	// NotebookKey has exactly one provider, the notebook plugin.
	NotebookKey = kernel.NewKey[Notebook]("notebook.service")
	// Preparing runs transforms in Order, then registration order. Transforms
	// prepare without mutation and may return deferred commits. An error aborts
	// Add with no append, commit or Added event. Notebook trims/rejects blanks at
	// order 10; workspace adds its optional length policy at order 20.
	Preparing = kernel.NewPipeline[string]("notebook.preparing")
	// Added broadcasts the final title after append and commit, synchronously in
	// subscription order. Listeners observe an addition; they cannot veto it.
	Added = kernel.NewTopic[string]("notebook.added")
	// Actions is a collection consumed by the host. Entries sort by Order, then
	// plugin install order, then registration order. Each entry disappears with
	// its owner's scope; cached entries must still be invoked through Invoke.
	Actions = kernel.NewPoint[Action]("notebook.actions")
)
