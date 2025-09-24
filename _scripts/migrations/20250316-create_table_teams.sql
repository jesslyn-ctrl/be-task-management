CREATE TABLE IF NOT EXISTS teams (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(255) NOT NULL,
   created_at TIMESTAMP NOT NULL DEFAULT NOW(),
   modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by NOT NULL,
   modified_by NOT NULL
);
