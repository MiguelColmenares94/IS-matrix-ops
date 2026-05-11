const { Router } = require('express')
const jwtMiddleware = require('../middleware/jwt')
const { computeStats } = require('./stats.service')
const { saveStatsComputation } = require('./stats.repository')

const router = Router()

function isValidMatrix(m) {
  return Array.isArray(m) && m.length > 0 && m.every(row => Array.isArray(row) && row.length > 0)
}

router.post('/', jwtMiddleware, async (req, res) => {
  const { q, r } = req.body
  const userId = req.user?.sub

  if (!isValidMatrix(q) || !isValidMatrix(r)) {
    await saveStatsComputation(userId, q || [], r || [], null, false, 'q and r must be non-empty arrays of arrays').catch(() => {})
    return res.status(400).json({ error: 'q and r must be non-empty arrays of arrays' })
  }

  try {
    const stats = computeStats(q, r)
    await saveStatsComputation(userId, q, r, stats, true, null)
    return res.json(stats)
  } catch (err) {
    await saveStatsComputation(userId, q, r, null, false, err.message).catch(() => {})
    return res.status(500).json({ error: 'computation failed' })
  }
})

module.exports = router
