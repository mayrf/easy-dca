package kraken

import (
	"testing"
)

func TestTrimFloat32ToOneDecimal(t *testing.T) {
	tests := []struct {
		input    float32
		want     float32
	}{
		{1.234, 1.2},
		{1.25, 1.2},
		{1.26, 1.2},
		{0.0, 0.0},
		{-1.27, -1.3},
		{2.99, 2.9},
	}
	for _, tc := range tests {
		got := trimFloat32ToOneDecimal(tc.input)
		if got != tc.want {
			t.Errorf("trimFloat32ToOneDecimal(%v) = %v; want %v", tc.input, got, tc.want)
		}
	}
} 