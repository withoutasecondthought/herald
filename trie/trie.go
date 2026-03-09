package trie

// BotEntry represents a known bot matched by the trie.
type BotEntry struct {
	Pattern  string
	Name     string
	Owner    string
	Category string
}

// node is a single node in the byte-level trie.
// Uses map[byte]*node instead of [256]*node for memory efficiency.
type node struct {
	children map[byte]*node
	entry    *BotEntry // non-nil means this node is a terminal
}

func newNode() *node {
	return &node{children: make(map[byte]*node, 4)} //nolint:mnd // initial map capacity
}

// Trie is a byte-level prefix trie for fast bot pattern matching.
type Trie struct {
	root *node
}

// New creates an empty trie.
func New() *Trie {
	return &Trie{root: newNode()}
}

// Insert adds a pattern and its associated bot entry to the trie.
func (t *Trie) Insert(pattern string, entry *BotEntry) {
	cur := t.root

	for i := range len(pattern) {
		b := pattern[i]

		child, ok := cur.children[b]
		if !ok {
			child = newNode()
			cur.children[b] = child
		}

		cur = child
	}

	cur.entry = entry
}

// Search scans the UA string for any known bot pattern.
// It tries to match from every position in the string and returns
// the first match found. Returns nil, false if no match.
func (t *Trie) Search(ua string) (*BotEntry, bool) {
	n := len(ua)

	for i := range n {
		if entry := t.matchAt(ua, i, n); entry != nil {
			return entry, true
		}
	}

	return nil, false
}

// matchAt tries to match a pattern starting at position start in ua.
// Returns the longest matching entry, or nil if no match.
func (t *Trie) matchAt(ua string, start, n int) *BotEntry {
	cur := t.root

	var matched *BotEntry

	for i := start; i < n; i++ {
		child, ok := cur.children[ua[i]]
		if !ok {
			break
		}

		cur = child

		if cur.entry != nil {
			matched = cur.entry
		}
	}

	return matched
}
