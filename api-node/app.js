require('dotenv').config()
const express = require('express')
const cors = require('cors')

const app = express()
app.use(express.json())
app.use(cors({ origin: process.env.ALLOWED_ORIGIN, methods: ['GET', 'POST', 'OPTIONS'] }))

const statsRouter = require('./src/stats/stats.router')
app.use('/api/v2/stats', statsRouter)

module.exports = app
