package engine

// ApplyRuntimeProfileToConfig maps a resolved runtime profile onto config.
func ApplyRuntimeProfileToConfig(config *Config, profile RuntimeProfile) {
	if config == nil {
		return
	}

	config.Framework = profile.Framework
	config.FrameworkVersion = profile.FrameworkVersion
	config.Stack.PHPVersion = profile.PHPVersion
	config.Stack.NodeVersion = profile.NodeVersion
	config.Stack.DBType = profile.DBType
	config.Stack.DBVersion = profile.DBVersion
	config.Stack.WebRoot = profile.WebRoot
	config.Stack.XdebugSession = profile.XdebugSession
	config.Stack.Services.WebServer = profile.WebServer
	config.Stack.Services.Cache = profile.Cache
	config.Stack.Services.Search = profile.Search
	config.Stack.Services.Queue = profile.Queue
	config.Stack.CacheVersion = profile.CacheVersion
	config.Stack.SearchVersion = profile.SearchVersion
	config.Stack.QueueVersion = profile.QueueVersion
	config.Stack.ComposerVersion = profile.ComposerVersion
}
