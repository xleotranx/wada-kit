// Extensions runs a headless two-plugin notebook. Run it from the kit checkout:
//
//	go run ./examples/extensions
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	contract "github.com/xleotranx/wada-kit/examples/extensions/internal/contract/notebook"
	"github.com/xleotranx/wada-kit/examples/extensions/internal/plugins/notebook"
	"github.com/xleotranx/wada-kit/examples/extensions/internal/plugins/workspace"
	"github.com/xleotranx/wada-kit/kernel"
)

func main() {
	if err := run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(out io.Writer) (err error) {
	k := kernel.New(kernel.Options{
		// This example has no workers; all calls stay on this goroutine.
		// A desktop app uses the shell's UI scheduler and settings store.
		Post: func(fn func()) { fn() }, Store: noSettings{}, Strict: true,
	})
	defer func() { err = errors.Join(err, k.Close(context.Background())) }()
	// Installation order is deliberately opposite to the service dependency.
	if err := k.Install(workspace.New, notebook.New); err != nil {
		return err
	}
	if err := k.Start(); err != nil {
		return err
	}
	for _, entry := range kernel.Items(k, contract.Actions) {
		var result string
		var actionErr error
		kernel.Invoke(k, entry, func(action contract.Action) { result, actionErr = action.Run() })
		if actionErr != nil {
			return fmt.Errorf("%s: %w", entry.ID, actionErr)
		}
		if _, err := fmt.Fprintf(out, "%s: %s\n", entry.Value.Title, result); err != nil {
			return err
		}
	}
	return nil
}

// noSettings makes the lack of persistence explicit; these plugins use no settings.
type noSettings struct{}

func (noSettings) Load(context.Context) ([]kernel.Value, error) { return nil, nil }
func (noSettings) Save(context.Context, []kernel.Value) error {
	return errors.New("this example does not persist settings")
}
