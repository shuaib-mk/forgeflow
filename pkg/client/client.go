package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
)

type Client struct {
	BaseURL, Token string
	HTTP           *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

type Error struct {
	Status                   int
	Code, Message, RequestID string
}

func (e *Error) Error() string {
	return fmt.Sprintf("ForgeFlow API: %s (status %d, request %s)", e.Message, e.Status, e.RequestID)
}

func (c *Client) Do(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		content, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(content)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		request.Header.Set("Authorization", "Bearer "+c.Token)
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		return fmt.Errorf("connect to ForgeFlow at %s: %w", c.BaseURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		var payload struct {
			Error struct{ Code, Message, RequestID string }
		}
		_ = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload)
		return &Error{Status: response.StatusCode, Code: payload.Error.Code, Message: payload.Error.Message, RequestID: payload.Error.RequestID}
	}
	if output != nil && response.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(output); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) Projects(ctx context.Context, organizationID models.ID) (models.Page[models.Project], error) {
	var page models.Page[models.Project]
	path := "/api/v1/projects?organizationId=" + url.QueryEscape(string(organizationID))
	err := c.Do(ctx, http.MethodGet, path, nil, &page)
	return page, err
}
func (c *Client) CreateProject(ctx context.Context, input any) (models.Project, error) {
	var project models.Project
	err := c.Do(ctx, http.MethodPost, "/api/v1/projects", input, &project)
	return project, err
}
func (c *Client) Tasks(ctx context.Context, projectID models.ID) (models.Page[models.Task], error) {
	var page models.Page[models.Task]
	err := c.Do(ctx, http.MethodGet, "/api/v1/projects/"+url.PathEscape(string(projectID))+"/tasks", nil, &page)
	return page, err
}
func (c *Client) CreateTask(ctx context.Context, projectID models.ID, input any) (models.Task, error) {
	var task models.Task
	err := c.Do(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(string(projectID))+"/tasks", input, &task)
	return task, err
}
func (c *Client) RunWorkflow(ctx context.Context, workflowID models.ID) (models.WorkflowRun, error) {
	var run models.WorkflowRun
	err := c.Do(ctx, http.MethodPost, "/api/v1/workflows/"+url.PathEscape(string(workflowID))+"/runs", map[string]any{}, &run)
	return run, err
}
func (c *Client) Run(ctx context.Context, runID models.ID) (models.WorkflowRun, error) {
	var run models.WorkflowRun
	err := c.Do(ctx, http.MethodGet, "/api/v1/runs/"+url.PathEscape(string(runID)), nil, &run)
	return run, err
}
func (c *Client) Logs(ctx context.Context, runID models.ID, after int) ([]string, error) {
	var result struct {
		Items []string `json:"items"`
	}
	err := c.Do(ctx, http.MethodGet, "/api/v1/runs/"+url.PathEscape(string(runID))+"/logs?after="+strconv.Itoa(after), nil, &result)
	return result.Items, err
}
