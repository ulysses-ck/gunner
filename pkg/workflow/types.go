package workflow

import (
	"fmt"

	"github.com/ulysses-ck/gunner/pkg/parser"
)

type Pipeline struct {
	Name         string            `yaml:"name"`
	Environments map[string]string `yaml:"environments"`
	Jobs         map[string]Job    `yaml:"jobs"`
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

func FromAST(ast *parser.Node) (*Pipeline, error) {
	data := ast.ToMap()

	pipeline := &Pipeline{
		Environments: make(map[string]string),
		Jobs:         make(map[string]Job),
	}

	if name, ok := data["name"].(string); ok {
		pipeline.Name = name
	}

	if envs, ok := data["environments"].(map[string]interface{}); ok {
		for k, v := range envs {
			if strVal, ok := v.(string); ok {
				pipeline.Environments[k] = strVal
			}
		}
	}

	if jobs, ok := data["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if jobMap, ok := jobData.(map[string]interface{}); ok {
				job := Job{
					Needs: []string{},
					Steps: []Step{},
				}

				if runsOn, ok := jobMap["runs-on"].(string); ok {
					job.RunsOn = runsOn
				}

				if stage, ok := jobMap["stage"].(string); ok {
					job.Stage = stage
				}

				if needs, ok := jobMap["needs"].([]interface{}); ok {
					for _, need := range needs {
						if needStr, ok := need.(string); ok {
							job.Needs = append(job.Needs, needStr)
						}
					}
				}

				if steps, ok := jobMap["steps"].([]interface{}); ok {
					for _, stepData := range steps {
						if stepMap, ok := stepData.(map[string]interface{}); ok {
							step := Step{
								With: make(map[string]string),
								Env:  make(map[string]string),
							}

							if name, ok := stepMap["name"].(string); ok {
								step.Name = name
							}

							if run, ok := stepMap["run"].(string); ok {
								step.Run = run
							}

							if uses, ok := stepMap["uses"].(string); ok {
								step.Uses = uses
							}

							if with, ok := stepMap["with"].(map[string]interface{}); ok {
								for k, v := range with {
									if strVal, ok := v.(string); ok {
										step.With[k] = strVal
									}
								}
							}

							if env, ok := stepMap["env"].(map[string]interface{}); ok {
								for k, v := range env {
									if strVal, ok := v.(string); ok {
										step.Env[k] = strVal
									}
								}
							}

							job.Steps = append(job.Steps, step)
						}
					}
				}

				pipeline.Jobs[jobName] = job
			}
		}
	}

	return pipeline, nil
}

func LoadPipeline(yamlContent string) (*Pipeline, error) {
	ast, err := parser.ParseYAML(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	pipeline, err := FromAST(ast)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %w", err)
	}

	return pipeline, nil
}
