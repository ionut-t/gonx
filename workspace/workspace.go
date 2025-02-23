package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	Name       string `json:"name"`
	OutputPath string `json:"outputPath"`
}

type Library struct {
	Name  string `json:"name"`
	Suite string `json:"suite"`
}

type E2EApp struct {
	Name  string `json:"name"`
	Suite string `json:"suite"`
}

func getAllApps() ([]Application, []Library, []E2EApp, error) {
	cmd := exec.Command("nx", "show", "projects")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, nil, nil, err
	}

	outputStr := string(output)
	rawApps := strings.Split(strings.TrimSpace(outputStr), "\n")

	slices.SortStableFunc(rawApps, func(a, b string) int {
		return strings.Compare(a, b)
	})

	apps, libs, e2eApps := parseApps(rawApps)

	return apps, libs, e2eApps, nil
}

func parseApps(allApps []string) ([]Application, []Library, []E2EApp) {
	var apps []Application
	var libs []Library
	var e2eApps []E2EApp

	for _, app := range allApps {
		cmd := exec.Command("nx", "show", "project", app, "--json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Failed to show project %s: %v\nOutput: %s", app, err, string(output))
			continue
		} else {
			var projectConfig ProjectConfig

			parseError := json.Unmarshal(output, &projectConfig)

			if parseError != nil {
				fmt.Printf("Failed to parse project config for %s %s", app, parseError)
				continue
			} else {
				projectType := projectConfig.ProjectType
				if strings.HasSuffix(app, ".e2e") {
					continue
				}

				if strings.Contains(app, "-e2e") {
					e2eApps = append(e2eApps, E2EApp{Name: app})
					continue
				}

				if projectType == "application" {
					apps = append(apps, Application{
						Name:       app,
						OutputPath: projectConfig.Targets.Build.Options.OutputPath,
					})
				}

				if projectType == "library" {
					libs = append(libs, Library{Name: app})
				}
			}
		}
	}

	return apps, libs, e2eApps
}

type DoneMsg struct {
	Workspace Workspace
}

type ErrMsg struct {
	Err error
}

func NewWorkspace() (*Workspace, error) {
	cwd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	workspace := Workspace{
		Name: filepath.Base(cwd),
	}

	apps, libs, e2eApps, err := getAllApps()
	if err != nil {
		return nil, err
	}

	workspace.Applications = apps
	workspace.Libraries = libs
	workspace.E2EApps = e2eApps

	return &workspace, nil
}
