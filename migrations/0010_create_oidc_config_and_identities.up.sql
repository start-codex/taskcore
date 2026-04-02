CREATE TABLE oidc_providers (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    slug          TEXT        NOT NULL UNIQUE,
    issuer_url    TEXT        NOT NULL,
    client_id     TEXT        NOT NULL,
    client_secret TEXT        NOT NULL,
    redirect_uri  TEXT        NOT NULL,
    scopes        TEXT        NOT NULL DEFAULT 'openid email profile',
    auto_register BOOLEAN    NOT NULL DEFAULT false,
    enabled       BOOLEAN    NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_identities (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    provider_id UUID        NOT NULL REFERENCES oidc_providers(id) ON DELETE CASCADE,
    subject     TEXT        NOT NULL,
    email       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, subject),
    UNIQUE (user_id, provider_id)
);

CREATE INDEX idx_user_identities_user ON user_identities(user_id);
CREATE INDEX idx_user_identities_provider_subject ON user_identities(provider_id, subject);
