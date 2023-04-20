package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_solve(t *testing.T) {
	tests := map[string]struct {
		target       int
		digits       []int
		wantLen      int
		wantFirst    string
		wantShortest string
	}{
		"a": {
			digits:       []int{5, 7, 9, 10, 15, 25},
			target:       93,
			wantLen:      68,
			wantFirst:    "((((25 * (15 + 10)) + -9) / 7) + 5)",
			wantShortest: "((9 * 7) + 25 + 5)",
		},
		"b": {

			digits:       []int{4, 5, 7, 8, 15, 20},
			target:       113,
			wantLen:      190,
			wantFirst:    "(((20 + -7) * 8) + 5 + 4)",
			wantShortest: "((15 * 7) + 8)",
		},
		"c": {
			digits:       []int{3, 4, 6, 9, 11, 15},
			target:       205,
			wantLen:      34,
			wantFirst:    "(((15 + 9 + -6) * 11) + 4 + 3)",
			wantShortest: "((9 * 6 * 4) + -11)",
		},
		"d": {
			digits:       []int{3, 5, 9, 11, 23, 25},
			target:       351,
			wantLen:      24,
			wantFirst:    "(((11 * 9) + 23 + -5) * 3)",
			wantShortest: "((25 + 11 + 3) * 9)",
		},
		"f": {
			digits:       []int{24, 8, 10, 20, 5, 15},
			target:       497,
			wantLen:      11,
			wantFirst:    "((24 * 20) + 15 + 10 + -8)",
			wantShortest: "((24 * 20) + 15 + 10 + -8)",
		},
	}
	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			got := solve(tt.target, tt.digits)

			if len(got) != tt.wantLen {
				t.Fatalf("solve() got %d results, want %d", len(got), tt.wantLen)
			}
			first := got[0]

			shortest, err := shortest(got)
			if err != nil {
				t.Fatalf("shortest() failed unexpectedly: %v", err)
			}

			if !cmp.Equal(first.String(), tt.wantFirst) {
				t.Errorf("solve() got[0] = %q, want %q", first.String(), tt.wantFirst)
			}
			if !cmp.Equal(shortest.String(), tt.wantShortest) {
				t.Errorf("solve() got[shortest] = %q, want %q", shortest.String(), tt.wantShortest)
			}
		})
	}
}
