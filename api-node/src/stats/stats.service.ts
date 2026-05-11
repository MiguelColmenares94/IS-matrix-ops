import { max, min, mean, sum } from 'mathjs'

type Matrix = number[][]

function flattenMatrix(matrix: Matrix): number[] {
  return matrix.flat()
}

export function isDiagonal(matrix: Matrix): boolean {
  for (let i = 0; i < matrix.length; i++) {
    for (let j = 0; j < matrix[i].length; j++) {
      if (i !== j && Math.abs(matrix[i][j]) >= 1e-9) return false
    }
  }
  return true
}

export interface Stats {
  max: number
  min: number
  avg: number
  sum: number
  q_diagonal: boolean
  r_diagonal: boolean
}

export function computeStats(q: Matrix, r: Matrix): Stats {
  const values = [...flattenMatrix(q), ...flattenMatrix(r)]
  return {
    max: max(values) as number,
    min: min(values) as number,
    avg: mean(values) as number,
    sum: sum(values) as number,
    q_diagonal: isDiagonal(q),
    r_diagonal: isDiagonal(r),
  }
}
