// Package runner provides utilities for testing competitive programming solutions.
package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const maxOutputLen = 2000

// RunFileTests runs all tests found in testdataDir.
// It discovers pairs of NN.in / NN.out files, feeds each .in to solve,
// and compares the output with the corresponding .out file.
// Output is normalized by trimming leading/trailing whitespace before comparison.
// If a .in file has no matching .out file, the solution is run and its output
// is printed with a "no expected output" notice — useful for quickly inspecting
// output on a new test case without creating the .out file first.
func RunFileTests(t *testing.T, solve func(in io.Reader, out io.Writer), testdataDir string) {
	t.Helper()

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("cannot read testdata dir %q: %v", testdataDir, err)
	}

	var inFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".in") {
			inFiles = append(inFiles, e.Name())
		}
	}
	sort.Strings(inFiles)

	if len(inFiles) == 0 {
		t.Logf("no test files found in %s", testdataDir)
		return
	}

	for _, inFile := range inFiles {
		name := strings.TrimSuffix(inFile, ".in")
		inPath := filepath.Join(testdataDir, inFile)
		outPath := filepath.Join(testdataDir, name+".out")

		t.Run(name, func(t *testing.T) {
			inData, err := os.ReadFile(inPath)
			if err != nil {
				t.Fatalf("cannot read input file %q: %v", inPath, err)
			}

			var got bytes.Buffer
			solve(bytes.NewReader(inData), &got)
			gotStr := normalize(got.String())

			outData, err := os.ReadFile(outPath)
			if err != nil {
				t.Logf("no expected output, got:\n%s", truncate(gotStr))
				return
			}

			wantStr := normalize(string(outData))
			if gotStr == wantStr {
				return
			}

			diffLine, wantLine, gotLine := firstDiff(wantStr, gotStr)
			t.Errorf("WRONG ANSWER on test %q\n"+
				"first diff at line %d:\n  want: %q\n  got:  %q\n"+
				"--- expected (truncated at %d chars) ---\n%s\n"+
				"--- got (truncated at %d chars) ---\n%s",
				name, diffLine, wantLine, gotLine,
				maxOutputLen, truncate(wantStr),
				maxOutputLen, truncate(gotStr),
			)
		})
	}
}

func normalize(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func truncate(s string) string {
	if len(s) <= maxOutputLen {
		return s
	}
	return fmt.Sprintf("%s\n... (truncated, %d total chars)", s[:maxOutputLen], len(s))
}

func firstDiff(want, got string) (lineNum int, wantLine, gotLine string) {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	n := len(wantLines)
	if len(gotLines) > n {
		n = len(gotLines)
	}

	for i := 0; i < n; i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w != g {
			return i + 1, w, g
		}
	}
	return n, "", ""
}
