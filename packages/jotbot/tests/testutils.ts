import type ts from 'typescript'
import { expect } from 'vitest'
import type { Comments } from '../src/comments'
import { getNodeComments, getRawIdentifiers } from '../src/comments'
import type { Fixture } from '../tests/fixtures'
import { loadFixture } from '../tests/fixtures'
import { RawIdentifier, describeIdentifier } from '../src/identifier'
import { parseIdentifier } from '../src/identifier'
import { findNode } from '../src/nodes'
import type { Findings } from '../src/finder'

export function withFixture(
  name: Fixture,
  fn: (fixture: ReturnType<typeof loadFixture>) => void,
) {
  const fixture = loadFixture(name)
  fn(fixture)
}

export function tryOut<Fn extends (...args: any[]) => any>(
  fn: Fn,
): [ReturnType<Fn>, null] | [null, Error] {
  try {
    return [fn(), null]
  } catch (err) {
    const error = err instanceof Error ? err : new Error(String(err))
    return [null, error]
  }
}

export function expectAddedComments(file: ts.SourceFile, comments: Comments) {
  const identifiers = getRawIdentifiers(comments)

  for (const raw of identifiers) {
    const ident = parseIdentifier(raw)
    const node = findNode(file, ident)

    expect(node).toBeDefined()

    const { commentTarget } = node!

    const { text } = getNodeComments(commentTarget)

    expect(text).toBe(comments[raw])
  }
}

export function without<
  const Rm extends readonly any[],
  const From extends readonly any[],
>(remove: Rm, from: From): Remove<Rm, From> {
  return from.filter((item) => !remove.includes(item)) as Remove<Rm, From>
}

export type Remove<
  Rm extends readonly any[],
  From extends readonly any[],
> = From extends [infer Head, ...infer Tail]
  ? Head extends Rm[number]
    ? Remove<Rm, Tail>
    : [Head, ...Remove<Rm, Tail>]
  : From

export function expectFindings(
  findings: Findings,
  want: Record<string, readonly RawIdentifier[]>,
) {
  for (const _findings of Object.values(findings)) {
    for (const finding of _findings) {
      const wantTarget = describeIdentifier(finding.identifier)
      expect(finding.target).toBe(wantTarget)
    }
  }

  for (const path of Object.keys(want)) want[path] = want[path].slice().sort()

  const got = Object.entries(findings).reduce<Record<string, RawIdentifier[]>>(
    (acc, [path, findings]) => {
      return {
        ...acc,
        [path]: findings.map((finding) => finding.identifier).sort(),
      }
    },
    {},
  )

  expect(got).toEqual(want)
}

export function expectNotFound(findings: Findings, dontWant: RawIdentifier[]) {
  dontWant = dontWant.slice().sort()

  const got = Object.values(findings)
    .flatMap((findings) => findings.map((finding) => finding.identifier))
    .sort()

  for (const identifier of dontWant) expect(got).not.toContain(identifier)
}

export function expectFiles(findings: Findings, want: readonly string[]) {
  const got = Object.keys(findings).sort()
  expect(got).toEqual(want.slice().sort())
}

export function expectNoFiles(findings: Findings, dontWant: readonly string[]) {
  expect(Object.keys(findings).slice()).not.toEqual(dontWant.slice().sort())
}

export function debugNode(n: ts.Node, msg?: string) {
  console.log(`\n${msg ?? 'Node:'}\n${n.getText()}`)
}
