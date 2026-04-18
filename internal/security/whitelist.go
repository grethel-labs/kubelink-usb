package security

// Whitelist tracks previously approved USB fingerprints.
type Whitelist struct {
	entries map[string]struct{}
}

func NewWhitelist() *Whitelist { return &Whitelist{entries: map[string]struct{}{}} }
func (w *Whitelist) Has(key string) bool {
	_, ok := w.entries[key]
	return ok
}
func (w *Whitelist) Add(key string) { w.entries[key] = struct{}{} }
