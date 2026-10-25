package thinking

// ═════════════════════════════════════════════════════════════════════
// WORKED EXAMPLE: read this whole file before doing the exercises.
// It's complete and its tests pass. It shows the 5-step method from
// GUIDE.md applied to a real problem.
// ═════════════════════════════════════════════════════════════════════
//
// THE PROBLEM (as a product manager might say it):
//
//   "Sometimes our payment app sends the same payment twice when the
//    network is slow. Find the duplicates so we can refund them."
//
// ─── STEP 1: UNDERSTAND ──────────────────────────────────────────────
//
// That sentence is fuzzy. Questions I'd ask (and the answers I got):
//
//   Q: What makes two payments "the same"?
//   A: Same account, same amount, and less than 60 seconds apart.
//
//   Q: Which one is the duplicate?
//   A: The later one. The first one is the real payment.
//
//   Q: Three identical payments 20 seconds apart?
//   A: The 2nd and 3rd are both duplicates.
//
//   Q: Exactly 60 seconds apart?
//   A: Not a duplicate. Strictly less than 60.
//
//   Q: Are the payments sorted by time?
//   A: No guarantee.
//
// Inputs:  a slice of Payment {ID, Account, AmountYen, Unix seconds}
// Output:  the IDs of the duplicates, in the order they happened
//
// Example by hand:
//
//   ID  Account  Amount  Time
//   p1  alice    500     1000
//   p2  bob      500     1010    different account → not a duplicate
//   p3  alice    500     1030    alice, 500, 30s after p1 → DUPLICATE
//   p4  alice    700     1035    different amount → not a duplicate
//   p5  alice    500     1095    65s after p1, but only 65s after p3...
//
// Wait: p5 is 65s after p1 but 65s after p3 too (1095-1030 = 65).
// Not a duplicate. But what if p5 were at 1080? Then it's 50s after p3
// (itself a duplicate). Is that a duplicate? I asked:
//
//   A: Yes. Compare with the previous matching payment, whether it was a
//      duplicate or not. (A slow network can retry several times.)
//
// That question only came up because I worked an example by hand. This is
// why step 1 matters.
//
// Edge cases: empty input, one payment, unsorted input, exactly 60s.
//
// ─── STEP 2: PLAN ────────────────────────────────────────────────────
//
// Simplest idea (brute force):
//   sort payments by time
//   for each payment:
//       look back at earlier payments
//       if the most recent one with the same account+amount is < 60s
//       older → duplicate
//
// That's correct but O(n²): fine for 100 payments, too slow for 1,000,000.
//
// Better idea: I only ever care about "the most recent earlier payment
// with the same account and amount." A map can remember exactly that:
//
//   sort payments by time
//   lastSeen := map from (account, amount) → time of the last payment
//   for each payment, in time order:
//       if key is in lastSeen and time - lastSeen[key] < 60 → duplicate
//       lastSeen[key] = time
//
// Sorting is O(n log n), the loop is O(n). Much better.
//
// Sub-problems:
//   1. sort a COPY by time (don't modify the caller's slice)
//   2. build a key from account + amount
//   3. the loop
//
// ─── STEPS 3 & 4: IMPLEMENT AND TEST ─────────────────────────────────
// Below. I wrote the brute force first, tested it, then wrote the fast
// version and tested that both give the same answers (see the test file).

import (
	"cmp"
	"slices"
)

type Payment struct {
	ID        string
	Account   string
	AmountYen int
	Unix      int64 // seconds
}

const duplicateWindowSeconds = 60

// FindDuplicatesBruteForce is the first, simple version. Keeping a simple
// correct version around is useful: tests can check the fast version
// against it.
func FindDuplicatesBruteForce(payments []Payment) []string {
	sorted := sortedByTime(payments)
	var dups []string
	for i, p := range sorted {
		// Walk backwards to find the most recent matching payment.
		for j := i - 1; j >= 0; j-- {
			prev := sorted[j]
			if prev.Account == p.Account && prev.AmountYen == p.AmountYen {
				if p.Unix-prev.Unix < duplicateWindowSeconds {
					dups = append(dups, p.ID)
				}
				break // only the most recent match counts
			}
		}
	}
	return dups
}

// FindDuplicates is the fast version: O(n log n).
func FindDuplicates(payments []Payment) []string {
	type key struct {
		account string
		amount  int
	}
	lastSeen := make(map[key]int64)

	var dups []string
	for _, p := range sortedByTime(payments) {
		k := key{p.Account, p.AmountYen}
		if last, ok := lastSeen[k]; ok && p.Unix-last < duplicateWindowSeconds {
			dups = append(dups, p.ID)
		}
		lastSeen[k] = p.Unix
	}
	return dups
}

// sortedByTime returns a sorted copy, leaving the caller's slice untouched.
// Ties are broken by ID so the result is always the same.
func sortedByTime(payments []Payment) []Payment {
	sorted := slices.Clone(payments)
	slices.SortFunc(sorted, func(a, b Payment) int {
		return cmp.Or(cmp.Compare(a.Unix, b.Unix), cmp.Compare(a.ID, b.ID))
	})
	return sorted
}

// ─── STEP 5: REFLECT ─────────────────────────────────────────────────
//
// - Using a struct as a map key is a neat trick: any comparable struct works.
// - "Remember the last X per group" → map. I'll see that pattern again
//   (it's the core of the sliding-window and TwoSum problems in 04-dsa).
// - Asking questions in step 1 found a rule (compare with the previous
//   duplicate too) that I would otherwise have gotten wrong.
// - Writing the brute force first was worth it: the test compares both
//   on 1,000 random cases, which caught a bug in my first fast version
//   (I had forgotten to update lastSeen for duplicates).
