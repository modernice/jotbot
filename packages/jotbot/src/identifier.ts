import ts from 'typescript'
import {
  getClassName,
  getClassNameOfMethod,
  getFunctionName,
  getInterfaceName,
  getInterfaceNameOfMethod,
  getMethodName,
  getOwnerName,
  getPropertyName,
  getVariableName,
  isExportedClass,
  isExportedFunction,
  isExportedInterface,
  isExportedVariable,
  isMethodOfExportedInterface,
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

export interface VariableIdentifier {
  type: 'var'
  path: string
  variableName: string
}

export interface FunctionIdentifier {
  type: 'func'
  path: string
  functionName: string
}

export interface ClassIdentifier {
  type: 'class'
  path: string
  className: string
}

export interface InterfaceIdentifier {
  type: 'iface'
  path: string
  interfaceName: string
}

export interface MethodIdentifier {
  type: 'method'
  path: string
  ownerName: string // className or interfaceName
  methodName: string
}

export interface PropertyIdentifier {
  type: 'prop'
  path: string
  ownerName: string // className or interfaceName
  propertyName: string
}

export type NaturalLanguageTarget<Symbols extends SymbolType = SymbolType> =
  | GlobalTarget<Symbols>
  | OwnerTarget<Symbols>

export type GlobalTarget<Symbol extends SymbolType = SymbolType> =
  `${SymbolToName[Symbol]} '${string}'`

export type OwnerTarget<Symbol extends SymbolType = SymbolType> =
  `${SymbolToName[Symbol]} '${string}' of '${string}`

interface SymbolToName extends Record<SymbolType, string> {
  var: 'variable'
  func: 'function'
  class: 'class'
  iface: 'interface'
  method: 'method'
  prop: 'property'
}

export function isVariableIdentifier(
  identifier: Identifier,
): identifier is VariableIdentifier {
  return identifier.type === 'var'
}

export function isFunctionIdentifier(
  identifier: Identifier,
): identifier is FunctionIdentifier {
  return identifier.type === 'func'
}

export function isClassIdentifier(
  identifier: Identifier,
): identifier is ClassIdentifier {
  return identifier.type === 'class'
}

export function isMethodIdentifier(
  identifier: Identifier,
): identifier is MethodIdentifier {
  return identifier.type === 'method'
}

export function createRawIdentifier(node: ts.Node): RawIdentifier {
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
    return `iface:${getInterfaceName(node)}`
  }

  if (isPublicMethodOfExportedClass(node)) {
    const className = getClassNameOfMethod(node)!
    return `method:${className}.${getMethodName(node)}`
  }

  if (isMethodOfExportedInterface(node)) {
    const interfaceName = getInterfaceNameOfMethod(node)!
    return `method:${interfaceName}.${getMethodName(node)}`
  }

  if (isPublicPropertyOfExportedOwner(node)) {
    const ownerName = getOwnerName(node)!
    return `prop:${ownerName}.${getPropertyName(node)}`
  }

  throw new Error(`Cannot create identifier for node:\n${node.getText()}`)
}

export function findNodeName(node: ts.Declaration) {
  return ts.getNameOfDeclaration(node)
}

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
  }
}

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

export function describeIdentifier(
  identifier: RawIdentifier,
): NaturalLanguageTarget {
  const [symbol, name] = identifier.split(':') as [SymbolType, string]

  switch (symbol) {
    case 'var':
      return `variable '${name}'`
    case 'func':
      return `function '${name}'`
    case 'class':
      return `class '${name}'`
    case 'iface':
      return `interface '${name}'`
    case 'method': {
      const [owner, method] = name.split('.')
      return `method '${method}' of '${owner}'`
    }
    case 'prop': {
      const [owner, prop] = name.split('.')
      return `property '${prop}' of '${owner}'`
    }
    default:
      throw new Error(`Unknown symbol type: ${symbol}`)
  }
}

export function isRawIdentifier(s: unknown): s is RawIdentifier {
  return (
    typeof s === 'string' &&
    (s.startsWith('var:') ||
      s.startsWith('func:') ||
      s.startsWith('class:') ||
      s.startsWith('iface:') ||
      s.startsWith('method:') ||
      s.startsWith('prop:'))
  )
}
