package workspace

type ProjectConfig struct {
	Name        string  `json:"name"`
	Schema      string  `json:"$schema"`
	ProjectType string  `json:"projectType"`
	Prefix      string  `json:"prefix"`
	SourceRoot  string  `json:"sourceRoot"`
	Tags        []any   `json:"tags"`
	Targets     Targets `json:"targets"`
}

type Assets struct {
	Glob  string `json:"glob"`
	Input string `json:"input"`
}

type BuildOptions struct {
	OutputPath          string   `json:"outputPath"`
	Index               string   `json:"index"`
	Browser             string   `json:"browser"`
	Polyfills           []string `json:"polyfills"`
	TsConfig            string   `json:"tsConfig"`
	InlineStyleLanguage string   `json:"inlineStyleLanguage"`
	Assets              []Assets `json:"assets"`
	Styles              []string `json:"styles"`
	Scripts             []any    `json:"scripts"`
}
type Budgets struct {
	Type           string `json:"type"`
	MaximumWarning string `json:"maximumWarning"`
	MaximumError   string `json:"maximumError"`
}

type Production struct {
	Budgets       []Budgets `json:"budgets"`
	OutputHashing string    `json:"outputHashing"`
}

type Development struct {
	Optimization    bool `json:"optimization"`
	ExtractLicenses bool `json:"extractLicenses"`
	SourceMap       bool `json:"sourceMap"`
}

type Configurations struct {
	Production  Production  `json:"production"`
	Development Development `json:"development"`
}

type Build struct {
	Executor             string         `json:"executor"`
	Outputs              []string       `json:"outputs"`
	Options              BuildOptions   `json:"options"`
	Configurations       Configurations `json:"configurations"`
	DefaultConfiguration string         `json:"defaultConfiguration"`
}
type Serve struct {
	Executor             string         `json:"executor"`
	Configurations       Configurations `json:"configurations"`
	DefaultConfiguration string         `json:"defaultConfiguration"`
}

type ExtractI18NOptions struct {
	BuildTarget string `json:"buildTarget"`
}
type ExtractI18N struct {
	Executor string             `json:"executor"`
	Options  ExtractI18NOptions `json:"options"`
}

type Lint struct {
	Executor string `json:"executor"`
}
type TestOptions struct {
	JestConfig string `json:"jestConfig"`
}

type Test struct {
	Executor string      `json:"executor"`
	Outputs  []string    `json:"outputs"`
	Options  TestOptions `json:"options"`
}

type ServeStaticOptions struct {
	BuildTarget    string `json:"buildTarget"`
	Port           int    `json:"port"`
	StaticFilePath string `json:"staticFilePath"`
	Spa            bool   `json:"spa"`
}

type ServeStatic struct {
	Executor string             `json:"executor"`
	Options  ServeStaticOptions `json:"options"`
}

type Targets struct {
	Build       Build       `json:"build"`
	Serve       Serve       `json:"serve"`
	ExtractI18N ExtractI18N `json:"extract-i18n"`
	Lint        Lint        `json:"lint"`
	Test        Test        `json:"test"`
	ServeStatic ServeStatic `json:"serve-static"`
}
