package internal

import (
	"fmt"
	"reflect"
)

// Max calls f for each element in slice.
// f should return a number indicating the length of the value at slice[i].
// Max then returns the max value returned by f.
// If slice is empty then 0 is returned.
func Max(slice interface{}, f func(i int) int) int {
	value := reflect.ValueOf(slice)

	if kind := value.Kind(); kind != reflect.Slice {
		panic(fmt.Sprintf("expected parameter slice to be of kind %s but got %s", reflect.Slice, kind))
	}

	var (
		max      *int
		sliceLen = value.Len()
	)
	for i := 0; i < sliceLen; i++ {
		v := f(i)
		if max == nil || v > *max {
			max = &v
		}
	}

	if max == nil {
		return 0
	}

	return *max
}
