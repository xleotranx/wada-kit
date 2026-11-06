package main

import (
	"testing"

	"github.com/xleotranx/wada-kit/kittest"
)

func TestArchitectureGuards(t *testing.T) {
	kittest.ColorsStayInThePalette(t, ".")
	kittest.PluginsSpawnTheirGoroutines(t, "internal/plugins")
}
