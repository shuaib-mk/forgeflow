package models

import "time"

type ID string

type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleViewer    Role = "viewer"
)

func (r Role) Valid() bool {
	return r == RoleOwner || r == RoleAdmin || r == RoleDeveloper || r == RoleViewer
}

type User struct {
	ID           ID        `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Organization struct {
	ID        ID        `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
}

type Membership struct {
	OrganizationID ID   `json:"organizationId"`
	UserID         ID   `json:"userId"`
	Role           Role `json:"role"`
}

type Project struct {
	ID             ID        `json:"id"`
	OrganizationID ID        `json:"organizationId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	CreatedBy      ID        `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Repository struct {
	ID            ID        `json:"id"`
	ProjectID     ID        `json:"projectId"`
	Name          string    `json:"name"`
	LocalPath     string    `json:"localPath"`
	DefaultBranch string    `json:"defaultBranch"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Branch struct {
	ID           ID        `json:"id"`
	RepositoryID ID        `json:"repositoryId"`
	Name         string    `json:"name"`
	HeadSHA      string    `json:"headSha"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Commit struct {
	ID           ID        `json:"id"`
	RepositoryID ID        `json:"repositoryId"`
	SHA          string    `json:"sha"`
	AuthorName   string    `json:"authorName"`
	AuthorEmail  string    `json:"authorEmail"`
	Message      string    `json:"message"`
	CommittedAt  time.Time `json:"committedAt"`
}

type TaskStatus string

const (
	TaskOpen       TaskStatus = "open"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskCanceled   TaskStatus = "canceled"
)

type Task struct {
	ID          ID         `json:"id"`
	ProjectID   ID         `json:"projectId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	AssigneeID  *ID        `json:"assigneeId,omitempty"`
	CreatedBy   ID         `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type RunStatus string

const (
	RunQueued    RunStatus = "queued"
	RunRunning   RunStatus = "running"
	RunSucceeded RunStatus = "succeeded"
	RunFailed    RunStatus = "failed"
	RunCanceled  RunStatus = "canceled"
	RunTimedOut  RunStatus = "timed_out"
)

type Workflow struct {
	ID        ID             `json:"id"`
	ProjectID ID             `json:"projectId"`
	Name      string         `json:"name"`
	Version   int            `json:"version"`
	Steps     []WorkflowStep `json:"steps"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type WorkflowStep struct {
	ID             string        `json:"id" yaml:"id"`
	Name           string        `json:"name" yaml:"name"`
	Command        string        `json:"command" yaml:"command"`
	Args           []string      `json:"args,omitempty" yaml:"args,omitempty"`
	DependsOn      []string      `json:"dependsOn,omitempty" yaml:"depends_on,omitempty"`
	Timeout        time.Duration `json:"timeout" yaml:"-"`
	TimeoutText    string        `json:"-" yaml:"timeout,omitempty"`
	Retries        int           `json:"retries" yaml:"retries,omitempty"`
	ContinueOnFail bool          `json:"continueOnFail" yaml:"continue_on_failure,omitempty"`
}

type WorkflowRun struct {
	ID          ID         `json:"id"`
	WorkflowID  ID         `json:"workflowId"`
	ProjectID   ID         `json:"projectId"`
	Status      RunStatus  `json:"status"`
	TriggeredBy ID         `json:"triggeredBy"`
	Attempt     int        `json:"attempt"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type Artifact struct {
	ID        ID        `json:"id"`
	RunID     ID        `json:"runId"`
	StepID    string    `json:"stepId"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	SizeBytes int64     `json:"sizeBytes"`
	SHA256    string    `json:"sha256"`
	CreatedAt time.Time `json:"createdAt"`
}

type TestRun struct {
	ID        ID        `json:"id"`
	RunID     ID        `json:"runId"`
	Framework string    `json:"framework"`
	Total     int       `json:"total"`
	Passed    int       `json:"passed"`
	Failed    int       `json:"failed"`
	Skipped   int       `json:"skipped"`
	CreatedAt time.Time `json:"createdAt"`
}

type TestResult struct {
	ID         ID            `json:"id"`
	TestRunID  ID            `json:"testRunId"`
	Suite      string        `json:"suite"`
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	DurationMS int64         `json:"durationMs"`
	Message    string        `json:"message,omitempty"`
}

type AuditEvent struct {
	ID             ID             `json:"id"`
	OrganizationID ID             `json:"organizationId"`
	ActorID        *ID            `json:"actorId,omitempty"`
	Action         string         `json:"action"`
	ResourceType   string         `json:"resourceType"`
	ResourceID     ID             `json:"resourceId"`
	Metadata       map[string]any `json:"metadata"`
	RequestID      string         `json:"requestId"`
	CreatedAt      time.Time      `json:"createdAt"`
}

type Notification struct {
	ID        ID         `json:"id"`
	UserID    ID         `json:"userId"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ReadAt    *time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type Plugin struct {
	ID          ID        `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ProjectMetric struct {
	ID        ID        `json:"id"`
	ProjectID ID        `json:"projectId"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	RecordedAt time.Time `json:"recordedAt"`
}

type Page[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

