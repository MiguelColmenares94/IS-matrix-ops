const { max, min, mean, sum } = require('mathjs')

function flattenMatrix(matrix) {
  return matrix.flat()
}

function isDiagonal(matrix) {
  for (let i = 0; i < matrix.length; i++) {
    for (let j = 0; j < matrix[i].length; j++) {
      if (i !== j && Math.abs(matrix[i][j]) >= 1e-9) return false
    }
  }
  return true
}

function computeStats(q, r) {
  const values = [...flattenMatrix(q), ...flattenMatrix(r)]
  return {
    max: max(values),
    min: min(values),
    avg: mean(values),
    sum: sum(values),
    q_diagonal: isDiagonal(q),
    r_diagonal: isDiagonal(r),
  }
}

module.exports = { computeStats, isDiagonal }
