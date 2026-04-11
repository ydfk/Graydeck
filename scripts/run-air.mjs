import { spawn } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { join } from 'node:path'

const cwd = process.cwd()
const cacheDir = join(cwd, '.gocache')
const tmpDir = join(cwd, 'tmp')

mkdirSync(cacheDir, { recursive: true })
mkdirSync(tmpDir, { recursive: true })

const child = spawn('air', ['-c', '.air.toml'], {
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
