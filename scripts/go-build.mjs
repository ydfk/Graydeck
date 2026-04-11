import { spawn } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { join } from 'node:path'

const cwd = process.cwd()
const cacheDir = join(cwd, '.gocache')

mkdirSync(cacheDir, { recursive: true })

const child = spawn('go', ['build', './...'], {
  cwd,
  stdio: 'inherit',
  shell: true,
  env: {
    ...process.env,
    GOCACHE: cacheDir,
  },
})

child.on('exit', (code) => {
  process.exit(code ?? 0)
})

child.on('error', () => {
  process.exit(1)
})
