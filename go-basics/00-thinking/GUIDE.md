# How to think like a programmer

Programming isn't typing code. It's **breaking a fuzzy problem into small,
precise steps**, then writing those steps down in a language a computer
understands. The typing is the easy part. This guide gives you a method to
follow every time, until it becomes automatic.

---

## The 5-step method

Use it for every exercise, every bug and every feature. Write the steps
down as comments before you write code.

### 1. Understand: what exactly is being asked?

You can't solve a problem you can't state. Most beginner mistakes happen
here, before any code exists.

- **Restate the problem in your own words.** If you can't, re-read it.
- **What are the inputs?** Types, ranges. Can they be empty? Negative? Huge?
- **What is the output?** Exact format. What happens on bad input?
- **Work 2–3 examples by hand**, on paper, including one weird one.
- **List the edge cases**: empty input, one element, duplicates, zero,
  negative, maximum size, invalid format, unicode.

> If you can't solve an example by hand, you can't tell a computer how to.

### 2. Plan: solve it in plain language first

- **Start with the simplest thing that could work** (brute force). Make it
  correct first, then make it fast.
- **Break it into sub-problems.** Each sub-problem becomes a function. If a
  step feels hard, break it down again.
- **Write pseudocode as comments:**

```go
// for each transaction:
//     find earlier transactions on the same account
//     if any has the same amount and is less than 60s older → duplicate
```

- **Choose data structures deliberately.** "I need to look things up by
  key" → map. "I need them in order" → slice plus sort. "I need the
  smallest one repeatedly" → heap.

### 3. Implement: small steps, run constantly

- Turn one comment into code, then **run it immediately**.
- Never write 50 lines before running anything. When it breaks, you won't
  know which line broke it.
- Print things (`fmt.Printf("%+v\n", x)`) to see what's really happening, not
  what you assume is happening.

### 4. Test: try to break your own code

- Run your hand-worked examples from step 1.
- Run every edge case from your list.
- Ask: "what input would make this crash or give a wrong answer?" Then try it.

### 5. Reflect: improve and learn

- Can it be simpler? Clearer names? Smaller functions?
- What's the time complexity? Can it be faster if needed?
- **What did you learn that applies to other problems?** Write it in your
  journal. This is how pattern recognition is built.

---

## When you're stuck

Being stuck is normal. It's most of the job. What matters is having a
process for getting unstuck:

1. **Solve a smaller version.** Can't handle a list of 1,000? Do it for 1
   element, then 2, then 3. Watch the pattern.
2. **Go back to examples.** Work one by hand, slowly, and write down every
   single micro-step you take. Those steps are your algorithm.
3. **Change the representation.** Draw it. Sort it first. Count things
   first. Put it in a map first.
4. **Explain it out loud** (rubber-duck debugging). Explaining forces you
   to notice the gap in your understanding.
5. **Take a break.** Seriously. Your brain keeps working on it.
6. **Time-box it.** Stuck for 30+ minutes with no progress? Ask for a hint.
   Then rewrite the solution yourself the next day, without looking.

## How to debug

Debugging is the scientific method applied to code:

1. **Reproduce it.** Find an input that always triggers the bug. Write it as
   a failing test.
2. **Read the error message, all of it.** Go tells you the file, the line,
   and often exactly what's wrong. `panic: index out of range [5] with
   length 5` means you read one past the end.
3. **Make a hypothesis.** "I think `end` is one too big when the page is the last page."
4. **Test the hypothesis** with a print, a debugger, or a smaller test.
5. **Fix it, then check the fix with the test** from step 1. Keep the test
   so the bug never returns.

Never change code randomly hoping it works. If you can't explain why the
fix works, you haven't found the bug.

## How to read code you didn't write

1. Start at the entry point (`main`, or the handler) and follow one path.
2. Read function names and signatures first; read bodies only when needed.
3. Run it. Change something small and see what happens.
4. Write down what each piece does, in one sentence.

## Habits of good programmers

- **Names are documentation.** `remainingBalance` beats `rb`. `isExpired` beats `flag`.
- **Small functions that do one thing.** If you need "and" to describe it, split it.
- **Handle errors where they happen.** Don't ignore them; decide what should happen.
- **Make it work, make it right, make it fast**, in that order.
- **Delete code.** The simplest solution that meets the requirements wins.
- **Read error messages and documentation before asking.**

---

## Exercises in this folder

1. **Read `worked_example.go` first.** It shows the whole method applied to
   one problem, with every thinking step written as comments.
2. Then do the exercises in `exercises.go`, `buggy.go` and `messy.go`:

```bash
cd ~/go-basics
go test ./00-thinking/ -v -run Receipt     # one at a time
```

| Exercise | Skill | File |
|---|---|---|
| 1. `FormatReceipt` | breaking a problem into small functions | `exercises.go` |
| 2. `ShiftMinutes` | finding edge cases | `exercises.go` |
| 3. `RunBank` | modelling state and reading a spec carefully | `exercises.go` |
| 4. `ToRoman` / `FromRoman` | finding the pattern; validating cleverly | `exercises.go` |
| 5. Fix 5 bugs | debugging method | `buggy.go` |
| 6. Refactor | making working code clean without breaking it | `messy.go` |

For exercises 1–4, **write steps 1 and 2 of the method as comments above
the function before writing any code.** When you show me your solution,
I'll review the comments as much as the code.
