package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type BootstrapRuntimeOptions struct {
	Source          string
	Clone           bool
	CodeOnly        bool
	Fresh           bool
	IncludeSample   bool
	DBImport        bool
	MediaSync       bool
	ComposerInstall bool
	AdminCreate     bool
	StreamDB        bool
	SkipUp          bool
	MetaPackage     string
	MetaVersion     string
	DBDump          string
	FixDeps         bool
	HyvaInstall     bool
	HyvaToken       string
	MageUsername    string
	MagePassword    string
	AssumeYes       bool
	IncludeProduct  bool
	Plan            bool
	NoNoise         bool
	NoPII           bool
	DeleteSync      bool
	NoCompress      bool
	ExcludePatterns []string
}

func resolveBootstrapOptions(cmd *cobra.Command) (BootstrapRuntimeOptions, error) {
	opts := BootstrapRuntimeOptions{
		Source:          normalizeBootstrapSource(bootstrapEnv),
		Clone:           bootstrapClone,
		CodeOnly:        bootstrapCodeOnly,
		Fresh:           bootstrapFresh,
		IncludeSample:   bootstrapIncludeSample,
		DBImport:        !bootstrapSkipDB,
		MediaSync:       !bootstrapSkipMedia,
		ComposerInstall: !bootstrapSkipComposer,
		AdminCreate:     !bootstrapSkipAdmin,
		StreamDB:        !bootstrapNoStreamDB,
		SkipUp:          bootstrapSkipUp,
		MetaPackage:     strings.TrimSpace(bootstrapMetaPackage),
		MetaVersion:     strings.TrimSpace(bootstrapFrameworkVersion),
		DBDump:          strings.TrimSpace(bootstrapDBDump),
		FixDeps:         bootstrapFixDeps,
		HyvaInstall:     bootstrapHyvaInstall,
		HyvaToken:       strings.TrimSpace(bootstrapHyvaToken),
		MageUsername:    strings.TrimSpace(bootstrapMageUsername),
		MagePassword:    strings.TrimSpace(bootstrapMagePassword),
		AssumeYes:       bootstrapAssumeYes,
		IncludeProduct:  bootstrapIncludeProduct,
		Plan:            bootstrapPlan,
		NoNoise:         bootstrapNoNoise,
		NoPII:           bootstrapNoPII,
		DeleteSync:      bootstrapDelete,
		NoCompress:      bootstrapNoCompress,
		ExcludePatterns: bootstrapExclude,
	}

	if opts.MetaPackage == "" {
		opts.MetaPackage = defaultBootstrapMetaPackage
	}
	if opts.HyvaToken == "" {
		opts.HyvaToken = defaultBootstrapHyvaToken
	}
	cloneFlagExplicit := false
	if cmd != nil {
		cloneFlagExplicit = cmd.Flags().Changed("clone")
	}

	if opts.MetaVersion != "" {
		comparison, comparable := compareNumericDotVersions(opts.MetaVersion, "2.0.0")
		if !comparable || comparison < 0 {
			return BootstrapRuntimeOptions{}, fmt.Errorf("invalid --framework-version value %q (must be Magento 2.0.0+)", opts.MetaVersion)
		}
	}
	if opts.Fresh && opts.Clone {
		if cloneFlagExplicit {
			return BootstrapRuntimeOptions{}, fmt.Errorf("--fresh and --clone cannot be used together")
		}
		opts.Clone = false
	}
	if opts.CodeOnly && !opts.Clone {
		return BootstrapRuntimeOptions{}, fmt.Errorf("--code-only requires --clone")
	}
	if opts.Fresh {
		opts.ComposerInstall = false
		opts.DBImport = false
		opts.MediaSync = false
	}
	if opts.Clone && opts.CodeOnly {
		opts.DBImport = false
		opts.MediaSync = false
	}

	return opts, nil
}

func normalizeBootstrapSource(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func DefaultBootstrapRuntimeOptionsForTest() BootstrapRuntimeOptions {
	return BootstrapRuntimeOptions{
		Source:          "",
		DBImport:        true,
		MediaSync:       true,
		ComposerInstall: true,
		AdminCreate:     true,
		StreamDB:        true,
		MetaPackage:     defaultBootstrapMetaPackage,
		HyvaToken:       defaultBootstrapHyvaToken,
	}
}

// ResetBootstrapFlags resets all package-level bootstrap flag variables.
// Used primarily for testing to ensure a clean state between runs.
func ResetBootstrapFlags() {
	bootstrapClone = false
	bootstrapCodeOnly = false
	bootstrapFresh = false
	bootstrapIncludeSample = false
	bootstrapSkipDB = false
	bootstrapSkipMedia = false
	bootstrapSkipComposer = false
	bootstrapSkipAdmin = false
	bootstrapNoStreamDB = false
	bootstrapEnv = ""
	bootstrapFramework = ""
	bootstrapFrameworkVersion = ""
	bootstrapSkipUp = false
	bootstrapMetaPackage = defaultBootstrapMetaPackage
	bootstrapDBDump = ""
	bootstrapFixDeps = false
	bootstrapHyvaInstall = false
	bootstrapHyvaToken = defaultBootstrapHyvaToken
	bootstrapMageUsername = ""
	bootstrapMagePassword = ""
	bootstrapAssumeYes = false
	bootstrapIncludeProduct = false
	bootstrapPlan = false
	bootstrapNoNoise = false
	bootstrapNoPII = false
	bootstrapDelete = false
	bootstrapNoCompress = false
	bootstrapExclude = []string{}

	bootstrapCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
}
