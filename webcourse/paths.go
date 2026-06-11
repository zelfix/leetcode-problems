package main

import (
	"errors"
	"os"
	"path/filepath"
)

// Paths holds the resolved locations the tool works with. The repo root is the
// directory that contains the problems dataset and the study-plan markdown.
type Paths struct {
	Root      string // repo root (contains problems/ and the plan .md)
	Problems  string // <root>/problems
	Plan      string // <root>/NeetCode_250_Study_Plan_2026-06-15.md
	Solutions string // <root>/solutions
	DB        string // <root>/webcourse/course.db
}

const planFileName = "NeetCode_250_Study_Plan_2026-06-15.md"

// findRoot walks up from the current working directory looking for the marker
// file merged_problems.json, so the tool works whether it is invoked from the
// repo root or from the webcourse/ subdirectory.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "merged_problems.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not locate repo root (merged_problems.json not found in any parent directory)")
		}
		dir = parent
	}
}

// resolvePaths builds the Paths set from the detected repo root.
func resolvePaths() (Paths, error) {
	root, err := findRoot()
	if err != nil {
		return Paths{}, err
	}
	return Paths{
		Root:      root,
		Problems:  filepath.Join(root, "problems"),
		Plan:      filepath.Join(root, planFileName),
		Solutions: filepath.Join(root, "solutions"),
		DB:        filepath.Join(root, "webcourse", "course.db"),
	}, nil
}
