// Package documents owns document access without knowing about plugins or UI.
package documents

import (
	"context"
	"embed"
	"io/fs"
)

// Repository implementations are worker-safe and return owned values.
type Repository interface {
	Load(context.Context, string) (string, error)
}

// Files reads relative paths from an app-owned filesystem. Use os.DirFS for
// on-disk documents; the caller chooses the root, never the view.
type Files struct{ FS fs.FS }

func (r Files) Load(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	body, err := fs.ReadFile(r.FS, id)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return string(body), nil
}

//go:embed samples/*.txt
var samples embed.FS

// Samples keeps the starter runnable without touching user files.
func Samples() Repository {
	files, err := fs.Sub(samples, "samples")
	if err != nil {
		panic(err)
	}
	return Files{FS: files}
}
