package cmd

// FindWailsCLIForTest exposes Wails binary discovery for external tests.
func FindWailsCLIForTest() (string, error) {
	return findWailsCLI()
}
