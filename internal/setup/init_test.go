package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRun_FreshInit(t *testing.T) {
	tmpdir := t.TempDir()
	homeDir := filepath.Join(tmpdir, "ohp1x_home")
	t.Setenv("OHP1X_HOME", homeDir)

	opts := InitOptions{Force: false}
	err := Run(opts)
	require.NoError(t, err)

	// Verify directory structure
	dirs := []string{
		homeDir,
		filepath.Join(homeDir, "presets"),
		filepath.Join(homeDir, "generated"),
		filepath.Join(homeDir, "local"),
	}
	for _, dir := range dirs {
		assert.DirExists(t, dir)
	}

	// Verify .gitignore
	gitignore := filepath.Join(homeDir, ".gitignore")
	assert.FileExists(t, gitignore)
	content, err := os.ReadFile(gitignore)
	require.NoError(t, err)
	assert.Equal(t, "local/\n", string(content))

	// Verify ohp1x.yaml
	manifest := filepath.Join(homeDir, "ohp1x.yaml")
	assert.FileExists(t, manifest)

	// Verify init scripts
	assert.FileExists(t, filepath.Join(homeDir, "init.zsh"))
	assert.FileExists(t, filepath.Join(homeDir, "init.bash"))

	// Verify git repo
	assert.DirExists(t, filepath.Join(homeDir, ".git"))
}

func TestRun_AlreadyInitialized_WithoutForce(t *testing.T) {
	tmpdir := t.TempDir()
	homeDir := filepath.Join(tmpdir, "ohp1x_home")
	t.Setenv("OHP1X_HOME", homeDir)

	// First init
	err := Run(InitOptions{Force: false})
	require.NoError(t, err)

	// Second init without --force should fail
	err = Run(InitOptions{Force: false})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

func TestRun_AlreadyInitialized_WithForce(t *testing.T) {
	tmpdir := t.TempDir()
	homeDir := filepath.Join(tmpdir, "ohp1x_home")
	t.Setenv("OHP1X_HOME", homeDir)

	// First init
	err := Run(InitOptions{Force: false})
	require.NoError(t, err)

	// Create a marker file
	markerFile := filepath.Join(homeDir, "marker.txt")
	err = os.WriteFile(markerFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Second init with --force should succeed and backup
	err = Run(InitOptions{Force: true})
	require.NoError(t, err)

	// Old marker should be in backup
	backups, _ := filepath.Glob(homeDir + ".backup-*")
	assert.GreaterOrEqual(t, len(backups), 1)

	// New home should not have marker
	assert.NoFileExists(t, markerFile)

	// New home should be initialized
	assert.FileExists(t, filepath.Join(homeDir, "ohp1x.yaml"))
}

func TestCreateDirectories(t *testing.T) {
	tmpdir := t.TempDir()
	homeDir := filepath.Join(tmpdir, "new_home")

	err := createDirectories(homeDir)
	require.NoError(t, err)

	for _, subdir := range []string{"", "presets", "generated", "local"} {
		dir := filepath.Join(homeDir, subdir)
		assert.DirExists(t, dir)
	}
}

func TestWriteManifest(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("OHP1X_HOME", tmpdir)

	err := writeManifest(tmpdir)
	require.NoError(t, err)

	manifestPath := filepath.Join(tmpdir, "ohp1x.yaml")
	assert.FileExists(t, manifestPath)

	content, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var manifest MinimalManifest
	err = yaml.Unmarshal(content, &manifest)
	require.NoError(t, err)

	assert.NotEmpty(t, manifest.Shell)
	assert.Empty(t, manifest.Presets)
}

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		expected string
	}{
		{"zsh path", "/bin/zsh", "zsh"},
		{"bash path", "/bin/bash", "bash"},
		{"fish path", "/usr/bin/fish", "fish"},
		{"empty", "", "zsh"}, // default
		{"unknown shell", "/bin/unknownshell", "zsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shellEnv)
			assert.Equal(t, tt.expected, detectShell())
		})
	}
}

func TestGenerateInitScripts(t *testing.T) {
	tmpdir := t.TempDir()

	err := generateInitScripts(tmpdir)
	require.NoError(t, err)

	zshPath := filepath.Join(tmpdir, "init.zsh")
	bashPath := filepath.Join(tmpdir, "init.bash")

	assert.FileExists(t, zshPath)
	assert.FileExists(t, bashPath)

	// Check content
	zshContent, err := os.ReadFile(zshPath)
	require.NoError(t, err)
	assert.Contains(t, string(zshContent), "OHP1X_HOME")

	bashContent, err := os.ReadFile(bashPath)
	require.NoError(t, err)
	assert.Contains(t, string(bashContent), "OHP1X_HOME")
}

func TestWriteGitignore(t *testing.T) {
	tmpdir := t.TempDir()

	err := writeGitignore(tmpdir)
	require.NoError(t, err)

	gitignorePath := filepath.Join(tmpdir, ".gitignore")
	assert.FileExists(t, gitignorePath)

	content, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "local/\n", string(content))
}

func TestRun_InitFromURL(t *testing.T) {
	tmpdir := t.TempDir()
	homeDir := filepath.Join(tmpdir, "ohp1x_home")
	t.Setenv("OHP1X_HOME", homeDir)

	// Create a temporary git repo to clone from
	sourceDir := filepath.Join(tmpdir, "source_repo")
	err := os.MkdirAll(sourceDir, 0755)
	require.NoError(t, err)

	// Initialize source repo
	cmd := exec.Command("git", "init")
	cmd.Dir = sourceDir
	err = cmd.Run()
	require.NoError(t, err)

	// Create a test file in source repo
	testFile := filepath.Join(sourceDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Commit the test file
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = sourceDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = sourceDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = sourceDir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = sourceDir
	_ = cmd.Run()

	// Clone from source repo
	opts := InitOptions{FromURL: sourceDir}
	err = Run(opts)
	require.NoError(t, err)

	// Verify cloned repo exists
	assert.DirExists(t, homeDir)
	assert.FileExists(t, filepath.Join(homeDir, "test.txt"))
	assert.DirExists(t, filepath.Join(homeDir, ".git"))

	// Verify init scripts were generated
	assert.FileExists(t, filepath.Join(homeDir, "init.zsh"))
	assert.FileExists(t, filepath.Join(homeDir, "init.bash"))
}
