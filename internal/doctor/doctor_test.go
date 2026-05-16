package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestRun_UninitializedHome(t *testing.T) {
	tmpdir := t.TempDir()
	nonexistSubdir := filepath.Join(tmpdir, "nonexist")
	t.Setenv("OHP1X_HOME", nonexistSubdir)

	report := Run()
	assert.False(t, report.OK)
	assert.NotEmpty(t, report.Issues)

	hasError := false
	for _, issue := range report.Issues {
		if issue.Level == "error" && contains(issue.Message, "Home directory not initialized") {
			hasError = true
			break
		}
	}
	assert.True(t, hasError, "Expected home dir not initialized error")
}

func TestRun_InitializedHome(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("OHP1X_HOME", tmpdir)

	// Create required directories
	dirs := []string{
		filepath.Join(tmpdir, "presets"),
		filepath.Join(tmpdir, "local"),
		filepath.Join(tmpdir, "generated"),
	}
	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}

	report := Run()

	// May have warnings (no package manager, etc) but no errors about missing dirs
	for _, issue := range report.Issues {
		assert.NotContains(t, issue.Message, "Missing directory")
	}
}

func TestRun_MissingSubdirectory(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("OHP1X_HOME", tmpdir)

	// Create only some required directories
	require.NoError(t, os.MkdirAll(filepath.Join(tmpdir, "presets"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpdir, "local"), 0755))
	// Missing "generated"

	report := Run()
	assert.False(t, report.OK)

	hasError := false
	for _, issue := range report.Issues {
		if issue.Level == "error" && contains(issue.Message, "Missing directory") && contains(issue.Message, "generated") {
			hasError = true
			break
		}
	}
	assert.True(t, hasError, "Expected error for missing 'generated' directory")
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		issues   []Issue
		expected bool
	}{
		{
			name:     "no issues",
			issues:   []Issue{},
			expected: false,
		},
		{
			name: "only warnings",
			issues: []Issue{
				{Level: "warning", Message: "test"},
			},
			expected: false,
		},
		{
			name: "with error",
			issues: []Issue{
				{Level: "warning", Message: "test"},
				{Level: "error", Message: "test error"},
			},
			expected: true,
		},
		{
			name: "only errors",
			issues: []Issue{
				{Level: "error", Message: "test error"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &Report{Issues: tt.issues}
			assert.Equal(t, tt.expected, hasErrors(report))
		})
	}
}

