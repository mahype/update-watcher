package npm

import (
	"encoding/json"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// npmOutdatedEntry represents a single package from "npm outdated -g --json".
type npmOutdatedEntry struct {
	Current string `json:"current"`
	Wanted  string `json:"wanted"`
	Latest  string `json:"latest"`
}

// parseOutdated parses the JSON output of "npm outdated -g --json" into Updates.
func parseOutdated(jsonData string) ([]checker.Update, error) {
	jsonData = strings.TrimSpace(jsonData)
	if jsonData == "" || jsonData == "{}" {
		return nil, nil
	}

	var outdated map[string]npmOutdatedEntry
	if err := json.Unmarshal([]byte(jsonData), &outdated); err != nil {
		return nil, err
	}

	var updates []checker.Update
	for name, entry := range outdated {
		if entry.Current == entry.Latest {
			continue
		}
		updates = append(updates, checker.Update{
			Name:           name,
			CurrentVersion: entry.Current,
			NewVersion:     entry.Latest,
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
		})
	}

	return updates, nil
}
