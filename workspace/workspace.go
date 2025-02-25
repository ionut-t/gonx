package workspace

import (
	"encoding/json"
	"fmt"
	"github.com/ionut-t/gonx/internal/constants"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ionut-t/gonx/utils"
)

type DoneMsg struct {
	Workspace Model
}

type ErrMsg struct {
	Err error
}

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

// CacheMetadata Cache-related structures
type CacheMetadata struct {
	Timestamp    time.Time         `json:"timestamp"`
	ProjectPaths map[string]string `json:"projectPaths"` // Map of project name to project.json path
}

type CachedWorkspace struct {
	Model    Model         `json:"model"`
	Metadata CacheMetadata `json:"metadata"`
}

func New() (*Model, error) {
	// Create channels for message passing
	doneChan := make(chan DoneMsg, 1)
	errChan := make(chan ErrMsg, 1)

	// Start workspace scan in a goroutine
	go func() {
		cwd, err := os.Getwd()
		if err != nil {
			errChan <- ErrMsg{Err: err}
			return
		}

		workspace := Model{
			Name: filepath.Base(cwd),
		}

		// Check if we can use the cache
		if cachedModel, isFresh := tryLoadFromCache(); isFresh {
			doneChan <- DoneMsg{Workspace: *cachedModel}
			return
		}

		apps, libs, e2eApps, err := getAllApps()
		if err != nil {
			errChan <- ErrMsg{Err: err}
			return
		}

		workspace.Applications = apps
		workspace.Libraries = libs
		workspace.E2EApps = e2eApps

		// Save to cache for next time
		saveToCache(&workspace)

		// Send the done message
		doneChan <- DoneMsg{Workspace: workspace}
	}()

	// Wait for either done or error message
	select {
	case done := <-doneChan:
		return &done.Workspace, nil
	case err := <-errChan:
		return nil, err.Err
	}
}

// getAllApps retrieves all projects from the NX workspace
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

// parseApps processes the list of projects with parallelized execution
func parseApps(allApps []string) ([]Application, []Library, []E2EApp) {
	var apps []Application
	var libs []Library
	var e2eApps []E2EApp
	var mutex sync.Mutex // Protects concurrent access to the slices

	// Early filter for e2e projects to avoid unnecessary processing
	var projectsToProcess []string
	for _, app := range allApps {
		if strings.HasSuffix(app, ".e2e") {
			continue // Skip these entirely
		}

		if strings.Contains(app, "-e2e") {
			e2eApps = append(e2eApps, E2EApp{
				Name: app,
				Type: E2EType,
			})
			continue // Already processed
		}

		projectsToProcess = append(projectsToProcess, app)
	}

	// Create channels for results and semaphore for limiting concurrency
	type projectResult struct {
		app         string
		projectType ProjectType
		outputPath  string
	}
	resultChan := make(chan projectResult, len(projectsToProcess))

	// Limit concurrency to avoid overwhelming the system
	semaphore := make(chan struct{}, 10) // Process 10 projects concurrently
	var wg sync.WaitGroup

	// Process each app concurrently
	for _, app := range projectsToProcess {
		wg.Add(1)
		go func(app string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			cmd := exec.Command("nx", "show", "project", app, "--json")
			output, err := cmd.CombinedOutput()

			if err != nil {
				fmt.Printf("Failed to show project %s: %v\nOutput: %s", app, err, string(output))
				resultChan <- projectResult{app: ""}
				return
			}

			var projectConfig ProjectConfig
			if err := json.Unmarshal(output, &projectConfig); err != nil {
				fmt.Printf("Failed to parse project config for %s: %v", app, err)
				resultChan <- projectResult{app: ""}
				return
			}

			resultChan <- projectResult{
				app:         app,
				projectType: ProjectType(projectConfig.ProjectType),
				outputPath:  projectConfig.Targets.Build.Options.OutputPath,
			}
		}(app)
	}

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		// Skip empty results (errors)
		if result.app == "" {
			continue
		}

		if result.projectType == ApplicationType {
			mutex.Lock()
			apps = append(apps, Application{
				Name:       result.app,
				OutputPath: result.outputPath,
				Type:       ApplicationType,
			})
			mutex.Unlock()
		}

		if result.projectType == LibraryType {
			mutex.Lock()
			libs = append(libs, Library{
				Name: filepath.Base(result.app),
				Type: LibraryType,
			})
			mutex.Unlock()
		}
	}

	return apps, libs, e2eApps
}

// tryLoadFromCache attempts to load a cached workspace if it's still valid
func tryLoadFromCache() (*Model, bool) {
	// Check if cache file exists
	cacheInfo, err := os.Stat(constants.CacheFilePath)
	if err != nil {
		return nil, false
	}

	// Read and parse cache
	data, err := os.ReadFile(constants.CacheFilePath)
	if err != nil {
		return nil, false
	}

	var cachedWorkspace CachedWorkspace
	if err := json.Unmarshal(data, &cachedWorkspace); err != nil {
		return nil, false
	}

	// Get all the project files and check if any are newer than our cache
	for _, projectPath := range cachedWorkspace.Metadata.ProjectPaths {
		projectInfo, err := os.Stat(projectPath)
		if err != nil {
			// Project file no longer exists or can't be accessed
			return nil, false
		}

		if projectInfo.ModTime().After(cacheInfo.ModTime()) {
			// A project file was modified after the cache was created
			return nil, false
		}
	}

	// Check if any new projects have been added by running nx show projects
	cmd := exec.Command("nx", "show", "projects")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If we can't run nx, assume cache is stale
		return nil, false
	}

	currentProjects := strings.Split(strings.TrimSpace(string(output)), "\n")
	cachedProjects := make([]string, 0, len(cachedWorkspace.Metadata.ProjectPaths))
	for project := range cachedWorkspace.Metadata.ProjectPaths {
		cachedProjects = append(cachedProjects, project)
	}

	// Sort both lists for comparison
	slices.Sort(currentProjects)
	slices.Sort(cachedProjects)

	if !slices.Equal(currentProjects, cachedProjects) {
		// Project list has changed
		return nil, false
	}

	return &cachedWorkspace.Model, true
}

// saveToCache saves the workspace model to cache with metadata
func saveToCache(model *Model) {
	// Create a map of project paths
	projectPaths := make(map[string]string)

	// Find paths to all project.json files
	for _, app := range model.Applications {
		projectPath := findProjectJsonPath(app.Name)
		if projectPath != "" {
			projectPaths[app.Name] = projectPath
		}
	}

	for _, lib := range model.Libraries {
		projectPath := findProjectJsonPath(lib.Name)
		if projectPath != "" {
			projectPaths[lib.Name] = projectPath
		}
	}

	for _, e2e := range model.E2EApps {
		projectPath := findProjectJsonPath(e2e.Name)
		if projectPath != "" {
			projectPaths[e2e.Name] = projectPath
		}
	}

	cache := CachedWorkspace{
		Model: *model,
		Metadata: CacheMetadata{
			Timestamp:    time.Now(),
			ProjectPaths: projectPaths,
		},
	}

	data, err := json.Marshal(cache)
	if err != nil {
		fmt.Printf("Warning: Failed to marshal model for caching: %v\n", err)
		return
	}

	if err := os.MkdirAll(constants.Folder, 0755); err != nil {
		fmt.Printf("Warning: Failed to create cache folder: %v\n", err)
	}

	if err := os.WriteFile(constants.CacheFilePath, data, 0644); err != nil {
		fmt.Printf("Warning: Failed to write cache file: %v\n", err)
	}
}

// findProjectJsonPath attempts to locate the project.json file for a given project
func findProjectJsonPath(projectName string) string {
	// First try direct path (nx 12+)
	path := filepath.Join(projectName, "project.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Try apps/libs directories (earlier nx versions)
	for _, dir := range []string{"apps", "libs"} {
		path = filepath.Join(dir, projectName, "project.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Use nx to get the project path
	cmd := exec.Command("nx", "show", "project", projectName, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	var projectConfig struct {
		Root string `json:"root"`
	}

	if err := json.Unmarshal(output, &projectConfig); err != nil {
		return ""
	}

	path = filepath.Join(projectConfig.Root, "project.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// GetProjects returns projects filtered by type
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
