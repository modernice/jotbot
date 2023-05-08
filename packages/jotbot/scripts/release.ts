import { fileURLToPath } from 'node:url'
import { execSync } from 'node:child_process'
import { SemanticVersion } from '@hediet/semver'
import { readPackageJSON } from 'pkg-types'

const packageRoot = fileURLToPath(new URL('..', import.meta.url))

export async function release(version: string) {
  const pkg = await readPackageJSON(packageRoot)
  const currentVersion = SemanticVersion.parse(pkg.version!)
  const newVersion = SemanticVersion.parse(version)

  if (newVersion.compareTo(currentVersion) < 1) {
    throw new Error(
      `New version ${newVersion} must be greater than current version ${currentVersion}.`,
    )
  }

  const v = newVersion.toString()
  execSync(`pnpm version ${v}`)

  execSync('git add package.json')
  execSync(`git commit -m 'chore(jotbot-ts): ${v}'`)
  execSync(`git push`)

  execSync(`pnpm publish --access public`)
}
