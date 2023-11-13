import ts from 'typescript'
import {
  getClassName,
  getFunctionName,
  getOwnerName,
  getVariableName,
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
import type { GlobalSymbol, SymbolType } from './symbols'

export type RawIdentifier<Symbols extends SymbolType = SymbolType> =
  | RawGlobalIdentifier<Symbols & GlobalSymbol>
  | RawMethodIdentifier
  | RawPropertyIdentifier

export type RawGlobalIdentifier<Symbols extends GlobalSymbol = GlobalSymbol> =
  `${Symbols}:${string}`

export type RawMethodIdentifier = `method:${string}.${string}`

export type RawPropertyIdentifier = `prop:${string}.${string}`

export type Identifier =
  | VariableIdentifier
  | FunctionIdentifier
  | ClassIdentifier
  | InterfaceIdentifier
  | MethodIdentifier
  | PropertyIdentifier
  | TypeIdentifier

/**
 * Represents a unique identifier for a variable within the codebase,
 * encapsulating its name and the path to its location. The `VariableIdentifier`
 * object includes a 'type' field set to `'var'`, indicating the kind of symbol
 * it represents, alongside a 'path' field describing where the variable is
 * defined and a 'variableName' field holding the name of the variable itself.
 * This type is part of a larger system used for identifying and working with
 * different kinds of symbols in a code analysis or transformation tool.
 */
export interface VariableIdentifier {
  type: 'var'
  path: string
  variableName: string
}

/**
 * Represents the unique signature of a function within the system,
 * encapsulating its name and the path where it is defined. It is used to
 * identify and reference functions across different parts of the application,
 * ensuring that each function can be distinctly recognized by its
 * `functionName` and associated `path`. This identifier is particularly useful
 * for tasks such as serialization, documentation generation, or when performing
 * analysis and transformations on the codebase.
 */
export interface FunctionIdentifier {
  type: 'func'
  path: string
  functionName: string
}

/**
 * ClassIdentifier represents a unique identifier for a class within a codebase.
 * It includes the 'class' type, the file path where the class is defined, and
 * the name of the class itself. This construct is used to uniquely identify and
 * reference classes across different parts of the system.
 */
export interface ClassIdentifier {
  type: 'class'
  path: string
  className: string
}

/**
 * InterfaceIdentifier represents a unique identifier for an interface within a
 * codebase. It includes the 'type' field with a value of 'iface', indicating
 * the identifier type, along with 'path' and 'interfaceName' fields that
 * specify the location and name of the interface respectively. This identifier
 * is used to uniquely distinguish interfaces across different modules and
 * namespaces.
 */
export interface InterfaceIdentifier {
  type: 'iface'
  path: string
  interfaceName: string
}

/**
 * Represents a method within a TypeScript codebase, uniquely identified by its
 * type, file path, owning class or interface name, and the method name itself.
 * This is part of a system used to reference and distinguish methods for
 * tooling purposes.
 */
export interface MethodIdentifier {
  type: 'method'
  path: string
  ownerName: string
  methodName: string
}

/**
 * Represents a unique identifier for a property within a given owner, such as a
 * class or an interface. `PropertyIdentifier` includes the type of the
 * identifier, the file path where the property is declared, the name of the
 * owner (class or interface), and the property name itself. This enables
 * precise referencing and manipulation of property symbols in code analysis
 * tools.
 */
export interface PropertyIdentifier {
  type: 'prop'
  path: string
  ownerName: string
  propertyName: string
}

/**
 * Represents an identifier for a type within the codebase, encapsulating the
 * specific name of the type as well as its location. The `TypeIdentifier`
 * object includes a `type` field set to `'type'`, a `path` field containing the
 * file path where the type is defined, and a `typeName` field holding the name
 * of the type itself. This identifier is typically used in scenarios where
 * types need to be referenced or processed in a structured manner.
 */
export interface TypeIdentifier {
  type: 'type'
  path: string
  typeName: string
}

/**
 * Determines whether a given {@link Identifier} is a {@link
 * VariableIdentifier}. Returns `true` if the identifier's type property is
 * equivalent to `'var'`, indicating that it represents a variable. Otherwise,
 * returns `false`.
 */
export function isVariableIdentifier(
  identifier: Identifier,
): identifier is VariableIdentifier {
  return identifier.type === 'var'
}

/**
 * Determines whether a given {@link Identifier} represents a function by
 * checking if its type property is equivalent to 'func'. If true, the
 * identifier can be safely cast to a {@link FunctionIdentifier}.
 */
export function isFunctionIdentifier(
  identifier: Identifier,
): identifier is FunctionIdentifier {
  return identifier.type === 'func'
}

/**
 * Determines if the provided {@link Identifier} is a {@link ClassIdentifier}.
 * Returns `true` if the {@link Identifier}'s type is 'class', indicating it
 * represents a class.
 */
export function isClassIdentifier(
  identifier: Identifier,
): identifier is ClassIdentifier {
  return identifier.type === 'class'
}

/**
 * Determines if the provided {@link Identifier} is a {@link MethodIdentifier}.
 * It checks the `type` property of the given identifier and returns `true` if
 * it matches 'method', indicating that the identifier represents a method.
 * Otherwise, it returns `false`.
 */
export function isMethodIdentifier(
  identifier: Identifier,
): identifier is MethodIdentifier {
  return identifier.type === 'method'
}

/**
 * Generates a raw textual representation of a TypeScript symbol from a given
 * node, if the symbol is exported. Depending on the node's kind and visibility,
 * it returns a {@link RawIdentifier} that follows a specific format for
 * different symbol types such as functions, variables, classes, interfaces,
 * methods, properties, or types. If the node does not represent an exported
 * symbol or its kind cannot be determined for identifier creation, it returns
 * `null`. In cases where the identifier cannot be created due to an unexpected
 * node structure, it throws an error.
 */
export function createRawIdentifier(node: ts.Node): RawIdentifier | null {
  if (isExportedFunction(node)) {
    return `func:${getFunctionName(node)!}`
  }

  if (isExportedVariable(node)) {
    return `var:${getVariableName(node)}`
  }

  if (isExportedClass(node)) {
    return `class:${getClassName(node)!}`
  }

  if (isExportedInterface(node)) {
    return `iface:${node.name.getText()}`
  }

  if (
    isPublicMethodOfExportedClass(node) ||
    isMethodOfExportedInterface(node) ||
    isMethodOfExportedTypeAlias(node)
  ) {
    const interfaceName = getOwnerName(node)!
    if (!interfaceName) {
      return null
    }
    return `method:${interfaceName}.${node.name.getText()}`
  }

  if (isPublicPropertyOfExportedOwner(node)) {
    const ownerName = getOwnerName(node)
    if (!ownerName) {
      return null
    }
    return `prop:${ownerName}.${node.name.getText()}`
  }

  if (isExportedType(node)) {
    return `type:${node.name.getText()}`
  }

  throw new Error(`Cannot create identifier for node:\n${node.getText()}`)
}

/**
 * Retrieves the name of the provided TypeScript declaration node. If the node
 * is a declaration with a name (such as a class, function, variable, etc.),
 * `findNodeName` will return that name. Otherwise, it returns `undefined`. This
 * function is useful when you need to extract the identifier name from a
 * TypeScript AST node that represents a declaration.
 */
export function findNodeName(node: ts.Declaration) {
  return ts.getNameOfDeclaration(node)
}

/**
 * Converts a given {@link Identifier} into a {@link RawIdentifier} by
 * assembling its components into a string representation that follows a
 * specific format based on the identifier type. Each identifier type is
 * prefixed with an indicative keyword (e.g., 'var', 'func', 'class', etc.)
 * followed by relevant naming details, separated by colons or periods as
 * necessary to denote scopes and ownership.
 */
export function formatIdentifier(identifier: Identifier): RawIdentifier {
  switch (identifier.type) {
    case 'var':
      return `var:${identifier.variableName}`
    case 'func':
      return `func: ${identifier.functionName}`
    case 'class':
      return `class:${identifier.className}`
    case 'iface':
      return `iface:${identifier.interfaceName}`
    case 'method':
      return `method:${identifier.ownerName}.${identifier.methodName}`
    case 'prop':
      return `prop:${identifier.ownerName}.${identifier.propertyName}`
    case 'type':
      return `type:${identifier.typeName}`
  }
}

/**
 * Parses a given raw identifier string into a structured {@link Identifier}
 * object, separating and classifying the type and path of the identifier based
 * on its prefix and structure. This function throws an error if the raw
 * identifier does not conform to expected formats or if parsing fails due to
 * missing or invalid components.
 */
export function parseIdentifier(identifier: RawIdentifier): Identifier {
  const [type, path] = identifier.split(':') as [SymbolType, string]

  if (!type || !path) {
    throw new Error(`Invalid identifier: ${identifier}`)
  }

  switch (type) {
    case 'var':
      return parseVariableIdentifier(path)
    case 'func':
      return parseFunctionIdentifier(path)
    case 'class':
      return parseClassIdentifier(path)
    case 'iface':
      return parseInterfaceIdentifier(path)
    case 'method':
      return parseMethodIdentifier(path)
    case 'prop':
      return parsePropertyIdentifier(path)
    case 'type':
      return parseTypeIdentifier(path)
  }
}

function parseMethodIdentifier(path: string): MethodIdentifier {
  const [ownerName, methodName] = path.split('.')
  if (!ownerName || !methodName)
    throw new Error(`Invalid method identifier: ${path}`)

  return {
    type: 'method',
    path,
    ownerName,
    methodName,
  }
}

function parseFunctionIdentifier(path: string): FunctionIdentifier {
  return {
    type: 'func',
    path,
    functionName: path,
  }
}

function parseClassIdentifier(path: string): ClassIdentifier {
  return {
    type: 'class',
    path,
    className: path,
  }
}

function parseInterfaceIdentifier(path: string): InterfaceIdentifier {
  return {
    type: 'iface',
    path,
    interfaceName: path,
  }
}

function parseVariableIdentifier(path: string): VariableIdentifier {
  return {
    type: 'var',
    path,
    variableName: path,
  }
}

function parsePropertyIdentifier(path: string): PropertyIdentifier {
  const [ownerName, propertyName] = path.split('.')
  if (!ownerName || !propertyName)
    throw new Error(`Invalid property identifier: ${path}`)

  return {
    type: 'prop',
    path,
    ownerName,
    propertyName,
  }
}

function parseTypeIdentifier(path: string): TypeIdentifier {
  return {
    type: 'type',
    path,
    typeName: path,
  }
}

/**
 * Determines if a given string is a {@link RawIdentifier} by checking if it
 * starts with one of the recognized prefixes for variable, function, class,
 * interface, method, property, or type identifiers. Returns `true` if the
 * string is a valid raw identifier, otherwise `false`.
 */
export function isRawIdentifier(s: string): s is RawIdentifier {
  return (
    s.startsWith('var:') ||
    s.startsWith('func:') ||
    s.startsWith('class:') ||
    s.startsWith('iface:') ||
    s.startsWith('method:') ||
    s.startsWith('prop:') ||
    s.startsWith('type:')
  )
}
