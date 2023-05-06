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

export interface TypeIdentifier {
  type: 'type'
  path: string
  typeName: string
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
    return `iface:${node.name.getText()}`
  }

  if (
    isPublicMethodOfExportedClass(node) ||
    isMethodOfExportedInterface(node) ||
    isMethodOfExportedTypeAlias(node)
  ) {
    const interfaceName = getOwnerName(node)!
    return `method:${interfaceName}.${node.name.getText()}`
  }

  if (isPublicPropertyOfExportedOwner(node)) {
    const ownerName = getOwnerName(node)!
    return `prop:${ownerName}.${node.name.getText()}`
  }

  if (isExportedType(node)) {
    return `type:${node.name.getText()}`
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
    case 'type':
      return `type:${identifier.typeName}`
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
