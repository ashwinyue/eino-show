-- Migration: FAQ Tables
-- Description: Create FAQ entries table and related enhancements
-- Date: 2026-01-17

DO $$ BEGIN RAISE NOTICE '[Migration FAQ] Starting FAQ tables setup...'; END $$;

-- ============================================================================
-- Section 1: FAQ Entries Table
-- ============================================================================

DO $$ BEGIN RAISE NOTICE '[Migration FAQ] Creating table: faq_entries'; END $$;
CREATE TABLE IF NOT EXISTS faq_entries (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    knowledge_id VARCHAR(36),
    chunk_id VARCHAR(36),
    tag_id BIGINT,
    standard_question TEXT NOT NULL,
    similar_questions TEXT DEFAULT '[]',
    negative_questions TEXT DEFAULT '[]',
    answers TEXT NOT NULL DEFAULT '[]',
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    is_recommended BOOLEAN NOT NULL DEFAULT false,
    index_mode VARCHAR(50) NOT NULL DEFAULT 'hybrid',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Add indexes for faq_entries
CREATE INDEX IF NOT EXISTS idx_faq_entries_tenant_id ON faq_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_faq_entries_knowledge_base_id ON faq_entries(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_faq_entries_knowledge_id ON faq_entries(knowledge_id);
CREATE INDEX IF NOT EXISTS idx_faq_entries_tag_id ON faq_entries(tag_id);
CREATE INDEX IF NOT EXISTS idx_faq_entries_is_enabled ON faq_entries(is_enabled);
CREATE INDEX IF NOT EXISTS idx_faq_entries_deleted_at ON faq_entries(deleted_at);
CREATE INDEX IF NOT EXISTS idx_faq_entries_updated_at ON faq_entries(updated_at DESC);

-- Add unique constraint on standard_question within knowledge_base
CREATE UNIQUE INDEX IF NOT EXISTS idx_faq_entries_kb_question 
    ON faq_entries(knowledge_base_id, standard_question) 
    WHERE deleted_at IS NULL;

-- Add foreign key constraints
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_faq_entries_tenant') THEN
        ALTER TABLE faq_entries ADD CONSTRAINT fk_faq_entries_tenant
            FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
        RAISE NOTICE '[Migration FAQ] Added foreign key constraint fk_faq_entries_tenant';
    END IF;
EXCEPTION
    WHEN undefined_table THEN
        RAISE NOTICE '[Migration FAQ] Skipping fk_faq_entries_tenant - tenants table not found';
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_faq_entries_kb') THEN
        ALTER TABLE faq_entries ADD CONSTRAINT fk_faq_entries_kb
            FOREIGN KEY (knowledge_base_id) REFERENCES knowledge_bases(id) ON DELETE CASCADE;
        RAISE NOTICE '[Migration FAQ] Added foreign key constraint fk_faq_entries_kb';
    END IF;
EXCEPTION
    WHEN undefined_table THEN
        RAISE NOTICE '[Migration FAQ] Skipping fk_faq_entries_kb - knowledge_bases table not found';
END $$;

-- Comments
COMMENT ON TABLE faq_entries IS 'FAQ entries for knowledge bases';
COMMENT ON COLUMN faq_entries.id IS 'Unique identifier (auto-increment)';
COMMENT ON COLUMN faq_entries.tenant_id IS 'Tenant ID that owns this FAQ';
COMMENT ON COLUMN faq_entries.knowledge_base_id IS 'Associated knowledge base ID';
COMMENT ON COLUMN faq_entries.knowledge_id IS 'Associated knowledge ID (optional)';
COMMENT ON COLUMN faq_entries.chunk_id IS 'Associated chunk ID (optional)';
COMMENT ON COLUMN faq_entries.tag_id IS 'Tag ID for categorization';
COMMENT ON COLUMN faq_entries.standard_question IS 'The standard question text';
COMMENT ON COLUMN faq_entries.similar_questions IS 'JSON array of similar question variations';
COMMENT ON COLUMN faq_entries.negative_questions IS 'JSON array of negative examples';
COMMENT ON COLUMN faq_entries.answers IS 'JSON array of answer texts';
COMMENT ON COLUMN faq_entries.is_enabled IS 'Whether this FAQ entry is active';
COMMENT ON COLUMN faq_entries.is_recommended IS 'Whether this FAQ is recommended/featured';
COMMENT ON COLUMN faq_entries.index_mode IS 'Index mode: hybrid, vector, keyword';

-- ============================================================================
-- Section 2: Add seq_id column to knowledge_tags for FAQ compatibility
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'knowledge_tags' AND column_name = 'seq_id'
    ) THEN
        ALTER TABLE knowledge_tags ADD COLUMN seq_id BIGSERIAL;
        RAISE NOTICE '[Migration FAQ] Added seq_id column to knowledge_tags';
    END IF;
EXCEPTION
    WHEN undefined_table THEN
        RAISE NOTICE '[Migration FAQ] Skipping knowledge_tags enhancement - table not found';
END $$;

-- ============================================================================
-- Section 3: Add last_faq_import_result column to knowledge_bases
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'knowledge_bases' AND column_name = 'last_faq_import_result'
    ) THEN
        ALTER TABLE knowledge_bases ADD COLUMN last_faq_import_result JSONB DEFAULT NULL;
        RAISE NOTICE '[Migration FAQ] Added last_faq_import_result column to knowledge_bases';
    END IF;
EXCEPTION
    WHEN undefined_table THEN
        RAISE NOTICE '[Migration FAQ] Skipping knowledge_bases enhancement - table not found';
END $$;

-- ============================================================================
-- Section 4: Workflow Checkpoints Table (if not exists)
-- ============================================================================

DO $$ BEGIN RAISE NOTICE '[Migration FAQ] Checking workflow_checkpoints table...'; END $$;
CREATE TABLE IF NOT EXISTS workflow_checkpoints (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(36) NOT NULL,
    checkpoint_type VARCHAR(50) NOT NULL,
    state JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_workflow_checkpoints_session ON workflow_checkpoints(session_id);
CREATE INDEX IF NOT EXISTS idx_workflow_checkpoints_type ON workflow_checkpoints(checkpoint_type);

-- ============================================================================
-- Section 5: Summaries Table (for context compression)
-- ============================================================================

DO $$ BEGIN RAISE NOTICE '[Migration FAQ] Checking summaries table...'; END $$;
CREATE TABLE IF NOT EXISTS summaries (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    message_ids TEXT NOT NULL DEFAULT '[]',
    token_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_summaries_session ON summaries(session_id);

DO $$ BEGIN RAISE NOTICE '[Migration FAQ] FAQ tables setup completed successfully!'; END $$;
