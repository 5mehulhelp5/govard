package desktop

import "govard/internal/engine"

// ResetStateForTest clears process-level caches used by desktop package.
func ResetStateForTest() {
	prefsMu.Lock()
	cachedPrefs = nil
	prefsMu.Unlock()
}

// ResolveRequestedLogTargetsForTest exposes log target normalization for tests.
func ResolveRequestedLogTargetsForTest(service string, discovered []string) []string {
	return resolveRequestedLogTargets(service, discovered)
}

// PrefixServiceLogLinesForTest exposes service-prefix formatting for tests.
func PrefixServiceLogLinesForTest(service string, raw string) string {
	return prefixServiceLogLines(service, raw)
}

// BuildOperationNotificationForTest exposes operation notification formatting for tests.
func BuildOperationNotificationForTest(event engine.OperationEvent) (OperationNotification, bool) {
	return buildOperationNotification(event)
}

// SelectOperationEventsSinceForTest exposes operation event cursor logic for tests.
func SelectOperationEventsSinceForTest(events []engine.OperationEvent, cursor string) ([]engine.OperationEvent, string) {
	return selectOperationEventsSince(events, cursor)
}

// OperationEventSignatureForTest exposes operation event signature generation for tests.
func OperationEventSignatureForTest(event engine.OperationEvent) string {
	return operationEventSignature(event)
}

// CalculateCPUPercentForTest exposes CPU percentage math for tests.
func CalculateCPUPercentForTest(
	currentUsage uint64,
	previousUsage uint64,
	currentSystem uint64,
	previousSystem uint64,
	onlineCPUs uint32,
	perCPUCount int,
) float64 {
	return calculateCPUPercentFromDeltas(
		currentUsage,
		previousUsage,
		currentSystem,
		previousSystem,
		onlineCPUs,
		perCPUCount,
	)
}

// BuildMetricsWarningsForTest exposes metrics warning generation for tests.
func BuildMetricsWarningsForTest(projects []ProjectResourceMetric, input []string) []string {
	return buildMetricsWarnings(projects, input)
}

// BytesToMBForTest exposes bytes-to-MB conversion for tests.
func BytesToMBForTest(bytes uint64) float64 {
	return bytesToMB(bytes)
}
