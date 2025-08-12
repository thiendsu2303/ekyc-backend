CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE ekyc_status AS ENUM ('CREATED','DOC_UPLOADED','SELFIE_UPLOADED','LIVENESS_PENDING','UNDER_REVIEW','APPROVED','REJECTED');
CREATE TYPE result_kind AS ENUM ('OCR','FACE','LIVENESS');
CREATE TYPE decision_status AS ENUM ('APPROVED','REVIEW','REJECTED');
CREATE TYPE artifact_type AS ENUM ('DOC_FRONT','DOC_BACK','PASSPORT','SELFIE','LIVENESS_CLIP','OTHER');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ekyc_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status ekyc_status NOT NULL DEFAULT 'CREATED',
    score INT,
    pending_steps JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE person_pii (
    session_id UUID PRIMARY KEY REFERENCES ekyc_sessions(id) ON DELETE CASCADE,
    full_name TEXT,
    id_number TEXT,
    dob DATE,
    issue_date DATE,
    expiry_date DATE,
    address_text TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ekyc_artifacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES ekyc_sessions(id) ON DELETE CASCADE,
    type artifact_type NOT NULL,
    s3_key TEXT NOT NULL,
    meta_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ekyc_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES ekyc_sessions(id) ON DELETE CASCADE,
    kind result_kind NOT NULL,
    payload_json JSONB,
    quality FLOAT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ekyc_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES ekyc_sessions(id) ON DELETE CASCADE,
    status decision_status NOT NULL,
    score INT,
    reasons_json JSONB,
    decided_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor TEXT,
    action TEXT,
    session_id UUID REFERENCES ekyc_sessions(id) ON DELETE CASCADE,
    meta_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user   ON ekyc_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON ekyc_sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_updated ON ekyc_sessions(updated_at);
CREATE INDEX IF NOT EXISTS idx_artifacts_session ON ekyc_artifacts(session_id);
CREATE INDEX IF NOT EXISTS idx_results_session ON ekyc_results(session_id);
CREATE INDEX IF NOT EXISTS idx_decisions_session ON ekyc_decisions(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_session ON audit_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_ekyc_sessions_updated_at 
    BEFORE UPDATE ON ekyc_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_person_pii_updated_at 
    BEFORE UPDATE ON person_pii 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
