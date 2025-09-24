CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() ,
    expired_access_date TIMESTAMP NOT NULL,
    expired_refresh_date TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    created_by UUID,
    modified_by UUID,
    CONSTRAINT unique_user_session UNIQUE (user_id, id)
    );

ALTER TABLE user_sessions
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN modified_at SET DEFAULT CURRENT_TIMESTAMP;