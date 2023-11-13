import { fileURLToPath } from 'node:url'
import { promisify } from 'node:util'
import { exec } from 'node:child_process'
import { SemanticVersion } from '@hediet/semver'
import { readPackageJSON } from 'pkg-types'

const packageRoot = fileURLToPath(new URL('..', import.meta.url))

/**
 * Performs the release process for a package by updating its version,
 * installing dependencies, building the project, committing changes, and
 * publishing to a package registry. It accepts a version string and checks that
 * it is greater than the current package version. If successful, it proceeds
 * with the release steps, otherwise, it throws an error. This operation assumes
 * that the working directory is a valid package repository and that all
 * necessary tools are available in the execution environment.
 * 
 * Note: The function assumes certain behaviors from external commands like
 * `pnpm` and `git`, and is intended to be used in a specific project setup.
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

  const v = newVersion.toString()
  await execute(`pnpm version v${v}`)

  await execute('pnpm i')
  await execute('pnpm build')

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
