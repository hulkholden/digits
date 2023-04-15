package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_solve(t *testing.T) {
	tests := map[string]struct {
		target int
		digits []int
		want   string
		wantOk bool
	}{
		"a": {
			digits: []int{5, 7, 9, 10, 15, 25},
			target: 93,
			want:   "(5 + (((25 * (10 + 15)) - 9) / 7))",
			wantOk: true,
		},
		"b": {

			digits: []int{4, 5, 7, 8, 15, 20},
			target: 113,
			want:   "(4 + (5 + (8 * (20 - 7))))",
			// TODO: this is correct but a much simpler solution is (8*15)-7.
			wantOk: true,
		},
		"c": {
			digits: []int{3, 4, 6, 9, 11, 15},
			target: 205,
			want:   "(3 + (4 + (11 * ((9 + 15) - 6))))",
			wantOk: true,
		},
		"d": {
			digits: []int{3, 5, 9, 11, 23, 25},
			target: 351,
			want:   "(3 * ((23 + (9 * 11)) - 5))",
			wantOk: true,
		},
		"f": {
			digits: []int{24, 8, 10, 20, 5, 15},
			target: 497,
			want:   "((10 + (15 + (24 * 20))) - 8)",
			wantOk: true,
		},
	}
	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			got, gotOk := solve(tt.target, tt.digits)
			if !cmp.Equal(got.String(), tt.want) {
				t.Errorf("solve() got = %v, want %v", got.String(), tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("solve() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
