package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// jsonString marshals v to a compact JSON string (empty array/object on error).
func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

// jsonParse unmarshals s into dst, ignoring empty input and errors.
func jsonParse(s string, dst any) {
	if strings.TrimSpace(s) == "" {
		return
	}
	_ = json.Unmarshal([]byte(s), dst)
}

// rawProblem mirrors the on-disk problems/*.json schema (only fields we use).
type rawProblem struct {
	Title        string            `json:"title"`
	FrontendID   string            `json:"frontend_id"`
	Difficulty   string            `json:"difficulty"`
	Slug         string            `json:"problem_slug"`
	Topics       []string          `json:"topics"`
	Description  string            `json:"description"`
	Examples     []Example         `json:"examples"`
	Constraints  []string          `json:"constraints"`
	FollowUps    []string          `json:"follow_ups"`
	Hints        []string          `json:"hints"`
	CodeSnippets map[string]string `json:"code_snippets"`
	Solution     string            `json:"solution"`
}

// buildSlugIndex maps a leetcode slug to its problems/NNNN-slug.json path.
func buildSlugIndex(problemsDir string) (map[string]string, error) {
	entries, err := os.ReadDir(problemsDir)
	if err != nil {
		return nil, err
	}
	idx := make(map[string]string, len(entries))
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		base := strings.TrimSuffix(name, ".json") // NNNN-slug
		if i := strings.Index(base, "-"); i >= 0 {
			slug := base[i+1:]
			idx[slug] = filepath.Join(problemsDir, name)
		}
	}
	return idx, nil
}

// loadProblemData reads a problems/*.json file into a Problem.
func loadProblemData(path string) (*Problem, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rp rawProblem
	if err := json.Unmarshal(b, &rp); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	fid, _ := strconv.Atoi(rp.FrontendID)
	p := &Problem{
		Slug:        rp.Slug,
		FrontendID:  fid,
		Title:       rp.Title,
		Difficulty:  rp.Difficulty,
		Topics:      rp.Topics,
		Description: rp.Description,
		Examples:    rp.Examples,
		Constraints: rp.Constraints,
		FollowUps:   rp.FollowUps,
		Hints:       rp.Hints,
		GoSnippet:   rp.CodeSnippets["golang"],
		Solution:    rp.Solution,
		HasData:     true,
	}
	return p, nil
}

var (
	reDay  = regexp.MustCompile(`(?m)^## Day (\d+) - (\d{4}-\d{2}-\d{2})\s*$`)
	reItem = regexp.MustCompile(`- \[[ x]\] (\S+) \[(.+?)\]\(https://leetcode\.com/problems/([a-z0-9-]+)/?\) - \*(.+?)\*`)
)

// difficultyForEmoji maps the plan's emoji to a difficulty label.
func difficultyForEmoji(emoji string) string {
	switch emoji {
	case "🟢":
		return "Easy"
	case "🟡":
		return "Medium"
	case "🔴":
		return "Hard"
	default:
		return "Unknown"
	}
}

// parsePlan parses the study-plan markdown into days and a problem map. For
// problems present in the dataset it loads full data; the rest become stubs.
func parsePlan(paths Paths) ([]PlanDay, map[string]*Problem, error) {
	src, err := os.ReadFile(paths.Plan)
	if err != nil {
		return nil, nil, err
	}
	idx, err := buildSlugIndex(paths.Problems)
	if err != nil {
		return nil, nil, err
	}

	text := string(src)
	dayLocs := reDay.FindAllStringSubmatchIndex(text, -1)

	problems := make(map[string]*Problem)
	var days []PlanDay

	for i, loc := range dayLocs {
		dayNum, _ := strconv.Atoi(text[loc[2]:loc[3]])
		date := text[loc[4]:loc[5]]

		// Section spans from end of this header to start of the next.
		start := loc[1]
		end := len(text)
		if i+1 < len(dayLocs) {
			end = dayLocs[i+1][0]
		}
		section := text[start:end]

		// Topic line.
		topic := ""
		if m := regexp.MustCompile(`\*\*Topic:\*\* (.+)`).FindStringSubmatch(section); m != nil {
			topic = strings.TrimSpace(m[1])
		}

		d := PlanDay{Day: dayNum, Date: date, Topic: topic}
		for _, im := range reItem.FindAllStringSubmatch(section, -1) {
			emoji, title, slug, itemTopic := im[1], im[2], im[3], im[4]
			p, ok := problems[slug]
			if !ok {
				p = resolveProblem(slug, title, emoji, itemTopic, idx)
				problems[slug] = p
			}
			d.Problems = append(d.Problems, *p)
		}
		days = append(days, d)
	}
	return days, problems, nil
}

// resolveProblem loads dataset data for a slug, or builds a stub from plan metadata.
func resolveProblem(slug, title, emoji, topic string, idx map[string]string) *Problem {
	url := "https://leetcode.com/problems/" + slug + "/"
	if path, ok := idx[slug]; ok {
		if p, err := loadProblemData(path); err == nil {
			p.Topic = topic
			p.URL = url
			if p.Slug == "" {
				p.Slug = slug
			}
			if p.Title == "" {
				p.Title = title
			}
			return p
		}
	}
	// Premium / missing problem: stub from plan metadata.
	return &Problem{
		Slug:       slug,
		Title:      title,
		Difficulty: difficultyForEmoji(emoji),
		Topic:      topic,
		URL:        url,
		HasData:    false,
	}
}

// cmdSeed (re)builds the problem/plan tables from the dataset.
func cmdSeed(paths Paths, _ []string) error {
	days, problems, err := parsePlan(paths)
	if err != nil {
		return err
	}
	st, err := openStore(paths.DB)
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.replacePlan(days, problems); err != nil {
		return err
	}
	fmt.Printf("Seeded %d days, %d unique problems into %s\n", len(days), len(problems), paths.DB)
	withData := 0
	for _, p := range problems {
		if p.HasData {
			withData++
		}
	}
	fmt.Printf("  %d with dataset content, %d stubs (premium/missing)\n", withData, len(problems)-withData)
	return nil
}
