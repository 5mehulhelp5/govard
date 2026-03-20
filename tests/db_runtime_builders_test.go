package tests

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestResolveDBImportReaderForTestFromFile(t *testing.T) {
	tempDir := t.TempDir()
	dumpPath := filepath.Join(tempDir, "dump.sql")
	wantBody := "SELECT 1;\n"
	if err := os.WriteFile(dumpPath, []byte(wantBody), 0o644); err != nil {
		t.Fatalf("write dump file: %v", err)
	}

	reader, closer, _, err := cmd.ResolveDBImportReaderForTest(cmd.DBCommandOptions{File: dumpPath})
	if err != nil {
		t.Fatalf("ResolveDBImportReaderForTest() error = %v", err)
	}
	if closer == nil {
		t.Fatal("expected closer for file-backed reader")
	}
	t.Cleanup(func() {
		_ = closer.Close()
	})

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read import reader: %v", err)
	}
	if string(body) != wantBody {
		t.Fatalf("reader body = %q, want %q", string(body), wantBody)
	}
}

func TestResolveDBImportReaderForTestRejectsMissingInputOnTerminal(t *testing.T) {
	defer cmd.SetStdinIsTerminalForTest(func() bool { return true })()

	_, _, _, err := cmd.ResolveDBImportReaderForTest(cmd.DBCommandOptions{})
	if err == nil {
		t.Fatal("expected missing input error")
	}
	if !strings.Contains(err.Error(), "no import input provided") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveDBImportReaderForTestUsesStdinWhenPiped(t *testing.T) {
	defer cmd.SetStdinIsTerminalForTest(func() bool { return false })()

	reader, closer, _, err := cmd.ResolveDBImportReaderForTest(cmd.DBCommandOptions{})
	if err != nil {
		t.Fatalf("ResolveDBImportReaderForTest() error = %v", err)
	}
	if closer != nil {
		t.Fatal("expected nil closer for stdin reader")
	}
	if reader != os.Stdin {
		t.Fatal("expected stdin reader when terminal detection is false")
	}
}

func TestBuildDBDumpCommandForTestLocalUsesDockerInspectAndCredentials(t *testing.T) {
	shimDir := installDockerInspectShim(t)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	args, err := cmd.BuildDBDumpCommandForTest(
		engine.Config{ProjectName: "sample-project"},
		cmd.DBCommandOptions{Environment: "local", Full: true},
	)
	if err != nil {
		t.Fatalf("BuildDBDumpCommandForTest() error = %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "docker exec -i -e MYSQL_PWD=devpass sample-project-db-1 mysqldump") {
		t.Fatalf("unexpected local dump command: %s", joined)
	}
	if !strings.Contains(joined, "--routines --events --triggers") {
		t.Fatalf("expected full dump flags in command: %s", joined)
	}
	if !strings.Contains(joined, " devdb") {
		t.Fatalf("expected resolved database name in command: %s", joined)
	}
}

func TestBuildDBImportCommandForTestLocalUsesDockerExecWithForceMode(t *testing.T) {
	shimDir := installDockerInspectShim(t)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	args, err := cmd.BuildDBImportCommandForTest(
		engine.Config{ProjectName: "sample-project"},
		cmd.DBCommandOptions{Environment: "local"},
	)
	if err != nil {
		t.Fatalf("BuildDBImportCommandForTest() error = %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "docker exec -i -e MYSQL_PWD=devpass sample-project-db-1 sh -lc") {
		t.Fatalf("unexpected local import command: %s", joined)
	}
	if !strings.Contains(joined, "-u 'devuser' 'devdb' -f") {
		t.Fatalf("expected mysql force import flags in command: %s", joined)
	}
}

func TestBuildDBDumpCommandForTestRemoteBuildsSSHCommand(t *testing.T) {
	config := engine.Config{
		Framework: "unknown",
		Remotes: map[string]engine.RemoteConfig{
			"dev": {
				Host: "example.com",
				User: "deploy",
				Capabilities: engine.RemoteCapabilities{
					DB: true,
				},
			},
		},
	}

	args, err := cmd.BuildDBDumpCommandForTest(config, cmd.DBCommandOptions{Environment: "dev"})
	if err != nil {
		t.Fatalf("BuildDBDumpCommandForTest() error = %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.HasPrefix(joined, "ssh ") {
		t.Fatalf("expected ssh command, got: %s", joined)
	}
	if !strings.Contains(joined, "deploy@example.com") {
		t.Fatalf("expected remote target in ssh command, got: %s", joined)
	}
	if !strings.Contains(joined, "mysqldump") {
		t.Fatalf("expected mysqldump command, got: %s", joined)
	}
}

func TestBuildDBImportCommandForTestRemoteBuildsSSHCommand(t *testing.T) {
	config := engine.Config{
		Framework: "unknown",
		Remotes: map[string]engine.RemoteConfig{
			"dev": {
				Host: "example.com",
				User: "deploy",
				Capabilities: engine.RemoteCapabilities{
					DB: true,
				},
			},
		},
	}

	args, err := cmd.BuildDBImportCommandForTest(config, cmd.DBCommandOptions{Environment: "dev"})
	if err != nil {
		t.Fatalf("BuildDBImportCommandForTest() error = %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.HasPrefix(joined, "ssh ") {
		t.Fatalf("expected ssh command, got: %s", joined)
	}
	if !strings.Contains(joined, "deploy@example.com") {
		t.Fatalf("expected remote target in ssh command, got: %s", joined)
	}
	if !strings.Contains(joined, "mysql") {
		t.Fatalf("expected mysql import command, got: %s", joined)
	}
}

func TestResolveDBRemoteForTestValidation(t *testing.T) {
	t.Run("UnknownRemote", func(t *testing.T) {
		_, err := cmd.ResolveDBRemoteForTest(engine.Config{Remotes: map[string]engine.RemoteConfig{}}, "dev", false)
		if err == nil {
			t.Fatal("expected unknown remote error")
		}
		if !strings.Contains(err.Error(), "unknown remote: dev") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("CapabilityMissing", func(t *testing.T) {
		config := engine.Config{
			Remotes: map[string]engine.RemoteConfig{
				"dev": {},
			},
		}
		_, err := cmd.ResolveDBRemoteForTest(config, "dev", false)
		if err == nil {
			t.Fatal("expected capability error")
		}
		if !strings.Contains(err.Error(), "does not allow db operations") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ProdBlockedForReadFlow", func(t *testing.T) {
		config := engine.Config{
			Remotes: map[string]engine.RemoteConfig{
				"prod": {
					Capabilities: engine.RemoteCapabilities{DB: true},
				},
			},
		}
		_, err := cmd.ResolveDBRemoteForTest(config, "prod", false)
		if err == nil {
			t.Fatal("expected write-protection error")
		}
		if !strings.Contains(err.Error(), "write-protected") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ProdAllowedForWriteFlow", func(t *testing.T) {
		config := engine.Config{
			Remotes: map[string]engine.RemoteConfig{
				"prod": {
					Capabilities: engine.RemoteCapabilities{DB: true},
				},
			},
		}
		_, err := cmd.ResolveDBRemoteForTest(config, "prod", true)
		if err != nil {
			t.Fatalf("expected forWrite flow to bypass protection, got: %v", err)
		}
	})
}

func installDockerInspectShim(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	script := `#!/bin/sh
set -eu
if [ "${1:-}" = "inspect" ] && [ "${2:-}" = "-f" ]; then
  case "${3:-}" in
    "{{.State.Running}}")
      echo "true"
      exit 0
      ;;
    "{{range .Config.Env}}{{println .}}{{end}}")
      cat <<'EOF'
MYSQL_USER=devuser
MYSQL_PASSWORD=devpass
MYSQL_DATABASE=devdb
EOF
      exit 0
      ;;
  esac
fi
echo "unexpected docker args: $*" >&2
exit 1
`
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write docker shim: %v", err)
	}
	return dir
}
