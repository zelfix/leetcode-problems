package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Problem is the denormalized record shown in the web UI and used by the CLI.
type Problem struct {
	Slug        string
	FrontendID  int
	Title       string
	Difficulty  string // Easy | Medium | Hard
	Topic       string // category from the study plan
	Topics      []string
	Description string
	Examples    []Example
	Constraints []string
	FollowUps   []string
	Hints       []string
	GoSnippet   string
	Solution    string
	URL         string
	HasData     bool

	// Progress fields (joined when listed).
	Status   string // todo | attempted | solved
	SolvedAt string
}

// Example mirrors one worked example from the dataset.
type Example struct {
	Num    int      `json:"example_num"`
	Text   string   `json:"example_text"`
	Images []string `json:"images"`
}

// PlanDay is one day of the study plan with its problems.
type PlanDay struct {
	Day      int
	Date     string
	Topic    string
	Problems []Problem
}

// Submission is one stored solution attempt.
type Submission struct {
	ID            int64
	Slug          string
	Code          string
	Language      string
	CompileOK     bool
	CompileOutput string
	CreatedAt     string
}

// Store wraps the SQLite database.
type Store struct {
	db *sql.DB
}

// openStore opens (creating if needed) the SQLite database and ensures the schema.
func openStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS problems (
  slug             TEXT PRIMARY KEY,
  frontend_id      INTEGER,
  title            TEXT,
  difficulty       TEXT,
  topic            TEXT,
  topics_json      TEXT,
  description      TEXT,
  examples_json    TEXT,
  constraints_json TEXT,
  follow_ups_json  TEXT,
  hints_json       TEXT,
  go_snippet       TEXT,
  solution         TEXT,
  url              TEXT,
  has_data         INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS plan_days (
  day   INTEGER PRIMARY KEY,
  date  TEXT,
  topic TEXT
);
CREATE TABLE IF NOT EXISTS plan_items (
  day      INTEGER,
  slug     TEXT,
  position INTEGER,
  PRIMARY KEY (day, slug)
);
CREATE TABLE IF NOT EXISTS submissions (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  slug           TEXT NOT NULL,
  code           TEXT NOT NULL,
  language       TEXT NOT NULL DEFAULT 'go',
  compile_ok     INTEGER NOT NULL DEFAULT 0,
  compile_output TEXT,
  created_at     TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS progress (
  slug               TEXT PRIMARY KEY,
  status             TEXT NOT NULL DEFAULT 'todo',
  solved_at          TEXT,
  last_submission_id INTEGER
);
CREATE INDEX IF NOT EXISTS idx_submissions_slug ON submissions(slug, created_at DESC);
`
	_, err := s.db.Exec(schema)
	return err
}

// --- seeding ---------------------------------------------------------------

// problemCount returns how many problems are currently seeded.
func (s *Store) problemCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM problems`).Scan(&n)
	return n, err
}

// replacePlan wipes and rewrites the problems/plan tables inside a transaction,
// preserving the user-owned submissions and progress tables.
func (s *Store) replacePlan(days []PlanDay, problems map[string]*Problem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range []string{`DELETE FROM problems`, `DELETE FROM plan_days`, `DELETE FROM plan_items`} {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	pStmt, err := tx.Prepare(`INSERT INTO problems
		(slug, frontend_id, title, difficulty, topic, topics_json, description,
		 examples_json, constraints_json, follow_ups_json, hints_json, go_snippet,
		 solution, url, has_data)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer pStmt.Close()

	for _, p := range problems {
		hasData := 0
		if p.HasData {
			hasData = 1
		}
		if _, err := pStmt.Exec(
			p.Slug, p.FrontendID, p.Title, p.Difficulty, p.Topic,
			jsonString(p.Topics), p.Description, jsonString(p.Examples),
			jsonString(p.Constraints), jsonString(p.FollowUps), jsonString(p.Hints),
			p.GoSnippet, p.Solution, p.URL, hasData,
		); err != nil {
			return fmt.Errorf("insert problem %s: %w", p.Slug, err)
		}
	}

	dStmt, err := tx.Prepare(`INSERT INTO plan_days (day, date, topic) VALUES (?,?,?)`)
	if err != nil {
		return err
	}
	defer dStmt.Close()
	iStmt, err := tx.Prepare(`INSERT INTO plan_items (day, slug, position) VALUES (?,?,?)`)
	if err != nil {
		return err
	}
	defer iStmt.Close()

	for _, d := range days {
		if _, err := dStmt.Exec(d.Day, d.Date, d.Topic); err != nil {
			return err
		}
		for pos, p := range d.Problems {
			if _, err := iStmt.Exec(d.Day, p.Slug, pos); err != nil {
				return err
			}
			// Ensure a progress row exists without clobbering existing status.
			if _, err := tx.Exec(
				`INSERT INTO progress (slug, status) VALUES (?, 'todo')
				 ON CONFLICT(slug) DO NOTHING`, p.Slug); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// --- queries ---------------------------------------------------------------

// allDays returns the full plan with problems and their current status.
func (s *Store) allDays() ([]PlanDay, error) {
	rows, err := s.db.Query(`SELECT day, date, topic FROM plan_days ORDER BY day`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []PlanDay
	for rows.Next() {
		var d PlanDay
		if err := rows.Scan(&d.Day, &d.Date, &d.Topic); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range days {
		probs, err := s.dayProblems(days[i].Day)
		if err != nil {
			return nil, err
		}
		days[i].Problems = probs
	}
	return days, nil
}

// dayProblems returns the problems for a single day with status info.
func (s *Store) dayProblems(day int) ([]Problem, error) {
	rows, err := s.db.Query(`
		SELECT p.slug, p.frontend_id, p.title, p.difficulty, p.topic, p.url, p.has_data,
		       COALESCE(pr.status,'todo'), COALESCE(pr.solved_at,'')
		FROM plan_items pi
		JOIN problems p ON p.slug = pi.slug
		LEFT JOIN progress pr ON pr.slug = p.slug
		WHERE pi.day = ?
		ORDER BY pi.position`, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Problem
	for rows.Next() {
		var p Problem
		var hasData int
		if err := rows.Scan(&p.Slug, &p.FrontendID, &p.Title, &p.Difficulty, &p.Topic,
			&p.URL, &hasData, &p.Status, &p.SolvedAt); err != nil {
			return nil, err
		}
		p.HasData = hasData == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// dayOf returns the plan day for a problem slug (0 if none).
func (s *Store) dayOf(slug string) (day int, date string, err error) {
	err = s.db.QueryRow(`
		SELECT pd.day, pd.date FROM plan_items pi
		JOIN plan_days pd ON pd.day = pi.day
		WHERE pi.slug = ?`, slug).Scan(&day, &date)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}
	return day, date, err
}

// getProblem loads a full problem record including status.
func (s *Store) getProblem(slug string) (*Problem, error) {
	var p Problem
	var hasData int
	var topicsJSON, examplesJSON, constraintsJSON, followUpsJSON, hintsJSON string
	err := s.db.QueryRow(`
		SELECT p.slug, p.frontend_id, p.title, p.difficulty, p.topic, p.topics_json,
		       p.description, p.examples_json, p.constraints_json, p.follow_ups_json,
		       p.hints_json, p.go_snippet, p.solution, p.url, p.has_data,
		       COALESCE(pr.status,'todo'), COALESCE(pr.solved_at,'')
		FROM problems p
		LEFT JOIN progress pr ON pr.slug = p.slug
		WHERE p.slug = ?`, slug).Scan(
		&p.Slug, &p.FrontendID, &p.Title, &p.Difficulty, &p.Topic, &topicsJSON,
		&p.Description, &examplesJSON, &constraintsJSON, &followUpsJSON,
		&hintsJSON, &p.GoSnippet, &p.Solution, &p.URL, &hasData,
		&p.Status, &p.SolvedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.HasData = hasData == 1
	jsonParse(topicsJSON, &p.Topics)
	jsonParse(examplesJSON, &p.Examples)
	jsonParse(constraintsJSON, &p.Constraints)
	jsonParse(followUpsJSON, &p.FollowUps)
	jsonParse(hintsJSON, &p.Hints)
	return &p, nil
}

// submissionsFor returns submissions for a slug, newest first.
func (s *Store) submissionsFor(slug string) ([]Submission, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, code, language, compile_ok, COALESCE(compile_output,''), created_at
		FROM submissions WHERE slug = ? ORDER BY created_at DESC, id DESC`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Submission
	for rows.Next() {
		var sub Submission
		var ok int
		if err := rows.Scan(&sub.ID, &sub.Slug, &sub.Code, &sub.Language, &ok, &sub.CompileOutput, &sub.CreatedAt); err != nil {
			return nil, err
		}
		sub.CompileOK = ok == 1
		out = append(out, sub)
	}
	return out, rows.Err()
}

// addSubmission records a submission and updates progress to solved.
func (s *Store) addSubmission(sub Submission) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	ok := 0
	if sub.CompileOK {
		ok = 1
	}
	res, err := tx.Exec(`INSERT INTO submissions (slug, code, language, compile_ok, compile_output, created_at)
		VALUES (?,?,?,?,?,?)`, sub.Slug, sub.Code, sub.Language, ok, sub.CompileOutput, sub.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// A submission that compiles counts as solved; otherwise it's an attempt.
	// Never downgrade a problem that was already solved by an earlier submission.
	status := "attempted"
	solvedAt := ""
	if sub.CompileOK {
		status = "solved"
		solvedAt = sub.CreatedAt
	}
	if _, err := tx.Exec(`
		INSERT INTO progress (slug, status, solved_at, last_submission_id)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			last_submission_id = excluded.last_submission_id,
			status   = CASE WHEN progress.status='solved' OR excluded.status='solved' THEN 'solved' ELSE excluded.status END,
			solved_at = COALESCE(NULLIF(progress.solved_at,''), NULLIF(excluded.solved_at,''))`,
		sub.Slug, status, solvedAt, id); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// ProgressStats summarizes how far along the user is.
type ProgressStats struct {
	Total  int
	Solved int
	ByDiff map[string][2]int // difficulty -> [solved, total]
}

func (s *Store) progressStats() (ProgressStats, error) {
	st := ProgressStats{ByDiff: map[string][2]int{}}
	rows, err := s.db.Query(`
		SELECT p.difficulty, COALESCE(pr.status,'todo')
		FROM problems p LEFT JOIN progress pr ON pr.slug = p.slug`)
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		var diff, status string
		if err := rows.Scan(&diff, &status); err != nil {
			return st, err
		}
		st.Total++
		cur := st.ByDiff[diff]
		cur[1]++
		if status == "solved" {
			st.Solved++
			cur[0]++
		}
		st.ByDiff[diff] = cur
	}
	return st, rows.Err()
}

// nowRFC3339 returns the current timestamp formatted for storage.
func nowRFC3339() string { return time.Now().Format(time.RFC3339) }
