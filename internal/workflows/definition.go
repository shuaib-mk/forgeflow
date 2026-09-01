package workflows

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
	"gopkg.in/yaml.v3"
)

const maxDefinitionBytes = 256 << 10

type Definition struct {
	Name  string                `yaml:"name" json:"name"`
	Steps []models.WorkflowStep `yaml:"steps" json:"steps"`
}

func Parse(reader io.Reader) (Definition, error) {
	decoder := yaml.NewDecoder(io.LimitReader(reader, maxDefinitionBytes+1))
	decoder.KnownFields(true)
	var definition Definition
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, fmt.Errorf("decode workflow: %w", err)
	}
	if err := Validate(&definition); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func Validate(definition *Definition) error {
	if strings.TrimSpace(definition.Name) == "" || len(definition.Name) > 100 {
		return errors.New("workflow name must contain 1 to 100 characters")
	}
	if len(definition.Steps) == 0 || len(definition.Steps) > 100 {
		return errors.New("workflow must contain between 1 and 100 steps")
	}
	byID := make(map[string]int, len(definition.Steps))
	for index := range definition.Steps {
		step := &definition.Steps[index]
		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("step %d: id is required", index+1)
		}
		if _, exists := byID[step.ID]; exists {
			return fmt.Errorf("step %q: duplicate id", step.ID)
		}
		if strings.TrimSpace(step.Name) == "" {
			step.Name = step.ID
		}
		if strings.TrimSpace(step.Command) == "" {
			return fmt.Errorf("step %q: command is required", step.ID)
		}
		if strings.ContainsRune(step.Command, '\x00') {
			return fmt.Errorf("step %q: command contains a null byte", step.ID)
		}
		if step.Retries < 0 || step.Retries > 10 {
			return fmt.Errorf("step %q: retries must be between 0 and 10", step.ID)
		}
		step.Timeout = 15 * time.Minute
		if step.TimeoutText != "" {
			timeout, err := time.ParseDuration(step.TimeoutText)
			if err != nil || timeout < time.Second || timeout > 24*time.Hour {
				return fmt.Errorf("step %q: timeout must be between 1s and 24h", step.ID)
			}
			step.Timeout = timeout
		}
		byID[step.ID] = index
	}
	for _, step := range definition.Steps {
		seen := map[string]bool{}
		for _, dependency := range step.DependsOn {
			if dependency == step.ID {
				return fmt.Errorf("step %q: cannot depend on itself", step.ID)
			}
			if _, exists := byID[dependency]; !exists {
				return fmt.Errorf("step %q: dependency %q does not exist", step.ID, dependency)
			}
			if seen[dependency] {
				return fmt.Errorf("step %q: duplicate dependency %q", step.ID, dependency)
			}
			seen[dependency] = true
		}
	}
	if cycle := findCycle(definition.Steps); len(cycle) > 0 {
		return fmt.Errorf("workflow dependency cycle: %s", strings.Join(cycle, " -> "))
	}
	return nil
}

func findCycle(steps []models.WorkflowStep) []string {
	dependencies := make(map[string][]string, len(steps))
	for _, step := range steps {
		dependencies[step.ID] = step.DependsOn
	}
	state := map[string]int{}
	stack := []string{}
	var visit func(string) []string
	visit = func(id string) []string {
		state[id] = 1
		stack = append(stack, id)
		for _, dependency := range dependencies[id] {
			if state[dependency] == 1 {
				start := 0
				for stack[start] != dependency {
					start++
				}
				return append(append([]string(nil), stack[start:]...), dependency)
			}
			if state[dependency] == 0 {
				if cycle := visit(dependency); cycle != nil {
					return cycle
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[id] = 2
		return nil
	}
	for id := range dependencies {
		if state[id] == 0 {
			if cycle := visit(id); cycle != nil {
				return cycle
			}
		}
	}
	return nil
}
