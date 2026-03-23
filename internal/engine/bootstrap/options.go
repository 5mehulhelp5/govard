package bootstrap

type Options struct {
	Source   string
	Clone    bool
	CodeOnly bool
	Fresh    bool
	Version  string
	Env      string
	Runner   func(command string) error

	// Database credentials for local configuration
	DBHost string
	DBUser string
	DBPass string
	DBName string

	// Environment configuration
	Domain string
}

func DefaultOptions() Options {
	return Options{Source: "staging"}
}

type FrameworkBootstrap interface {
	Name() string
	SupportsFreshInstall() bool
	SupportsClone() bool
	FreshCommands() []string
	CreateProject(projectDir string) error
	Install(projectDir string) error
	Configure(projectDir string) error
	PostClone(projectDir string) error
}
