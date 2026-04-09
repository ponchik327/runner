package runner

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// StressConfig holds the configuration for a stress test run.
type StressConfig struct {
	// Generate produces a random test case as a string given an rng source.
	Generate func(rng *rand.Rand) string

	// BruteForce is the slow but obviously correct reference solution.
	BruteForce func(in io.Reader, out io.Writer)

	// Optimized is the fast solution being verified.
	Optimized func(in io.Reader, out io.Writer)

	// NumTests is the number of random test cases to run. Defaults to 1000.
	NumTests int

	// Seed is the RNG seed. Use 0 for a random seed derived from the current time.
	Seed int64

	// TimeLimit is the per-test time limit applied to Optimized only.
	// Defaults to 2 seconds.
	TimeLimit time.Duration

	// TestdataDir is where failed.in is written on failure.
	// Defaults to "testdata" relative to the test working directory.
	TestdataDir string
}

// RunStressTest runs a stress test according to cfg.
// It generates NumTests random inputs, runs both BruteForce and Optimized,
// and compares their outputs. On the first mismatch (WRONG ANSWER) or
// timeout (TLE), it fails the test with full details and saves the
// failing input to testdata/failed.in for easy reproduction.
func RunStressTest(t *testing.T, cfg StressConfig) {
	t.Helper()

	if cfg.NumTests <= 0 {
		cfg.NumTests = 1000
	}
	if cfg.TimeLimit <= 0 {
		cfg.TimeLimit = 2 * time.Second
	}
	if cfg.TestdataDir == "" {
		cfg.TestdataDir = "testdata"
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	t.Logf("stress test seed: %d (use Seed: %d to reproduce)", cfg.Seed, cfg.Seed)

	rng := rand.New(rand.NewSource(cfg.Seed))

	for i := 1; i <= cfg.NumTests; i++ {
		input := cfg.Generate(rng)

		var bruteOut bytes.Buffer
		cfg.BruteForce(strings.NewReader(input), &bruteOut)
		bruteResult := normalize(bruteOut.String())

		var optOut bytes.Buffer
		start := time.Now()
		cfg.Optimized(strings.NewReader(input), &optOut)
		elapsed := time.Since(start)
		optResult := normalize(optOut.String())

		if elapsed > cfg.TimeLimit {
			saveFailedInput(t, cfg.TestdataDir, input)
			t.Fatalf("TLE on test #%d: optimized took %v (limit %v)\ninput:\n%s",
				i, elapsed, cfg.TimeLimit, truncate(input))
		}

		if bruteResult != optResult {
			saveFailedInput(t, cfg.TestdataDir, input)
			t.Fatalf("WRONG ANSWER on test #%d\ninput:\n%s\nexpected (brute):\n%s\ngot (optimized):\n%s",
				i, truncate(input), truncate(bruteResult), truncate(optResult))
		}

		t.Logf("test #%d: OK (%dms)", i, elapsed.Milliseconds())
	}
}

func saveFailedInput(t *testing.T, dir, input string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Logf("warning: cannot create testdata dir: %v", err)
		return
	}

	path := filepath.Join(dir, "failed.in")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Logf("warning: cannot write failed.in: %v", err)
		return
	}

	t.Logf("failing input saved to %s", path)
}
