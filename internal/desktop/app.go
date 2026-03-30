package desktop

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"sync"

	"govard/internal/engine"
)

func (app *App) GetUserInfo() (res UserInfo, err error) {
	defer RecoverPanic(&err, "GetUserInfo")
	res = UserInfo{
		Username: "unknown",
		Name:     "Unknown User",
	}
	u, errCurrent := user.Current()
	if errCurrent != nil {
		return res, errCurrent
	}
	res.Username = u.Username
	res.Name = u.Name
	if res.Name == "" {
		res.Name = u.Username
	}
	return res, nil
}

var Version = "1.37.0"

type App struct {
	ctx context.Context

	Settings    *SettingsService
	Onboarding  *OnboardingService
	Environment *EnvironmentService
	Remote      *RemoteService
	System      *SystemService
	Logs        *LogService
	Global      *GlobalServiceService

	notifyMu     sync.Mutex
	notifyCancel context.CancelFunc
}

func NewApp() *App {
	return &App{
		Settings:    NewSettingsService(),
		Onboarding:  NewOnboardingService(),
		Environment: NewEnvironmentService(),
		Remote:      NewRemoteService(),
		System:      NewSystemService(),
		Logs:        NewLogService(),
		Global:      NewGlobalServiceService(),
	}
}

func (app *App) GetVersion() (v string, err error) {
	defer RecoverPanic(&err, "GetVersion")
	return Version, nil
}

func (app *App) Startup(ctx context.Context) {
	app.ctx = ctx
	app.Settings.Setup(ctx)
	app.Onboarding.Setup(ctx)
	app.Environment.Setup(ctx)
	app.Remote.Setup(ctx)
	app.System.Setup(ctx)
	app.Logs.Setup(ctx)
	app.Global.Setup(ctx)

	app.startOperationNotificationWatcher()
}

func (app *App) showWindow() {
	if app == nil || app.ctx == nil {
		return
	}
	showApplication(app.ctx)
}

func (app *App) hideWindow(ctx context.Context) {
	if app == nil {
		return
	}
	targetCtx := ctx
	if targetCtx == nil {
		targetCtx = app.ctx
	}
	if targetCtx == nil {
		return
	}
	hideApplication(targetCtx)
}

func (app *App) BeforeClose(ctx context.Context) bool {
	settings, err := app.Settings.GetSettings()
	if err != nil {
		return false
	}
	if settings.RunInBackground {
		app.hideWindow(ctx)
		return true // prevent close
	}
	return false // allow close
}

func (app *App) Shutdown(ctx context.Context) {
	_ = ctx
	app.stopOperationNotificationWatcher()
}

func (app *App) Status() string {
	return "Govard Desktop ready."
}

func (app *App) OpenDocs(path string) (res string, err error) {
	defer RecoverPanic(&err, "OpenDocs")
	if path == "" {
		return "", fmt.Errorf("no docs path provided")
	}
	if errOpen := openDocs(app.ctx, path); errOpen != nil {
		return "", fmt.Errorf("failed to open docs: %w", errOpen)
	}
	return "Opening docs...", nil
}

func (app *App) QuickAction(action string) (res string, err error) {
	defer RecoverPanic(&err, "QuickAction")
	return quickAction(app.ctx, action, "")
}

func (app *App) QuickActionForProject(action string, project string) (res string, err error) {
	defer RecoverPanic(&err, "QuickActionForProject")
	return quickAction(app.ctx, action, project)
}

func (app *App) GetShellUser(project string) (res string, err error) {
	defer RecoverPanic(&err, "GetShellUser")
	return getShellUser(project)
}

func (app *App) SetShellUser(project string, user string) (res string, err error) {
	defer RecoverPanic(&err, "SetShellUser")
	if errSet := setShellUser(project, user); errSet != nil {
		return "", errSet
	}
	if user == "" {
		return "Cleared shell user for " + project, nil
	}
	return "Saved shell user for " + project, nil
}

func (app *App) ResetShellUsers() (res string, err error) {
	defer RecoverPanic(&err, "ResetShellUsers")
	if errReset := resetShellUsers(); errReset != nil {
		return "", errReset
	}
	return "Shell user preferences reset", nil
}

func (app *App) DeleteProject(projectQuery string) (res string, err error) {
	defer RecoverPanic(&err, "DeleteProject")
	if projectQuery == "" {
		return "", fmt.Errorf("project name or path is required")
	}

	root, err := resolveProjectRootForRemotes(projectQuery)
	if err != nil {
		return "", err
	}

	// We use the application context for the deletion process
	if err := engine.DeleteProject(app.ctx, root, os.Stdout, os.Stderr); err != nil {
		return "", err
	}
	return "Project deleted successfully", nil
}
