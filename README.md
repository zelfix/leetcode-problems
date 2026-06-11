# LeetCode Problems — JSON Dataset & Local Study Course

This repository provides a dataset of LeetCode problems in JSON format **and** a self-contained local study tool built on top of it for grinding the [NeetCode 250](NeetCode_250_Study_Plan_2026-06-15.md) plan in Go.

- **Dataset** — every problem as a separate `.json` file in `problems/`, plus a merged `merged_problems.json`. (Schema documented [below](#dataset-structure).)
- **Study course** (`webcourse/`) — a Go CLI + web server that scaffolds, compile-checks, and tracks your Go solutions, backed by a local SQLite DB.
- **Claude coach** (`.claude/skills/neetcode-coach/`) — a Socratic coaching skill for Claude Code that nudges you through each problem without spoiling the answer.

## Quick Start — Local Study Course

**Prerequisites:** [Go](https://go.dev/dl/) **1.26+** (the only toolchain you need; the SQLite driver is cgo-free, so no C compiler is required).

```bash
# 1. Clone and enter the repo
git clone git@github.com:zelfix/leetcode-problems.git
cd leetcode-problems

# 2. Build the CLI (run from webcourse/)
cd webcourse
go build -o neet .

# 3. Seed the local database from the plan + problems/ dataset (250 problems)
./neet seed
```

That's it. The database (`webcourse/course.db`) and the `neet` binary are **generated locally** and are intentionally git-ignored — a fresh clone never ships them, you build and seed your own. This keeps your personal progress off the shared repo.

### Daily workflow

```bash
./neet today              # show today's (or next) plan problems
./neet new two-sum        # scaffold solutions/two-sum.go from the Go starter
#   ...solve it in your editor...
./neet submit two-sum     # save to DB, mark solved, compile-check
./neet status             # overall progress
./neet serve              # browse the plan & track progress at http://localhost:8080
```

The tool auto-detects the repo root (the directory containing `merged_problems.json`), so you can run `neet` from the repo root or from `webcourse/`. Your solutions live in `solutions/<slug>.go` and **are** tracked in git, so you can commit your work; the DB and binary are not.

### The Claude coach skill

`.claude/skills/neetcode-coach/SKILL.md` turns Claude Code into a patient, Socratic coach. It reads the problem's hints/editorial and your current `solutions/<slug>.go`, then guides you one step at a time, only revealing the full solution if you explicitly ask. Trigger it in Claude Code with `/neetcode-coach <slug>` (or just *"coach me on two-sum"* while editing a solution file).

For the full command reference and how `submit`'s compile-check works, see [`webcourse/README.md`](webcourse/README.md).

## Dataset Structure

- `problems/`: Contains individual LeetCode problems as separate `.json` files. Each file is named with the problem's ID and a slug (e.g., `0001-two-sum.json`).
- `merged_problems.json`: A single file containing all problems merged into a list.

## Updated JSON Schema for Each Problem

Each problem JSON file contains the following fields:

- `title`: The name of the problem (e.g., "Container With Most Water").
- `problem_id`: The internal problem ID (string).
- `frontend_id`: The LeetCode frontend ID (string).
- `difficulty`: The difficulty level (`Easy`, `Medium`, or `Hard`).
- `problem_slug`: The URL-friendly name (e.g., `container-with-most-water`).
- `topics`: Array of topic tags (e.g., `Array`, `Two Pointers`).
- `description`: The full problem statement, usually in Markdown format.
- `examples`: Array of example objects, each with:
    - `example_num`: Example number
    - `example_text`: Input/output and explanation
    - `images`: Array of image URLs (if available)
- `constraints`: Array of constraints or limits for the problem.
- `follow_ups`: Array of follow-up questions (if any).
- `hints`: Array of hints for solving the problem.
- `code_snippets`: Object containing starter code for various languages (e.g., `python`, `cpp`, `java`, etc.)
- `solutions`: HTML string containing editorial content for some problems.

## Example Problem JSON

```json
{
  "title": "Container With Most Water",
  "problem_id": "11",
  "frontend_id": "11",
  "difficulty": "Medium",
  "problem_slug": "container-with-most-water",
  "topics": [
      "Array",
      "Two Pointers",
      "Greedy"
  ],
  "description": "You are given an integer array height of length n...",
  "examples": [
    {
      "example_num": 1,
      "example_text": "Input: height = [1,8,6,2,5,4,8,3,7]\\nOutput: 49\\nExplanation: ...",
      "images": ["https://s3-lc-upload.s3.amazonaws.com/uploads/2018/07/17/question_11.jpg"]
    },
    {
      "example_num": 2,
      "example_text": "Input: height = [1,1]\\nOutput: 1",
      "images": ["https://s3-lc-upload.s3.amazonaws.com/uploads/2018/07/17/question_11.jpg"]
    }
  ],
  "constraints": [
    "n == height.length",
    "2 <= n <= 105",
    "0 <= height[i] <= 104"
  ],
  "follow_ups": [],
  "hints": [
    "If you simulate the problem, it will be O(n^2) which is not efficient.",
    "Try to use two-pointers...",
    "How can you calculate the amount of water at each step?"
  ],
  "code_snippets": {
    "cpp": "class Solution {\npublic:\n    int maxArea(vector<int>& height) {\n        \n    }\n};",
    "java": "class Solution {\n    public int maxArea(int[] height) {\n        \n    }\n}",
    "python": "class Solution(object):\n    def maxArea(self, height):\n        \"\"\"\n        :type height: List[int]\n        :rtype: int\n        \"\"\"\n        ",
    "python3": "class Solution:\n    def maxArea(self, height: List[int]) -> int:\n        ",
    "c": "int maxArea(int* height, int heightSize) {\n    \n}",
    "csharp": "public class Solution {\n    public int MaxArea(int[] height) {\n        \n    }\n}",
    "javascript": "/**\n * @param {number[]} height\n * @return {number}\n */\nvar maxArea = function(height) {\n    \n};",
    "typescript": "function maxArea(height: number[]): number {\n    \n};",
    "php": "class Solution {\n\n    /**\n     * @param Integer[] $height\n     * @return Integer\n     */\n    function maxArea($height) {\n        \n    }\n}",
    "swift": "class Solution {\n    func maxArea(_ height: [Int]) -> Int {\n        \n    }\n}",
    "kotlin": "class Solution {\n    fun maxArea(height: IntArray): Int {\n        \n    }\n}",
    "dart": "class Solution {\n  int maxArea(List<int> height) {\n    \n  }\n}",
    "golang": "func maxArea(height []int) int {\n    \n}",
    "ruby": "# @param {Integer[]} height\n# @return {Integer}\ndef max_area(height)\n    \nend",
    "scala": "object Solution {\n    def maxArea(height: Array[Int]): Int = {\n        \n    }\n}",
    "rust": "impl Solution {\n    pub fn max_area(height: Vec<i32>) -> i32 {\n        \n    }\n}",
    "racket": "(define/contract (max-area height)\n  (-> (listof exact-integer?) exact-integer?)\n  )",
    "erlang": "-spec max_area(Height :: [integer()]) -> integer().\nmax_area(Height) ->\n  .",
    "elixir": "defmodule Solution do\n  @spec max_area(height :: [integer]) :: integer\n  def max_area(height) do\n    \n  end\nend"
  }
}
```

## Notes
- Some fields (like `solutions`, `images`, `follow_ups`) may be missing for certain problems.

## Usage

You can use this dataset for:
- Building practice tools
- Analyzing problem trends
- Interview preparation
- Educational projects

Feel free to contribute or suggest improvements!
