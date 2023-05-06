import ts from 'typescript'
import {
  hasComments,
  isExportedClass,
  isExportedFunction,
  isExportedInterface,
  isExportedType,
  isExportedVariable,
  isMethodOfExportedInterface,
  isMethodOfExportedTypeAlias,
  isPublicMethodOfExportedClass,
  isPublicPropertyOfExportedOwner,
} from './nodes'
import type { RawIdentifier } from './identifier'
import { createRawIdentifier } from './identifier'
import type { SymbolType } from './symbols'
import { configureSymbols } from './symbols'
import { createSourceFile } from './parse'

export interface WithSymbolsOption<Symbols extends SymbolType = SymbolType> {
  symbols?: readonly Symbols[]
}

export type GlobOption = string | readonly string[]

export interface FinderOptions<Symbols extends SymbolType = SymbolType>
  extends WithSymbolsOption<Symbols> {
  includeDocumented?: boolean
}

export function createFinder<Symbols extends SymbolType = SymbolType>(
  options?: FinderOptions<Symbols>,
) {
  function find(code: string) {
    const nodes = findNodes(createSourceFile('', code), options)
    return nodes.map((node): RawIdentifier<Symbols> => {
      const identifier = createRawIdentifier(node) as RawIdentifier<Symbols>
      return identifier
    })
  }

  return {
    find,
  }
}

function findNodes<Symbols extends SymbolType = SymbolType>(
  node: ts.Node,
  options?: FinderOptions<Symbols>,
): ts.Node[] {
  const found: ts.Node[] = []

  function traverse(node: ts.Node) {
    if (ts.isSourceFile(node)) {
      ts.forEachChild(node, traverse)
      return
    }

    if (!isSupportedNode(node, options)) {
      ts.forEachChild(node, traverse)
      return
    }

    if (!hasComments(node) || options?.includeDocumented) {
      found.push(node)
    }

    ts.forEachChild(node, traverse)
  }

  traverse(node)

  return [...found.values()]
}

export function printFindings(findings: RawIdentifier[]) {
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
        tests.push(
          isPublicMethodOfExportedClass,
          isMethodOfExportedInterface,
          isMethodOfExportedTypeAlias,
        )
        break
      case 'prop':
        tests.push(isPublicPropertyOfExportedOwner)
        break
      case 'type':
        tests.push(isExportedType)
        break
    }
  }

  return tests.some((test) => test(node))
}
