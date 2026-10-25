package dsa

// ═════════════════════════════════════════════════════════════
// 10. DESIGNING DATA STRUCTURES
// These combine the basics into structures that real backends use.
// ═════════════════════════════════════════════════════════════

// LRUCache holds up to `capacity` keys. When full, Put evicts the Least
// Recently Used key. Both Get and Put count as "using" a key.
// Target: Get and Put in O(1). Use a map (key → node) plus a doubly linked
// list ordered by recent use. Build the list yourself (no container/list).
// The test does 300,000 operations, so scanning a slice to evict will time out.
// This is how in-memory caches (and Redis's eviction policy) work.
type LRUCache struct {
	// TODO: add fields
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{} // TODO
}

// Get returns the value and true, or false if the key isn't cached.
func (c *LRUCache) Get(key int) (int, bool) {
	return 0, false // TODO
}

// Put inserts or updates a key, evicting the least recently used one if
// over capacity.
func (c *LRUCache) Put(key, value int) {
	// TODO
}

// Trie (prefix tree) stores words so you can quickly find all words with a
// given prefix. Each node has children (one per next letter) and a flag
// marking where a word ends. Search-box autocomplete works this way.
type Trie struct {
	// TODO: add fields
}

func NewTrie() *Trie {
	return &Trie{} // TODO
}

func (t *Trie) Insert(word string) {
	// TODO
}

// Search reports whether the exact word was inserted.
func (t *Trie) Search(word string) bool {
	return false // TODO
}

// StartsWith reports whether any inserted word starts with prefix.
func (t *Trie) StartsWith(prefix string) bool {
	return false // TODO
}

// WithPrefix returns all inserted words starting with prefix, in
// alphabetical order (autocomplete).
func (t *Trie) WithPrefix(prefix string) []string {
	return nil // TODO
}
