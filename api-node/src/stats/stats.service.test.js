import { computeStats, isDiagonal } from './stats.service.js'

describe('computeStats', () => {
  const q = [[1, 0], [0, 1]]
  const r = [[2, 3], [0, 4]]

  test('returns correct max, min, avg, sum', () => {
    const stats = computeStats(q, r)
    expect(stats.max).toBe(4)
    expect(stats.min).toBe(0)
    expect(stats.sum).toBeCloseTo(1 + 0 + 0 + 1 + 2 + 3 + 0 + 4)
    expect(stats.avg).toBeCloseTo(stats.sum / 8)
  })

  test('identity matrix is diagonal', () => {
    expect(isDiagonal([[1, 0], [0, 1]])).toBe(true)
  })

  test('non-diagonal matrix returns false', () => {
    expect(isDiagonal([[1, 2], [0, 1]])).toBe(false)
  })

  test('near-zero off-diagonal (< 1e-9) passes diagonal check', () => {
    expect(isDiagonal([[1, 1e-10], [1e-10, 1]])).toBe(true)
  })

  test('q_diagonal and r_diagonal in result', () => {
    const stats = computeStats([[1, 0], [0, 2]], [[3, 1], [0, 4]])
    expect(stats.q_diagonal).toBe(true)
    expect(stats.r_diagonal).toBe(false)
  })
})
