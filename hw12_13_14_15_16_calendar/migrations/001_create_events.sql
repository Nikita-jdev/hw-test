CREATE TABLE IF NOT EXISTS events (
                                      id       BIGINT PRIMARY KEY,
                                      title    TEXT NOT NULL,
                                      start_at TIMESTAMP NOT NULL,
                                      duration BIGINT NOT NULL,
                                      user_id  BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_start_at ON events (start_at);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events (user_id);
