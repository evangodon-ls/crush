package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/home"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func() string
	}{
		{
			name:  "regular path unchanged",
			input: "/absolute/path",
			expected: func() string {
				return "/absolute/path"
			},
		},
		{
			name:  "tilde expansion",
			input: "~/documents",
			expected: func() string {
				return home.Dir() + "/documents"
			},
		},
		{
			name:  "tilde only",
			input: "~",
			expected: func() string {
				return home.Dir()
			},
		},
		{
			name:  "environment variable expansion",
			input: "$HOME",
			expected: func() string {
				return os.Getenv("HOME")
			},
		},
		{
			name:  "relative path unchanged",
			input: "relative/path",
			expected: func() string {
				return "relative/path"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			expected := tt.expected()

			// Skip test if environment variable is not set
			if strings.HasPrefix(tt.input, "$") && expected == "" {
				t.Skip("Environment variable not set")
			}

			if result != expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, expected)
			}
		})
	}
}

func TestGetGlobalContext(t *testing.T) {
	t.Run("returns empty string when file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_DATA_HOME", tmpDir)

		result := getGlobalContext()
		require.Empty(t, result)
	})

	t.Run("returns content when file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_DATA_HOME", tmpDir)

		globalPath := config.GlobalContextPath()
		require.NoError(t, os.MkdirAll(filepath.Dir(globalPath), 0o755))

		content := "# Global coding preferences\n- Use tabs for indentation\n- Prefer functional programming"
		require.NoError(t, os.WriteFile(globalPath, []byte(content), 0o644))

		result := getGlobalContext()
		require.Contains(t, result, content)
		require.Contains(t, result, "# From:")
	})

	t.Run("returns empty string when file is empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_DATA_HOME", tmpDir)

		globalPath := config.GlobalContextPath()
		require.NoError(t, os.MkdirAll(filepath.Dir(globalPath), 0o755))
		require.NoError(t, os.WriteFile(globalPath, []byte(""), 0o644))

		result := getGlobalContext()
		require.Empty(t, result)
	})
}

func TestCoderPromptWithGlobalContext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	workDir := t.TempDir()
	_, err := config.Init(workDir, filepath.Join(workDir, ".crush"), false)
	require.NoError(t, err)

	t.Run("includes global context when file exists", func(t *testing.T) {
		globalPath := config.GlobalContextPath()
		require.NoError(t, os.MkdirAll(filepath.Dir(globalPath), 0o755))

		globalContent := "# Global preferences\n- Always use semantic commits"
		require.NoError(t, os.WriteFile(globalPath, []byte(globalContent), 0o644))

		prompt := CoderPrompt("anthropic")
		require.Contains(t, prompt, "Global Context")
		require.Contains(t, prompt, globalContent)
	})

	t.Run("includes both global and project context", func(t *testing.T) {
		globalPath := config.GlobalContextPath()
		require.NoError(t, os.MkdirAll(filepath.Dir(globalPath), 0o755))

		globalContent := "# Global preferences\n- Use tabs"
		require.NoError(t, os.WriteFile(globalPath, []byte(globalContent), 0o644))

		projectContextPath := filepath.Join(workDir, "CRUSH.md")
		projectContent := "# Project preferences\n- Use spaces"
		require.NoError(t, os.WriteFile(projectContextPath, []byte(projectContent), 0o644))

		prompt := CoderPrompt("anthropic", projectContextPath)
		require.Contains(t, prompt, "Global Context")
		require.Contains(t, prompt, globalContent)
		require.Contains(t, prompt, "Project Context")
		require.Contains(t, prompt, projectContent)
	})
}
