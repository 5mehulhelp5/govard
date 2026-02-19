package updater

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pterm/pterm"
)

func CheckForUpdates(current string) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/ddtcorex/govard/releases/latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	if release.TagName != "" && release.TagName != "v"+current {
		pterm.Warning.Printf("A new version of Govard is available: %s (current: %s)\n", release.TagName, current)
		pterm.Info.Println("Run 'govard self-update' to upgrade.")
	}
}
