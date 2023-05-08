import { fileURLToPath } from 'node:url'
import { promisify } from 'node:util'
import { exec } from 'node:child_process'
import { SemanticVersion } from '@hediet/semver'
import { readPackageJSON } from 'pkg-types'

const packageRoot = fileURLToPath(new URL('..', import.meta.url))

/**
 * Updates the current version of the package to the specified `version`,
 * builds, and publishes it.
 * Throws an error if the new version is not greater than the current version.
 * @param version - The new version string to update the package to.
 */
export async function release(version: string) {
  const pkg = await readPackageJSON(packageRoot)
  const currentVersion = SemanticVersion.parse(pkg.version!.replace('v', ''))
  const newVersion = SemanticVersion.parse(version.replace('v', ''))

  if (newVersion.compareTo(currentVersion) < 1) {
    throw new Error(
      `New version v${newVersion} must be greater than current version v${currentVersion}.`,
    )
  }

  await execute('pnpm i')
  await execute('pnpm build')

  const v = newVersion.toString()
  await execute(`pnpm version v${v}`)

  await execute('git add package.json')
  await execute(`git commit -m 'chore(jotbot-ts): v${v}'`)
  await execute(`git push`)

  await execute(`pnpm publish --access public --no-git-checks`)
}

const _exec = promisify(exec)
async function execute(cmd: string) {
  const { stdout } = await _exec(cmd)
  process.stdout.write(stdout)
  return stdout
}

if (import.meta.url.startsWith('file:')) {
  const modulePath = fileURLToPath(import.meta.url)
  if (process.argv.length > 2 && process.argv[1] === modulePath) {
    release(process.argv[2])
  }
}
