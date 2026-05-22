import { cpSync, existsSync, mkdirSync, rmSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const cwd = process.cwd()
const source = join(cwd, 'web', 'dist')
const target = join(cwd, 'internal', 'webui', 'dist')

if (!existsSync(source)) {
  console.error('web/dist 不存在，请先执行 pnpm run build:web')
  process.exit(1)
}

rmSync(target, { recursive: true, force: true })
mkdirSync(target, { recursive: true })
cpSync(source, target, { recursive: true })
writeFileSync(join(target, '.gitignore'), '*\n!.gitignore\n!keep.txt\n')
writeFileSync(join(target, 'keep.txt'), 'placeholder\n')
