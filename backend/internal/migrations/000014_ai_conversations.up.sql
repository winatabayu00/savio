CREATE TABLE ai_conversations (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(120) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ai_conversations_owner_updated_idx
    ON ai_conversations(workspace_id, user_id, updated_at DESC);

CREATE TABLE ai_messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    response JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ai_messages_role_ck CHECK (role IN ('USER', 'ASSISTANT')),
    CONSTRAINT ai_messages_content_ck CHECK (char_length(content) BETWEEN 1 AND 10000)
);

CREATE INDEX ai_messages_conversation_created_idx
    ON ai_messages(conversation_id, created_at, id);
