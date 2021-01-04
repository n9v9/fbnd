package internal

import "testing"

func TestMax(t *testing.T) {
	type testCase struct {
		name        string
		input       interface{}
		shouldPanic bool
		want        int
	}

	testCases := []testCase{
		{
			name:  "NilSlice",
			input: []int(nil),
			want:  0,
		},
		{
			name:  "EmptySlice",
			input: []int{},
			want:  0,
		},
		{
			name:  "NonEmptySliceMaxIsPositive",
			input: []int{1, 2, -2},
			want:  2,
		},
		{
			name:  "NonEmptySliceMaxIsNegative",
			input: []int{-3, -10, -4},
			want:  -3,
		},
		{
			name:        "NoSlice",
			input:       "This is no slice",
			shouldPanic: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if test.shouldPanic && err == nil {
					t.Fatalf("should panic but did not")
				} else if !test.shouldPanic && err != nil {
					t.Fatalf("should not panic but did: %v", err)
				}
			}()

			got := Max(test.input, func(i int) int { return test.input.([]int)[i] })

			if test.want != got {
				t.Fatalf("want %d, got %d", test.want, got)
			}
		})
	}
}
