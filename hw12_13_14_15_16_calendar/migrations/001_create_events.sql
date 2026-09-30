CREATE TABLE IF NOT EXISTS events (
                                      id            BIGINT PRIMARY KEY,
                                      title         TEXT NOT NULL,
                                      start_at      TIMESTAMP NOT NULL,
                                      duration      BIGINT NOT NULL,
                                      user_id       BIGINT NOT NULL,
                                      notify_before BIGINT NOT NULL DEFAULT 0,
                                      notified      BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_events_start_at ON events (start_at);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events (user_id);

CREATE TABLE IF NOT EXISTS notifications (
                                             id       BIGSERIAL PRIMARY KEY,
                                             event_id BIGINT NOT NULL,
                                             title    TEXT NOT NULL,
                                             date     TIMESTAMP NOT NULL,
                                             user_id  BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
