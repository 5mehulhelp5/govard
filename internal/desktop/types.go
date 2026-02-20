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

type ResourceMetricsSnapshot struct {
	UpdatedAt string                  `json:"updatedAt"`
	Summary   ResourceMetricsSummary  `json:"summary"`
	Projects  []ProjectResourceMetric `json:"projects"`
	Warnings  []string                `json:"warnings"`
}

type ResourceMetricsSummary struct {
	ActiveProjects int     `json:"activeProjects"`
	CPUPercent     float64 `json:"cpuPercent"`
	MemoryMB       float64 `json:"memoryMB"`
	NetRxMB        float64 `json:"netRxMB"`
	NetTxMB        float64 `json:"netTxMB"`
	OOMProjects    int     `json:"oomProjects"`
}

type ProjectResourceMetric struct {
	Project       string  `json:"project"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryMB      float64 `json:"memoryMB"`
	MemoryPercent float64 `json:"memoryPercent"`
	NetRxMB       float64 `json:"netRxMB"`
	NetTxMB       float64 `json:"netTxMB"`
	OOMKilled     bool    `json:"oomKilled"`
}
