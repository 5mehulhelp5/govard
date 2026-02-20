package desktop

import (
	"context"
	"sync"
	"time"
)

type App struct {
	ctx          context.Context
	streamMu     sync.Mutex
	streamCancel context.CancelFunc
	notifyMu     sync.Mutex
	notifyCancel context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (app *App) Startup(ctx context.Context) {
	app.ctx = ctx
	app.startOperationNotificationWatcher()
}

func (app *App) Shutdown(ctx context.Context) {
	_ = ctx
	app.stopOperationNotificationWatcher()
}

func (app *App) Status() string {
	return "Govard Desktop bootstrap is ready."
}

func (app *App) GetDashboard() Dashboard {
	dashboard, err := buildDashboard()
	if err != nil {
		dashboard = Dashboard{
			ActiveEnvironments: 0,
			RunningServices:    0,
			QueuedTasks:        0,
			ActiveSummary:      "No environments detected",
			ServicesSummary:    "Docker unavailable",
			QueueSummary:       "Queue idle",
			Environments:       []Environment{},
			Warnings: []string{
				"Docker unavailable. Showing cached or mock data.",
			},
		}
	}
	return dashboard
}

func (app *App) GetResourceMetrics() ResourceMetricsSnapshot {
	snapshot, err := buildResourceMetrics()
	if err == nil {
		return snapshot
	}
	return ResourceMetricsSnapshot{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Summary:   ResourceMetricsSummary{},
		Projects:  []ProjectResourceMetric{},
		Warnings:  []string{"Metrics unavailable: " + err.Error()},
	}
}

func (app *App) ToggleEnvironment(name string) string {
	message, err := toggleEnvironment(name)
	if err != nil {
		return "Failed to toggle " + name + ": " + err.Error()
	}
	return message
}

func (app *App) OpenEnvironment(name string) string {
	url, err := environmentURL(name)
	if err != nil {
		return "Unable to determine URL for " + name + ": " + err.Error()
	}
	if err := openURLWithPreferences(app.ctx, url); err != nil {
		return "Open " + url + " manually"
	}
	return "Opening " + url + "..."
}

func (app *App) QuickAction(action string) string {
	if message, err := quickAction(app.ctx, action, ""); err == nil {
		return message
	}
	return "Action received: " + action
}

func (app *App) QuickActionForProject(action string, project string) string {
	if message, err := quickAction(app.ctx, action, project); err == nil {
		return message
	}
	return "Action received: " + action
}

func (app *App) GetLogs(project string) string {
	logs, err := getLogs(project, 200)
	if err != nil {
		return "Failed to load logs: " + err.Error()
	}
	return logs
}

func (app *App) GetLogsForService(project string, service string) string {
	logs, err := getLogsForService(project, service, 200)
	if err != nil {
		return "Failed to load logs: " + err.Error()
	}
	return logs
}

func (app *App) OpenShell(project string) string {
	if err := openShell(project); err != nil {
		return "Failed to open shell: " + err.Error()
	}
	return "Opened shell for " + project
}

func (app *App) OpenShellForService(project string, service string, user string, shell string) string {
	if err := openShellForService(project, service, user, shell); err != nil {
		return "Failed to open shell: " + err.Error()
	}
	return "Opened shell for " + project
}

func (app *App) GetShellUser(project string) string {
	user, err := getShellUser(project)
	if err != nil {
		return ""
	}
	return user
}

func (app *App) SetShellUser(project string, user string) string {
	if err := setShellUser(project, user); err != nil {
		return "Failed to save shell user: " + err.Error()
	}
	if user == "" {
		return "Cleared shell user for " + project
	}
	return "Saved shell user for " + project
}

func (app *App) ResetShellUsers() string {
	if err := resetShellUsers(); err != nil {
		return "Failed to reset shell users: " + err.Error()
	}
	return "Shell user preferences reset"
}

func (app *App) GetSettings() DesktopSettings {
	settings, err := getSettings()
	if err != nil {
		return DesktopSettings{}
	}
	return settings
}

func (app *App) UpdateSettings(theme string, proxyTarget string, preferredBrowser string) string {
	settings := DesktopSettings{
		Theme:            theme,
		ProxyTarget:      proxyTarget,
		PreferredBrowser: preferredBrowser,
	}
	if err := setSettings(settings); err != nil {
		return "Failed to save settings: " + err.Error()
	}
	return "Settings updated"
}

func (app *App) ResetSettings() string {
	if err := resetSettings(); err != nil {
		return "Failed to reset settings: " + err.Error()
	}
	return "Settings reset"
}

func (app *App) OpenDocs(path string) string {
	if path == "" {
		return "No docs path provided"
	}
	if err := openDocs(app.ctx, path); err != nil {
		return "Failed to open docs: " + err.Error()
	}
	return "Opening docs..."
}

func (app *App) StartEnvironment(project string) string {
	message, err := startEnvironment(project)
	if err != nil {
		return "Failed to start " + project + ": " + err.Error()
	}
	return message
}

func (app *App) StopEnvironment(project string) string {
	message, err := stopEnvironment(project)
	if err != nil {
		return "Failed to stop " + project + ": " + err.Error()
	}
	return message
}

func (app *App) StartLogStream(project string) string {
	app.streamMu.Lock()
	defer app.streamMu.Unlock()

	if app.streamCancel != nil {
		app.streamCancel()
		app.streamCancel = nil
	}

	ctx := app.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	streamCtx, cancel := context.WithCancel(ctx)
	app.streamCancel = cancel

	go streamLogs(streamCtx, ctx, project, "")
	return "Live logs started"
}

func (app *App) StartLogStreamForService(project string, service string) string {
	app.streamMu.Lock()
	defer app.streamMu.Unlock()

	if app.streamCancel != nil {
		app.streamCancel()
		app.streamCancel = nil
	}

	ctx := app.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	streamCtx, cancel := context.WithCancel(ctx)
	app.streamCancel = cancel

	go streamLogs(streamCtx, ctx, project, service)
	return "Live logs started"
}

func (app *App) StopLogStream() string {
	app.streamMu.Lock()
	defer app.streamMu.Unlock()

	if app.streamCancel != nil {
		app.streamCancel()
		app.streamCancel = nil
		return "Live logs stopped"
	}
	return "Live logs already stopped"
}
