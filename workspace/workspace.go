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

type Model struct {
	Name         string        `json:"name"`
	Applications []Application `json:"applications"`
	Libraries    []Library     `json:"libraries"`
	E2EApps      []E2EApp      `json:"e2eApps"`
}

func (m *Model) String() string {
	return utils.PrettyJSON(m)
}

type ProjectType string

const (
	ApplicationType ProjectType = "application"
	LibraryType     ProjectType = "library"
	E2EType         ProjectType = "e2e"
)

type Project interface {
	GetName() string
	GetType() ProjectType
}

type Application struct {
	Name       string      `json:"name"`
	OutputPath string      `json:"outputPath"`
	Type       ProjectType `json:"type"`
}

func (a Application) GetName() string {
	return a.Name
}

func (a Application) GetType() ProjectType {
	return a.Type
}

type Library struct {
	Name  string      `json:"name"`
	Suite string      `json:"suite"`
	Type  ProjectType `json:"type"`
}

func (l Library) GetName() string {
	return l.Name
}

func (l Library) GetType() ProjectType {
	return l.Type
}

type E2EApp struct {
	Name  string      `json:"name"`
	Suite string      `json:"suite"`
	Type  ProjectType `json:"type"`
}

func (e E2EApp) GetName() string {
	return e.Name
}

func (e E2EApp) GetType() ProjectType {
	return e.Type
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

			_ = json.Unmarshal(output, &projectConfig)

			// Ignore parsing errors for now since the project configuration it's different for each project
			// TODO: Find a way to properly parse the project configuration for each project
			//if parseError != nil {
			//	fmt.Printf("Failed to parse project config for %s %s", app, parseError)
			//	continue
			//}
			projectType := projectConfig.ProjectType
			if strings.HasSuffix(app, ".e2e") {
				continue
			}

			if strings.Contains(app, "-e2e") {
				e2eApps = append(e2eApps, E2EApp{
					Name: app,
					Type: E2EType,
				})
				continue
			}

			if ProjectType(projectType) == ApplicationType {
				apps = append(apps, Application{
					Name:       app,
					OutputPath: projectConfig.Targets.Build.Options.OutputPath,
					Type:       ApplicationType,
				})
			}

			if ProjectType(projectType) == LibraryType {
				libs = append(libs, Library{
					Name: filepath.Base(app),
					Type: LibraryType,
				})
			}
		}
	}

	return apps, libs, e2eApps
}

type DoneMsg struct {
	Workspace Model
}

type ErrMsg struct {
	Err error
}

func New() (*Model, error) {
	cwd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	workspace := Model{
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

func (m Model) GetProjects(includeTypes []ProjectType) []Project {
	var projects []Project

	if len(includeTypes) == 0 {
		return projects
	}

	if slices.Contains(includeTypes, ApplicationType) {
		for _, app := range m.Applications {
			projects = append(projects, app)
		}
	}

	if slices.Contains(includeTypes, LibraryType) {
		for _, lib := range m.Libraries {
			projects = append(projects, lib)
		}
	}

	if slices.Contains(includeTypes, E2EType) {
		for _, e2eApp := range m.E2EApps {
			projects = append(projects, e2eApp)
		}
	}

	return projects
}
