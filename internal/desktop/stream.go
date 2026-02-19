package desktop

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
)

func streamLogs(ctx context.Context, appCtx context.Context, project string, service string) {
	info, err := loadProjectInfo(project)
	if err != nil {
		emitEvent(appCtx, "logs:error", "Failed to load project: "+err.Error())
		return
	}

	containerName := resolveLogContainer(info, service)
	cmd := exec.CommandContext(ctx, "docker", "logs", "--tail", "100", "-f", containerName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		emitEvent(appCtx, "logs:error", "Failed to stream logs: "+err.Error())
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		emitEvent(appCtx, "logs:error", "Failed to stream logs: "+err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		emitEvent(appCtx, "logs:error", "Failed to start log stream: "+err.Error())
		return
	}

	emitEvent(appCtx, "logs:status", fmt.Sprintf("Streaming logs from %s", containerName))

	done := make(chan struct{}, 2)
	go scanPipe(appCtx, stdout, "logs:line", done)
	go scanPipe(appCtx, stderr, "logs:line", done)

	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
	case <-done:
	}
}

func scanPipe(appCtx context.Context, pipe interface{}, event string, done chan<- struct{}) {
	reader, ok := pipe.(interface {
		Read(p []byte) (n int, err error)
	})
	if !ok {
		emitEvent(appCtx, "logs:error", "Failed to read log stream")
		done <- struct{}{}
		return
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		emitEvent(appCtx, event, scanner.Text())
	}
	done <- struct{}{}
}
