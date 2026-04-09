# runner

Go-пакет для тестирования решений алгоритмических задач.

---

## 1. Создать воркспейс

```bash
go run github.com/ponchik327/runner/cmd/new@latest -p 7 -dir contest -mod contest
cd contest
go mod tidy
```

Флаги:
- `-p 7` — количество задач (по умолчанию 7)
- `-dir contest` — имя создаваемой папки
- `-mod contest` — имя Go-модуля

Создаётся структура:

```
contest/
├── go.mod
├── Makefile
├── p1/
│   ├── main.go         ← решение задачи
│   ├── main_test.go    ← файловые тесты
│   ├── stress_test.go  ← стресс-тест
│   └── testdata/       ← входные/выходные данные
├── p2/
│   └── ...
└── p7/
    └── ...
```

---

## 2. Написать решение

Откройте `p1/main.go` и напишите решение в функцию `solve`:

```go
func solve(in io.Reader, out io.Writer) {
    var n int
    fmt.Fscan(in, &n)
    fmt.Fprintln(out, n*2)
}
```

Читайте из `in`, пишите в `out`. Функцию `main` не трогайте.

---

## 3. Файловые тесты

Добавьте тест:

```bash
make add-test P=1 ID=01
```

Создадутся пустые файлы `p1/testdata/01.in` и `p1/testdata/01.out`. Заполните их входными и ожидаемыми выходными данными.

Запустить тесты:

```bash
make test P=1
```

Если файла `.out` нет — решение всё равно запустится и выведет результат (удобно для быстрой проверки).

---

## 4. Стресс-тест

Откройте `p1/stress_test.go`. Там уже есть шаблон — нужно реализовать две функции:

**`generate`** — генерирует случайный входной тест:

```go
func generate(rng *rand.Rand) string {
    n := rng.Intn(100) + 1
    return fmt.Sprintf("%d\n", n)
}
```

**`brute`** — медленное, но заведомо верное решение:

```go
func brute(in io.Reader, out io.Writer) {
    var n int
    fmt.Fscan(in, &n)
    fmt.Fprintln(out, n*2)
}
```

Затем уберите строку `t.Skip(...)` в `TestStress` и запустите:

```bash
make stress P=1
```

При расхождении ответов тест упадёт и сохранит проблемный вход в `p1/testdata/failed.in`.
Seed для воспроизведения логируется — укажите `Seed: N` в `StressConfig` для детерминированного прогона.

---

## Все команды Makefile

```
make run P=1               # запустить решение задачи 1 интерактивно
make test P=1              # файловые тесты задачи 1
make test-all              # файловые тесты всех задач
make stress P=1            # стресс-тест задачи 1
make add-test P=1 ID=02    # создать p1/testdata/02.in и 02.out
make run-test P=1 ID=01    # прогнать задачу 1 на p1/testdata/01.in
```