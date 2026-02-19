package desktop

type Dashboard struct {
	ActiveEnvironments int           `json:"activeEnvironments"`
	RunningServices    int           `json:"runningServices"`
	QueuedTasks        int           `json:"queuedTasks"`
	ActiveSummary      string        `json:"activeSummary"`
	ServicesSummary    string        `json:"servicesSummary"`
	QueueSummary       string        `json:"queueSummary"`
	Environments       []Environment `json:"environments"`
	Warnings           []string      `json:"warnings"`
}

type DesktopSettings struct {
	Theme            string `json:"theme"`
	ProxyTarget      string `json:"proxyTarget"`
	PreferredBrowser string `json:"preferredBrowser"`
}

type Environment struct {
	Project        string   `json:"project"`
	Domain         string   `json:"domain"`
	Name           string   `json:"name"`
	Framework      string   `json:"framework"`
	PHP            string   `json:"php"`
	Database       string   `json:"database"`
	Services       []string `json:"services"`
	ServiceTargets []string `json:"serviceTargets"`
	Status         string   `json:"status"`
}
