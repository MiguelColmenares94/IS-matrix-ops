import pool from '../db/postgres'
import { Stats } from './stats.service'

type Matrix = number[][]

export async function saveStatsComputation(
  userId: string | undefined,
  q: Matrix,
  r: Matrix,
  stats: Stats | null,
  success: boolean,
  errorMsg: string | null
): Promise<void> {
  await pool.query(
    'CALL save_stats_computation($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)',
    [
      userId,
      JSON.stringify(q),
      JSON.stringify(r),
      stats?.max ?? null,
      stats?.min ?? null,
      stats?.avg ?? null,
      stats?.sum ?? null,
      stats?.q_diagonal ?? null,
      stats?.r_diagonal ?? null,
      success,
      errorMsg ?? null,
    ]
  )
}
