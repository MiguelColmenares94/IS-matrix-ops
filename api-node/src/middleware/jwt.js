const jwt = require('jsonwebtoken')

module.exports = (req, res, next) => {
  const auth = req.headers.authorization || ''
  if (!auth.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'missing token' })
  }
  try {
    const decoded = jwt.verify(auth.slice(7), process.env.JWT_SECRET, { algorithms: ['HS256'] })
    req.user = decoded
    next()
  } catch {
    res.status(401).json({ error: 'invalid token' })
  }
}
