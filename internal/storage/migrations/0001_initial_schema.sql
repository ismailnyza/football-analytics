PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS worlds (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    seed INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS branches (
    id INTEGER PRIMARY KEY,
    world_id INTEGER NOT NULL,
    parent_branch_id INTEGER,
    name TEXT NOT NULL,
    created_from_match_id INTEGER,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (world_id) REFERENCES worlds(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_branch_id) REFERENCES branches(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS clubs (
    id INTEGER PRIMARY KEY,
    branch_id INTEGER NOT NULL,
    ext_key TEXT,
    name TEXT NOT NULL,
    short_name TEXT NOT NULL,
    country TEXT,
    division TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_clubs_branch_name
    ON clubs(branch_id, name);

CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY,
    branch_id INTEGER NOT NULL,
    ext_key TEXT,
    club_id INTEGER,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    preferred_name TEXT,
    date_of_birth TEXT,
    nationality TEXT,
    primary_position TEXT NOT NULL,
    secondary_positions TEXT NOT NULL DEFAULT '',
    overall INTEGER NOT NULL DEFAULT 0,
    potential INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    FOREIGN KEY (club_id) REFERENCES clubs(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_players_branch_club
    ON players(branch_id, club_id);

CREATE TABLE IF NOT EXISTS seasons (
    id INTEGER PRIMARY KEY,
    branch_id INTEGER NOT NULL,
    label TEXT NOT NULL,
    start_year INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_seasons_branch_label
    ON seasons(branch_id, label);

CREATE TABLE IF NOT EXISTS fixtures (
    id INTEGER PRIMARY KEY,
    season_id INTEGER NOT NULL,
    matchday INTEGER NOT NULL,
    home_club_id INTEGER NOT NULL,
    away_club_id INTEGER NOT NULL,
    scheduled_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (season_id) REFERENCES seasons(id) ON DELETE CASCADE,
    FOREIGN KEY (home_club_id) REFERENCES clubs(id) ON DELETE CASCADE,
    FOREIGN KEY (away_club_id) REFERENCES clubs(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fixtures_season_pair
    ON fixtures(season_id, matchday, home_club_id, away_club_id);

CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY,
    fixture_id INTEGER NOT NULL,
    branch_id INTEGER NOT NULL,
    home_goals INTEGER NOT NULL DEFAULT 0,
    away_goals INTEGER NOT NULL DEFAULT 0,
    tick_count INTEGER NOT NULL DEFAULT 900,
    seed INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed',
    simulated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (fixture_id) REFERENCES fixtures(id) ON DELETE CASCADE,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_matches_fixture_branch
    ON matches(fixture_id, branch_id);

CREATE TABLE IF NOT EXISTS lineup_entries (
    id INTEGER PRIMARY KEY,
    match_id INTEGER NOT NULL,
    club_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    slot_code TEXT NOT NULL,
    is_starter INTEGER NOT NULL DEFAULT 1,
    minute_on INTEGER NOT NULL DEFAULT 0,
    minute_off INTEGER,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (club_id) REFERENCES clubs(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lineup_entries_match_player
    ON lineup_entries(match_id, player_id);

CREATE TABLE IF NOT EXISTS match_events (
    id INTEGER PRIMARY KEY,
    match_id INTEGER NOT NULL,
    tick INTEGER NOT NULL,
    minute INTEGER NOT NULL,
    team_club_id INTEGER,
    player_id INTEGER,
    event_type TEXT NOT NULL,
    pitch_zone TEXT,
    payload_json TEXT NOT NULL DEFAULT '{}',
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (team_club_id) REFERENCES clubs(id) ON DELETE SET NULL,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_match_events_match_tick
    ON match_events(match_id, tick);

CREATE TABLE IF NOT EXISTS player_match_stats (
    id INTEGER PRIMARY KEY,
    match_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    club_id INTEGER NOT NULL,
    minutes_played INTEGER NOT NULL DEFAULT 0,
    goals INTEGER NOT NULL DEFAULT 0,
    assists INTEGER NOT NULL DEFAULT 0,
    shots INTEGER NOT NULL DEFAULT 0,
    tackles INTEGER NOT NULL DEFAULT 0,
    saves INTEGER NOT NULL DEFAULT 0,
    rating REAL NOT NULL DEFAULT 0,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (club_id) REFERENCES clubs(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_player_match_stats_unique
    ON player_match_stats(match_id, player_id);

CREATE TABLE IF NOT EXISTS raw_sources (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS raw_payloads (
    id INTEGER PRIMARY KEY,
    source_id INTEGER NOT NULL,
    external_id TEXT NOT NULL,
    fetched_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload_json TEXT NOT NULL,
    checksum TEXT,
    FOREIGN KEY (source_id) REFERENCES raw_sources(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_raw_payloads_source_external
    ON raw_payloads(source_id, external_id, fetched_at);
