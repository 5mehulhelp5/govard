package cmd

import (
	"fmt"
	"govard/internal/engine"
	"govard/internal/proxy"
	"govard/internal/updater"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [flags]",
	Short: "Start the development environment",
	Long: `Starts all Docker containers required for the current project.
It automatically handles framework detection, configuration validation,
Docker Compose blueprint rendering, host mapping, and proxy registration.

The startup process follows these stages:
1. Detect: Identifies framework context from project files.
2. Validate: Checks Docker status, port conflicts, and layered config.
3. Render: Generates the specialized Docker Compose file.
4. Start: Runs 'docker compose up' in detached mode.
5. Verify: Maps the .test domain to 127.0.0.1 and registers it with the Govard Proxy.

Case Studies:
- Standard Startup: Simply run 'govard env up' to get your full stack running.
- Low Resource Mode: Use --quickstart if you have limited RAM or only need PHP/Web server.
- Fresh Install Recovery: If containers are broken, 'govard env up' re-renders and restarts them.`,
	Example: `  # Start the environment normally
  govard env up

  # Fast startup: skip heavy services like Elasticsearch, Varnish, Redis
  govard env up --quickstart`,
	RunE: runUpCommand,
}

type upPipelineStage struct {
	Name         string
	OnFailureTip string
	Run          func() error
}

type upRuntimeContext struct {
	Cwd           string
	Config        engine.Config
	Compose       string
	Loaded        []string
	Metadata      engine.ProjectMetadata
	Profile       string
	Quickstart    bool
	Pull          bool
	FallbackLocal bool
	RemoveOrphans bool
	ForceRecreate bool
	UpdateLock    bool
	Out           io.Writer
	Err           io.Writer
}

func buildUpPipelineStages(cmd *cobra.Command, context *upRuntimeContext) []upPipelineStage {
	return []upPipelineStage{
		{
			Name:         "Detect",
			OnFailureTip: "govard init",
			Run: func() error {
				context.Metadata = engine.DetectFramework(context.Cwd)
				if strings.TrimSpace(context.Metadata.Version) == "" {
					pterm.Info.Printf("Detected framework: %s\n", context.Metadata.Framework)
				} else {
					pterm.Info.Printf("Detected framework: %s (%s)\n", context.Metadata.Framework, context.Metadata.Version)
				}
				return nil
			},
		},
		{
			Name:         "Validate",
			OnFailureTip: "govard doctor",
			Run: func() error {
				var err error
				context.Config, context.Loaded, err = engine.LoadConfigFromDirWithProfile(context.Cwd, true, context.Profile)
				if err != nil {
					return fmt.Errorf("load layered config: %w", err)
				}
				if warnings := CheckMagentoRuntimeSync(context.Config, context.Metadata); len(warnings) > 0 {
					for _, warning := range warnings {
						pterm.Warning.Println(warning)
					}
				}
				if context.Quickstart {
					ApplyQuickstartProfile(&context.Config)
					pterm.Info.Println("Quickstart profile enabled: optional services reduced for faster first run.")
				}
				context.Compose = engine.ComposeFilePathWithProfile(context.Cwd, context.Config.ProjectName, context.Profile)
				pterm.Info.Printf("Loaded config layers: %d\n", len(context.Loaded))
				lockWarnings, lockErr := evaluateUpLockPolicy(context.Cwd, context.Config, context.UpdateLock)
				for _, warning := range lockWarnings {
					pterm.Warning.Println(warning)
				}
				if len(lockWarnings) > 0 {
					pterm.Warning.Println("Run `govard lock check` for full lockfile compliance details.")
				}
				if lockErr != nil {
					return lockErr
				}

				if err := engine.CheckDockerStatus(cmd.Context()); err != nil {
					return fmt.Errorf("docker daemon is not ready: %w", err)
				}
				if err := engine.CheckDockerComposePlugin(cmd.Context()); err != nil {
					return err
				}

				if !engine.CheckPortForGovardProxy(cmd.Context(), "80") {
					pterm.Warning.Println("Port 80 is currently in use. Proxy routing may fail.")
				}
				if !engine.CheckPortForGovardProxy(cmd.Context(), "443") {
					pterm.Warning.Println("Port 443 is currently in use. HTTPS proxy routing may fail.")
				}
				if err := engine.CheckDiskScratchWrite(); err != nil {
					return fmt.Errorf("disk scratch check failed: %w", err)
				}
				if err := engine.CheckNetworkConnectivity(); err != nil {
					pterm.Warning.Printf("Network outbound probe failed: %v\n", err)
				}
				return nil
			},
		},
		{
			Name:         "Render",
			OnFailureTip: "govard doctor --fix",
			Run: func() error {
				if err := engine.RunHooks(context.Config, engine.HookPreUp, context.Out, context.Err); err != nil {
					return fmt.Errorf("pre-up hooks failed: %w", err)
				}

				if err := engine.EnsureGlobalProxy(); err != nil {
					pterm.Warning.Printf("Govard Proxy not ready: %v\n", err)
				}

				if err := engine.RenderBlueprintWithProfile(context.Cwd, context.Config, context.Profile); err != nil {
					return fmt.Errorf("render blueprint: %w", err)
				}
				pterm.Success.Printf("Rendered compose file: %s\n", context.Compose)
				return nil
			},
		},
		{
			Name:         "Pull",
			OnFailureTip: "govard doctor --fix",
			Run: func() error {
				if !context.Pull {
					return nil
				}
				err := engine.RunCompose(cmd.Context(), engine.ComposeOptions{
					ProjectDir:  context.Cwd,
					ProjectName: context.Config.ProjectName,
					ComposeFile: context.Compose,
					Args:        []string{"pull"},
					Stdout:      context.Out,
					Stderr:      context.Err,
				})
				if err != nil {
					if !context.FallbackLocal {
						return fmt.Errorf("docker compose pull failed: %w", err)
					}

					pterm.Warning.Printf("docker compose pull failed: %v\n", err)
					pterm.Info.Println("Attempting local Govard image build fallback...")

					built, fallbackErr := engine.FallbackBuildMissingGovardImagesFromCompose(context.Compose, context.Out, context.Err)
					if fallbackErr != nil {
						return fmt.Errorf("docker compose pull failed: %w (local fallback failed: %v)", err, fallbackErr)
					}
					if len(built) == 0 {
						pterm.Warning.Println("No missing Govard-managed images required local build. Continuing with current local cache.")
						return nil
					}

					pterm.Success.Printf("Local fallback built %d image(s): %s\n", len(built), strings.Join(built, ", "))
				}
				return nil
			},
		},
		{
			Name:         "Start",
			OnFailureTip: "govard doctor",
			Run: func() error {
				upArgs := []string{"up", "-d"}
				if context.RemoveOrphans {
					upArgs = append(upArgs, "--remove-orphans")
				}
				if context.ForceRecreate {
					upArgs = append(upArgs, "--force-recreate")
				}

				err := engine.RunCompose(cmd.Context(), engine.ComposeOptions{
					ProjectDir:  context.Cwd,
					ProjectName: context.Config.ProjectName,
					ComposeFile: context.Compose,
					Args:        upArgs,
					Stdout:      context.Out,
					Stderr:      context.Err,
				})
				if err != nil {
					if !context.FallbackLocal {
						return fmt.Errorf("docker compose up failed: %w", err)
					}

					pterm.Warning.Printf("docker compose up failed: %v\n", err)
					pterm.Info.Println("Attempting local Govard image build fallback...")

					built, fallbackErr := engine.FallbackBuildMissingGovardImagesFromCompose(context.Compose, context.Out, context.Err)
					if fallbackErr != nil {
						return fmt.Errorf("docker compose up failed: %w (local fallback failed: %v)", err, fallbackErr)
					}
					if len(built) == 0 {
						return fmt.Errorf("docker compose up failed: %w", err)
					}

					pterm.Success.Printf("Local fallback built %d image(s): %s\n", len(built), strings.Join(built, ", "))
					pterm.Info.Println("Retrying docker compose up after local fallback build...")

					retryErr := engine.RunCompose(cmd.Context(), engine.ComposeOptions{
						ProjectDir:  context.Cwd,
						ProjectName: context.Config.ProjectName,
						ComposeFile: context.Compose,
						Args:        upArgs,
						Stdout:      context.Out,
						Stderr:      context.Err,
					})
					if retryErr != nil {
						return fmt.Errorf("docker compose up failed after local fallback retry: %w", retryErr)
					}
				}
				return nil
			},
		},
		{
			Name:         "Verify",
			OnFailureTip: "govard doctor",
			Run: func() error {
				target := ResolveUpProxyTarget(context.Config)
				allDomains := context.Config.AllDomains()

				var proxyErr error
				if len(allDomains) > 0 {
					proxyErr = proxy.RegisterDomains(allDomains, target)
				}

				for _, domain := range allDomains {
					if engine.IsDomainResolvableLocally(domain) {
						pterm.Success.Printf("Domain %s already resolves locally\n", domain)
					} else if err := engine.AddHostsEntry(domain); err != nil {
						pterm.Warning.Printf("Could not update hosts file for %s: %v\n", domain, err)
						pterm.Info.Printf("Please manually add '127.0.0.1 %s' to your hosts file.\n", domain)
					} else {
						pterm.Success.Printf("Domain %s mapped to 127.0.0.1\n", domain)
					}

					if proxyErr != nil {
						pterm.Warning.Printf("Could not register domain %s with Govard Proxy: %v\n", domain, proxyErr)
					} else {
						pterm.Success.Printf("Domain %s registered with Govard Proxy -> %s\n", domain, target)
					}
				}

				if err := engine.RunHooks(context.Config, engine.HookPostUp, context.Out, context.Err); err != nil {
					return fmt.Errorf("post-up hooks failed: %w", err)
				}

				if err := engine.RefreshPMAActiveProjects(); err != nil {
					pterm.Warning.Printf("Could not refresh PMA active projects: %v\n", err)
				}
				return nil
			},
		},
	}
}

func runUpPipeline(stages []upPipelineStage) error {
	total := len(stages)
	for index, stage := range stages {
		step := fmt.Sprintf("[%d/%d] %s", index+1, total, stage.Name)
		pterm.Info.Printf("%s...\n", step)

		started := time.Now()
		if err := stage.Run(); err != nil {
			pterm.Error.Printf("%s failed (%s): %v\n", step, time.Since(started).Round(time.Millisecond), err)
			if stage.OnFailureTip != "" {
				pterm.Info.Printf("Suggested next command: %s\n", stage.OnFailureTip)
			}
			return err
		}
		pterm.Success.Printf("%s completed (%s)\n", step, time.Since(started).Round(time.Millisecond))
	}
	return nil
}

// ResolveUpProxyTarget resolves the upstream container for proxy registration.
func ResolveUpProxyTarget(config engine.Config) string {
	target := config.ProjectName + "-web-1"
	if config.Stack.Features.Varnish {
		target = config.ProjectName + "-varnish-1"
	}
	return target
}

// ApplyQuickstartProfile trims optional runtime services for faster first startup.
func ApplyQuickstartProfile(config *engine.Config) {
	if config == nil {
		return
	}

	config.Stack.Features.Xdebug = false
	config.Stack.Features.Varnish = false

	config.Stack.Services.Search = "none"
	config.Stack.SearchVersion = ""
	config.Stack.Features.Elasticsearch = false

	config.Stack.Services.Cache = "none"
	config.Stack.CacheVersion = ""
	config.Stack.Features.Redis = false

	config.Stack.Services.Queue = "none"
	config.Stack.QueueVersion = ""
}

// CheckMagentoRuntimeSync checks the detected Magento framework against configured
// runtime and returns warnings if there's a mismatch (without auto-tuning).
func CheckMagentoRuntimeSync(config engine.Config, metadata engine.ProjectMetadata) []string {
	if config.Framework != "magento2" {
		return nil
	}

	version := strings.TrimSpace(metadata.Version)
	if version == "" {
		version = strings.TrimSpace(config.FrameworkVersion)
	}

	profileResult, err := engine.ResolveRuntimeProfile("magento2", version)
	if err != nil {
		return nil // skip check if profile fails to resolve
	}

	var warnings []string
	p := profileResult.Profile

	if p.PHPVersion != "" && config.Stack.PHPVersion != p.PHPVersion {
		warnings = append(warnings, fmt.Sprintf("PHP %s (expected %s)", config.Stack.PHPVersion, p.PHPVersion))
	}
	if p.Search != "" && config.Stack.Services.Search != "none" && config.Stack.Services.Search != p.Search {
		warnings = append(warnings, fmt.Sprintf("Search %s (expected %s)", config.Stack.Services.Search, p.Search))
	}

	if len(warnings) > 0 {
		return []string{fmt.Sprintf(
			"Magento %s expects different services: %s. Run 'govard doctor --fix' to align.",
			version, strings.Join(warnings, ", "),
		)}
	}

	return nil
}

func shouldPreserveConfiguredDB(existingType, existingVersion, tunedType, tunedVersion string) bool {
	existingType = strings.ToLower(strings.TrimSpace(existingType))
	tunedType = strings.ToLower(strings.TrimSpace(tunedType))
	existingVersion = strings.TrimSpace(existingVersion)
	tunedVersion = strings.TrimSpace(tunedVersion)

	if existingType == "" || existingVersion == "" || tunedType == "" || tunedVersion == "" {
		return false
	}
	if existingType != tunedType {
		return false
	}

	comparison, comparable := compareNumericDotVersions(existingVersion, tunedVersion)
	return comparable && comparison > 0
}

func compareNumericDotVersions(left, right string) (int, bool) {
	return engine.CompareNumericDotVersions(left, right)
}

func runUpCommand(cmd *cobra.Command, args []string) (err error) {
	updater.CheckForUpdates(Version)
	startedAt := time.Now()

	quickstart, _ := cmd.Flags().GetBool("quickstart")
	profile, _ := cmd.Flags().GetString("profile")
	pull, _ := cmd.Flags().GetBool("pull")
	fallbackLocalBuild := boolFlagOrDefault(cmd, "fallback-local-build", true)
	removeOrphans, _ := cmd.Flags().GetBool("remove-orphans")
	forceRecreate, _ := cmd.Flags().GetBool("force-recreate")
	updateLock, _ := cmd.Flags().GetBool("update-lock")
	cwd, _ := os.Getwd()
	context := upRuntimeContext{
		Cwd:           cwd,
		Profile:       profile,
		Quickstart:    quickstart,
		Pull:          pull,
		FallbackLocal: fallbackLocalBuild,
		RemoveOrphans: removeOrphans,
		ForceRecreate: forceRecreate,
		UpdateLock:    updateLock,
		Out:           cmd.OutOrStdout(),
		Err:           cmd.ErrOrStderr(),
	}
	defer func() {
		status := engine.OperationStatusSuccess
		message := "up completed"
		if err != nil {
			status = engine.OperationStatusFailure
			message = err.Error()
		}
		writeOperationEventBestEffort(
			"up.run",
			status,
			context.Config,
			"",
			"",
			message,
			"",
			time.Since(startedAt),
		)
		if err == nil {
			trackProjectRegistryBestEffort(context.Config, cwd, "up")
		}
	}()
	stages := buildUpPipelineStages(cmd, &context)
	if err = runUpPipeline(stages); err != nil {
		return err
	}
	pterm.Success.Printf("Environment is up and running at https://%s\n", context.Config.Domain)
	return nil
}

func addUpFlags(command *cobra.Command) {
	command.Flags().Bool("quickstart", false, "Use a minimal runtime profile for faster first run")
	command.Flags().String("profile", "", "Environment scope (profile) to use")
	command.Flags().Bool("pull", false, "Pull latest images before starting")
	command.Flags().Bool("fallback-local-build", true, "When pull/start fails due missing Govard images, build missing Govard-managed images locally and retry")
	command.Flags().Bool("remove-orphans", false, "Remove containers for services not defined in the compose file")
	command.Flags().Bool("force-recreate", false, "Recreate containers even if their configuration and image haven't changed")
	command.Flags().Bool("update-lock", false, "Automatically update govard.lock if mismatches are found")
}

func init() {
	addUpFlags(upCmd)
}
