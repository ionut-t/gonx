package workspace

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ionut-t/gonx/utils"
)

type Workspace struct {
	Name         string        `json:"name"`
	Applications []Application `json:"applications"`
	Libraries    []Library     `json:"libraries"`
	E2EApps      []E2EApp      `json:"e2eApps"`
}

func (w *Workspace) String() string {
	return utils.PrettyJSON(w)
}

type Application struct {
	Name  string `json:"name"`
	Suite string `json:"suite"`
}

type Library struct {
	Name  string `json:"name"`
	Suite string `json:"suite"`
}

type E2EApp struct {
	Name  string `json:"name"`
	Suite string `json:"suite"`
}

func getAllApps() (map[string][]string, error) {
	cmd := exec.Command("nx", "show", "projects")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	outputStr := string(output)
	rawApps := strings.Split(strings.TrimSpace(outputStr), "\n")

	apps := filterApps(rawApps)

	return apps, nil
}

func filterApps(allApps []string) map[string][]string {
	var apps []string
	var libs []string
	var e2eApps []string

	for _, app := range allApps {
		cmd := exec.Command("nx", "show", "project", app, "--json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Failed to show project %s: %v\nOutput: %s", app, err, string(output))
			continue
		} else {
			var projectConfig map[string]interface{}

			parseError := json.Unmarshal(output, &projectConfig)

			if parseError != nil {
				fmt.Printf("Failed to parse project config for %s %s", app, parseError)
				continue
			} else {
				if projectType, ok := projectConfig["projectType"].(string); ok {
					if strings.HasSuffix(app, ".e2e") {
						continue
					}

					if strings.Contains(app, "-e2e") {
						e2eApps = append(e2eApps, app)
						continue
					}

					if projectType == "application" {
						apps = append(apps, app)
					}

					if projectType == "library" {
						libs = append(libs, app)
					}
				} else {
					fmt.Printf("Failed to assert projectType for %s", app)
				}
			}
		}

	}

	return map[string][]string{
		"applications": apps,
		"libraries":    libs,
		"e2eApps":      e2eApps,
	}
}

type DoneMsg struct {
	Workspace Workspace
}

type ErrMsg struct {
	Err error
}

func NewWorkspace() (*Workspace, error) {
	workspace := Workspace{
		Name: "My Workspace",
	}

	//workspaceModel := Model{
	//	suspense:  suspense.CreateSuspenseModel(true, "Scanning workspace"),
	//	Workspace: workspace,
	//}

	apps, err := getAllApps()
	if err != nil {
		return nil, err
	}

	for _, app := range apps["applications"] {
		workspace.Applications = append(workspace.Applications, Application{Name: app})
	}

	for _, lib := range apps["libraries"] {
		workspace.Libraries = append(workspace.Libraries, Library{Name: lib})
	}

	for _, e2eApp := range apps["e2eApps"] {
		workspace.E2EApps = append(workspace.E2EApps, E2EApp{Name: e2eApp})
	}

	return &workspace, nil
}
