import ts from 'typescript'
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
import type { NaturalLanguageTarget, RawIdentifier } from './identifier'
import { createRawIdentifier, describeIdentifier } from './identifier'
import type { SymbolType } from './symbols'
import { configureSymbols } from './symbols'
import { createSourceFile } from './parse'

export interface Finding<Symbols extends SymbolType = SymbolType> {
  identifier: RawIdentifier<Symbols>
  target: NaturalLanguageTarget<Symbols>
}

export interface WithSymbolsOption<Symbols extends SymbolType = SymbolType> {
  symbols?: readonly Symbols[]
}

export type GlobOption = string | readonly string[]

export type FinderOptions<Symbols extends SymbolType = SymbolType> =
  WithSymbolsOption<Symbols>

export function createFinder<Symbols extends SymbolType = SymbolType>(
  options?: FinderOptions<Symbols>,
) {
  function find(code: string) {
    const nodes = findUncommentedNodes(createSourceFile('', code), options)
    return nodes.map((node): Finding<Symbols> => {
      const identifier = createRawIdentifier(node) as RawIdentifier<Symbols>
      return {
        identifier,
        target: describeIdentifier(
          identifier,
        ) as NaturalLanguageTarget<Symbols>,
      }
    })
  }

  return {
    find,
  }
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

export function printFindings(findings: Finding[]) {
  return JSON.stringify(findings, null, 2)
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
