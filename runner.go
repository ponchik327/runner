// Пакет runner предоставляет утилиты для тестирования решений алгоритмических задач.
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

// максимальная длина вывода при отображении в ошибке теста
const maxOutputLen = 2000

// RunFileTests запускает все тесты из директории testdataDir.
// Ищет пары файлов NN.in / NN.out, подаёт каждый .in в solve
// и сравнивает результат с соответствующим .out.
// Вывод нормализуется перед сравнением: обрезается trailing whitespace и \r\n.
// Если файл .out отсутствует — решение запускается и вывод печатается без сравнения,
// что удобно для быстрой проверки на новом тесте.
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

// normalize приводит строку к нормализованному виду:
// заменяет \r\n на \n, обрезает trailing whitespace на каждой строке
// и убирает leading/trailing пустые строки.
func normalize(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// truncate обрезает строку до maxOutputLen символов и добавляет пометку об усечении.
func truncate(s string) string {
	if len(s) <= maxOutputLen {
		return s
	}
	return fmt.Sprintf("%s\n... (truncated, %d total chars)", s[:maxOutputLen], len(s))
}

// firstDiff возвращает номер первой отличающейся строки и её содержимое в want и got.
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