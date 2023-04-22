import { relative } from 'node:path'
import ts from 'typescript'
import { minimatch } from 'minimatch'
import {
  hasComments,
  isExportedClass,
  isExportedFunction,
  isExportedInterface,
  isExportedVariable,
  isMethodOfExportedInterface,
  isPublicMethodOfExportedClass,
  isPublicPropertyOfExportedOwner,
} from './nodes'
import type { RawIdentifier } from './identifier'
import { createRawIdentifier } from './identifier'
import type { SymbolType } from './symbols'
import { configureSymbols } from './symbols'
import { toArray } from './utils'

export type Findings = Record<string, Finding[]>

export interface Finding<Symbols extends SymbolType = SymbolType> {
  identifier: RawIdentifier<Symbols>
  node: ts.Node
}

export interface WithSymbolsOption<Symbols extends SymbolType = SymbolType> {
  symbols?: readonly Symbols[]
}

export interface WithIncludeOptions {
  include?: GlobOption
  exclude?: GlobOption
}

export type GlobOption = string | readonly string[]

export type FinderOptions<Symbols extends SymbolType = SymbolType> =
  WithSymbolsOption<Symbols> & WithIncludeOptions

export const defaultExclude = [
  '**/dist/**',
  '**/node_modules/**',
  '**/tests/**',
  '**/__tests__/**',
  '**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts}',
  '**/*.d.ts',
] as const

export function createFinder<Symbols extends SymbolType = SymbolType>(
  root: string,
  options?: FinderOptions,
) {
  let files = ts.sys.readDirectory(root, [
    '.js',
    '.mjs',
    '.cjs',
    '.ts',
    '.mts',
    '.cts',
  ])

  const exclude = toArray(options?.exclude ?? defaultExclude)

  if (options?.include) {
    files = files.filter((file) =>
      toArray(options.include).some((pattern) =>
        minimatch(stripRoot(file, root), pattern),
      ),
    )
  }

  if (exclude.length) {
    files = files.filter(
      (file) =>
        !exclude.some((pattern) => minimatch(stripRoot(file, root), pattern)),
    )
  }

  function findUncommented() {
    const results = files
      .map((path) => ({
        path: relative(root, path),
        findings: findUncommentedInPath(path, options),
      }))
      .filter((result) => !!result.findings?.length) as Array<{
      path: string
      findings: Array<Finding<Symbols>>
    }>

    return results.reduce<Record<string, Array<Finding<Symbols>>>>(
      (acc, result) => {
        return {
          ...acc,
          [result.path]: result.findings,
        }
      },
      {},
    )
  }

  return {
    files,
    findUncommented,
  }
}

function findUncommentedInPath<Symbols extends SymbolType = SymbolType>(
  path: string,
  options?: WithSymbolsOption<Symbols>,
) {
  const content = ts.sys.readFile(path)
  if (!content) {
    return null
  }

  const file = ts.createSourceFile(path, content, ts.ScriptTarget.Latest, true)
  const nodes = findUncommentedNodes(file, options)

  return nodes.map((node) => ({ node, identifier: createRawIdentifier(node) }))
}

function findUncommentedNodes<Symbols extends SymbolType = SymbolType>(
  node: ts.Node,
  options?: WithSymbolsOption<Symbols>,
): ts.Node[] {
  const uncommented: ts.Node[] = []

  function traverse(node: ts.Node) {
    if (ts.isSourceFile(node)) {
      ts.forEachChild(node, traverse)
      return
    }

    if (!isSupportedNode(node, options)) {
      ts.forEachChild(node, traverse)
      return
    }

    if (!hasComments(node)) {
      uncommented.push(node)
    }

    ts.forEachChild(node, traverse)
  }

  traverse(node)

  return [...uncommented.values()]
}

export function printFindings(findings: Findings) {
  const withoutNodes = removeNodesFromFindings(findings)
  return JSON.stringify(withoutNodes, null, 2)
}

export function removeNodesFromFindings(findings: Findings) {
  const out = {} as Record<string, RawIdentifier[]>
  for (const [file, _findings] of Object.entries(findings)) {
    out[file] = _findings.map((finding) => finding.identifier)
  }
  return out
}

function isSupportedNode<Symbols extends SymbolType = SymbolType>(
  node: ts.Node,
  options?: WithSymbolsOption<Symbols>,
) {
  const tests = [] as Array<(node: ts.Node) => boolean>

  for (const symbol of configureSymbols(options?.symbols ?? [])) {
    switch (symbol) {
      case 'var':
        tests.push(isExportedVariable)
        break
      case 'func':
        tests.push(isExportedFunction)
        break
      case 'class':
        tests.push(isExportedClass)
        break
      case 'iface':
        tests.push(isExportedInterface)
        break
      case 'method':
        tests.push(isPublicMethodOfExportedClass, isMethodOfExportedInterface)
        break
      case 'prop':
        tests.push(isPublicPropertyOfExportedOwner)
        break
    }
  }

  return tests.some((test) => test(node))
}

function stripRoot(file: string, root: string) {
  return file.replace(root, '').replace(/^[\/\\]+/, '')
}
