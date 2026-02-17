package homebrew

import (
	"encoding/json"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// brew outdated --json=v2 output structure.
type brewOutdatedOutput struct {
	Formulae []brewFormula `json:"formulae"`
	Casks    []brewCask    `json:"casks"`
}

type brewFormula struct {
	Name              string   `json:"name"`
	InstalledVersions []string `json:"installed_versions"`
	CurrentVersion    string   `json:"current_version"`
	Pinned            bool     `json:"pinned"`
}

type brewCask struct {
	Name              string   `json:"name"`
	InstalledVersions []string `json:"installed_versions"`
	CurrentVersion    string   `json:"current_version"`
}

// parseOutdated parses the JSON output of "brew outdated --json=v2" into Updates.
func parseOutdated(jsonData string, includeCasks bool) ([]checker.Update, error) {
	jsonData = strings.TrimSpace(jsonData)
	if jsonData == "" || jsonData == "{}" {
		return nil, nil
	}

	var output brewOutdatedOutput
	if err := json.Unmarshal([]byte(jsonData), &output); err != nil {
		return nil, err
	}

	var updates []checker.Update

	for _, f := range output.Formulae {
		if f.Pinned {
			continue
		}
		current := ""
		if len(f.InstalledVersions) > 0 {
			current = f.InstalledVersions[0]
		}
		updates = append(updates, checker.Update{
			Name:           f.Name,
			CurrentVersion: current,
			NewVersion:     f.CurrentVersion,
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
			Source:         "formulae",
		})
	}

	if includeCasks {
		for _, c := range output.Casks {
			current := ""
			if len(c.InstalledVersions) > 0 {
				current = c.InstalledVersions[0]
			}
			updates = append(updates, checker.Update{
				Name:           c.Name,
				CurrentVersion: current,
				NewVersion:     c.CurrentVersion,
				Type:           checker.UpdateTypeRegular,
				Priority:       checker.PriorityNormal,
				Source:         "casks",
			})
		}
	}

	return updates, nil
}
