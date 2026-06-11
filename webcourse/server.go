package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// server bundles the store and parsed templates.
type server struct {
	st   *Store
	tmpl *template.Template
}

// templateFuncs are helpers available inside HTML templates.
var templateFuncs = template.FuncMap{
	"lower": strings.ToLower,
	"add1":  func(i int) int { return i + 1 },
	// html renders trusted dataset HTML (hints contain <code> tags).
	"html": func(s string) template.HTML { return template.HTML(s) },
	// nl2br converts newlines to <br> for plain-text fields.
	"nl2br": func(s string) template.HTML {
		return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
	},
	"shortTime": func(s string) string {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return s
		}
		return t.Format("2006-01-02 15:04")
	},
}

// cmdServe starts the local web server.
func cmdServe(paths Paths, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	port := fs.Int("port", 8080, "port to listen on")
	if err := fs.Parse(args); err != nil {
		return err
	}

	st, err := ensureSeeded(paths)
	if err != nil {
		return err
	}
	defer st.Close()

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return err
	}

	s := &server{st: st, tmpl: tmpl}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /problem/{slug}", s.handleProblem)
	mux.HandleFunc("GET /day/{day}", s.handleDay)

	addr := fmt.Sprintf("localhost:%d", *port)
	fmt.Printf("NeetCode 250 course running at http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// indexData is the template payload for the plan overview.
type indexData struct {
	Days    []PlanDay
	Stats   ProgressStats
	Percent float64
	Today   string
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	days, err := s.st.allDays()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	stats, err := s.st.progressStats()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	data := indexData{
		Days:    days,
		Stats:   stats,
		Percent: percent(stats.Solved, stats.Total),
		Today:   time.Now().Format("2006-01-02"),
	}
	s.render(w, "index.html", data)
}

// problemData is the template payload for a single problem.
type problemData struct {
	Problem     *Problem
	Day         int
	Date        string
	Submissions []Submission
}

func (s *server) handleProblem(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	p, err := s.st.getProblem(slug)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if p == nil {
		http.NotFound(w, r)
		return
	}
	day, date, err := s.st.dayOf(slug)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	subs, err := s.st.submissionsFor(slug)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "problem.html", problemData{Problem: p, Day: day, Date: date, Submissions: subs})
}

func (s *server) handleDay(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.PathValue("day"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	probs, err := s.st.dayProblems(n)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if len(probs) == 0 {
		http.NotFound(w, r)
		return
	}
	// Reuse index template with a single day for simplicity.
	day := PlanDay{Day: n, Problems: probs}
	if len(probs) > 0 {
		day.Topic = probs[0].Topic
	}
	if _, date, _ := s.st.dayOf(probs[0].Slug); date != "" {
		day.Date = date
	}
	s.render(w, "day.html", day)
}

func (s *server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
