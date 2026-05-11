import { Router, Request, Response } from 'express'
import jwtMiddleware from '../middleware/jwt'
import { computeStats } from './stats.service'
import { saveStatsComputation } from './stats.repository'

const router = Router()

type Matrix = number[][]

function isValidMatrix(m: unknown): m is Matrix {
  return Array.isArray(m) && m.length > 0 && (m as unknown[]).every(
    row => Array.isArray(row) && (row as unknown[]).length > 0
  )
}

router.post('/', jwtMiddleware, async (req: Request, res: Response): Promise<void> => {
  const { q, r } = req.body as { q: unknown; r: unknown }
  const userId = (req as Request & { user?: { sub?: string } }).user?.sub

  if (!isValidMatrix(q) || !isValidMatrix(r)) {
    await saveStatsComputation(userId, q as Matrix || [], r as Matrix || [], null, false, 'q and r must be non-empty arrays of arrays').catch(() => {})
    res.status(400).json({ error: 'q and r must be non-empty arrays of arrays' })
    return
  }

  try {
    const stats = computeStats(q, r)
    await saveStatsComputation(userId, q, r, stats, true, null)
    res.json(stats)
  } catch (err) {
    const msg = err instanceof Error ? err.message : 'unknown error'
    await saveStatsComputation(userId, q, r, null, false, msg).catch(() => {})
    res.status(500).json({ error: 'computation failed' })
  }
})

export default router
