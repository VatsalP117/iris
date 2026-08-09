ALTER TABLE events ADD COLUMN local_day TEXT NOT NULL DEFAULT '';

UPDATE events
SET local_day = strftime('%Y-%m-%d', occurred_at_us / 1000000, 'unixepoch')
WHERE local_day = '';

CREATE INDEX idx_events_site_name_day
    ON events(site_id, event_name, local_day);

CREATE TABLE daily_visitors (
    site_id           TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    day               TEXT NOT NULL,
    visitor_id        TEXT NOT NULL,
    PRIMARY KEY (site_id, day, visitor_id)
);

CREATE TABLE daily_sessions (
    site_id           TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    day               TEXT NOT NULL,
    session_id        TEXT NOT NULL,
    PRIMARY KEY (site_id, day, session_id)
);
