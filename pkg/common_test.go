package pkg_test

import (
	"testing"

	"github.com/gwuah/accounts/pkg"
	"github.com/stretchr/testify/require"
)

func TestConvertToCents(t *testing.T) {
	type TestCase struct {
		input  float64
		output int64
	}

	cases := []TestCase{
		{
			input:  14.58,
			output: 1458,
		},
		{
			input:  1458,
			output: 145800,
		},
	}

	for _, tc := range cases {
		require.Equal(t, tc.output, pkg.ConvertToCents(tc.input))
	}
}

func TestConvertToUnit(t *testing.T) {
	type TestCase struct {
		input  int64
		output float64
	}

	cases := []TestCase{
		{
			input:  1458,
			output: 14.58,
		},
		{
			input:  145800,
			output: 1458,
		},
	}

	for _, tc := range cases {
		require.Equal(t, tc.output, pkg.ConvertToUnit(tc.input))
	}
}
