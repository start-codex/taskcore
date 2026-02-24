CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE workspaces (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ
);

CREATE TABLE app_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ
);

CREATE TABLE workspace_members (
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  key TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  UNIQUE (workspace_id, key),
  CHECK (char_length(key) BETWEEN 2 AND 10),
  CHECK (key = UPPER(key))
);

CREATE TABLE project_issue_counters (
  project_id UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
  last_number INT NOT NULL CHECK (last_number >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_members (
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('admin', 'member', 'viewer')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  PRIMARY KEY (project_id, user_id)
);

CREATE TABLE statuses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT NOT NULL CHECK (category IN ('todo', 'doing', 'done')),
  position INT NOT NULL CHECK (position >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  UNIQUE (project_id, name),
  UNIQUE (project_id, position)
);

CREATE TABLE issue_types (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  icon TEXT,
  level INT NOT NULL CHECK (level >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  UNIQUE (project_id, name)
);

CREATE TABLE boards (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('kanban', 'scrum')),
  filter_query TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  UNIQUE (project_id, name)
);

CREATE TABLE board_columns (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  position INT NOT NULL CHECK (position >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  UNIQUE (board_id, name),
  UNIQUE (board_id, position)
);

CREATE TABLE board_column_statuses (
  board_column_id UUID NOT NULL REFERENCES board_columns(id) ON DELETE CASCADE,
  status_id UUID NOT NULL REFERENCES statuses(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (board_column_id, status_id)
);

CREATE TABLE issues (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  number INT NOT NULL CHECK (number > 0),
  issue_type_id UUID NOT NULL REFERENCES issue_types(id),
  status_id UUID NOT NULL REFERENCES statuses(id),
  parent_issue_id UUID REFERENCES issues(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
  assignee_id UUID REFERENCES app_users(id) ON DELETE SET NULL,
  reporter_id UUID NOT NULL REFERENCES app_users(id),
  due_date DATE,
  status_position INT NOT NULL DEFAULT 0 CHECK (status_position >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  archived_at TIMESTAMPTZ,
  UNIQUE (project_id, number)
);

CREATE TABLE issue_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  actor_id UUID NOT NULL REFERENCES app_users(id),
  event_type TEXT NOT NULL CHECK (event_type IN ('created', 'updated', 'moved', 'commented')),
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_statuses_project_position
  ON statuses (project_id, position);

CREATE INDEX idx_issues_project_status_position
  ON issues (project_id, status_id, status_position);
CREATE UNIQUE INDEX uq_issues_active_status_position
  ON issues (project_id, status_id, status_position)
  WHERE archived_at IS NULL;

CREATE INDEX idx_issues_assignee ON issues (assignee_id);
CREATE INDEX idx_issues_parent_issue ON issues (parent_issue_id);
CREATE INDEX idx_issue_events_issue_created_at ON issue_events (issue_id, created_at DESC);
CREATE INDEX idx_board_columns_board_position ON board_columns (board_id, position);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION validate_issue_integrity()
RETURNS trigger AS $$
DECLARE
  v_project_from_type UUID;
  v_project_from_status UUID;
  v_parent_project UUID;
  v_parent_type_id UUID;
  v_parent_level INT;
  v_child_level INT;
  v_cycle_found BOOLEAN;
BEGIN
  SELECT project_id, level INTO v_project_from_type, v_child_level
  FROM issue_types
  WHERE id = NEW.issue_type_id;

  IF v_project_from_type IS NULL OR v_project_from_type <> NEW.project_id THEN
    RAISE EXCEPTION 'issue_type_id % does not belong to project_id %', NEW.issue_type_id, NEW.project_id;
  END IF;

  SELECT project_id INTO v_project_from_status
  FROM statuses
  WHERE id = NEW.status_id;

  IF v_project_from_status IS NULL OR v_project_from_status <> NEW.project_id THEN
    RAISE EXCEPTION 'status_id % does not belong to project_id %', NEW.status_id, NEW.project_id;
  END IF;

  IF NEW.parent_issue_id IS NOT NULL THEN
    SELECT project_id, issue_type_id INTO v_parent_project, v_parent_type_id
    FROM issues
    WHERE id = NEW.parent_issue_id;

    IF v_parent_project IS NULL OR v_parent_project <> NEW.project_id THEN
      RAISE EXCEPTION 'parent_issue_id % must belong to same project_id %', NEW.parent_issue_id, NEW.project_id;
    END IF;

    SELECT level INTO v_parent_level FROM issue_types WHERE id = v_parent_type_id;
    IF v_parent_level IS NOT NULL AND v_child_level <= v_parent_level THEN
      RAISE EXCEPTION 'child issue level (%) must be greater than parent level (%)', v_child_level, v_parent_level;
    END IF;

    WITH RECURSIVE ancestors AS (
      SELECT i.id, i.parent_issue_id
      FROM issues i
      WHERE i.id = NEW.parent_issue_id
      UNION ALL
      SELECT p.id, p.parent_issue_id
      FROM issues p
      JOIN ancestors a ON a.parent_issue_id = p.id
    )
    SELECT EXISTS(
      SELECT 1 FROM ancestors WHERE id = NEW.id
    ) INTO v_cycle_found;

    IF v_cycle_found THEN
      RAISE EXCEPTION 'cycle detected in issue hierarchy for issue %', NEW.id;
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION validate_board_column_status_project()
RETURNS trigger AS $$
DECLARE
  v_board_project UUID;
  v_status_project UUID;
BEGIN
  SELECT b.project_id INTO v_board_project
  FROM boards b
  JOIN board_columns c ON c.board_id = b.id
  WHERE c.id = NEW.board_column_id;

  SELECT project_id INTO v_status_project
  FROM statuses
  WHERE id = NEW.status_id;

  IF v_board_project IS NULL OR v_status_project IS NULL OR v_board_project <> v_status_project THEN
    RAISE EXCEPTION 'board column and status must belong to the same project';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_set_updated_at_workspaces
BEFORE UPDATE ON workspaces
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_app_users
BEFORE UPDATE ON app_users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_workspace_members
BEFORE UPDATE ON workspace_members
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_projects
BEFORE UPDATE ON projects
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_project_issue_counters
BEFORE UPDATE ON project_issue_counters
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_project_members
BEFORE UPDATE ON project_members
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_statuses
BEFORE UPDATE ON statuses
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_issue_types
BEFORE UPDATE ON issue_types
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_boards
BEFORE UPDATE ON boards
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_board_columns
BEFORE UPDATE ON board_columns
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_issues
BEFORE UPDATE ON issues
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_validate_issue_integrity
BEFORE INSERT OR UPDATE ON issues
FOR EACH ROW EXECUTE FUNCTION validate_issue_integrity();

CREATE TRIGGER trg_validate_board_column_status_project
BEFORE INSERT OR UPDATE ON board_column_statuses
FOR EACH ROW EXECUTE FUNCTION validate_board_column_status_project();
