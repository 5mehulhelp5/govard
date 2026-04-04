package bootstrap

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func removeProjectContents(projectDir string) error {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read project directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == ".govard" || entry.Name() == ".govard.yml" {
			continue
		}
		targetPath := filepath.Join(projectDir, entry.Name())
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func runStagedCreateProject(projectDir string, runner func(command string) error, createInStage func(stageDir string) error, runnerCommand string) error {
	if err := removeProjectContents(projectDir); err != nil {
		return err
	}

	stageDir, err := os.MkdirTemp(projectDir, "govard-create-project-")
	if err != nil {
		return fmt.Errorf("create staged project directory: %w", err)
	}
	defer os.RemoveAll(stageDir)

	if runner != nil {
		if strings.TrimSpace(runnerCommand) == "" {
			return fmt.Errorf("runner command is required for staged project creation")
		}
		if err := runner(buildStagedRunnerCommand(projectDir, stageDir, runnerCommand)); err != nil {
			return err
		}
	} else {
		if createInStage == nil {
			return fmt.Errorf("staged project creator is required")
		}
		if err := createInStage(stageDir); err != nil {
			return err
		}
	}

	if err := syncStagedProject(stageDir, projectDir); err != nil {
		return err
	}

	return nil
}

func buildStagedRunnerCommand(projectDir, stageDir, runnerCommand string) string {
	containerStageDir := "/var/www/html"
	if relStageDir, err := filepath.Rel(projectDir, stageDir); err == nil && relStageDir != "." {
		containerStageDir = path.Join(containerStageDir, filepath.ToSlash(relStageDir))
	}

	return fmt.Sprintf(
		"export GOVARD_STAGE_DIR=%s GOVARD_STAGE_HOST_DIR=%s; %s",
		shellQuote(containerStageDir),
		shellQuote(stageDir),
		runnerCommand,
	)
}

func syncStagedProject(stageDir, projectDir string) error {
	entries, err := os.ReadDir(stageDir)
	if err != nil {
		return fmt.Errorf("read staged project directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(stageDir, entry.Name())
		dstPath := filepath.Join(projectDir, entry.Name())
		if err := copyProjectPath(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func copyProjectPath(srcPath, dstPath string) error {
	info, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("stat staged path %s: %w", srcPath, err)
	}

	switch mode := info.Mode(); {
	case mode.IsDir():
		if err := os.MkdirAll(dstPath, mode.Perm()); err != nil {
			return fmt.Errorf("create project directory %s: %w", dstPath, err)
		}
		entries, err := os.ReadDir(srcPath)
		if err != nil {
			return fmt.Errorf("read staged directory %s: %w", srcPath, err)
		}
		for _, entry := range entries {
			if err := copyProjectPath(filepath.Join(srcPath, entry.Name()), filepath.Join(dstPath, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	case mode&os.ModeSymlink != 0:
		target, err := os.Readlink(srcPath)
		if err != nil {
			return fmt.Errorf("read staged symlink %s: %w", srcPath, err)
		}
		if err := os.Symlink(target, dstPath); err != nil {
			return fmt.Errorf("create project symlink %s: %w", dstPath, err)
		}
		return nil
	default:
		return copyProjectFile(srcPath, dstPath, info.Mode().Perm())
	}
}

func copyProjectFile(srcPath, dstPath string, mode os.FileMode) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open staged file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create project parent directory %s: %w", filepath.Dir(dstPath), err)
	}

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create project file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy staged file %s: %w", srcPath, err)
	}

	return nil
}

func runComposerProjectCommand(projectDir string, runner func(command string) error, args ...string) error {
	if runner != nil {
		commandArgs := make([]string, len(args))
		for i, arg := range args {
			commandArgs[i] = shellQuote(arg)
		}
		return runner("composer " + strings.Join(commandArgs, " "))
	}

	cmd := exec.Command("composer", args...)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runPHPProjectScript(projectDir string, runner func(command string) error, scriptPath string, args ...string) error {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil
	}

	if runner != nil {
		runnerScriptPath := scriptPath
		if relScriptPath, err := filepath.Rel(projectDir, scriptPath); err == nil && !strings.HasPrefix(relScriptPath, "..") {
			runnerScriptPath = path.Join("/var/www/html", filepath.ToSlash(relScriptPath))
		}

		commandArgs := []string{"php", shellQuote(runnerScriptPath)}
		for _, arg := range args {
			commandArgs = append(commandArgs, shellQuote(arg))
		}
		return runner(strings.Join(commandArgs, " "))
	}

	cmd := exec.Command("php", append([]string{scriptPath}, args...)...)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runPHPOneLiner(projectDir string, runner func(command string) error, code string) error {
	if runner != nil {
		return runner("php -r " + shellQuote(code))
	}

	cmd := exec.Command("php", "-r", code)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func RunStagedCreateProjectForTest(projectDir string, runner func(command string) error, createInStage func(stageDir string) error, runnerCommand string) error {
	return runStagedCreateProject(projectDir, runner, createInStage, runnerCommand)
}
