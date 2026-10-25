// DON'T EDIT THIS FILE. These tests check your answers in exercises.go.
package basics

import (
	"math"
	"maps"
	"slices"
	"testing"
)

func Test01_Describe(t *testing.T) {
	if got := Describe("Myke", 30); got != "Myke is 30 years old" {
		t.Errorf(`Describe("Myke", 30) = %q, want "Myke is 30 years old"`, got)
	}
}

func Test02_CelsiusToFahrenheit(t *testing.T) {
	cases := map[float64]float64{0: 32, 100: 212, -40: -40, 37: 98.6}
	for c, want := range cases {
		if got := CelsiusToFahrenheit(c); math.Abs(got-want) > 0.001 {
			t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", c, got, want)
		}
	}
}

func Test03_SplitBill(t *testing.T) {
	cases := []struct{ total, people, each, rem int }{
		{1000, 3, 333, 1},
		{900, 3, 300, 0},
		{5, 10, 0, 5},
	}
	for _, c := range cases {
		each, rem := SplitBill(c.total, c.people)
		if each != c.each || rem != c.rem {
			t.Errorf("SplitBill(%d, %d) = (%d, %d), want (%d, %d)", c.total, c.people, each, rem, c.each, c.rem)
		}
	}
}

func Test04_Grade(t *testing.T) {
	cases := map[int]string{100: "A", 90: "A", 89: "B", 75: "B", 74: "C", 50: "C", 49: "F", 0: "F"}
	for score, want := range cases {
		if got := Grade(score); got != want {
			t.Errorf("Grade(%d) = %q, want %q", score, got, want)
		}
	}
}

func Test05_SumTo(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 4: 10, 100: 5050}
	for n, want := range cases {
		if got := SumTo(n); got != want {
			t.Errorf("SumTo(%d) = %d, want %d", n, got, want)
		}
	}
}

func Test06_CountVowels(t *testing.T) {
	cases := map[string]int{"": 0, "hello": 2, "AEIOU": 5, "rhythm": 0, "Golang Is Fun": 4}
	for s, want := range cases {
		if got := CountVowels(s); got != want {
			t.Errorf("CountVowels(%q) = %d, want %d", s, got, want)
		}
	}
}

func Test06_Reverse(t *testing.T) {
	cases := map[string]string{"": "", "abc": "cba", "Go!": "!oG", "こんにちは": "はちにんこ"}
	for s, want := range cases {
		if got := Reverse(s); got != want {
			t.Errorf("Reverse(%q) = %q, want %q", s, got, want)
		}
	}
}

func Test07_Evens(t *testing.T) {
	got := Evens([]int{1, 2, 3, 4, 5, 6, -2})
	want := []int{2, 4, 6, -2}
	if !slices.Equal(got, want) {
		t.Errorf("Evens = %v, want %v", got, want)
	}
	if got := Evens([]int{1, 3}); len(got) != 0 {
		t.Errorf("Evens([1 3]) = %v, want empty", got)
	}
}

func Test07_Max(t *testing.T) {
	if got, ok := Max([]int{3, 9, -1, 7}); got != 9 || !ok {
		t.Errorf("Max([3 9 -1 7]) = (%d, %v), want (9, true)", got, ok)
	}
	if got, ok := Max([]int{-5, -2}); got != -2 || !ok {
		t.Errorf("Max([-5 -2]) = (%d, %v), want (-2, true). Hint: don't start from 0", got, ok)
	}
	if got, ok := Max(nil); got != 0 || ok {
		t.Errorf("Max(empty) = (%d, %v), want (0, false)", got, ok)
	}
}

func Test08_WordCount(t *testing.T) {
	got := WordCount("Go is fun go  IS great")
	want := map[string]int{"go": 2, "is": 2, "fun": 1, "great": 1}
	if !maps.Equal(got, want) {
		t.Errorf("WordCount = %v, want %v", got, want)
	}
}

func Test09_Rectangle(t *testing.T) {
	r := Rectangle{Width: 3, Height: 4}
	if got := r.Area(); got != 12 {
		t.Errorf("Area() = %v, want 12", got)
	}
	if got := r.Perimeter(); got != 14 {
		t.Errorf("Perimeter() = %v, want 14", got)
	}
}

func Test10_Double(t *testing.T) {
	n := 21
	Double(&n)
	if n != 42 {
		t.Errorf("after Double, n = %d, want 42", n)
	}
}

func Test10_Counter(t *testing.T) {
	var c Counter
	c.Inc()
	c.Inc()
	c.Inc()
	if c.Value() != 3 {
		t.Errorf("after 3 Inc() calls, Value() = %d, want 3", c.Value())
	}
}

func Test11_Divide(t *testing.T) {
	if got, err := Divide(10, 2); got != 5 || err != nil {
		t.Errorf("Divide(10, 2) = (%d, %v), want (5, nil)", got, err)
	}
	if _, err := Divide(1, 0); err == nil {
		t.Error("Divide(1, 0) returned no error, want an error")
	}
}

func Test11_ParseAge(t *testing.T) {
	if got, err := ParseAge("30"); got != 30 || err != nil {
		t.Errorf(`ParseAge("30") = (%d, %v), want (30, nil)`, got, err)
	}
	for _, bad := range []string{"abc", "", "-1", "200", "3.5"} {
		if _, err := ParseAge(bad); err == nil {
			t.Errorf("ParseAge(%q) returned no error, want an error", bad)
		}
	}
}

func Test12_Shapes(t *testing.T) {
	if got := (Circle{Radius: 2}).Area(); math.Abs(got-4*math.Pi) > 0.001 {
		t.Errorf("Circle{2}.Area() = %v, want %v", got, 4*math.Pi)
	}
	shapes := []Shape{Rectangle{Width: 2, Height: 5}, Circle{Radius: 1}, Rectangle{Width: 1, Height: 1}}
	if got, want := TotalArea(shapes), 11+math.Pi; math.Abs(got-want) > 0.001 {
		t.Errorf("TotalArea = %v, want %v", got, want)
	}
}
