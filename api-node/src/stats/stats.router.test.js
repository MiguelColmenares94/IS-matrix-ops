import request from 'supertest'
import jwt from 'jsonwebtoken'
import pool from '../db/postgres.js'
import app from '../../app.js'

const skip = !process.env.DATABASE_URL

function makeToken(sub = 'test-user-id') {
  return jwt.sign({ sub, email: 'test@test.com' }, process.env.JWT_SECRET || 'test', {
    algorithm: 'HS256',
    expiresIn: '10m',
  })
}

beforeEach(async () => {
  if (skip) return
  await pool.query('TRUNCATE stats_computations CASCADE')
})

afterAll(async () => {
  if (!skip) await pool.end()
})

const validQ = [[1, 0], [0, 1]]
const validR = [[2, 3], [0, 4]]

describe('POST /api/v2/stats', () => {
  test('valid Q/R with JWT returns 200 and all stats', async () => {
    if (skip) return
    const token = makeToken()
    const res = await request(app)
      .post('/api/v2/stats')
      .set('Authorization', `Bearer ${token}`)
      .send({ q: validQ, r: validR })
    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty('max')
    expect(res.body).toHaveProperty('min')
    expect(res.body).toHaveProperty('avg')
    expect(res.body).toHaveProperty('sum')
    expect(res.body).toHaveProperty('q_diagonal')
    expect(res.body).toHaveProperty('r_diagonal')
  })

  test('no JWT returns 401', async () => {
    const res = await request(app).post('/api/v2/stats').send({ q: validQ, r: validR })
    expect(res.status).toBe(401)
  })

  test('missing q field returns 400', async () => {
    const token = makeToken()
    const res = await request(app)
      .post('/api/v2/stats')
      .set('Authorization', `Bearer ${token}`)
      .send({ r: validR })
    expect(res.status).toBe(400)
  })
})
