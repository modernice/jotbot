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
 * WithSymbolsOption allows specifying an array of {@link SymbolType} to filter
 * the supported nodes while traversing TypeScript code. This is used in
 * conjunction with other options to customize the behavior of the finder
 * function.
 */
export interface WithSymbolsOption<Symbols extends SymbolType = SymbolType> {
  /**
   * Optional property that specifies an array of {@link SymbolType} to be
   * considered when searching for supported nodes within the context of {@link
   * WithSymbolsOption}. If not provided, all available symbol types will be
   * used.
   */
  symbols?: readonly Symbols[]
}

/**
 * GlobOption represents a string or an array of strings that define the
 * patterns for matching files and directories.
 */
export type GlobOption = string | readonly string[]

/**
 * FinderOptions is an interface that extends {@link WithSymbolsOption} and
 * allows to customize the behavior of the symbol finder. It includes an
 * optional `includeDocumented` property, which when set to true, will include
 * documented symbols in the search results.
 */
export interface FinderOptions<Symbols extends SymbolType = SymbolType>
  extends WithSymbolsOption<Symbols> {
  /**
   * Determines whether to include documented nodes in the final result. If set
   * to `true`, both documented and undocumented nodes will be included. If not
   * specified or set to `false`, only undocumented nodes will be included.
   */
  includeDocumented?: boolean
}

/**
 * Creates a finder object with a `find` method, which searches for TypeScript
 * nodes in the given code based on the provided options.
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
 * Converts an array of {@link RawIdentifier} findings into a formatted JSON
 * string.
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
