package tests

import (
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestResolveToolExecutionForPHPFramework(t *testing.T) {
	containerName, workdir, user := cmd.ResolveToolExecutionForTest(engine.Config{
		ProjectName: "sample-project",
		Framework:   "laravel",
	}, "composer")

	if containerName != "sample-project-php-1" {
		t.Fatalf("expected php container, got %s", containerName)
	}
	if workdir != "/var/www/html" {
		t.Fatalf("expected php workdir, got %s", workdir)
	}
	if user == "" {
		t.Fatal("expected php framework command to keep a concrete user")
	}
}

func TestResolveToolExecutionForEmdashUsesWebContainer(t *testing.T) {
	containerName, workdir, user := cmd.ResolveToolExecutionForTest(engine.Config{
		ProjectName: "sample-project",
		Framework:   "emdash",
	}, "pnpm")

	if containerName != "sample-project-web-1" {
		t.Fatalf("expected web container, got %s", containerName)
	}
	if workdir != "/app" {
		t.Fatalf("expected /app workdir, got %s", workdir)
	}
	if user != "" {
		t.Fatalf("expected empty user override for emdash, got %s", user)
	}
}
