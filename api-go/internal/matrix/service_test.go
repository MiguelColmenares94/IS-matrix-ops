package matrix_test

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/is-matrix-ops/api-go/internal/matrix"
)

var svc = &matrix.Service{}

func TestValidateMatrix_Valid3x3(t *testing.T) {
	input := [][]interface{}{{1.0, 2.0, 3.0}, {4.0, 5.0, 6.0}, {7.0, 8.0, 9.0}}
	data, err := svc.ValidateMatrix(input)
	require.NoError(t, err)
	assert.Equal(t, 3, len(data))
	assert.Equal(t, 3, len(data[0]))
}

func TestValidateMatrix_Empty(t *testing.T) {
	_, err := svc.ValidateMatrix([][]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateMatrix_JaggedRows(t *testing.T) {
	input := [][]interface{}{{1.0, 2.0, 3.0}, {3.0, 4.0}}
	_, err := svc.ValidateMatrix(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rectangular")
	assert.Contains(t, err.Error(), "row 2")
}

func TestValidateMatrix_NonNumeric(t *testing.T) {
	input := [][]interface{}{{1.0, "x"}, {3.0, 4.0}}
	_, err := svc.ValidateMatrix(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-numeric")
}

func TestValidateMatrix_NaN(t *testing.T) {
	input := [][]interface{}{{math.NaN(), 2.0}, {3.0, 4.0}}
	_, err := svc.ValidateMatrix(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-numeric")
}

func TestValidateMatrix_MoreColsThanRows(t *testing.T) {
	// 2 rows, 4 cols → m < n
	input := [][]interface{}{{1.0, 2.0, 3.0, 4.0}, {5.0, 6.0, 7.0, 8.0}}
	_, err := svc.ValidateMatrix(input)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "rows as columns") || strings.Contains(err.Error(), "2×4"))
}

func TestFactorizeQR_3x3(t *testing.T) {
	data := [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	result, err := svc.FactorizeQR(data)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Q))
	assert.Equal(t, 3, len(result.Q[0]))
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			dot := 0.0
			for k := 0; k < 3; k++ {
				dot += result.Q[k][i] * result.Q[k][j]
			}
			expected := 0.0
			if i == j {
				expected = 1.0
			}
			assert.InDelta(t, expected, dot, 1e-9)
		}
	}
}

func TestFactorizeQR_NonSquare(t *testing.T) {
	data := [][]float64{{1, 2}, {3, 4}, {5, 6}}
	result, err := svc.FactorizeQR(data)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Q))
	assert.Equal(t, 3, len(result.R))
	assert.Equal(t, 2, len(result.R[0]))
}
