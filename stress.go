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

// StressConfig содержит параметры стресс-теста.
type StressConfig struct {
	// Generate генерирует случайный входной тест в виде строки.
	Generate func(rng *rand.Rand) string

	// BruteForce — медленное, но заведомо верное эталонное решение.
	BruteForce func(in io.Reader, out io.Writer)

	// Optimized — быстрое решение, которое проверяется.
	Optimized func(in io.Reader, out io.Writer)

	// NumTests — количество случайных тестов. По умолчанию 1000.
	NumTests int

	// Seed — начальное значение генератора случайных чисел.
	// При значении 0 используется текущее время.
	Seed int64

	// TimeLimit — лимит времени на один тест для Optimized.
	// По умолчанию 2 секунды.
	TimeLimit time.Duration

	// TestdataDir — директория, куда сохраняется failed.in при падении.
	// По умолчанию "testdata".
	TestdataDir string
}

// RunStressTest запускает стресс-тест согласно cfg.
// Генерирует NumTests случайных входов, прогоняет BruteForce и Optimized,
// сравнивает их вывод. При первом расхождении (WRONG ANSWER) или превышении
// лимита времени (TLE) тест падает и сохраняет проблемный вход в failed.in.
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

// saveFailedInput сохраняет проблемный входной тест в dir/failed.in.
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