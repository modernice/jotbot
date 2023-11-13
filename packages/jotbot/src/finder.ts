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

/**
 * The `WithSymbolsOption` interface provides an optional `symbols` property
 * that can be set with an array of specific symbol types, derived from the
 * {@link SymbolType}. This allows for fine-tuning the behavior of certain
 * functions by specifying which symbol types should be considered in their
 * processing.
 */
export interface WithSymbolsOption<Symbols extends SymbolType = SymbolType> {
  symbols?: readonly Symbols[]
}

export type GlobOption = string | readonly string[]

/**
 * `FinderOptions` is an interface that specifies the settings for a finder
 * operation. It extends {@link WithSymbolsOption} by including an optional
 * `includeDocumented` flag that determines whether documented nodes should be
 * included in the search results. The type parameter `Symbols` extends {@link
 * SymbolType} and allows for customization of the symbols considered during the
 * finding process.
 */
export interface FinderOptions<Symbols extends SymbolType = SymbolType>
  extends WithSymbolsOption<Symbols> {
  includeDocumented?: boolean
}

/**
 * Creates a `find` function encapsulated within a returned object that can be
 * used to identify and filter nodes in a given piece of TypeScript code based
 * on the provided {@link FinderOptions}. The `find` function parses the code
 * into an abstract syntax tree, examines each node to determine if it matches
 * the specified criteria, including symbol types and documentation status, and
 * returns an array of {@link RawIdentifier}s representing the identified nodes.
 */
export function createFinder<Symbols extends SymbolType = SymbolType>(
  options?: FinderOptions<Symbols>,
) {
  function find(code: string) {
    const nodes = findNodes(createSourceFile('', code), options)
    return nodes
      .map((node): RawIdentifier<Symbols> | null => {
        const identifier = createRawIdentifier(
          node,
        ) as RawIdentifier<Symbols> | null
        return identifier
      })
      .filter((ident): ident is RawIdentifier<Symbols> => !!ident)
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

/**
 * printFindings serializes an array of {@link RawIdentifier} objects into a
 * pretty-printed JSON string.
 */
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
