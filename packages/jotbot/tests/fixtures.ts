import { fileURLToPath } from 'node:url'
import { join } from 'node:path'
import ts, { sys } from 'typescript'
import { expect } from 'vitest'
import type { Comments } from '../src/comments'
import { applyComments } from '../src/comments'
import type { RawIdentifier } from '../src/identifier'
import { expectAddedComments, tryOut } from './testutils'

export const fixtures = [
  'basic',
  'unexported',
  'iface',
  'exclude-symbols',
  'tree',
  'test-files',
  'dts',
  'js',
] as const

export type Fixture = (typeof fixtures)[number]

const fixturesRoot = fileURLToPath(new URL('./fixtures', import.meta.url))

export function fixtureRoot(name: Fixture) {
  return join(fixturesRoot, name)
}

export function loadFixture(name: Fixture) {
  const dir = fixtureRoot(name)
  const paths = sys.readDirectory(dir)
  const fileList = paths
    .map((path) => ({ path, file: loadFile(path)! }))
    .filter((f) => !!f.file)

  const files = fileList.reduce<Record<string, ts.SourceFile>>(
    (acc, { path, file }) => ({ ...acc, [path]: file }),
    {},
  )

  const filePaths = Object.keys(files)

  function getFile(path: string): ts.SourceFile | null {
    const fullPath = join(dir, path)
    return files[fullPath] ?? null
  }

  function testPatch(
    path: string,
    tests: Record<
      RawIdentifier,
      {
        comment: string
        want: string
      }
    >,
  ) {
    const file = getFile(path)
    if (file == null)
      throw new Error(`Could not find file '${path}' in fixture.`)

    const comments = Object.entries(tests).reduce<Comments>(
      (acc, [raw, { comment }]) => ({ ...acc, [raw]: comment }),
      {},
    )

    const want = Object.entries(tests).reduce<Comments>(
      (acc, [raw, { want }]) => ({ ...acc, [raw]: want }),
      {},
    )

    const [result, err] = tryOut(() => applyComments(file, comments))

    expect(err).toBeNull()
    expectAddedComments(result!.patched, want)
  }

  return {
    files,
    filePaths,
    getFile,
    testPatch,
  }
}

function loadFile(path: string) {
  const contents = sys.readFile(path)
  if (!contents) {
    return null
  }

  return ts.createSourceFile(path, contents, ts.ScriptTarget.Latest, true)
}
