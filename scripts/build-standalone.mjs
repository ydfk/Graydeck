import { mkdirSync } from 'node:fs'
import { basename, join } from 'node:path'
import { spawnSync } from 'node:child_process'

const cwd = process.cwd()
const cacheDir = join(cwd, '.gocache')
const outputDir = join(cwd, 'dist')
const options = parseArgs(process.argv.slice(2))
const target = parseTarget(options)
const version = normalizeVersion(options.version ?? process.env.GRAYDECK_VERSION ?? '0.0.0-dev')
const binaryName = target.os === 'windows'
  ? `graydeck-${target.os}-${target.arch}.exe`
  : `graydeck-${target.os}-${target.arch}`

mkdirSync(cacheDir, { recursive: true })
mkdirSync(outputDir, { recursive: true })

run('pnpm', ['run', 'build:web'])
run('node', ['./scripts/sync-web-assets.mjs'])
run('go', ['build', '-trimpath', '-ldflags', `-X mihomo-manager/internal/buildinfo.Version=${version}`, '-o', join(outputDir, binaryName), './cmd/managerd'], {
  GOOS: target.os,
  GOARCH: target.arch,
}, false)

function run(command, args, extraEnv = {}, shell = true) {
  const result = spawnSync(command, args, {
    cwd,
    stdio: 'inherit',
    shell,
    env: {
      ...process.env,
      GOCACHE: cacheDir,
      ...extraEnv,
    },
  })

  if (result.status !== 0) {
    process.exit(result.status ?? 1)
  }
}

console.log(`standalone binary: ${join('dist', basename(binaryName))}`)

function parseTarget(options) {
  const targetValue = options.target ?? options.t
  let os = options.os
  let arch = options.arch

  if (targetValue) {
    const parts = String(targetValue).split(/[/-]/).filter(Boolean)
    os = os || parts[0]
    arch = arch || parts[1]
  }

  return {
    os: normalizeOS(os || defaultGOOS()),
    arch: normalizeArch(arch || process.arch),
  }
}

function normalizeVersion(value) {
  const version = String(value).trim().replace(/^v/, '')
  if (/^\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$/.test(version)) {
    return version
  }

  console.error(`版本号必须使用语义化版本：${value}`)
  process.exit(1)
}

function parseArgs(args) {
  const options = {}

  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index]
    if (!arg.startsWith('--')) {
      continue
    }

    const raw = arg.slice(2)
    const [key, inlineValue] = raw.split('=', 2)
    if (inlineValue !== undefined) {
      options[key] = inlineValue
      continue
    }

    const next = args[index + 1]
    if (next && !next.startsWith('--')) {
      options[key] = next
      index += 1
    } else {
      options[key] = 'true'
    }
  }

  return options
}

function defaultGOOS() {
  if (process.platform === 'win32') {
    return 'windows'
  }

  return process.platform
}

function normalizeOS(value) {
  const os = String(value).toLowerCase()
  if (os === 'win32' || os === 'win') {
    return 'windows'
  }
  if (os === 'darwin' || os === 'mac' || os === 'macos') {
    return 'darwin'
  }
  if (['linux', 'windows', 'freebsd'].includes(os)) {
    return os
  }

  console.error(`不支持的目标系统：${value}`)
  process.exit(1)
}

function normalizeArch(value) {
  const arch = String(value).toLowerCase()
  if (arch === 'x64' || arch === 'x86_64') {
    return 'amd64'
  }
  if (arch === 'arm64' || arch === 'aarch64') {
    return 'arm64'
  }
  if (['amd64', '386', 'arm'].includes(arch)) {
    return arch
  }

  console.error(`不支持的目标架构：${value}`)
  process.exit(1)
}
