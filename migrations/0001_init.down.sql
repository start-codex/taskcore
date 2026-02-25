DROP TRIGGER IF EXISTS trg_validate_board_column_status_project ON board_column_statuses;
DROP TRIGGER IF EXISTS trg_validate_issue_integrity ON issues;

DROP TRIGGER IF EXISTS trg_set_updated_at_issues ON issues;
DROP TRIGGER IF EXISTS trg_set_updated_at_board_columns ON board_columns;
DROP TRIGGER IF EXISTS trg_set_updated_at_boards ON boards;
DROP TRIGGER IF EXISTS trg_set_updated_at_issue_types ON issue_types;
DROP TRIGGER IF EXISTS trg_set_updated_at_statuses ON statuses;
DROP TRIGGER IF EXISTS trg_set_updated_at_project_members ON project_members;
DROP TRIGGER IF EXISTS trg_set_updated_at_project_issue_counters ON project_issue_counters;
DROP TRIGGER IF EXISTS trg_set_updated_at_projects ON projects;
DROP TRIGGER IF EXISTS trg_set_updated_at_workspace_members ON workspace_members;
DROP TRIGGER IF EXISTS trg_set_updated_at_app_users ON app_users;
DROP TRIGGER IF EXISTS trg_set_updated_at_workspaces ON workspaces;

DROP FUNCTION IF EXISTS validate_board_column_status_project();
DROP FUNCTION IF EXISTS validate_issue_integrity();
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS issue_events;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS board_column_statuses;
DROP TABLE IF EXISTS board_columns;
DROP TABLE IF EXISTS boards;
DROP TABLE IF EXISTS issue_types;
DROP TABLE IF EXISTS statuses;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS project_issue_counters;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS app_users;
DROP TABLE IF EXISTS workspaces;

DROP EXTENSION IF EXISTS pgcrypto;
