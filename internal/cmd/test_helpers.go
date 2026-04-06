package cmd

import (
	"os/exec"
	"time"

	"govard/internal/engine"
)

type UpReadinessCheckForTest struct {
	Service       string
	ContainerName string
}

// FindWailsCLIForTest exposes Wails binary discovery for external tests.
func FindWailsCLIForTest() (string, error) {
	return findWailsCLI()
}

// DesktopBinaryArgsForTest exposes govard-desktop argument construction for tests.
func DesktopBinaryArgsForTest(background bool) []string {
	return buildDesktopBinaryArgs(background)
}

// FindDesktopBinaryForTest exposes desktop binary discovery for external tests.
func FindDesktopBinaryForTest() (string, error) {
	return findDesktopBinary()
}

// SetDesktopExecutablePathForTest overrides os.Executable usage for tests.
func SetDesktopExecutablePathForTest(fn func() (string, error)) func() {
	previous := desktopExecutablePath
	desktopExecutablePath = fn
	return func() {
		desktopExecutablePath = previous
	}
}

// SetDesktopLookPathForTest overrides exec.LookPath usage for tests.
func SetDesktopLookPathForTest(fn func(file string) (string, error)) func() {
	previous := desktopBinaryLookPath
	desktopBinaryLookPath = fn
	return func() {
		desktopBinaryLookPath = previous
	}
}

// ExecLookPathForTest delegates to exec.LookPath for test stubs.
func ExecLookPathForTest(file string) (string, error) {
	return exec.LookPath(file)
}

// DesktopProductionBuildTagsForTest exposes desktop production build tags.
func DesktopProductionBuildTagsForTest() string {
	return desktopBuildTags(true)
}

// BuildUpReadinessChecksForTest exposes startup readiness planning for tests.
func BuildUpReadinessChecksForTest(config engine.Config) []UpReadinessCheckForTest {
	checks := buildUpReadinessChecks(config)
	result := make([]UpReadinessCheckForTest, 0, len(checks))
	for _, check := range checks {
		result = append(result, UpReadinessCheckForTest{
			Service:       check.Service,
			ContainerName: check.ContainerName,
		})
	}
	return result
}

// WaitForUpRuntimeReadinessForTest exposes readiness waiting for tests.
func WaitForUpRuntimeReadinessForTest(config engine.Config, timeout time.Duration) error {
	return waitForUpRuntimeReadiness(config, timeout)
}

// SetUpReadinessProbeRunnerForTest overrides the probe runner used by readiness checks.
func SetUpReadinessProbeRunnerForTest(fn func(containerName string, probeArgs []string) error) func() {
	previous := upReadinessProbeRunner
	if fn == nil {
		upReadinessProbeRunner = previous
		return func() {
			upReadinessProbeRunner = previous
		}
	}
	upReadinessProbeRunner = fn
	return func() {
		upReadinessProbeRunner = previous
	}
}

// SetUpReadinessProbeIntervalForTest overrides readiness retry intervals.
func SetUpReadinessProbeIntervalForTest(interval time.Duration) func() {
	previous := upReadinessProbeInterval
	upReadinessProbeInterval = interval
	return func() {
		upReadinessProbeInterval = previous
	}
}

// SetUpReadinessSleepForTest overrides readiness sleeping behavior.
func SetUpReadinessSleepForTest(fn func(time.Duration)) func() {
	previous := upReadinessSleep
	if fn == nil {
		upReadinessSleep = time.Sleep
		return func() {
			upReadinessSleep = previous
		}
	}
	upReadinessSleep = fn
	return func() {
		upReadinessSleep = previous
	}
}
