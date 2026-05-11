import { Request, Response, NextFunction } from 'express'
import jwt from 'jsonwebtoken'

export default (req: Request, res: Response, next: NextFunction): void => {
  const auth = req.headers.authorization || ''
  if (!auth.startsWith('Bearer ')) {
    res.status(401).json({ error: 'missing token' })
    return
  }
  try {
    const decoded = jwt.verify(auth.slice(7), process.env.JWT_SECRET as string, {
      algorithms: ['HS256'],
    })
    ;(req as Request & { user: unknown }).user = decoded
    next()
  } catch {
    res.status(401).json({ error: 'invalid token' })
  }
}
