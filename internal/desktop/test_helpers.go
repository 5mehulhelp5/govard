package desktop

// ResetStateForTest clears process-level caches used by desktop package.
func ResetStateForTest() {
	prefsMu.Lock()
	cachedPrefs = nil
	prefsMu.Unlock()
}
