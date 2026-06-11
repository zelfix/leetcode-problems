package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CompileResult captures the outcome of a best-effort compile check.
type CompileResult struct {
	OK     bool
	Output string
}

var reFuncMain = regexp.MustCompile(`(?m)^\s*func\s+main\s*\(`)

// helpersSrc defines the LeetCode-standard linked-list and tree node types,
// injected only when the submitted code references them without defining them.
const helpersSrc = `package main

type ListNode struct {
	Val  int
	Next *ListNode
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}
`

// compileCheck builds the submitted Go code in an isolated temp module using
// the system go toolchain. It is best-effort: LeetCode snippets are bare
// functions, so we wrap them in package main, add an empty main if absent, and
// inject ListNode/TreeNode helpers when referenced but not defined.
func compileCheck(code string) CompileResult {
	dir, err := os.MkdirTemp("", "neet-compile-*")
	if err != nil {
		return CompileResult{OK: false, Output: "could not create temp dir: " + err.Error()}
	}
	defer os.RemoveAll(dir)

	// Ensure the code is in package main.
	body := code
	if !regexp.MustCompile(`(?m)^\s*package\s+\w+`).MatchString(body) {
		body = "package main\n\n" + body
	} else {
		body = regexp.MustCompile(`(?m)^\s*package\s+\w+`).ReplaceAllString(body, "package main")
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module neetsubmission\n\ngo 1.26\n"), 0o644); err != nil {
		return CompileResult{OK: false, Output: err.Error()}
	}
	if err := os.WriteFile(filepath.Join(dir, "solution.go"), []byte(body), 0o644); err != nil {
		return CompileResult{OK: false, Output: err.Error()}
	}

	// Empty main so `go build` has an entrypoint, if the user didn't supply one.
	if !reFuncMain.MatchString(code) {
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
			return CompileResult{OK: false, Output: err.Error()}
		}
	}

	// Inject node helpers when referenced but not defined by the submission.
	needHelpers := false
	helpers := helpersSrc
	if strings.Contains(code, "ListNode") && !strings.Contains(code, "type ListNode") {
		needHelpers = true
	} else {
		helpers = strings.Replace(helpers, "\ntype ListNode struct {\n\tVal  int\n\tNext *ListNode\n}\n", "", 1)
	}
	if strings.Contains(code, "TreeNode") && !strings.Contains(code, "type TreeNode") {
		needHelpers = true
	} else {
		helpers = strings.Replace(helpers, "\ntype TreeNode struct {\n\tVal   int\n\tLeft  *TreeNode\n\tRight *TreeNode\n}\n", "", 1)
	}
	if needHelpers {
		if err := os.WriteFile(filepath.Join(dir, "helpers.go"), []byte(helpers), 0o644); err != nil {
			return CompileResult{OK: false, Output: err.Error()}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", os.DevNull, "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GO111MODULE=on")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return CompileResult{OK: false, Output: msg}
	}
	return CompileResult{OK: true, Output: strings.TrimSpace(string(out))}
}
