-- Reference copy of the schema. The authoritative migrations that actually run
-- are embedded in internal/db/migrations/ and applied automatically on startup.

-- 001_init.sql
CREATE TABLE IF NOT EXISTS users (
  id            BIGSERIAL PRIMARY KEY,
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS snippets (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title         TEXT NOT NULL,
  content       TEXT NOT NULL,
  language      TEXT DEFAULT 'text',
  is_public     BOOLEAN DEFAULT FALSE,
  share_slug    TEXT UNIQUE,
  workspace_id  BIGINT REFERENCES workspaces(id) ON DELETE SET NULL,
  search_vector TSVECTOR,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
  id         BIGSERIAL PRIMARY KEY,
  snippet_id BIGINT NOT NULL REFERENCES snippets(id) ON DELETE CASCADE,
  name       TEXT NOT NULL
);

-- 002_workspaces.sql
CREATE TABLE IF NOT EXISTS workspaces (
  id          BIGSERIAL PRIMARY KEY,
  owner_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  invite_code TEXT UNIQUE NOT NULL,
  is_public   BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workspace_members (
  workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role         TEXT NOT NULL DEFAULT 'member', -- 'admin' | 'member'
  joined_at    TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (workspace_id, user_id)
);
