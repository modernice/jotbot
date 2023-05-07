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
 * The `iface:WithSymbolsOption` interface is an optional configuration option
 * for the `createFinder` function that allows specifying an array of symbol
 * types to include in the search. This interface extends the
 * `WithSymbolsOption` interface and is generic over a `SymbolType`. The symbols
 * included can be any of: variables, functions, classes, interfaces, methods,
 * properties, and types.
 */
export interface WithSymbolsOption<Symbols extends SymbolType = SymbolType> {
  /**
   * The "symbols" property is an optional array of symbol types used to filter
   * the nodes returned by the finder. The finder will only return nodes that
   * match at least one of the specified symbol types.
   */
  symbols?: readonly Symbols[]
}

/**
 * The `GlobOption` type represents a string or an array of strings that can be
 * used to specify file paths or patterns when searching for nodes in TypeScript
 * code. It is used as a parameter in functions like `createFinder` to specify
 * the files to search through.
 */
export type GlobOption = string | readonly string[]

/**
 * The `iface:FinderOptions` interface defines options for a finder function
 * that searches a TypeScript AST for nodes matching certain criteria. It
 * extends the `WithSymbolsOption` interface to include an option for whether to
 * include only documented nodes. The `createFinder` function takes in an
 * optional `FinderOptions` object and returns a finder object with a `find`
 * method that uses the provided options to search a given code string for
 * matching nodes. The `printFindings` function takes in an array of raw
 * identifiers and returns a formatted JSON string of the findings.
 */
export interface FinderOptions<Symbols extends SymbolType = SymbolType>
  extends WithSymbolsOption<Symbols> {
  /**
   * Property `includeDocumented` is an optional boolean flag in the
   * `FinderOptions` interface that, when set to `true`, includes only nodes in
   * the search results that have associated comments. When set to `false` or
   * omitted, all nodes that meet the other search criteria will be included in
   * the results regardless of whether they have comments.
   */
  includeDocumented?: boolean
}

/**
 * Creates a finder object with a "find" function that takes a string of code
 * and returns an array of RawIdentifier objects representing the symbols found
 * in the code. The type of symbols to search for can be optionally specified in
 * the options parameter. If includeDocumented is true in the options parameter,
 * only symbols with documentation comments will be included in the results.
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
 * Prints the findings of identified nodes as a JSON string with two-space
 * indentation.
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
