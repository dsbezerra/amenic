package mathutil

// Max returns the larger of a and b.
func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Min returns the smaller of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// MaxInt64 returns the larger of a and b.
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

// MinInt64 returns the smaller of a and b.
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}

// Clamp returns a value restricted between lo and hi.
func Clamp(v, lo, hi int) int {
	return Min(Max(v, lo), hi)
}

// ClampInt64 returns a value restricted between lo and hi.
func ClampInt64(v, lo, hi int64) int64 {
	return MinInt64(MaxInt64(v, lo), hi)
}
