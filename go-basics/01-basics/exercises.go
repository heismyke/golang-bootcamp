// Package basics: YOUR CODE GOES IN THIS FILE.
//
// Each function below is unfinished. Replace the body (the line with
// "TODO") with your own code. Then run the tests from the go-basics folder:
//
//	go test ./01-basics/ -v
//
// Work top to bottom. A test turns from FAIL to PASS when your function is
// correct. Don't edit exercises_test.go.
package basics

// ─────────────────────────────────────────────────────────────
// 1. Variables, strings and formatting
// ─────────────────────────────────────────────────────────────

// Describe returns a sentence like "Myke is 30 years old".
// Hint: fmt.Sprintf("%s is %d years old", ...)  (add "fmt" to the imports)
func Describe(name string, age int) string {
	return "" // TODO
}

// ─────────────────────────────────────────────────────────────
// 2. Number types: float64
// ─────────────────────────────────────────────────────────────

// CelsiusToFahrenheit converts a temperature. Formula: F = C × 9/5 + 32
// Careful: in Go, 9/5 with integers is 1, not 1.8. Use 9.0/5.0.
func CelsiusToFahrenheit(c float64) float64 {
	return 0 // TODO
}

// ─────────────────────────────────────────────────────────────
// 3. Integers, division (/) and remainder (%), multiple return values
// ─────────────────────────────────────────────────────────────

// SplitBill splits a bill in yen evenly. It returns how much each person
// pays and the yen left over. SplitBill(1000, 3) returns (333, 1).
func SplitBill(totalYen, people int) (each int, remainder int) {
	return 0, 0 // TODO
}

// ─────────────────────────────────────────────────────────────
// 4. if / else if / else
// ─────────────────────────────────────────────────────────────

// Grade returns "A" for 90 and above, "B" for 75–89, "C" for 50–74, and
// "F" for anything below 50.
func Grade(score int) string {
	return "" // TODO
}

// ─────────────────────────────────────────────────────────────
// 5. for loops
// ─────────────────────────────────────────────────────────────

// SumTo returns 1 + 2 + ... + n. SumTo(4) = 10. SumTo(0) = 0.
func SumTo(n int) int {
	return 0 // TODO
}

// ─────────────────────────────────────────────────────────────
// 6. Strings, bytes and runes
// ─────────────────────────────────────────────────────────────

// CountVowels counts a, e, i, o, u (upper or lower case).
// Hint: for _, r := range s { ... } loops over characters (runes).
func CountVowels(s string) int {
	return 0 // TODO
}

// Reverse reverses a string. It must also work for Japanese:
// Reverse("こんにちは") == "はちにんこ".
// Hint: convert to []rune first. Why does reversing bytes break Japanese?
func Reverse(s string) string {
	return "" // TODO
}

// ─────────────────────────────────────────────────────────────
// 7. Slices
// ─────────────────────────────────────────────────────────────

// Evens returns only the even numbers, in the same order.
// Hint: var out []int, then out = append(out, n)
func Evens(nums []int) []int {
	return nil // TODO
}

// Max returns the largest number and true. For an empty slice it returns
// 0 and false (there is no maximum).
func Max(nums []int) (int, bool) {
	return 0, false // TODO
}

// ─────────────────────────────────────────────────────────────
// 8. Maps
// ─────────────────────────────────────────────────────────────

// WordCount counts how often each word appears, ignoring case.
// WordCount("Go is fun go") == map[string]int{"go": 2, "is": 1, "fun": 1}
// Hint: strings.Fields and strings.ToLower (add "strings" to the imports)
func WordCount(text string) map[string]int {
	return nil // TODO
}

// ─────────────────────────────────────────────────────────────
// 9. Structs and methods
// ─────────────────────────────────────────────────────────────

type Rectangle struct {
	Width  float64
	Height float64
}

// Area returns width × height.
func (r Rectangle) Area() float64 {
	return 0 // TODO
}

// Perimeter returns the distance around the rectangle.
func (r Rectangle) Perimeter() float64 {
	return 0 // TODO
}

// ─────────────────────────────────────────────────────────────
// 10. Pointers
// ─────────────────────────────────────────────────────────────

// Double doubles the number that n points to. It returns nothing; it
// changes the original variable.
func Double(n *int) {
	// TODO
}

type Counter struct {
	count int
}

// Inc adds 1 to the counter. Notice the receiver is *Counter (a pointer).
// What would happen if it were Counter instead?
func (c *Counter) Inc() {
	// TODO
}

func (c *Counter) Value() int { return c.count }

// ─────────────────────────────────────────────────────────────
// 11. Errors
// ─────────────────────────────────────────────────────────────

// Divide returns a / b, or an error if b is 0.
// Hint: errors.New("cannot divide by zero")  (add "errors" to the imports)
func Divide(a, b int) (int, error) {
	return 0, nil // TODO
}

// ParseAge turns text like "30" into the number 30. It returns an error if
// the text isn't a number, or if the age is negative or over 150.
// Hint: strconv.Atoi
func ParseAge(s string) (int, error) {
	return 0, nil // TODO
}

// ─────────────────────────────────────────────────────────────
// 12. Interfaces
// ─────────────────────────────────────────────────────────────

// Shape is anything with an Area. Rectangle already satisfies it, because
// it has an Area() float64 method. In Go you never write "implements".
type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

// Area returns π × r². Hint: math.Pi
func (c Circle) Area() float64 {
	return 0 // TODO
}

// TotalArea adds up the area of every shape, whatever kind it is.
func TotalArea(shapes []Shape) float64 {
	return 0 // TODO
}
