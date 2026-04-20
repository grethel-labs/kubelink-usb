package security

// Whitelist tracks previously approved USB fingerprints.
type Whitelist struct {
	entries map[string]struct{}
}

// NewWhitelist returns an empty Whitelist ready for use.
func NewWhitelist() *Whitelist { return &Whitelist{entries: map[string]struct{}{}} }

// Has reports whether the given fingerprint is in the whitelist.
func (w *Whitelist) Has(key string) bool {
	_, ok := w.entries[key]
	return ok
}

// Add inserts a fingerprint into the whitelist.
func (w *Whitelist) Add(key string) { w.entries[key] = struct{}{} }
