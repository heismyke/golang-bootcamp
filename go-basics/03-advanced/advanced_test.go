// DON'T EDIT THIS FILE. These tests check your answers.
package advanced

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// within fails the test if f doesn't finish in time, so a deadlock gives a
// helpful message instead of hanging forever.
func within(t *testing.T, d time.Duration, hint string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("timed out after %v. %s", d, hint)
	}
}

func gen(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

// ─── Part 1: Interfaces ───────────────────────────────────────

func TestInterfaces01_Stringer(t *testing.T) {
	cases := map[Yen]string{0: "¥0", 999: "¥999", 1000: "¥1,000", 1234567: "¥1,234,567"}
	for y, want := range cases {
		if got := fmt.Sprint(y); got != want {
			t.Errorf("fmt.Sprint(Yen(%d)) = %q, want %q", int64(y), got, want)
		}
	}
}

func TestInterfaces02_CustomError(t *testing.T) {
	if err := FindExpense(1); err != nil {
		t.Errorf("FindExpense(1) = %v, want nil", err)
	}
	err := fmt.Errorf("handler: %w", FindExpense(7)) // wrapped, like real code
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("errors.As could not find a *NotFoundError in %v", err)
	}
	if nf.ID != 7 || nf.Resource != "expense" || nf.Error() != "expense 7 not found" {
		t.Errorf("got %+v with message %q, want ID 7, Resource expense, message \"expense 7 not found\"", nf, nf.Error())
	}
}

func TestInterfaces03_TypeSwitch(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{5, "int: 5"},
		{"hi", "string: hi"},
		{[]int{1, 2, 3}, "slice of 3 ints"},
		{errors.New("boom"), "error: boom"},
		{nil, "nil"},
		{3.14, "unknown"},
	}
	for _, c := range cases {
		if got := TypeName(c.in); got != c.want {
			t.Errorf("TypeName(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestInterfaces04_Writer(t *testing.T) {
	var buf bytes.Buffer
	w := UpperWriter{W: &buf}
	n, err := fmt.Fprintf(w, "hello %s", "go")
	if err != nil || n != 8 {
		t.Errorf("Fprintf returned (%d, %v), want (8, nil)", n, err)
	}
	if buf.String() != "HELLO GO" {
		t.Errorf("wrote %q, want %q", buf.String(), "HELLO GO")
	}
}

func TestInterfaces05_Reader(t *testing.T) {
	data, err := io.ReadAll(&RepeatReader{B: 'a', N: 5000})
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if len(data) != 5000 || strings.Trim(string(data), "a") != "" {
		t.Errorf("got %d bytes, want 5000 bytes of 'a' (ReadAll calls Read many times with small buffers)", len(data))
	}
	if data, _ := io.ReadAll(&RepeatReader{B: 'x', N: 0}); len(data) != 0 {
		t.Errorf("N=0 should produce nothing, got %d bytes", len(data))
	}
}

type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }

func TestInterfaces06_Clock(t *testing.T) {
	at := func(h, m int) fakeClock { return fakeClock{time.Date(2026, 1, 1, h, m, 0, 0, time.UTC)} }
	cases := []struct {
		clock fakeClock
		want  string
	}{
		{at(6, 0), "Good morning"}, {at(11, 59), "Good morning"},
		{at(12, 0), "Good afternoon"}, {at(17, 59), "Good afternoon"},
		{at(18, 0), "Good evening"}, {at(23, 30), "Good evening"},
	}
	for _, c := range cases {
		if got := Greeting(c.clock); got != c.want {
			t.Errorf("Greeting at %s = %q, want %q", c.clock.t.Format("15:04"), got, c.want)
		}
	}
}

// ─── Part 2: Generics ─────────────────────────────────────────

func TestGenerics07_Map(t *testing.T) {
	if got := Map([]int{1, 2, 3}, strconv.Itoa); !slices.Equal(got, []string{"1", "2", "3"}) {
		t.Errorf("Map ints to strings = %v", got)
	}
	if got := Map([]string{"go", "rust"}, func(s string) int { return len(s) }); !slices.Equal(got, []int{2, 4}) {
		t.Errorf("Map strings to lengths = %v", got)
	}
}

func TestGenerics08_Filter(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	if !slices.Equal(got, []int{2, 4, 6}) {
		t.Errorf("Filter evens = %v, want [2 4 6]", got)
	}
}

func TestGenerics09_Reduce(t *testing.T) {
	if got := Reduce([]int{1, 2, 3, 4}, 0, func(a, n int) int { return a + n }); got != 10 {
		t.Errorf("Reduce sum = %d, want 10", got)
	}
	got := Reduce([]string{"a", "b", "c"}, "", func(a, s string) string { return a + s })
	if got != "abc" {
		t.Errorf("Reduce concat = %q, want abc", got)
	}
}

func TestGenerics10_Stack(t *testing.T) {
	var s Stack[string]
	if _, ok := s.Pop(); ok {
		t.Error("Pop on an empty stack returned ok = true")
	}
	s.Push("a")
	s.Push("b")
	if s.Len() != 2 {
		t.Errorf("Len = %d, want 2", s.Len())
	}
	if v, ok := s.Pop(); v != "b" || !ok {
		t.Errorf("Pop = (%q, %v), want (b, true)", v, ok)
	}
	if v, ok := s.Pop(); v != "a" || !ok {
		t.Errorf("Pop = (%q, %v), want (a, true)", v, ok)
	}
	if s.Len() != 0 {
		t.Errorf("Len after popping everything = %d, want 0", s.Len())
	}
}

func TestGenerics11_Sum(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum ints = %d, want 6", got)
	}
	if got := Sum([]float64{0.5, 0.25}); got != 0.75 {
		t.Errorf("Sum floats = %v, want 0.75", got)
	}
	if got := Sum([]Yen{100, 250}); got != 350 {
		t.Errorf("Sum Yen = %v, want ¥350", int64(got))
	}
}

// ─── Part 3: Concurrency ──────────────────────────────────────

func TestConcurrency12_FetchAll(t *testing.T) {
	urls := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	slowFetch := func(u string) string { time.Sleep(100 * time.Millisecond); return strings.ToUpper(u) }

	within(t, 3*time.Second, "FetchAll never returned. Did you call wg.Wait() without wg.Add()?", func() {
		start := time.Now()
		got := FetchAll(urls, slowFetch)
		elapsed := time.Since(start)
		if want := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}; !slices.Equal(got, want) {
			t.Errorf("FetchAll = %v, want %v (same order as input)", got, want)
		}
		if elapsed > 400*time.Millisecond {
			t.Errorf("took %v. Fetches must run concurrently (about 100ms total, not 1s)", elapsed)
		}
	})
}

func TestConcurrency13_SafeCounter(t *testing.T) {
	c := NewSafeCounter()
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			for range 100 {
				c.Inc("hits")
			}
		})
	}
	wg.Wait()
	if got := c.Value("hits"); got != 10_000 {
		t.Errorf("Value = %d, want 10000 (100 goroutines × 100 increments)", got)
	}
	if got := c.Value("missing"); got != 0 {
		t.Errorf("Value(missing) = %d, want 0", got)
	}
}

func TestConcurrency14_Count(t *testing.T) {
	within(t, time.Second, "Count's channel never closed, or Count returned nil. Did you close(ch) after sending?", func() {
		var got []int
		for v := range Count(5) {
			got = append(got, v)
		}
		if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
			t.Errorf("Count(5) produced %v, want [1 2 3 4 5]", got)
		}
	})
}

func TestConcurrency15_Square(t *testing.T) {
	within(t, time.Second, "Square's output never closed. Close it after the input loop ends.", func() {
		var got []int
		for v := range Square(gen(1, 2, 3, 4)) {
			got = append(got, v)
		}
		if !slices.Equal(got, []int{1, 4, 9, 16}) {
			t.Errorf("Square produced %v, want [1 4 9 16]", got)
		}
	})
}

func TestConcurrency16_Merge(t *testing.T) {
	within(t, time.Second, "Merge's output never closed. Close it only after ALL inputs are drained.", func() {
		var got []int
		for v := range Merge(gen(1, 2, 3), gen(10, 20), gen(100)) {
			got = append(got, v)
		}
		slices.Sort(got)
		if !slices.Equal(got, []int{1, 2, 3, 10, 20, 100}) {
			t.Errorf("Merge produced %v (sorted), want [1 2 3 10 20 100]", got)
		}
	})
}

func TestConcurrency17_WithTimeout(t *testing.T) {
	within(t, 2*time.Second, "WithTimeout hung.", func() {
		v, err := WithTimeout(func() int { return 42 }, 100*time.Millisecond)
		if v != 42 || err != nil {
			t.Errorf("fast function: got (%d, %v), want (42, nil)", v, err)
		}

		start := time.Now()
		_, err = WithTimeout(func() int { time.Sleep(500 * time.Millisecond); return 1 }, 50*time.Millisecond)
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("slow function: err = %v, want ErrTimeout", err)
		}
		if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
			t.Errorf("slow function: returned after %v, should give up after about 50ms", elapsed)
		}
	})
}

func TestConcurrency18_WorkerPool(t *testing.T) {
	jobs := make([]int, 20)
	for i := range jobs {
		jobs[i] = i + 1
	}
	var running, maxRunning atomic.Int32
	work := func(n int) int {
		cur := running.Add(1)
		for {
			m := maxRunning.Load()
			if cur <= m || maxRunning.CompareAndSwap(m, cur) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		running.Add(-1)
		return n * n
	}

	within(t, 3*time.Second, "ProcessJobs hung. Did you close the jobs channel so workers can exit?", func() {
		start := time.Now()
		got := ProcessJobs(jobs, 4, work)
		elapsed := time.Since(start)

		want := Map(jobs, func(n int) int { return n * n })
		if !slices.Equal(got, want) {
			t.Errorf("results = %v, want %v (same order as jobs)", got, want)
		}
		if m := maxRunning.Load(); m > 4 {
			t.Errorf("%d jobs ran at once, the limit is 4 workers", m)
		} else if m < 2 {
			t.Errorf("only %d job ran at a time. Jobs should run in parallel", m)
		}
		if elapsed > 300*time.Millisecond {
			t.Errorf("took %v. 20 jobs × 20ms over 4 workers should take about 100ms", elapsed)
		}
	})
}

func TestConcurrency19_Produce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := Produce(ctx)

	within(t, time.Second, "Produce didn't send 0, 1, 2, ... (or returned nil).", func() {
		for want := range 5 {
			got, ok := <-ch
			if !ok || got != want {
				t.Errorf("received (%d, %v), want (%d, true)", got, ok, want)
				return
			}
		}
	})
	cancel()
	within(t, time.Second, "after cancel(), Produce must stop and close its channel (goroutine leak!).", func() {
		for range ch {
		}
	})
}

func TestConcurrency20_Lazy(t *testing.T) {
	var calls atomic.Int32
	l := &Lazy[string]{New: func() string {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		return "config loaded"
	}}

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if got := l.Get(); got != "config loaded" {
				t.Errorf("Get() = %q, want %q", got, "config loaded")
			}
		})
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Errorf("New was called %d times, want exactly 1", n)
	}
}
