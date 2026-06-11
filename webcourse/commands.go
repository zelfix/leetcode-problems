package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ensureSeeded opens the store and seeds it if empty.
func ensureSeeded(paths Paths) (*Store, error) {
	st, err := openStore(paths.DB)
	if err != nil {
		return nil, err
	}
	n, err := st.problemCount()
	if err != nil {
		st.Close()
		return nil, err
	}
	if n == 0 {
		days, problems, err := parsePlan(paths)
		if err != nil {
			st.Close()
			return nil, err
		}
		if err := st.replacePlan(days, problems); err != nil {
			st.Close()
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "(seeded %d problems on first run)\n", len(problems))
	}
	return st, nil
}

// goSnippetScaffold builds the contents of a fresh solution file.
func goSnippetScaffold(p *Problem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "// %s [%s] — %s\n", p.Title, p.Difficulty, p.Topic)
	fmt.Fprintf(&b, "// %s\n", p.URL)
	b.WriteString("//\n// Submit with:  neet submit " + p.Slug + "\n\n")
	b.WriteString("package main\n\n")
	snippet := strings.TrimSpace(p.GoSnippet)
	if snippet == "" {
		snippet = "// (no Go snippet available for this problem — write your solution here)"
	}
	b.WriteString(snippet)
	b.WriteString("\n")
	return b.String()
}

// cmdNew scaffolds solutions/<slug>.go from the dataset Go snippet.
func cmdNew(paths Paths, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: neet new <slug>")
	}
	slug := args[0]
	st, err := ensureSeeded(paths)
	if err != nil {
		return err
	}
	defer st.Close()

	p, err := st.getProblem(slug)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("unknown problem slug %q (not in the study plan)", slug)
	}

	if err := os.MkdirAll(paths.Solutions, 0o755); err != nil {
		return err
	}
	target := filepath.Join(paths.Solutions, slug+".go")
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%s already exists — not overwriting", target)
	}
	if err := os.WriteFile(target, []byte(goSnippetScaffold(p)), 0o644); err != nil {
		return err
	}
	fmt.Printf("Created %s\n", target)
	fmt.Printf("Edit it, then run:  neet submit %s\n", slug)
	return nil
}

// cmdSubmit reads a solution file, stores it, and compile-checks it.
func cmdSubmit(paths Paths, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: neet submit <slug> [file]")
	}
	slug := args[0]
	file := filepath.Join(paths.Solutions, slug+".go")
	if len(args) >= 2 {
		file = args[1]
	}

	st, err := ensureSeeded(paths)
	if err != nil {
		return err
	}
	defer st.Close()

	p, err := st.getProblem(slug)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("unknown problem slug %q (not in the study plan)", slug)
	}

	code, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not read solution file %s: %w", file, err)
	}

	fmt.Printf("Compiling %s ...\n", file)
	res := compileCheck(string(code))

	sub := Submission{
		Slug:          slug,
		Code:          string(code),
		Language:      "go",
		CompileOK:     res.OK,
		CompileOutput: res.Output,
		CreatedAt:     nowRFC3339(),
	}
	id, err := st.addSubmission(sub)
	if err != nil {
		return err
	}

	if res.OK {
		fmt.Printf("✅ Compiles. Submission #%d saved and marked solved.\n", id)
	} else {
		fmt.Printf("⚠️  Saved submission #%d, but it did NOT compile:\n\n%s\n\n", id, indent(res.Output))
		fmt.Println("(The submission is still recorded — fix and submit again.)")
	}
	fmt.Printf("View: http://localhost:8080/problem/%s\n", slug)
	return nil
}

// indent prefixes each line with two spaces for terminal display.
func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}

// cmdToday prints the plan day matching today (or the nearest upcoming day).
func cmdToday(paths Paths, _ []string) error {
	st, err := ensureSeeded(paths)
	if err != nil {
		return err
	}
	defer st.Close()

	days, err := st.allDays()
	if err != nil {
		return err
	}
	if len(days) == 0 {
		fmt.Println("No plan loaded.")
		return nil
	}

	today := time.Now().Format("2006-01-02")
	chosen := -1
	for i, d := range days {
		if d.Date == today {
			chosen = i
			break
		}
		if d.Date > today { // first upcoming day
			chosen = i
			break
		}
	}
	if chosen == -1 {
		chosen = len(days) - 1 // plan finished; show last day
	}

	d := days[chosen]
	label := "Today"
	if d.Date != today {
		label = "Next up"
	}
	fmt.Printf("%s — Day %d (%s) · %s\n", label, d.Day, d.Date, d.Topic)
	for _, p := range d.Problems {
		fmt.Printf("  %s %-8s %s  [%s]\n", statusMark(p.Status), "("+p.Difficulty+")", p.Title, p.Slug)
	}
	fmt.Printf("\nStart one with:  neet new <slug>\n")
	return nil
}

// cmdStatus prints overall progress.
func cmdStatus(paths Paths, _ []string) error {
	st, err := ensureSeeded(paths)
	if err != nil {
		return err
	}
	defer st.Close()

	stats, err := st.progressStats()
	if err != nil {
		return err
	}
	fmt.Printf("Progress: %d / %d solved (%.0f%%)\n", stats.Solved, stats.Total, percent(stats.Solved, stats.Total))
	diffs := make([]string, 0, len(stats.ByDiff))
	for d := range stats.ByDiff {
		diffs = append(diffs, d)
	}
	sort.Strings(diffs)
	for _, d := range diffs {
		v := stats.ByDiff[d]
		fmt.Printf("  %-8s %d / %d\n", d, v[0], v[1])
	}
	return nil
}

func statusMark(status string) string {
	switch status {
	case "solved":
		return "✅"
	case "attempted":
		return "🔄"
	default:
		return "⬜"
	}
}

func percent(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}
