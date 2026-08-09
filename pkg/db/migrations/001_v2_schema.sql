CREATE TABLE sites (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    retention_days  INTEGER NOT NULL DEFAULT 365 CHECK (retention_days > 0),
    created_at_us   INTEGER NOT NULL,
    disabled_at_us  INTEGER
);

CREATE TABLE site_domains (
    site_id         TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    hostname        TEXT NOT NULL,
    is_primary      INTEGER NOT NULL DEFAULT 0 CHECK (is_primary IN (0, 1)),
    created_at_us   INTEGER NOT NULL,
    PRIMARY KEY (site_id, hostname),
    UNIQUE (hostname)
);

CREATE TABLE ingest_keys (
    id              TEXT PRIMARY KEY,
    site_id         TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    key_hash        TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    created_at_us   INTEGER NOT NULL,
    revoked_at_us   INTEGER
);

CREATE TABLE events (
    seq             INTEGER PRIMARY KEY AUTOINCREMENT,
    id              TEXT NOT NULL UNIQUE,
    event_name      TEXT NOT NULL CHECK (length(event_name) BETWEEN 1 AND 128),
    site_id         TEXT NOT NULL REFERENCES sites(id),
    occurred_at_us  INTEGER NOT NULL,
    received_at_us  INTEGER NOT NULL,
    timestamp       DATETIME NOT NULL,
    url             TEXT NOT NULL DEFAULT '',
    domain          TEXT NOT NULL,
    pathname        TEXT NOT NULL DEFAULT '/',
    referrer        TEXT NOT NULL DEFAULT '',
    referrer_host   TEXT NOT NULL DEFAULT '',
    screen_width    INTEGER NOT NULL DEFAULT 0 CHECK (screen_width BETWEEN 0 AND 100000),
    session_id      TEXT NOT NULL DEFAULT '',
    visitor_id      TEXT NOT NULL DEFAULT '',
    properties      TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(properties)),
    schema_version  INTEGER NOT NULL DEFAULT 1 CHECK (schema_version > 0),
    sdk_version     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_events_site_time
    ON events(site_id, occurred_at_us);
CREATE INDEX idx_events_site_name_time
    ON events(site_id, event_name, occurred_at_us);
CREATE INDEX idx_events_site_session_time
    ON events(site_id, session_id, occurred_at_us);
CREATE INDEX idx_events_site_visitor_time
    ON events(site_id, visitor_id, occurred_at_us);

CREATE TABLE sessions (
    site_id             TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    session_id          TEXT NOT NULL,
    visitor_id          TEXT NOT NULL DEFAULT '',
    started_at_us       INTEGER NOT NULL,
    ended_at_us         INTEGER NOT NULL,
    entry_pathname      TEXT NOT NULL DEFAULT '/',
    exit_pathname       TEXT NOT NULL DEFAULT '/',
    referrer_host       TEXT NOT NULL DEFAULT '',
    pageviews           INTEGER NOT NULL DEFAULT 0,
    event_count         INTEGER NOT NULL DEFAULT 0,
    is_bounce           INTEGER NOT NULL DEFAULT 1 CHECK (is_bounce IN (0, 1)),
    projection_version  INTEGER NOT NULL,
    PRIMARY KEY (site_id, session_id)
);

CREATE INDEX idx_sessions_site_start
    ON sessions(site_id, started_at_us);
CREATE INDEX idx_sessions_site_visitor_start
    ON sessions(site_id, visitor_id, started_at_us);

CREATE TABLE daily_site_metrics (
    site_id           TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    day               TEXT NOT NULL,
    pageviews         INTEGER NOT NULL DEFAULT 0,
    custom_events     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (site_id, day)
);

CREATE TABLE daily_page_metrics (
    site_id           TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    day               TEXT NOT NULL,
    pathname          TEXT NOT NULL,
    pageviews         INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (site_id, day, pathname)
);

CREATE TABLE daily_referrer_visitors (
    site_id           TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    day               TEXT NOT NULL,
    referrer_host     TEXT NOT NULL,
    visitor_id        TEXT NOT NULL,
    PRIMARY KEY (site_id, day, referrer_host, visitor_id)
);

CREATE TABLE projection_checkpoints (
    name              TEXT PRIMARY KEY,
    last_seq          INTEGER NOT NULL DEFAULT 0,
    version           INTEGER NOT NULL,
    updated_at_us     INTEGER NOT NULL
);
