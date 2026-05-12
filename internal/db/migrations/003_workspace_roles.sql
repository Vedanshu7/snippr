-- Add role to members (default 'member' for existing rows).
ALTER TABLE workspace_members
  ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'member';

-- Promote all existing workspace owners to admin.
UPDATE workspace_members wm
SET role = 'admin'
FROM workspaces w
WHERE wm.workspace_id = w.id AND wm.user_id = w.owner_id;

-- Add public/private visibility to workspaces (default private).
ALTER TABLE workspaces
  ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE;
