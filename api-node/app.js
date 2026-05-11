import 'dotenv/config'
import express from 'express'
import cors from 'cors'
import statsRouter from './src/stats/stats.router.js'

const app = express()
app.use(express.json())
app.use(cors({ origin: process.env.ALLOWED_ORIGIN, methods: ['GET', 'POST', 'OPTIONS'] }))

app.use('/api/v2/stats', statsRouter)

export default app
