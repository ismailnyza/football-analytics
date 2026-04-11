package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/storage/migrations"
)

// Store is the SQLite-backed implementation of the repository contracts.
type Store struct {
	db beginTxer
}

// NewStore wraps a standard library SQL database handle.
func NewStore(db *sql.DB) *Store {
	return &Store{db: sqlDB{DB: db}}
}

func newStoreFromBeginTxer(db beginTxer) *Store {
	return &Store{db: db}
}

// Migrate applies embedded SQL migrations in filename order.
func (s *Store) Migrate(ctx context.Context) error {
	migs, err := migrations.Load()
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const bootstrap = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	if _, err := tx.ExecContext(ctx, bootstrap); err != nil {
		return fmt.Errorf("bootstrap schema_migrations: %w", err)
	}

	for _, migration := range migs {
		var applied int
		if err := tx.QueryRowContext(
			ctx,
			`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`,
			migration.Name,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", migration.Name, err)
		}
		if applied > 0 {
			continue
		}

		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Name, err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO schema_migrations(version) VALUES (?)`,
			migration.Name,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", migration.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

func (s *Store) CreateWorld(ctx context.Context, world domain.World) (domain.World, error) {
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO worlds(name, seed) VALUES (?, ?)`,
		world.Name,
		world.Seed,
	)
	if err != nil {
		return domain.World{}, fmt.Errorf("insert world: %w", err)
	}
	world.ID, err = res.LastInsertId()
	if err != nil {
		return domain.World{}, fmt.Errorf("world last insert id: %w", err)
	}
	return world, nil
}

func (s *Store) CreateBranch(ctx context.Context, branch domain.Branch) (domain.Branch, error) {
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO branches(world_id, parent_branch_id, name, created_from_match_id) VALUES (?, ?, ?, ?)`,
		branch.WorldID,
		toNullInt64(branch.ParentBranchID),
		branch.Name,
		toNullInt64(branch.CreatedFromMatchID),
	)
	if err != nil {
		return domain.Branch{}, fmt.Errorf("insert branch: %w", err)
	}
	branch.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Branch{}, fmt.Errorf("branch last insert id: %w", err)
	}
	return branch, nil
}

func (s *Store) ListBranches(ctx context.Context, worldID int64) ([]domain.Branch, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, world_id, parent_branch_id, name, created_from_match_id, created_at
		FROM branches WHERE world_id = ? ORDER BY id`,
		worldID,
	)
	if err != nil {
		return nil, fmt.Errorf("query branches: %w", err)
	}
	defer rows.Close()

	var branches []domain.Branch
	for rows.Next() {
		var branch domain.Branch
		var parentID sql.NullInt64
		var matchID sql.NullInt64
		if err := rows.Scan(
			&branch.ID,
			&branch.WorldID,
			&parentID,
			&branch.Name,
			&matchID,
			&branch.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan branch: %w", err)
		}
		branch.ParentBranchID = fromNullInt64(parentID)
		branch.CreatedFromMatchID = fromNullInt64(matchID)
		branches = append(branches, branch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate branches: %w", err)
	}
	return branches, nil
}

func (s *Store) CreateClub(ctx context.Context, club domain.Club) (domain.Club, error) {
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO clubs(branch_id, ext_key, name, short_name, country, division) VALUES (?, ?, ?, ?, ?, ?)`,
		club.BranchID,
		nullIfEmpty(club.ExtKey),
		club.Name,
		club.ShortName,
		nullIfEmpty(club.Country),
		nullIfEmpty(club.Division),
	)
	if err != nil {
		return domain.Club{}, fmt.Errorf("insert club: %w", err)
	}
	club.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Club{}, fmt.Errorf("club last insert id: %w", err)
	}
	return club, nil
}

func (s *Store) ListClubsByBranch(ctx context.Context, branchID int64) ([]domain.Club, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, branch_id, COALESCE(ext_key, ''), name, short_name, COALESCE(country, ''), COALESCE(division, '')
		FROM clubs WHERE branch_id = ? ORDER BY id`,
		branchID,
	)
	if err != nil {
		return nil, fmt.Errorf("query clubs: %w", err)
	}
	defer rows.Close()

	var clubs []domain.Club
	for rows.Next() {
		var club domain.Club
		if err := rows.Scan(
			&club.ID,
			&club.BranchID,
			&club.ExtKey,
			&club.Name,
			&club.ShortName,
			&club.Country,
			&club.Division,
		); err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		clubs = append(clubs, club)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clubs: %w", err)
	}
	return clubs, nil
}

func (s *Store) CreatePlayer(ctx context.Context, player domain.Player) (domain.Player, error) {
	secondary, err := encodePositions(player.SecondaryPositions)
	if err != nil {
		return domain.Player{}, err
	}
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO players(
			branch_id, ext_key, club_id, first_name, last_name, preferred_name, date_of_birth,
			nationality, primary_position, secondary_positions, overall, potential
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		player.BranchID,
		nullIfEmpty(player.ExtKey),
		toNullInt64(player.ClubID),
		player.FirstName,
		player.LastName,
		nullIfEmpty(player.PreferredName),
		nullTime(player.DateOfBirth),
		nullIfEmpty(player.Nationality),
		string(player.PrimaryPosition),
		secondary,
		player.Attributes.Overall,
		player.Attributes.Potential,
	)
	if err != nil {
		return domain.Player{}, fmt.Errorf("insert player: %w", err)
	}
	player.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Player{}, fmt.Errorf("player last insert id: %w", err)
	}
	return player, nil
}

func (s *Store) ListPlayersByClub(ctx context.Context, branchID, clubID int64) ([]domain.Player, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT
			id, branch_id, club_id, COALESCE(ext_key, ''), first_name, last_name, COALESCE(preferred_name, ''),
			date_of_birth, COALESCE(nationality, ''), primary_position, secondary_positions, overall, potential
		FROM players
		WHERE branch_id = ? AND club_id = ?
		ORDER BY id`,
		branchID,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("query players: %w", err)
	}
	defer rows.Close()

	var players []domain.Player
	for rows.Next() {
		player, err := scanPlayer(rows)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate players: %w", err)
	}
	return players, nil
}

func (s *Store) CreateSeason(ctx context.Context, season domain.Season) (domain.Season, error) {
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO seasons(branch_id, label, start_year) VALUES (?, ?, ?)`,
		season.BranchID,
		season.Label,
		season.StartYear,
	)
	if err != nil {
		return domain.Season{}, fmt.Errorf("insert season: %w", err)
	}
	season.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Season{}, fmt.Errorf("season last insert id: %w", err)
	}
	return season, nil
}

func (s *Store) CreateFixture(ctx context.Context, fixture domain.Fixture) (domain.Fixture, error) {
	if err := fixture.Validate(); err != nil {
		return domain.Fixture{}, err
	}
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO fixtures(season_id, matchday, home_club_id, away_club_id, scheduled_at)
		VALUES (?, ?, ?, ?, ?)`,
		fixture.SeasonID,
		fixture.Matchday,
		fixture.HomeClubID,
		fixture.AwayClubID,
		nullTime(fixture.Scheduled),
	)
	if err != nil {
		return domain.Fixture{}, fmt.Errorf("insert fixture: %w", err)
	}
	fixture.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Fixture{}, fmt.Errorf("fixture last insert id: %w", err)
	}
	return fixture, nil
}

func (s *Store) ListFixturesBySeason(ctx context.Context, seasonID int64) ([]domain.Fixture, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, season_id, matchday, home_club_id, away_club_id, scheduled_at
		FROM fixtures WHERE season_id = ? ORDER BY matchday, id`,
		seasonID,
	)
	if err != nil {
		return nil, fmt.Errorf("query fixtures: %w", err)
	}
	defer rows.Close()

	var fixtures []domain.Fixture
	for rows.Next() {
		var fixture domain.Fixture
		var scheduled sql.NullTime
		if err := rows.Scan(
			&fixture.ID,
			&fixture.SeasonID,
			&fixture.Matchday,
			&fixture.HomeClubID,
			&fixture.AwayClubID,
			&scheduled,
		); err != nil {
			return nil, fmt.Errorf("scan fixture: %w", err)
		}
		if scheduled.Valid {
			fixture.Scheduled = scheduled.Time
		}
		fixtures = append(fixtures, fixture)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fixtures: %w", err)
	}
	return fixtures, nil
}

func (s *Store) CreateMatch(ctx context.Context, match domain.Match) (domain.Match, error) {
	if match.TickCount == 0 {
		match.TickCount = 900
	}
	if strings.TrimSpace(match.Status) == "" {
		match.Status = "completed"
	}
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO matches(fixture_id, branch_id, home_goals, away_goals, tick_count, seed, status, simulated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		match.FixtureID,
		match.BranchID,
		match.HomeGoals,
		match.AwayGoals,
		match.TickCount,
		match.Seed,
		match.Status,
		nullTime(match.SimulatedAt),
	)
	if err != nil {
		return domain.Match{}, fmt.Errorf("insert match: %w", err)
	}
	match.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Match{}, fmt.Errorf("match last insert id: %w", err)
	}
	return match, nil
}

func (s *Store) SaveMatchEvents(ctx context.Context, events []domain.MatchEvent) error {
	for _, event := range events {
		if _, err := s.db.ExecContext(
			ctx,
			`INSERT INTO match_events(match_id, tick, minute, team_club_id, player_id, event_type, pitch_zone, payload_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			event.MatchID,
			event.Tick,
			event.Minute,
			toNullInt64(event.TeamClubID),
			toNullInt64(event.PlayerID),
			event.Type,
			nullIfEmpty(event.PitchZone),
			payloadOrDefault(event.PayloadRaw),
		); err != nil {
			return fmt.Errorf("insert match event: %w", err)
		}
	}
	return nil
}

func (s *Store) ListMatchEvents(ctx context.Context, matchID int64) ([]domain.MatchEvent, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, match_id, tick, minute, team_club_id, player_id, event_type, COALESCE(pitch_zone, ''), payload_json
		FROM match_events WHERE match_id = ? ORDER BY tick, id`,
		matchID,
	)
	if err != nil {
		return nil, fmt.Errorf("query match events: %w", err)
	}
	defer rows.Close()

	var events []domain.MatchEvent
	for rows.Next() {
		var event domain.MatchEvent
		var teamClubID sql.NullInt64
		var playerID sql.NullInt64
		if err := rows.Scan(
			&event.ID,
			&event.MatchID,
			&event.Tick,
			&event.Minute,
			&teamClubID,
			&playerID,
			&event.Type,
			&event.PitchZone,
			&event.PayloadRaw,
		); err != nil {
			return nil, fmt.Errorf("scan match event: %w", err)
		}
		event.TeamClubID = fromNullInt64(teamClubID)
		event.PlayerID = fromNullInt64(playerID)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate match events: %w", err)
	}
	return events, nil
}

func scanPlayer(scanner interface{ Scan(dest ...any) error }) (domain.Player, error) {
	var player domain.Player
	var clubID sql.NullInt64
	var birthDate sql.NullTime
	var primary string
	var secondary string
	if err := scanner.Scan(
		&player.ID,
		&player.BranchID,
		&clubID,
		&player.ExtKey,
		&player.FirstName,
		&player.LastName,
		&player.PreferredName,
		&birthDate,
		&player.Nationality,
		&primary,
		&secondary,
		&player.Attributes.Overall,
		&player.Attributes.Potential,
	); err != nil {
		return domain.Player{}, fmt.Errorf("scan player: %w", err)
	}
	player.ClubID = fromNullInt64(clubID)
	if birthDate.Valid {
		player.DateOfBirth = birthDate.Time
	}
	position, err := domain.ParsePosition(primary)
	if err != nil {
		return domain.Player{}, err
	}
	player.PrimaryPosition = position
	player.SecondaryPositions, err = decodePositions(secondary)
	if err != nil {
		return domain.Player{}, err
	}
	return player, nil
}

func encodePositions(positions []domain.Position) (string, error) {
	if len(positions) == 0 {
		return "", nil
	}
	parts := make([]string, 0, len(positions))
	for _, position := range positions {
		if !position.Valid() {
			return "", fmt.Errorf("encode positions: invalid position %q", position)
		}
		parts = append(parts, string(position))
	}
	return strings.Join(parts, ","), nil
}

func decodePositions(raw string) ([]domain.Position, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	positions := make([]domain.Position, 0, len(parts))
	for _, part := range parts {
		position, err := domain.ParsePosition(part)
		if err != nil {
			return nil, fmt.Errorf("decode positions: %w", err)
		}
		positions = append(positions, position)
	}
	return positions, nil
}

func toNullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func fromNullInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func nullIfEmpty(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value, Valid: true}
}

func payloadOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "{}"
	}
	return value
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
