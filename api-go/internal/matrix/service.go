package matrix

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

var ErrInvalidMatrix = errors.New("invalid matrix")

var exampleMatrix = map[string]interface{}{
	"matrix": [][]int{{1, 2}, {3, 4}, {5, 6}},
}

type QRResult struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

type Service struct{}

func (s *Service) ValidateMatrix(input [][]interface{}) ([][]float64, error) {
	if len(input) == 0 || len(input[0]) == 0 {
		return nil, fmt.Errorf("matrix must not be empty")
	}
	m := len(input)
	n := len(input[0])
	result := make([][]float64, m)
	for i, row := range input {
		if len(row) != n {
			return nil, fmt.Errorf("matrix must be rectangular: row %d has %d elements, expected %d", i+1, len(row), n)
		}
		result[i] = make([]float64, n)
		for j, v := range row {
			f, ok := toFloat64(v)
			if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
				return nil, fmt.Errorf("matrix contains non-numeric value at row %d, column %d", i+1, j+1)
			}
			result[i][j] = f
		}
	}
	if m < n {
		return nil, fmt.Errorf("QR decomposition requires m ≥ n (rows ≥ columns), but got %d×%d matrix", m, n)
	}
	return result, nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

func (s *Service) FactorizeQR(data [][]float64) (*QRResult, error) {
	m := len(data)
	n := len(data[0])
	flat := make([]float64, m*n)
	for i, row := range data {
		for j, v := range row {
			flat[i*n+j] = v
		}
	}
	A := mat.NewDense(m, n, flat)
	var qr mat.QR
	qr.Factorize(A)

	var Q mat.Dense
	qr.QTo(&Q)
	qRows, qCols := Q.Dims()
	qOut := make([][]float64, qRows)
	for i := range qOut {
		qOut[i] = make([]float64, qCols)
		for j := range qOut[i] {
			qOut[i][j] = Q.At(i, j)
		}
	}

	var R mat.Dense
	qr.RTo(&R)
	rRows, rCols := R.Dims()
	rOut := make([][]float64, rRows)
	for i := range rOut {
		rOut[i] = make([]float64, rCols)
		for j := range rOut[i] {
			rOut[i][j] = R.At(i, j)
		}
	}

	return &QRResult{Q: qOut, R: rOut}, nil
}
