// DON'T EDIT THIS FILE. These tests check your answers.
package thinking

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// safe runs f and turns a panic into a test failure, so one crashing
// function doesn't stop every other test.
func safe(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s panicked: %v", name, r)
		}
	}()
	f()
}

// ─── Worked example (already passes) ──────────────────────────

func TestWorkedExample(t *testing.T) {
	payments := []Payment{
		{"p5", "alice", 500, 1080}, // unsorted on purpose
		{"p1", "alice", 500, 1000},
		{"p2", "bob", 500, 1010},
		{"p3", "alice", 500, 1030},
		{"p4", "alice", 700, 1035},
		{"p6", "alice", 500, 1140}, // exactly 60s after p5: not a duplicate
	}
	want := []string{"p3", "p5"}
	if got := FindDuplicates(payments); !slices.Equal(got, want) {
		t.Errorf("FindDuplicates = %v, want %v", got, want)
	}
	if got := FindDuplicatesBruteForce(payments); !slices.Equal(got, want) {
		t.Errorf("FindDuplicatesBruteForce = %v, want %v", got, want)
	}
	if payments[0].ID != "p5" {
		t.Error("the caller's slice was reordered")
	}

	// Check the fast version against the simple one on random data.
	r := rand.New(rand.NewPCG(7, 7))
	for range 1000 {
		var ps []Payment
		for i := range r.IntN(30) {
			ps = append(ps, Payment{
				ID:        fmt.Sprintf("p%02d", i),
				Account:   []string{"a", "b"}[r.IntN(2)],
				AmountYen: []int{100, 200}[r.IntN(2)],
				Unix:      int64(r.IntN(300)),
			})
		}
		if a, b := FindDuplicates(ps), FindDuplicatesBruteForce(ps); !slices.Equal(a, b) {
			t.Fatalf("fast and brute force disagree on %v: %v vs %v", ps, a, b)
		}
	}
}

// ─── 1. Receipt ───────────────────────────────────────────────

func line(label, amount string) string { return fmt.Sprintf("%-14s%10s", label, amount) }

func TestReceipt(t *testing.T) {
	got := FormatReceipt([]Item{{"Coffee", 400, 2}, {"Sandwich", 650, 1}})
	want := strings.Join([]string{
		line("Coffee x2", "¥800"),
		line("Sandwich x1", "¥650"),
		strings.Repeat("-", 24),
		line("Subtotal", "¥1,450"),
		line("Tax 10%", "¥145"),
		line("Total", "¥1,595"),
	}, "\n")
	if got != want {
		t.Errorf("basic receipt:\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestReceiptRules(t *testing.T) {
	got := FormatReceipt([]Item{
		{"Strawberry Cake", 1299, 3},
		{"Free sample", 0, 1},
		{"Ghost", 500, 0},
		{"Tea", 333, 1},
	})
	want := strings.Join([]string{
		line("Strawberry x3", "¥3,897"),
		line("Free sampl x1", "¥0"), // "Free sample" is 11 characters, cut to 10
		line("Tea x1", "¥333"),
		strings.Repeat("-", 24),
		line("Subtotal", "¥4,230"),
		line("Tax 10%", "¥423"),
		line("Total", "¥4,653"),
	}, "\n")
	if got != want {
		t.Errorf("long names, qty 0, rounding:\ngot:\n%s\n\nwant:\n%s", got, want)
	}

	big := FormatReceipt([]Item{{"Laptop", 1234567, 1}})
	if !strings.Contains(big, "¥1,234,567") || !strings.Contains(big, "¥123,456") || !strings.Contains(big, "¥1,358,023") {
		t.Errorf("big amounts need several commas, and tax rounds down:\n%s", big)
	}
	for _, items := range [][]Item{nil, {{"Ghost", 100, 0}}} {
		if got := FormatReceipt(items); got != "no items" {
			t.Errorf("FormatReceipt(%v) = %q, want \"no items\"", items, got)
		}
	}
}

// ─── 2. Shift minutes ─────────────────────────────────────────

func TestShift(t *testing.T) {
	good := map[string]int{
		"09:00-17:30":   510,
		"22:00-02:00":   240,
		"00:00-23:59":   1439,
		"23:59-00:00":   1,
		"12:00-11:59":   1439,
		" 09:00-10:00 ": 60,
		"08:15-08:45":   30,
	}
	for in, want := range good {
		got, err := ShiftMinutes(in)
		if err != nil || got != want {
			t.Errorf("ShiftMinutes(%q) = (%d, %v), want (%d, nil)", in, got, err, want)
		}
	}
	bad := []string{
		"09:00-09:00", "9:00-17:00", "09:00-7:00", "25:00-01:00", "24:00-01:00",
		"09:60-10:00", "09:00", "09:00-", "-09:00", "", "aa:bb-cc:dd",
		"09:00 - 10:00", "09:00-10:00-11:00", "09-00-10-00", "0900-1000",
		"-1:00-02:00", "09:0a-10:00", "+9:00-10:00",
	}
	for _, in := range bad {
		if got, err := ShiftMinutes(in); err == nil {
			t.Errorf("ShiftMinutes(%q) = %d with no error, want an error", in, got)
		}
	}
}

// ─── 3. Bank ──────────────────────────────────────────────────

func TestBank(t *testing.T) {
	steps := []struct{ cmd, want string }{
		{"open alice", "ok"},
		{"open bob", "ok"},
		{"open alice", "error: account exists"},
		{"deposit alice 100", "ok"},
		{"withdraw alice 30", "ok"},
		{"balance alice", "alice: 70"},
		{"transfer alice bob 20", "ok"},
		{"balance alice", "alice: 50"},
		{"balance bob", "bob: 20"},
		{"withdraw bob 21", "error: insufficient funds"},
		{"balance bob", "bob: 20"},
		{"transfer alice bob 51", "error: insufficient funds"},
		{"transfer alice alice 5", "error: same account"},
		{"transfer alice carol 5", "error: no such account"},
		{"deposit carol 5", "error: no such account"},
		{"deposit alice -5", "error: invalid amount"},
		{"deposit alice 0", "error: invalid amount"},
		{"deposit alice 1.5", "error: invalid amount"},
		{"deposit alice ten", "error: invalid amount"},
		{"deposit carol ten", "error: no such account"},
		{"deposit alice", "error: wrong number of arguments"},
		{"balance", "error: wrong number of arguments"},
		{"open", "error: wrong number of arguments"},
		{"open a b", "error: wrong number of arguments"},
		{"steal alice 100", "error: unknown command"},
		{"", "error: unknown command"},
		{"withdraw alice 50", "ok"},
		{"balance alice", "alice: 0"},
	}
	cmds := make([]string, len(steps))
	for i, s := range steps {
		cmds[i] = s.cmd
	}
	var got []string
	safe(t, "RunBank", func() { got = RunBank(cmds) })
	if len(got) != len(steps) {
		t.Fatalf("got %d output lines for %d commands", len(got), len(steps))
	}
	for i, s := range steps {
		if got[i] != s.want {
			t.Errorf("command %d %q: got %q, want %q", i+1, s.cmd, got[i], s.want)
		}
	}
}

// ─── 4. Roman numerals ────────────────────────────────────────

func TestRoman(t *testing.T) {
	cases := map[int]string{
		1: "I", 3: "III", 4: "IV", 9: "IX", 14: "XIV", 40: "XL", 90: "XC",
		400: "CD", 444: "CDXLIV", 1994: "MCMXCIV", 2026: "MMXXVI", 3999: "MMMCMXCIX",
	}
	for n, want := range cases {
		if got, err := ToRoman(n); got != want || err != nil {
			t.Errorf("ToRoman(%d) = (%q, %v), want (%q, nil)", n, got, err, want)
		}
		if got, err := FromRoman(want); got != n || err != nil {
			t.Errorf("FromRoman(%q) = (%d, %v), want (%d, nil)", want, got, err, n)
		}
	}
	for _, n := range []int{0, -1, 4000} {
		if _, err := ToRoman(n); err == nil {
			t.Errorf("ToRoman(%d) should be an error", n)
		}
	}
	for n := 1; n <= 3999; n++ {
		s, _ := ToRoman(n)
		if back, err := FromRoman(s); back != n || err != nil {
			t.Fatalf("round trip failed for %d: ToRoman gives %q, FromRoman gives (%d, %v)", n, s, back, err)
		}
	}
	for _, bad := range []string{"", "IIII", "VV", "IC", "IL", "XM", "MCMC", "iv", "ABC", "MMMM", "IXI", "VX", "XXXX", "DD", "IIV"} {
		if got, err := FromRoman(bad); err == nil {
			t.Errorf("FromRoman(%q) = %d with no error, want an error", bad, got)
		}
	}
}

// ─── 5. Buggy ─────────────────────────────────────────────────

func TestBuggyAverage(t *testing.T) {
	safe(t, "Average", func() {
		if got := Average([]int{1, 2}); got != 1.5 {
			t.Errorf("Average([1 2]) = %v, want 1.5", got)
		}
		if got := Average([]int{-3, 3, 6}); got != 2 {
			t.Errorf("Average([-3 3 6]) = %v, want 2", got)
		}
	})
	safe(t, "Average(empty)", func() {
		if got := Average(nil); got != 0 {
			t.Errorf("Average(empty) = %v, want 0", got)
		}
	})
}

func TestBuggyMedian(t *testing.T) {
	safe(t, "Median", func() {
		cases := []struct {
			in   []int
			want float64
		}{{[]int{5, 1, 3}, 3}, {[]int{4, 1, 3, 2}, 2.5}, {[]int{7}, 7}, {nil, 0}, {[]int{10, 20}, 15}}
		for _, c := range cases {
			if got := Median(slices.Clone(c.in)); got != c.want {
				t.Errorf("Median(%v) = %v, want %v", c.in, got, c.want)
			}
		}
	})
	safe(t, "Median (input unchanged)", func() {
		in := []int{3, 1, 2}
		Median(in)
		if !slices.Equal(in, []int{3, 1, 2}) {
			t.Errorf("Median changed the caller's slice to %v", in)
		}
	})
}

func TestBuggyPaginate(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	cases := []struct {
		page, size int
		want       []string
	}{
		{1, 2, []string{"a", "b"}},
		{2, 2, []string{"c", "d"}},
		{3, 2, []string{"e"}},
		{4, 2, []string{}},
		{1, 10, []string{"a", "b", "c", "d", "e"}},
	}
	for _, c := range cases {
		safe(t, fmt.Sprintf("Paginate(page=%d, size=%d)", c.page, c.size), func() {
			if got := Paginate(items, c.page, c.size); !slices.Equal(got, c.want) {
				t.Errorf("Paginate(page=%d, size=%d) = %v, want %v", c.page, c.size, got, c.want)
			}
		})
	}
}

func TestBuggyLeapYear(t *testing.T) {
	cases := map[int]bool{2024: true, 2023: false, 1900: false, 2000: true, 2100: false, 2400: true, 1996: true}
	for y, want := range cases {
		if got := IsLeapYear(y); got != want {
			t.Errorf("IsLeapYear(%d) = %v, want %v", y, got, want)
		}
	}
}

func TestBuggyCamelToSnake(t *testing.T) {
	cases := map[string]string{
		"userName":            "user_name",
		"UserID":              "user_id",
		"HTTPServer":          "http_server",
		"getHTTPResponseCode": "get_http_response_code",
		"id":                  "id",
		"ID":                  "id",
		"simple":              "simple",
		"A":                   "a",
		"parseJSON":           "parse_json",
	}
	for in, want := range cases {
		safe(t, "CamelToSnake", func() {
			if got := CamelToSnake(in); got != want {
				t.Errorf("CamelToSnake(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

// ─── 6. Messy (already passes: keep it passing while you refactor) ───

func TestMessy(t *testing.T) {
	lines := []string{
		"# name,category,amount",
		"lunch, Food, 1200",
		"train,transport,450",
		"",
		"dinner,FOOD,2300",
		"rent,housing,80000",
		"broken line",
		"coffee,food,-50",
		"taxi,transport,abc",
		"snack,,300",
		"bus, Transport ,450",
		"a,b,c,d",
		"book,education,2300",
	}
	want := strings.Join([]string{
		fmt.Sprintf("%-12s %8d (%d)", "housing", 80000, 1),
		fmt.Sprintf("%-12s %8d (%d)", "food", 3500, 2),
		fmt.Sprintf("%-12s %8d (%d)", "education", 2300, 1),
		fmt.Sprintf("%-12s %8d (%d)", "transport", 900, 2),
		fmt.Sprintf("%-12s %8d", "TOTAL", 86700),
		"skipped: 5",
	}, "\n")
	if got := Summarize(lines); got != want {
		t.Errorf("Summarize:\ngot:\n%s\n\nwant:\n%s", got, want)
	}
	if got := Summarize(nil); got != fmt.Sprintf("%-12s %8d\nskipped: 0", "TOTAL", 0) {
		t.Errorf("Summarize(nil) = %q", got)
	}
}
