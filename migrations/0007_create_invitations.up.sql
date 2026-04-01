CREATE TABLE invitations (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email        TEXT        NOT NULL,
    role         TEXT        NOT NULL CHECK (role IN ('admin', 'member')),
    invited_by   UUID        NOT NULL REFERENCES app_users(id),
    token_hash   TEXT        NOT NULL UNIQUE,
    status       TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),
    expires_at   TIMESTAMPTZ NOT NULL,
    accepted_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_invitations_workspace_email_pending
    ON invitations (workspace_id, email) WHERE status = 'pending';
CREATE INDEX idx_invitations_token ON invitations(token_hash);
CREATE INDEX idx_invitations_workspace ON invitations(workspace_id);
