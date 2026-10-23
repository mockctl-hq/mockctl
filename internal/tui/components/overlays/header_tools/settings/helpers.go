package settings

import "time"

// --- Timeout index helpers ---

var timeoutValues = []time.Duration{5 * time.Second, 15 * time.Second, 30 * time.Second, 60 * time.Second}

func timeoutToIndex(d time.Duration) int {
	for i, v := range timeoutValues {
		if d == v {
			return i
		}
	}
	return 1 // default 15s
}

func indexToTimeout(i int) time.Duration {
	if i >= 0 && i < len(timeoutValues) {
		return timeoutValues[i]
	}
	return 15 * time.Second
}

// --- History limit helpers ---

var historyLimitValues = []int{100, 1000, 5000}

func historyLimitToIndex(n int) int {
	for i, v := range historyLimitValues {
		if n == v {
			return i
		}
	}
	return 1 // default 1000
}

func indexToHistoryLimit(i int) int {
	if i >= 0 && i < len(historyLimitValues) {
		return historyLimitValues[i]
	}
	return 1000
}

// --- Tab size helpers ---

var tabSizeValues = []int{2, 4, 8}

func tabSizeToIndex(n int) int {
	for i, v := range tabSizeValues {
		if n == v {
			return i
		}
	}
	return 0 // default 2
}

func indexToTabSize(i int) int {
	if i >= 0 && i < len(tabSizeValues) {
		return tabSizeValues[i]
	}
	return 2
}
