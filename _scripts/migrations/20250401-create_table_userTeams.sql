CREATE TABLE IF NOT EXISTS user_teams (
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
);

ALTER TABLE user_teams
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

ALTER TABLE user_teams
    ADD CONSTRAINT unique_user_team UNIQUE (user_id, team_id);