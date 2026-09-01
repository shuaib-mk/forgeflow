CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id uuid PRIMARY KEY,
  email text NOT NULL,
  display_name text NOT NULL CHECK (length(display_name) BETWEEN 1 AND 100),
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_unique ON users (lower(email));

CREATE TABLE organizations (
  id uuid PRIMARY KEY,
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TYPE organization_role AS ENUM ('owner','admin','developer','viewer');
CREATE TABLE memberships (
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role organization_role NOT NULL,
  PRIMARY KEY (organization_id,user_id)
);

CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  token_hash text NOT NULL UNIQUE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);

CREATE TABLE projects (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
  description text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (organization_id,slug)
);
CREATE INDEX projects_organization_created_idx ON projects(organization_id,created_at DESC);

CREATE TABLE repositories (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL,
  local_path text NOT NULL,
  default_branch text NOT NULL DEFAULT 'main',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(project_id,name),
  UNIQUE(local_path)
);

CREATE TABLE branches (
  id uuid PRIMARY KEY,
  repository_id uuid NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
  name text NOT NULL,
  head_sha char(40) NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(repository_id,name)
);

CREATE TABLE commits (
  id uuid PRIMARY KEY,
  repository_id uuid NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
  sha char(40) NOT NULL,
  author_name text NOT NULL,
  author_email text NOT NULL,
  message text NOT NULL,
  committed_at timestamptz NOT NULL,
  UNIQUE(repository_id,sha)
);
CREATE INDEX commits_repository_time_idx ON commits(repository_id,committed_at DESC);

CREATE TYPE task_status AS ENUM ('open','in_progress','done','canceled');
CREATE TABLE tasks (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title text NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
  description text NOT NULL DEFAULT '' CHECK (length(description) <= 10000),
  status task_status NOT NULL DEFAULT 'open',
  assignee_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tasks_project_status_idx ON tasks(project_id,status,created_at DESC);

CREATE TABLE workflows (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  version integer NOT NULL DEFAULT 1 CHECK (version > 0),
  steps jsonb NOT NULL CHECK (jsonb_typeof(steps) = 'array'),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(project_id,name,version)
);

CREATE TYPE run_status AS ENUM ('queued','running','succeeded','failed','canceled','timed_out');
CREATE TABLE workflow_runs (
  id uuid PRIMARY KEY,
  workflow_id uuid NOT NULL REFERENCES workflows(id),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  status run_status NOT NULL DEFAULT 'queued',
  triggered_by uuid NOT NULL REFERENCES users(id),
  attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
  started_at timestamptz,
  finished_at timestamptz,
  error text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (finished_at IS NULL OR started_at IS NOT NULL)
);
CREATE INDEX workflow_runs_project_status_idx ON workflow_runs(project_id,status,created_at DESC);

CREATE TABLE run_logs (
  run_id uuid NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
  step_id text NOT NULL,
  sequence integer NOT NULL CHECK (sequence >= 0),
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(run_id,step_id,sequence)
);

CREATE TABLE artifacts (
  id uuid PRIMARY KEY,
  run_id uuid NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
  step_id text NOT NULL,
  name text NOT NULL,
  path text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
  sha256 char(64) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE test_runs (
  id uuid PRIMARY KEY,
  run_id uuid NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
  framework text NOT NULL,
  total integer NOT NULL CHECK (total >= 0),
  passed integer NOT NULL CHECK (passed >= 0),
  failed integer NOT NULL CHECK (failed >= 0),
  skipped integer NOT NULL CHECK (skipped >= 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (passed + failed + skipped <= total)
);

CREATE TABLE test_results (
  id uuid PRIMARY KEY,
  test_run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
  suite text NOT NULL,
  name text NOT NULL,
  status text NOT NULL CHECK(status IN ('passed','failed','skipped')),
  duration_ms bigint NOT NULL CHECK(duration_ms >= 0),
  message text NOT NULL DEFAULT ''
);

CREATE TABLE audit_events (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  actor_id uuid REFERENCES users(id) ON DELETE SET NULL,
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id uuid NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}',
  request_id text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_events_org_time_idx ON audit_events(organization_id,created_at DESC);

CREATE TABLE notifications (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind text NOT NULL,
  title text NOT NULL,
  body text NOT NULL,
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notifications_user_unread_idx ON notifications(user_id,created_at DESC) WHERE read_at IS NULL;

CREATE TABLE plugins (
  id uuid PRIMARY KEY,
  name text NOT NULL UNIQUE,
  version text NOT NULL,
  description text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE project_metrics (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL,
  value double precision NOT NULL,
  unit text NOT NULL,
  recorded_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX project_metrics_project_name_time_idx ON project_metrics(project_id,name,recorded_at DESC);

