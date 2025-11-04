package workflow

type Pipeline struct {
	Name        string            `yaml:"name"`
	Environment map[string]string `yaml:"environments"`
	Jobs        map[string]Job    `yaml:"jobs"`
}

type Job struct {
	RunsOn string   `yaml:"runs-on"`
	Stage  string   `yaml:"stage"`
	Needs  []string `yaml:"needs"`
	Steps  []Step   `yaml:"steps"`
}

type Step struct {
	Name string            `yaml:"name"`
	Run  string            `yaml:"run,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
	Env  map[string]string `yaml:"env,omitempty"`
}
