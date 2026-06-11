// Command neet is a local NeetCode-250 study tool: a web server for browsing the
// study plan and problems, plus console subcommands for scaffolding and
// submitting Go solutions, all backed by a shared SQLite database.
package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Fprint(os.Stderr, `neet — local NeetCode 250 study tool

Usage:
  neet seed                 Build/refresh the SQLite DB from the plan and problems/ dataset
  neet serve [-port 8080]   Start the local web server (auto-seeds if DB is empty)
  neet new <slug>           Scaffold solutions/<slug>.go from the Go snippet
  neet submit <slug> [file] Submit a solution (saves code, marks solved, compile-checks)
  neet today                Show today's problems and their status
  neet status               Show overall progress

Run from the repo root or from webcourse/.
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	paths, err := resolvePaths()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "seed":
		err = cmdSeed(paths, args)
	case "serve":
		err = cmdServe(paths, args)
	case "new":
		err = cmdNew(paths, args)
	case "submit":
		err = cmdSubmit(paths, args)
	case "today":
		err = cmdToday(paths, args)
	case "status":
		err = cmdStatus(paths, args)
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", cmd)
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
