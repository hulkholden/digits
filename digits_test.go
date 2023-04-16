package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_solve(t *testing.T) {
	tests := map[string]struct {
		target    int
		digits    []int
		wantLen   int
		wantFirst string
	}{
		"a": {
			digits:    []int{5, 7, 9, 10, 15, 25},
			target:    93,
			wantLen:   146,
			wantFirst: "(5 + (((25 * (10 + 15)) - 9) / 7))",
		},
		"b": {

			digits:    []int{4, 5, 7, 8, 15, 20},
			target:    113,
			wantFirst: "(4 + (5 + (8 * (20 - 7))))",
			wantLen:   332,
			// TODO: this is correct but a much simpler solution is (8*15)-7.
		},
		"c": {
			digits:    []int{3, 4, 6, 9, 11, 15},
			target:    205,
			wantLen:   88,
			wantFirst: "(3 + (4 + (11 * ((9 + 15) - 6))))",
		},
		"d": {
			digits:    []int{3, 5, 9, 11, 23, 25},
			target:    351,
			wantLen:   98,
			wantFirst: "(3 * ((23 + (9 * 11)) - 5))",
		},
		"f": {
			digits:    []int{24, 8, 10, 20, 5, 15},
			target:    497,
			wantLen:   31,
			wantFirst: "((10 + (15 + (20 * 24))) - 8)",
		},
	}
	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			got := solve(tt.target, tt.digits)

			if len(got) != tt.wantLen {
				t.Fatalf("solve() got %d results, want %d", len(got), tt.wantLen)
			}
			first := got[0]

			if !cmp.Equal(first.String(), tt.wantFirst) {
				t.Errorf("solve() got = %v, want %v", first.String(), tt.wantFirst)
			}
		})
	}
}
