package bootstrap

type Options struct {
	Source   string
	Clone    bool
	CodeOnly bool
	Fresh    bool
	Version  string
	Env      string
}

func DefaultOptions() Options {
	return Options{Source: "staging"}
}
