# NeetCode 250 — Local Study Course

A self-contained local tool for working through the
[`NeetCode_250_Study_Plan_2026-06-15.md`](../NeetCode_250_Study_Plan_2026-06-15.md)
plan in Go. It combines:

- a **web server** to browse the day-by-day plan, read each problem (description,
  examples, constraints, hints, Go starter, editorial), and track progress &
  submission history;
- a **console utility** to scaffold and submit Go solutions (stored in SQLite,
  marked solved, and compile-checked);
- a **Claude skill** (`neetcode-coach`) that gently, Socratically guides you.

Everything is backed by one SQLite file (`webcourse/course.db`). Pure Go — the
only dependency is the cgo-free `modernc.org/sqlite` driver.

## Setup

```bash
cd webcourse
go build -o neet .
./neet seed        # build the DB from the plan + problems/ dataset (250 problems)
```

`seed` is safe to re-run: it rebuilds the plan/problem tables but **preserves**
your submissions and progress.

## Daily workflow

```bash
./neet today                 # show today's (or next) problems
./neet new two-sum           # scaffold solutions/two-sum.go from the Go starter
#   ...solve it in your editor...
./neet submit two-sum        # save to DB, mark solved, compile-check
./neet status                # overall progress
./neet serve                 # browse at http://localhost:8080
```

The tool auto-detects the repo root (the directory containing
`merged_problems.json`), so you can run it from the repo root or from
`webcourse/`.

## Commands

| Command | What it does |
|---|---|
| `neet seed` | (Re)build problem/plan tables from the dataset. Preserves progress. |
| `neet serve [-port 8080]` | Start the web server (auto-seeds if empty). |
| `neet new <slug>` | Create `solutions/<slug>.go` from the Go snippet. Won't overwrite. |
| `neet submit <slug> [file]` | Save a submission, mark solved, run a best-effort `go build` check. Default file `solutions/<slug>.go`. |
| `neet today` | Today's plan day and problem statuses. |
| `neet status` | Solved/total, broken down by difficulty. |

## How submit works

There are no test cases in the dataset, so `submit` does **not** run your code
against test inputs. Instead it:

1. saves the full source + timestamp into SQLite (`submissions`),
2. marks the problem **solved** if it compiles, else **attempted**,
3. runs a best-effort compile check: the code is wrapped in `package main` in a
   temp module, `ListNode`/`TreeNode` helpers are injected if referenced, and
   `go build` is run. Compile output (errors) is stored and shown on the
   problem page.

A non-compiling submission is still recorded — fix it and submit again.

## The coach skill

`.claude/skills/neetcode-coach/SKILL.md` makes Claude act as a Socratic coach: it
reads the problem's hints/editorial and your current `solutions/<slug>.go`, then
nudges you one step at a time. It only reveals the full solution if you
explicitly ask. Trigger it by asking Claude things like *"coach me on two-sum"*
or *"I'm stuck on this problem"* while editing a solution file.

## Notes

- 243/250 problems have full dataset content; 7 are premium/locked and show as
  stubs (🔒) with plan metadata + a LeetCode link only.
- `webcourse/course.db` and the `neet` binary are git-ignored; your `solutions/`
  are tracked so you can commit your progress.
