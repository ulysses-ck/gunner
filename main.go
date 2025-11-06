package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ulysses-ck/gunner/pkg/workflow"
)

func main() {
	yamlContent := `name: citest
environments:
  builder: debian

jobs:
  test-build:
    runs-on: builder
    stage: build
    needs: 
      - setup
    steps:
      - name: Compile Go
        run: go build .
      - name: Run tests
        uses: actions/test
        with:
          coverage: true
        env:
          GO_VERSION: 1.21
  
  deploy:
    runs-on: builder
    stage: deploy
    needs:
      - test-build
    steps:
      - name: Deploy to production
        run: ./deploy.sh`

	pipeline, err := workflow.LoadPipeline(yamlContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading pipeline: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== PIPELINE PARSED ===")
	fmt.Printf("Name: %s\n\n", pipeline.Name)

	fmt.Println("Environments:")
	for k, v := range pipeline.Environments {
		fmt.Printf("  %s: %s\n", k, v)
	}

	fmt.Println("\nJobs:")
	for jobName, job := range pipeline.Jobs {
		fmt.Printf("\n  %s:\n", jobName)
		fmt.Printf("    runs-on: %s\n", job.RunsOn)
		fmt.Printf("    stage: %s\n", job.Stage)

		if len(job.Needs) > 0 {
			fmt.Printf("    needs: %v\n", job.Needs)
		}

		fmt.Printf("    steps:\n")
		for i, step := range job.Steps {
			fmt.Printf("      %d. %s\n", i+1, step.Name)
			if step.Run != "" {
				fmt.Printf("         run: %s\n", step.Run)
			}
			if step.Uses != "" {
				fmt.Printf("         uses: %s\n", step.Uses)
			}
			if len(step.With) > 0 {
				fmt.Printf("         with:\n")
				for k, v := range step.With {
					fmt.Printf("           %s: %s\n", k, v)
				}
			}
			if len(step.Env) > 0 {
				fmt.Printf("         env:\n")
				for k, v := range step.Env {
					fmt.Printf("           %s: %s\n", k, v)
				}
			}
		}
	}

	fmt.Println("\n\n=== AS JSON ===")
	jsonBytes, err := json.MarshalIndent(pipeline, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}
