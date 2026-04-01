package cmd

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"govard/internal/engine"
	"govard/internal/engine/remote"

	"github.com/docker/go-units"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

var dbCmd = &cobra.Command{
	Use:   "db [connect|import|dump|query|info|top]",
	Short: "Interact with the database container",
	Long: `Manage your project's database. Supports connecting to the container shell,
importing SQL dumps, and creating backups. Works for both local and remote environments.

Storage Behavior for Dumps:
- Local Environment: Dumps are saved to the project's local 'var/' directory.
- Remote Environment (default): Dumps are stored on the remote server (usually ~/backup/).
- Remote Environment (+ --local): Dumps are streamed directly to the project's local 'var/' directory.

Dumps are comprehensive (including routines and triggers) by default to ensure full portability.`,
	Example: `  # Open an interactive MySQL shell locally
  govard db connect

  # Connect to the staging database via SSH tunnel
  govard db connect --environment staging

  # Import a local SQL file with clean reset (drop/recreate DB)
  govard db import --file backup.sql --drop

  # Stream a dump from production into your local database
  govard db import --stream-db --environment prod --drop

  # Create a database dump on the remote server (saved to ~/backup/ on remote)
  govard db dump --environment staging

  # Create a database dump from remote and save it to local 'var/' directory
  govard db dump --environment staging --local

  # Create a local dump excluding noise tables (cron, cache, logs...)
  govard db dump --no-noise

  # Create a local dump excluding noise + PII tables (customers, orders...)
  govard db dump --no-noise --no-pii

  # Execute a SQL query
  govard db query "SELECT * FROM core_config_data LIMIT 5"

  govard db info`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one argument (subcommand)")
		}
		subcommand := strings.ToLower(strings.TrimSpace(args[0]))
		if subcommand == "query" && len(args) < 2 {
			return errors.New("query subcommand requires a SQL query argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		subcommand := strings.ToLower(strings.TrimSpace(args[0]))
		if err := runDBSubcommand(cmd, subcommand, args[1:]); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	dbCmd.Flags().StringP("environment", "e", "local", "Target environment (local, staging, prod, etc.)")
	dbCmd.Flags().StringP("file", "f", "", "Database dump file (import or dump output)")
	dbCmd.Flags().String("profile", "", "Environment scope (profile) to use")
	dbCmd.Flags().Bool("stream-db", false, "For import: stream dump from remote environment into local database")
	dbCmd.Flags().BoolP("no-noise", "N", false, "For dump: exclude ephemeral tables (cron, cache, session, logs...)")
	dbCmd.Flags().BoolP("no-pii", "P", false, "For dump: exclude PII/sensitive tables (customers, orders...)")
	dbCmd.Flags().BoolP("sanitize", "S", false, "Alias for --no-pii")
	dbCmd.Flags().Bool("drop", false, "For import: drop and recreate the database before importing")
	dbCmd.Flags().Bool("local", false, "For dump/import: force local file operations for remote environments")
	dbCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
}

type DBCommandOptions struct {
	Environment string
	File        string
	Profile     string
	StreamDB    bool
	NoNoise     bool
	NoPII       bool
	Drop        bool
	Local       bool
	AssumeYes   bool
}

type dbCommandOptions = DBCommandOptions

var mysqlDatabaseNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var stdinIsTerminalFn = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// ValidateDBCommandOptions validates DB command option combinations.
func ValidateDBCommandOptions(subcommand string, options DBCommandOptions) error {
	return validateDBCommandOptions(subcommand, options)
}

func runDBSubcommand(cmd *cobra.Command, subcommand string, extraArgs []string) (err error) {
	startedAt := time.Now()
	operationStatus := engine.OperationStatusFailure
	operationCategory := ""
	operationMessage := ""
	operationConfig := engine.Config{}
	operationSource := ""
	operationDestination := ""
	defer func() {
		if err != nil && operationMessage == "" {
			operationMessage = err.Error()
		}
		if err == nil && operationStatus == engine.OperationStatusFailure {
			operationStatus = engine.OperationStatusSuccess
		}
		if err != nil && operationCategory == "" {
			operationCategory = classifyCommandError(err)
		}
		writeOperationEventBestEffort(
			"db."+subcommand,
			operationStatus,
			operationConfig,
			operationSource,
			operationDestination,
			operationMessage,
			operationCategory,
			time.Since(startedAt),
		)
		if err == nil {
			cwd, _ := os.Getwd()
			trackProjectRegistryBestEffort(operationConfig, cwd, "db-"+subcommand)
		}
	}()

	auditStatus := remote.RemoteAuditStatusFailure
	auditCategory := ""
	auditMessage := ""
	auditRemote := ""
	auditSource := ""
	auditDestination := ""

	options, err := readDBCommandOptions(cmd)
	if err != nil {
		writeRemoteAuditEvent(remote.AuditEvent{
			Operation:  "db." + subcommand,
			Status:     remote.RemoteAuditStatusFailure,
			Category:   "validation",
			DurationMS: time.Since(startedAt).Milliseconds(),
			Message:    err.Error(),
		})
		return err
	}
	if err := validateDBCommandOptions(subcommand, options); err != nil {
		writeRemoteAuditEvent(remote.AuditEvent{
			Operation:  "db." + subcommand,
			Status:     remote.RemoteAuditStatusFailure,
			Category:   "validation",
			Remote:     options.Environment,
			DurationMS: time.Since(startedAt).Milliseconds(),
			Message:    err.Error(),
		})
		return err
	}
	shouldAudit := options.Environment != "local" || options.StreamDB
	if shouldAudit {
		if options.StreamDB {
			auditSource = options.Environment
			auditDestination = "local"
			operationSource = options.Environment
			operationDestination = "local"
		} else {
			auditRemote = options.Environment
			operationDestination = options.Environment
		}
		defer func() {
			if err != nil && auditMessage == "" {
				auditMessage = err.Error()
			}
			if err == nil && auditStatus == remote.RemoteAuditStatusFailure {
				auditStatus = remote.RemoteAuditStatusSuccess
			}
			if err != nil && auditCategory == "" {
				auditCategory = classifyCommandError(err)
			}
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:   "db." + subcommand,
				Status:      auditStatus,
				Category:    auditCategory,
				Remote:      auditRemote,
				Source:      auditSource,
				Destination: auditDestination,
				DurationMS:  time.Since(startedAt).Milliseconds(),
				Message:     auditMessage,
			})
		}()
	}

	config, err := loadFullConfigWithProfile(options.Profile)
	if err != nil {

		return err
	}
	operationConfig = config
	if options.Environment != "local" {
		if remoteName, ok := findRemoteByNameOrEnvironment(config, options.Environment); ok {
			options.Environment = remoteName
		}
	}
	operationSource = options.Environment
	switch subcommand {
	case "connect":
		err = runDBConnect(cmd, config, options)
		if err == nil {
			auditStatus = remote.RemoteAuditStatusSuccess
			auditMessage = "db connect completed"
			operationStatus = engine.OperationStatusSuccess
			operationMessage = "db connect completed"
		}
		return err
	case "dump":
		err = runDBDump(cmd, config, options)
		if err == nil {
			auditStatus = remote.RemoteAuditStatusSuccess
			auditMessage = "db dump completed"
			operationStatus = engine.OperationStatusSuccess
			operationMessage = "db dump completed"
		}
		return err
	case "import":
		err = runDBImport(cmd, config, options)
		if err == nil {
			auditStatus = remote.RemoteAuditStatusSuccess
			if options.StreamDB {
				auditMessage = "db stream import completed"
				operationMessage = "db stream import completed"
			} else {
				auditMessage = "db import completed"
				operationMessage = "db import completed"
			}
			operationStatus = engine.OperationStatusSuccess
		}
		return err
	case "top":
		err = runDBTop(cmd, config, options)
		if err == nil {
			auditStatus = remote.RemoteAuditStatusSuccess
			auditMessage = "db top completed"
			operationStatus = engine.OperationStatusSuccess
			operationMessage = "db top completed"
		}
		return err
	case "query":
		err = runDBQuery(cmd, config, options, extraArgs)
		if err == nil {
			auditStatus = remote.RemoteAuditStatusSuccess
			auditMessage = "db query completed"
			operationStatus = engine.OperationStatusSuccess
			operationMessage = "db query completed"
		}
		return err
	case "info":
		err = runDBInfo(cmd, config, options)
		if err == nil {
			auditStatus = remote.RemoteAuditStatusSuccess
			auditMessage = "db info completed"
			operationStatus = engine.OperationStatusSuccess
			operationMessage = "db info completed"
		}
		return err
	default:
		return fmt.Errorf("unknown db subcommand: %s", subcommand)
	}
}

func readDBCommandOptions(cmd *cobra.Command) (dbCommandOptions, error) {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return dbCommandOptions{}, err
	}
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return dbCommandOptions{}, err
	}
	streamDB, err := cmd.Flags().GetBool("stream-db")
	if err != nil {
		return dbCommandOptions{}, err
	}
	drop, _ := cmd.Flags().GetBool("drop")
	local, _ := cmd.Flags().GetBool("local")
	noNoise, err := cmd.Flags().GetBool("no-noise")
	if err != nil {
		return dbCommandOptions{}, err
	}
	noPII, err := cmd.Flags().GetBool("no-pii")
	if err != nil {
		return dbCommandOptions{}, err
	}
	sanitizePII, err := cmd.Flags().GetBool("sanitize")
	if err != nil {
		return dbCommandOptions{}, err
	}
	profile, _ := cmd.Flags().GetString("profile")
	assumeYes, _ := cmd.Flags().GetBool("yes")
	return dbCommandOptions{
		Environment: strings.ToLower(strings.TrimSpace(environment)),
		File:        strings.TrimSpace(file),
		Profile:     profile,
		StreamDB:    streamDB,
		NoNoise:     noNoise,
		NoPII:       noPII || sanitizePII,
		Drop:        drop,
		Local:       local,
		AssumeYes:   assumeYes,
	}, nil
}

func validateDBCommandOptions(subcommand string, options dbCommandOptions) error {
	if options.Environment == "" {
		return errors.New("environment cannot be empty")
	}

	switch subcommand {
	case "connect":
		if options.File != "" || options.StreamDB || options.NoNoise || options.NoPII || options.Drop || options.Local {
			return errors.New("connect does not support --file, --stream-db, --no-noise, --no-pii, --drop, or --local")
		}
	case "dump":
		if options.StreamDB {
			return errors.New("--stream-db is only supported by db import")
		}
	case "import":
		if (options.NoNoise || options.NoPII) && !options.StreamDB {
			return errors.New("--no-noise and --no-pii are only supported by db dump or stream-db import")
		}
		if options.StreamDB && options.Environment == "local" {
			return errors.New("--stream-db requires a remote --environment source")
		}
	case "query", "info":
		if options.File != "" || options.StreamDB || options.NoNoise || options.NoPII || options.Drop || options.Local {
			return errors.New("query and info do not support --file, --stream-db, --no-noise, --no-pii, --drop, or --local")
		}
	default:
		return fmt.Errorf("unknown db subcommand: %s", subcommand)
	}
	return nil
}

// ResetDBFlagsForTest resets db command flags to defaults for tests.
func ResetDBFlagsForTest() {
	dbCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
}

func runDBHooks(config engine.Config, pre string, post string, cmd *cobra.Command, action func() error) error {
	if err := engine.RunHooks(config, pre, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("%s hooks failed: %w", pre, err)
	}
	if err := action(); err != nil {
		return err
	}
	if err := engine.RunHooks(config, post, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("%s hooks failed: %w", post, err)
	}
	return nil
}

func resolveDBImportReader(options dbCommandOptions) (io.Reader, io.Closer, int64, error) {
	if options.File != "" {
		path := filepath.Clean(options.File)
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("open import file: %w", err)
		}
		info, _ := file.Stat()
		return file, file, info.Size(), nil
	}

	if stdinIsTerminal() {
		return nil, nil, 0, errors.New("no import input provided; use --file or pipe SQL via stdin")
	}

	// Try to get size if stdin is redirected from a file
	if info, err := os.Stdin.Stat(); err == nil {
		// If it's not a char device (terminal) and not a pipe (S_IFIFO), it might be a regular file.
		// On Linux, a redirected file has Mode().IsRegular() or just a non-zero size.
		if (info.Mode() & os.ModeCharDevice) == 0 {
			return os.Stdin, nil, info.Size(), nil
		}
	}

	return os.Stdin, nil, 0, nil
}

func formatBytesCount(count, total int) string {
	return fmt.Sprintf("%s/%s", units.HumanSize(float64(count)), units.HumanSize(float64(total)))
}

type dbProgressReader struct {
	reader io.Reader
	bar    *pterm.ProgressbarPrinter
	total  int64
	label  string
}

func (r *dbProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 && r.bar != nil {
		r.bar.Add(n)
		current := float64(r.bar.Current)
		total := float64(r.total)
		r.bar.UpdateTitle(fmt.Sprintf("%s [%s/%s]", r.label, units.HumanSize(current), units.HumanSize(total)))
	}
	return n, err
}

func stdinIsTerminal() bool {
	return stdinIsTerminalFn()
}

func buildDBDumpCommand(config engine.Config, options dbCommandOptions) (*exec.Cmd, string, error) {
	suffix := "sql.gz"
	timestamp := time.Now().Format("20060102T150405")
	defaultFilename := fmt.Sprintf("%s_%s-%s.%s", config.ProjectName, options.Environment, timestamp, suffix)

	if options.Environment == "local" {
		containerName := dbContainerName(config)
		if err := ensureLocalDBRunning(containerName); err != nil {
			return nil, "", err
		}
		credentials := resolveLocalDBCredentials(config, containerName)
		return buildLocalDBDumpCommand(containerName, credentials, options.NoNoise, options.NoPII, config.Framework), "", nil
	}

	remoteCfg, err := resolveDBRemote(config, options.Environment, false)
	if err != nil {
		return nil, "", err
	}
	credentials, probeErr := resolveRemoteDBCredentials(config, options.Environment, remoteCfg)
	if probeErr != nil {
		pterm.Warning.Println(formatRemoteDBProbeWarning(options.Environment, probeErr))
	}

	dumpStr := buildRemoteMySQLDumpCommandString(credentials, options.NoNoise, options.NoPII, config.Framework, true)

	// If not local, we dump to a file on the remote server
	if !options.Local {
		remoteFile := options.File
		if remoteFile == "" {
			remoteFile = filepath.Join("~/backup", defaultFilename)
		}
		// We need to wrap the command to create the directory and redirect output
		// Note: we use base64 or complex quoting if needed, but here simple redirection should work if we quote the filename
		// Using sh -c to allow redirects and mkdir -p on the remote
		quotedFile := quoteRemotePath(remoteFile)
		remoteCmd := fmt.Sprintf("mkdir -p $(dirname %s) && { %s; } | gzip > %s", quotedFile, dumpStr, quotedFile)
		return remote.BuildSSHExecCommand(options.Environment, remoteCfg, true, remoteCmd), remoteFile, nil
	}

	return remote.BuildSSHExecCommand(options.Environment, remoteCfg, true, dumpStr), "", nil
}

func quoteRemotePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return "$HOME/" + engine.ShellQuote(path[2:])
	}
	return engine.ShellQuote(path)
}

func buildDBImportCommand(config engine.Config, options dbCommandOptions) (*exec.Cmd, error) {
	if options.Environment == "local" {
		containerName := dbContainerName(config)
		if err := ensureLocalDBRunning(containerName); err != nil {
			return nil, err
		}
		return buildLocalDBImportCommand(containerName, resolveLocalDBCredentials(config, containerName)), nil
	}

	remoteCfg, err := resolveDBRemote(config, options.Environment, true)
	if err != nil {
		return nil, err
	}
	credentials, probeErr := resolveRemoteDBCredentials(config, options.Environment, remoteCfg)
	if probeErr != nil {
		pterm.Warning.Println(formatRemoteDBProbeWarning(options.Environment, probeErr))
	}
	return remote.BuildSSHExecCommand(options.Environment, remoteCfg, true, buildRemoteMySQLImportCommandString(credentials)), nil
}

func dbContainerName(config engine.Config) string {
	return fmt.Sprintf("%s-db-1", config.ProjectName)
}

func ensureLocalDBRunning(containerName string) error {
	check := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	output, err := check.Output()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return fmt.Errorf("database container %s is not running", containerName)
	}
	return nil
}

func resolveDBRemote(config engine.Config, name string, forWrite bool) (engine.RemoteConfig, error) {
	_, remoteCfg, err := ensureRemoteKnown(config, name)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	if !engine.RemoteCapabilityEnabled(remoteCfg, engine.RemoteCapabilityDB) {
		return engine.RemoteConfig{}, fmt.Errorf(
			"remote '%s' does not allow db operations (capabilities: %s)",
			name,
			strings.Join(engine.RemoteCapabilityList(remoteCfg), ","),
		)
	}
	if !forWrite {
		if blocked, reason := engine.RemoteWriteBlocked(name, remoteCfg); blocked {
			return engine.RemoteConfig{}, fmt.Errorf("remote environment '%s' is write-protected: %s", name, reason)
		}
	}
	return remoteCfg, nil
}

func runDumpToWriter(dumpCmd *exec.Cmd, writer io.Writer, sanitize bool, stderr io.Writer) error {
	stdout, err := dumpCmd.StdoutPipe()
	if err != nil {
		return err
	}
	dumpCmd.Stderr = stderr
	if err := dumpCmd.Start(); err != nil {
		return err
	}

	// Automatic gzip detection to handle remote compressed streams
	var finalReader io.Reader
	if r, isGzipped, err := detectGzipReader(stdout); err == nil && isGzipped {
		gzReader, err := gzip.NewReader(r)
		if err != nil {
			return fmt.Errorf("create gzip reader for dump stream: %w", err)
		}
		defer func() { _ = gzReader.Close() }()
		finalReader = gzReader
	} else {
		finalReader = r
	}

	var copyErr error
	if sanitize {
		copyErr = engine.SanitizeSQLDump(finalReader, writer)
	} else {
		_, copyErr = io.Copy(writer, finalReader)
	}

	waitErr := dumpCmd.Wait()
	if copyErr != nil {
		return copyErr
	}
	return waitErr
}

func RunImportFromReader(importCmd *exec.Cmd, reader io.Reader, sanitize bool, stdout io.Writer, stderr io.Writer) error {
	return RunImportFromReaderWithProgress(importCmd, reader, 0, sanitize, stdout, stderr)
}

func RunImportFromReaderWithProgress(importCmd *exec.Cmd, reader io.Reader, totalSize int64, sanitize bool, stdout io.Writer, stderr io.Writer) error {
	stdin, err := importCmd.StdinPipe()
	if err != nil {
		return err
	}
	importCmd.Stdout = stdout
	importCmd.Stderr = stderr

	// Check for gzip via magic number or file extension
	isGzip := false
	var readerWithPeek = reader
	if r, ok := reader.(*os.File); ok && strings.HasSuffix(strings.ToLower(r.Name()), ".gz") {
		isGzip = true
	} else if r, g, err := detectGzipReader(reader); err == nil {
		readerWithPeek = r
		isGzip = g
	}

	// Progress tracking and Gzip detection logic
	var bar *pterm.ProgressbarPrinter
	var finalReader io.Reader

	// Heuristic: If we are reading a compressed source and totalSize equals the source size,
	// we track progress on the COMPRESSED reader (common for local .sql.gz files).
	// Otherwise, we track on the UNCOMPRESSED stream (common for remote stream syncs).
	trackCompressed := false
	if totalSize > 0 && isGzip {
		if f, ok := reader.(*os.File); ok {
			if stat, err := f.Stat(); err == nil && stat.Size() == totalSize {
				trackCompressed = true
			}
		}
	}

	if totalSize > 0 {
		bar, _ = pterm.DefaultProgressbar.WithTotal(int(totalSize)).
			WithTitle(fmt.Sprintf("Importing DB [0/%s]", units.HumanSize(float64(totalSize)))).
			WithShowCount(false).
			Start()

		if trackCompressed {
			// track progress on the COMPRESSED source (local .sql.gz file)
			readerWithPeek = &dbProgressReader{reader: readerWithPeek, bar: bar, total: totalSize, label: "Importing DB"}
		}
	}

	if isGzip {
		gzReader, err := gzip.NewReader(readerWithPeek)
		if err != nil {
			return fmt.Errorf("create gzip reader: %w", err)
		}
		defer func() { _ = gzReader.Close() }()
		finalReader = gzReader
	} else {
		finalReader = readerWithPeek
	}

	if totalSize > 0 && !trackCompressed {
		// Track against uncompressed stream (remote sync case)
		finalReader = &dbProgressReader{reader: finalReader, bar: bar, total: totalSize, label: "Importing DB"}
	}

	if err := importCmd.Start(); err != nil {
		return err
	}

	importPrefix := "SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET AUTOCOMMIT=0; SET SQL_MODE='NO_AUTO_VALUE_ON_ZERO';\n"
	importSuffix := "\nCOMMIT; SET FOREIGN_KEY_CHECKS=1; SET UNIQUE_CHECKS=1; SET AUTOCOMMIT=1;\n"

	// Wrap the reader with performance-optimized session variables
	finalReaderWithWrappers := io.MultiReader(
		strings.NewReader(importPrefix),
		finalReader,
		strings.NewReader(importSuffix),
	)

	var copyErr error
	if sanitize {
		copyErr = engine.SanitizeSQLDump(finalReaderWithWrappers, stdin)
	} else {
		_, copyErr = io.Copy(stdin, finalReaderWithWrappers)
	}

	if bar != nil {
		bar.Add(int(totalSize) - bar.Current)
		_, _ = bar.Stop()
	}

	closeErr := stdin.Close()
	waitErr := importCmd.Wait()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return waitErr
}

func RunDumpToImport(dumpCmd *exec.Cmd, importCmd *exec.Cmd, sanitize bool, stdout io.Writer, stderr io.Writer) error {
	return RunDumpToImportWithProgress(dumpCmd, importCmd, 0, sanitize, stdout, stderr)
}

func RunDumpToImportWithProgress(dumpCmd *exec.Cmd, importCmd *exec.Cmd, totalSize int64, sanitize bool, stdout io.Writer, stderr io.Writer) error {
	dumpStdout, err := dumpCmd.StdoutPipe()
	if err != nil {
		return err
	}
	importStdin, err := importCmd.StdinPipe()
	if err != nil {
		return err
	}

	// Capture stderr to prevent it from breaking the progress bar UI
	dumpStderr := &bytes.Buffer{}
	importStderr := &bytes.Buffer{}
	dumpCmd.Stderr = dumpStderr
	importCmd.Stdout = stdout
	importCmd.Stderr = importStderr

	if err := dumpCmd.Start(); err != nil {
		return err
	}
	if err := importCmd.Start(); err != nil {
		if dumpCmd.Process != nil {
			_ = dumpCmd.Process.Kill()
		}
		_ = dumpCmd.Wait()
		return err
	}

	importPrefix := "SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET AUTOCOMMIT=0; SET SQL_MODE='NO_AUTO_VALUE_ON_ZERO';\n"
	importSuffix := "\nCOMMIT; SET FOREIGN_KEY_CHECKS=1; SET UNIQUE_CHECKS=1; SET AUTOCOMMIT=1;\n"

	var bar *pterm.ProgressbarPrinter
	var reader io.Reader = dumpStdout
	isGzip := false

	// Automatic gzip detection for the stream
	if r, g, err := detectGzipReader(dumpStdout); err == nil {
		reader = r
		isGzip = g
	}

	if isGzip {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("create gzip reader for stream: %w", err)
		}
		defer func() { _ = gzReader.Close() }()
		reader = gzReader
	}

	// Progress tracking logic
	if totalSize > 0 {
		bar, _ = pterm.DefaultProgressbar.WithTotal(int(totalSize)).
			WithTitle(fmt.Sprintf("Syncing DB [0/%s]", units.HumanSize(float64(totalSize)))).
			WithShowCount(false).
			Start()

		// In RunDumpToImport, the source is always a pipe from the remote dump command.
		// If totalSize is provided here, it's always the logical/uncompressed size
		// from GetDatabaseSize(), so we ALWAYS track uncompressed progress.
		reader = &dbProgressReader{reader: reader, bar: bar, total: totalSize, label: "Syncing DB"}
	}

	// Wrap the reader with performance-optimized session variables
	finalReader := io.MultiReader(
		strings.NewReader(importPrefix),
		reader,
		strings.NewReader(importSuffix),
	)

	var copyErr error
	if sanitize {
		copyErr = engine.SanitizeSQLDump(finalReader, importStdin)
	} else {
		_, copyErr = io.Copy(importStdin, finalReader)
	}

	if bar != nil {
		bar.Add(int(totalSize) - bar.Current)
		_, _ = bar.Stop()
	}

	closeErr := importStdin.Close()
	dumpErr := dumpCmd.Wait()
	importErr := importCmd.Wait()

	if copyErr != nil {
		// If copy failed, check if it was due to a process termination
		if dumpErr != nil {
			return fmt.Errorf("database dump failed: %w\nOutput: %s", dumpErr, dumpStderr.String())
		}
		if importErr != nil {
			return fmt.Errorf("database import failed: %w\nOutput: %s", importErr, importStderr.String())
		}
		return copyErr
	}

	if closeErr != nil {
		return closeErr
	}
	if dumpErr != nil {
		return fmt.Errorf("database dump failed: %w\nOutput: %s", dumpErr, dumpStderr.String())
	}
	if importErr != nil {
		return fmt.Errorf("database import failed: %w\nOutput: %s", importErr, importStderr.String())
	}
	return nil
}

func detectGzipReader(r io.Reader) (io.Reader, bool, error) {
	br := bufio.NewReader(r)
	peek, err := br.Peek(2)
	if err != nil && err != io.EOF {
		return br, false, err
	}
	if len(peek) == 2 && peek[0] == 0x1f && peek[1] == 0x8b {
		return br, true, nil
	}
	return br, false, nil
}

// SetStdinIsTerminalForTest overrides terminal detection for tests.
func SetStdinIsTerminalForTest(fn func() bool) func() {
	previous := stdinIsTerminalFn
	if fn != nil {
		stdinIsTerminalFn = fn
	}
	return func() {
		stdinIsTerminalFn = previous
	}
}

// ResolveDBImportReaderForTest exposes resolveDBImportReader for tests in /tests.
func ResolveDBImportReaderForTest(options DBCommandOptions) (io.Reader, io.Closer, int64, error) {
	return resolveDBImportReader(options)
}

// BuildDBDumpCommandForTest exposes buildDBDumpCommand args for tests in /tests.
func BuildDBDumpCommandForTest(config engine.Config, options DBCommandOptions) ([]string, error) {
	command, _, err := buildDBDumpCommand(config, options)
	if err != nil {
		return nil, err
	}
	return command.Args, nil
}

// BuildDBImportCommandForTest exposes buildDBImportCommand args for tests in /tests.
func BuildDBImportCommandForTest(config engine.Config, options DBCommandOptions) ([]string, error) {
	command, err := buildDBImportCommand(config, options)
	if err != nil {
		return nil, err
	}
	return command.Args, nil
}

// ResolveDBRemoteForTest exposes resolveDBRemote for tests in /tests.
func ResolveDBRemoteForTest(config engine.Config, name string, forWrite bool) (engine.RemoteConfig, error) {
	return resolveDBRemote(config, name, forWrite)
}

func classifyCommandError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "unknown remote"),
		strings.Contains(message, "requires"),
		strings.Contains(message, "does not support"),
		strings.Contains(message, "does not allow"),
		strings.Contains(message, "blocks db write operations"),
		strings.Contains(message, "environment cannot be empty"),
		strings.Contains(message, "database container"):
		return "validation"
	default:
		return remote.ClassifyFailure(err, message).Category
	}
}

func runDBTop(cmd *cobra.Command, config engine.Config, options dbCommandOptions) error {
	var remoteCfg engine.RemoteConfig
	var credentials dbCredentials
	var err error

	if options.Environment == "local" {
		containerName := dbContainerName(config)
		if err := ensureLocalDBRunning(containerName); err != nil {
			return err
		}
		credentials = resolveLocalDBCredentials(config, containerName)
	} else {
		remoteCfg, err = resolveDBRemote(config, options.Environment, false)
		if err != nil {
			return err
		}
		credentials, err = resolveRemoteDBCredentials(config, options.Environment, remoteCfg)
		if err != nil {
			return err
		}
	}

	pterm.Info.Println("Starting db top. Press Ctrl+C to exit.")
	area, _ := pterm.DefaultArea.Start()
	defer func() { _ = area.Stop() }()

	// Handle graceful exit on Ctrl+C
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			query := "SHOW FULL PROCESSLIST"
			var out []byte
			var cmdErr error

			// Build command string
			var cmdStr string
			if credentials.Password != "" {
				cmdStr = fmt.Sprintf("mysql -u%s -p%s -BN -e %s", engine.ShellQuote(credentials.Username), engine.ShellQuote(credentials.Password), engine.ShellQuote(query))
			} else {
				cmdStr = fmt.Sprintf("mysql -u%s -BN -e %s", engine.ShellQuote(credentials.Username), engine.ShellQuote(query))
			}

			if options.Environment == "local" {
				containerName := dbContainerName(config)
				out, cmdErr = exec.Command("docker", "exec", containerName, "sh", "-c", cmdStr).CombinedOutput()
			} else {
				sshCmd := remote.BuildSSHExecCommand(options.Environment, remoteCfg, true, cmdStr)
				out, cmdErr = sshCmd.CombinedOutput()
			}

			if cmdErr != nil {
				area.Update(pterm.Red(fmt.Sprintf("Error fetching processlist: %v\nOutput: %s", cmdErr, string(out))))
			} else {
				tableStr, err := formatProcessListTable(string(out))
				if err != nil {
					area.Update(pterm.Red(fmt.Sprintf("Error formatting table: %v", err)))
				} else {
					area.Update(tableStr)
				}
			}
		}
	}
}

func formatProcessListTable(raw string) (string, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 {
		return "No processes found.", nil
	}

	data := [][]string{}
	// Headers for SHOW FULL PROCESSLIST are: Id, User, Host, db, Command, Time, State, Info
	data = append(data, []string{"ID", "User", "Host", "DB", "Command", "Time", "State", "Info"})

	for _, line := range lines {
		// mysql -BN output is tab-separated
		parts := strings.Split(line, "\t")
		data = append(data, parts)
	}

	table, err := pterm.DefaultTable.WithHasHeader().WithData(data).Srender()
	if err != nil {
		return "", err
	}
	return table, nil
}
