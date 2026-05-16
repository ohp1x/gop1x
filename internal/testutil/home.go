package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func SetupTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("OHP1X_HOME", home)
	must(t, os.MkdirAll(filepath.Join(home, "presets"), 0755))
	must(t, os.MkdirAll(filepath.Join(home, "local"), 0755))
	must(t, os.MkdirAll(filepath.Join(home, "generated"), 0755))
	return home
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
