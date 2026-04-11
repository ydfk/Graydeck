import { spawn } from 'node:child_process'

const children = []

function start(name, command, args) {
  const child = spawn(command, args, {
    cwd: process.cwd(),
    stdio: 'inherit',
    shell: true,
  })

  child.on('exit', (code) => {
    if (code && code !== 0) {
      process.exitCode = code
      shutdown()
    }
  })

  child.on('error', () => {
    console.error(`[${name}] failed to start`)
    process.exitCode = 1
    shutdown()
  })

  children.push(child)
}

function shutdown() {
  for (const child of children) {
    if (!child.killed) {
      child.kill()
    }
  }
}

process.on('SIGINT', shutdown)
process.on('SIGTERM', shutdown)

start('web', 'pnpm', ['run', 'dev:web'])
start('server', 'pnpm', ['run', 'dev:server'])
