package internal

// Max calls f for each element in slice and returns the max value returned by f.
// If slice is empty then 0 is returned.
func Max[T any](slice []T, f func(value *T) int) int {
	var max *int

	for i := 0; i < len(slice); i++ {
		if v := f(&slice[i]); max == nil || v > *max {
			max = &v
		}
	}

	if max == nil {
		return 0
	}

	return *max
}
