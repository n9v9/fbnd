package main

import "testing"

func TestMax(t *testing.T) {
	type testCase struct {
		name  string
		input []int
		want  int
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
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			if got := Max(test.input, func(i *int) int { return *i }); test.want != got {
				t.Fatalf("want %d, got %d", test.want, got)
			}
		})
	}
}
