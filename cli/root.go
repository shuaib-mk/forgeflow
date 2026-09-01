package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/internal/projects"
	"github.com/forgeflow/forgeflow/internal/tasks"
	"github.com/forgeflow/forgeflow/internal/workflows"
	"github.com/forgeflow/forgeflow/pkg/client"
	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/spf13/cobra"
)

func NewRoot(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{Use: "forgeflow", Short: "Manage local engineering workflows", SilenceUsage: true, SilenceErrors: true}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(initCommand(), projectCommand(), taskCommand(), workflowCommand(), runCommand(), doctorCommand(), configCommand())
	return root
}

func Execute() error { return NewRoot(os.Stdout, os.Stderr).Execute() }
func apiClient() (Config, *client.Client, error) {
	config, err := loadConfig()
	if err != nil {
		return Config{}, nil, err
	}
	return config, client.New(config.APIURL, config.Token), nil
}

func initCommand() *cobra.Command {
	return &cobra.Command{Use: "init", Short: "Create a starter workflow in the current repository", RunE: func(cmd *cobra.Command, _ []string) error {
		path := filepath.Join(".forgeflow", "workflow.yaml")
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		content := []byte("name: test-and-build\nsteps:\n  - id: test\n    command: go\n    args: [test, ./...]\n    timeout: 10m\n  - id: build\n    command: go\n    args: [build, ./...]\n    depends_on: [test]\n")
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Created", path)
		return nil
	}}
}

func projectCommand() *cobra.Command {
	command := &cobra.Command{Use: "project", Short: "Manage projects"}
	var name, slug, description, organization string
	create := &cobra.Command{Use: "create", Short: "Create a project", RunE: func(cmd *cobra.Command, _ []string) error {
		_, api, err := apiClient()
		if err != nil {
			return err
		}
		project, err := api.CreateProject(cmd.Context(), projects.CreateInput{OrganizationID: models.ID(organization), Name: name, Slug: slug, Description: description})
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), project)
	}}
	create.Flags().StringVar(&organization, "organization", "", "organization ID")
	create.Flags().StringVar(&name, "name", "", "project name")
	create.Flags().StringVar(&slug, "slug", "", "project slug")
	create.Flags().StringVar(&description, "description", "", "project description")
	_ = create.MarkFlagRequired("organization")
	_ = create.MarkFlagRequired("name")
	_ = create.MarkFlagRequired("slug")
	list := &cobra.Command{Use: "list", Short: "List projects", RunE: func(cmd *cobra.Command, _ []string) error {
		config, api, err := apiClient()
		if err != nil {
			return err
		}
		id := organization
		if id == "" {
			id = config.OrganizationID
		}
		page, err := api.Projects(cmd.Context(), models.ID(id))
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), page)
	}}
	list.Flags().StringVar(&organization, "organization", "", "organization ID")
	command.AddCommand(create, list)
	return command
}

func taskCommand() *cobra.Command {
	command := &cobra.Command{Use: "task", Short: "Manage project tasks"}
	var projectID, title, description string
	create := &cobra.Command{Use: "create", Short: "Create a task", RunE: func(cmd *cobra.Command, _ []string) error {
		_, api, err := apiClient()
		if err != nil {
			return err
		}
		task, err := api.CreateTask(cmd.Context(), models.ID(projectID), tasks.CreateInput{Title: title, Description: description})
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), task)
	}}
	create.Flags().StringVar(&projectID, "project", "", "project ID")
	create.Flags().StringVar(&title, "title", "", "task title")
	create.Flags().StringVar(&description, "description", "", "task description")
	_ = create.MarkFlagRequired("project")
	_ = create.MarkFlagRequired("title")
	list := &cobra.Command{Use: "list", Short: "List tasks", RunE: func(cmd *cobra.Command, _ []string) error {
		_, api, err := apiClient()
		if err != nil {
			return err
		}
		page, err := api.Tasks(cmd.Context(), models.ID(projectID))
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), page)
	}}
	list.Flags().StringVar(&projectID, "project", "", "project ID")
	_ = list.MarkFlagRequired("project")
	command.AddCommand(create, list)
	return command
}

func workflowCommand() *cobra.Command {
	command := &cobra.Command{Use: "workflow", Short: "Validate and run workflows"}
	validate := &cobra.Command{Use: "validate <file>", Args: cobra.ExactArgs(1), Short: "Validate a workflow definition", RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()
		definition, err := workflows.Parse(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q is valid (%d steps)\n", definition.Name, len(definition.Steps))
		return nil
	}}
	var workflowID string
	run := &cobra.Command{Use: "run", Short: "Queue a workflow run", RunE: func(cmd *cobra.Command, _ []string) error {
		_, api, err := apiClient()
		if err != nil {
			return err
		}
		result, err := api.RunWorkflow(cmd.Context(), models.ID(workflowID))
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), result)
	}}
	run.Flags().StringVar(&workflowID, "workflow", "", "workflow ID")
	_ = run.MarkFlagRequired("workflow")
	command.AddCommand(validate, run)
	return command
}

func runCommand() *cobra.Command {
	command := &cobra.Command{Use: "run", Short: "Inspect workflow runs"}
	var runID string
	status := &cobra.Command{Use: "status", Short: "Show run status", RunE: func(cmd *cobra.Command, _ []string) error {
		_, api, err := apiClient()
		if err != nil {
			return err
		}
		run, err := api.Run(cmd.Context(), models.ID(runID))
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), run)
	}}
	logs := &cobra.Command{Use: "logs", Short: "Show run logs", RunE: func(cmd *cobra.Command, _ []string) error {
		_, api, err := apiClient()
		if err != nil {
			return err
		}
		lines, err := api.Logs(cmd.Context(), models.ID(runID), -1)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), strings.Join(lines, ""))
		return nil
	}}
	for _, child := range []*cobra.Command{status, logs} {
		child.Flags().StringVar(&runID, "run", "", "run ID")
		_ = child.MarkFlagRequired("run")
	}
	command.AddCommand(status, logs)
	return command
}

func doctorCommand() *cobra.Command {
	return &cobra.Command{Use: "doctor", Short: "Check local configuration and API connectivity", RunE: func(cmd *cobra.Command, _ []string) error {
		config, _, err := apiClient()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(config.APIURL, "/")+"/health", nil)
		if err != nil {
			return err
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return fmt.Errorf("API unavailable: %w", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("API health returned %s", response.Status)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "ForgeFlow doctor\n  API: ok (%s)\n  CLI: %s/%s\n", config.APIURL, runtime.GOOS, runtime.GOARCH)
		return nil
	}}
}

func configCommand() *cobra.Command {
	command := &cobra.Command{Use: "config", Short: "Manage CLI configuration"}
	show := &cobra.Command{Use: "show", Short: "Show configuration with the token redacted", RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}
		if config.Token != "" {
			config.Token = "[REDACTED]"
		}
		return printJSON(cmd.OutOrStdout(), config)
	}}
	var apiURL, token, organization string
	set := &cobra.Command{Use: "set", Short: "Update CLI configuration", RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("api-url") {
			config.APIURL = apiURL
		}
		if cmd.Flags().Changed("token") {
			config.Token = token
		}
		if cmd.Flags().Changed("organization") {
			config.OrganizationID = organization
		}
		return saveConfig(config)
	}}
	set.Flags().StringVar(&apiURL, "api-url", "", "API base URL")
	set.Flags().StringVar(&token, "token", "", "session token")
	set.Flags().StringVar(&organization, "organization", "", "default organization ID")
	command.AddCommand(show, set)
	return command
}

func printJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
