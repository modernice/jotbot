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

/**
 * The type "RawIdentifier" represents a string that identifies a TypeScript
 * node or symbol. It can be one of three subtypes: "RawGlobalIdentifier",
 * "RawMethodIdentifier", or "RawPropertyIdentifier". The "RawGlobalIdentifier"
 * subtype represents a global symbol, while the "RawMethodIdentifier" and
 * "RawPropertyIdentifier" subtypes represent methods and properties of classes
 * or interfaces respectively. The type can be converted to an "Identifier"
 * object using the "parseIdentifier()" function and vice versa using the
 * "formatIdentifier()" function.
 */
export type RawIdentifier<Symbols extends SymbolType = SymbolType> =
  | RawGlobalIdentifier<Symbols & GlobalSymbol>
  | RawMethodIdentifier
  | RawPropertyIdentifier

/**
 * The type "RawGlobalIdentifier" represents a string that starts with a global
 * symbol followed by a colon and a string identifier. It is used as part of the
 * larger system for identifying TypeScript nodes and their properties.
 */
export type RawGlobalIdentifier<Symbols extends GlobalSymbol = GlobalSymbol> =
  `${Symbols}:${string}`

/**
 * The "RawMethodIdentifier" type is a string literal type that represents a
 * method identifier in raw form. It has the format of
 * "method:{ownerName}.{methodName}". This type is used as part of the larger
 * identifier system to represent various types of identifiers for TypeScript
 * nodes.
 */
export type RawMethodIdentifier = `method:${string}.${string}`

/**
 * The type "RawPropertyIdentifier" represents a string literal that starts with
 * "prop:" and is followed by the name of an owner (class or interface) and a
 * property name separated by a dot. It is one of three possible types for the
 * RawIdentifier type, which is used to represent various types of identifiers
 * in TypeScript code. The RawPropertyIdentifier is used specifically to
 * represent property identifiers in the code.
 */
export type RawPropertyIdentifier = `prop:${string}.${string}`

/**
 * The "Identifier" type is a union type that represents different kinds of
 * identifiers in TypeScript, such as variable, function, class, interface,
 * method, property, and type identifiers. It also includes functions for
 * creating, formatting, and parsing raw identifiers. Raw identifiers are
 * string-based representations of identifiers that include a prefix indicating
 * the type of the identifier and a path to the identifier in the source code.
 * The "Identifier" type and its related functions are used in various parts of
 * the library to identify and manipulate symbols and nodes in TypeScript code.
 */
export type Identifier =
  | VariableIdentifier
  | FunctionIdentifier
  | ClassIdentifier
  | InterfaceIdentifier
  | MethodIdentifier
  | PropertyIdentifier
  | TypeIdentifier

/**
 * The "iface:VariableIdentifier" interface represents an identifier for an
 * exported interface in TypeScript code. It contains information about the type
 * of the identifier, the path to the file where the interface is defined, and
 * the name of the exported interface.
 */
export interface VariableIdentifier {
  /**
   * The property "VariableIdentifier.type" is a string literal that indicates
   * the type of an identifier object. Specifically, for a VariableIdentifier
   * object, the type property's value is always set to the string "var".
   */
  type: 'var'
  /**
   * The "VariableIdentifier.path" property is a string that represents the path
   * to a variable declaration within a module. It is used in conjunction with
   * other properties of the "VariableIdentifier" interface to uniquely identify
   * a variable in the module.
   */
  path: string
  /**
   * Describes the property "VariableIdentifier.variableName", which is a string
   * representing the name of a variable in an identifier object. This property
   * is used in conjunction with other properties in the identifier object to
   * uniquely identify a code entity such as a variable, function, class, or
   * interface.
   */
  variableName: string
}

/**
 * The "iface:FunctionIdentifier" is an interface identifier that represents a
 * TypeScript interface. It contains the type, path, and interfaceName
 * properties and is used to identify interfaces in a TypeScript codebase.
 */
export interface FunctionIdentifier {
  /**
   * The property "FunctionIdentifier.type" is a string literal that specifies
   * the type of an identifier object. It can be one of the seven possible
   * values: 'var', 'func', 'class', 'iface', 'method', 'prop', or 'type'. This
   * property is used to differentiate between different types of identifiers
   * and to provide type safety when working with them.
   */
  type: 'func'
  /**
   * The "FunctionIdentifier.path" property is a string that represents the path
   * of the function within the source code. It is used as part of an identifier
   * to uniquely identify a function within a project.
   */
  path: string
  /**
   * The property "FunctionIdentifier.functionName" is a string that represents
   * the name of a function in the context of a path.
   */
  functionName: string
}

/**
 * Describes an interface identifier, which is part of the `Identifier` union
 * type. An interface identifier has a `type` of `'iface'`, and contains
 * information about the path and name of the interface. It is used in various
 * functions throughout the library for identifying TypeScript nodes.
 */
export interface ClassIdentifier {
  /**
   * The `ClassIdentifier.type` property is a string literal that specifies the
   * type of an identifier object as "class". This property is used to identify
   * and distinguish class identifiers from other types of identifiers in the
   * code.
   */
  type: 'class'
  /**
   * The property "ClassIdentifier.path" is a string that represents the path of
   * a class in a TypeScript module. It is used as part of an identifier to
   * uniquely identify a class in the module.
   */
  path: string
  /**
   * The property "ClassIdentifier.className" is a string property that
   * represents the name of a class. It is used as a part of an identifier to
   * uniquely identify a class in the TypeScript code.
   */
  className: string
}

/**
 * Represents an interface identifier, which is used to identify and reference
 * TypeScript interfaces. It has a 'type' property of value 'iface', a 'path'
 * property that contains the full path of the interface, and an 'interfaceName'
 * property that contains the name of the interface. It is used as part of a
 * union type called Identifier, which is used to identify variables, functions,
 * classes, interfaces, methods, properties, and types.
 */
export interface InterfaceIdentifier {
  /**
   * The property "InterfaceIdentifier.type" is a string literal type that
   * specifies the type of an identifier object. In the case of an
   * InterfaceIdentifier, it will always be set to "iface".
   */
  type: 'iface'
  /**
   * The "InterfaceIdentifier.path" property is a string that represents the
   * path of an interface identifier. It is used as part of the raw identifier
   * format and can be parsed to obtain an InterfaceIdentifier object.
   */
  path: string
  /**
   * The property "InterfaceIdentifier.interfaceName" is a string property that
   * represents the name of an interface. It is used as a part of an identifier
   * to uniquely identify an interface in the codebase.
   */
  interfaceName: string
}

/**
 * The "iface:MethodIdentifier" is an interface that represents the identifier
 * for a method of an exported interface. It contains the type, path, ownerName
 * (which can be either className or interfaceName), and methodName. The
 * "parseMethodIdentifier" function can be used to parse a string into a
 * MethodIdentifier object, and the "isMethodIdentifier" function can be used to
 * check if an Identifier object is a MethodIdentifier.
 */
export interface MethodIdentifier {
  /**
   * The `MethodIdentifier.type` property is a string literal that indicates the
   * type of the identifier, which is always set to `'method'` for a
   * `MethodIdentifier` object. This property is used to distinguish between
   * different types of identifiers when working with them in TypeScript code.
   */
  type: 'method'
  /**
   * The "MethodIdentifier.path" property is a string that represents the path
   * of a method within a class or interface. The path consists of the name of
   * the owner (class or interface) followed by a dot and the name of the
   * method.
   */
  path: string
  /**
   * The property "MethodIdentifier.ownerName" represents the name of the class
   * or interface that owns the method. It is a string value that is used to
   * identify the owner of a method within an {@link Identifier}.
   */
  ownerName: string // className or interfaceName
  methodName: string
}

/**
 * The `iface:PropertyIdentifier` interface represents an identifier for a
 * property of an exported interface. It contains information about the type of
 * identifier (`type`), the path to the source file of the property (`path`),
 * the name of the owning interface (`ownerName`) and the name of the property
 * (`propertyName`). This interface is used in conjunction with other identifier
 * interfaces to identify different types of symbols in TypeScript source code.
 */
export interface PropertyIdentifier {
  /**
   * The `PropertyIdentifier.type` property is a string literal type that
   * indicates the type of identifier for a property. It can have the value
   * `'prop'`. This property is used in conjunction with other properties in the
   * `PropertyIdentifier` interface to uniquely identify a property.
   */
  type: 'prop'
  /**
   * The "PropertyIdentifier.path" property is a string that represents the
   * fully qualified path of a property in a class or interface, including the
   * name of the owner (class or interface) and the name of the property. It is
   * used as part of an identifier to uniquely identify a property in TypeScript
   * code.
   */
  path: string
  /**
   * The "ownerName" property of the "PropertyIdentifier" interface represents
   * the name of the class or interface that owns the property. It is used in
   * conjunction with the "propertyName" property to uniquely identify a
   * property.
   */
  ownerName: string // className or interfaceName
  propertyName: string
}

/**
 * The "iface:TypeIdentifier" interface represents an identifier for an exported
 * TypeScript interface. It contains information about the path and name of the
 * interface.
 */
export interface TypeIdentifier {
  /**
   * The "TypeIdentifier.type" property is a string literal type that indicates
   * the type of an Identifier object as "type". It is used in conjunction with
   * other properties of the Identifier object, such as "path" and "typeName",
   * to uniquely identify a TypeScript symbol.
   */
  type: 'type'
  /**
   * The property `TypeIdentifier.path` is a string that represents the path of
   * the type identifier.
   */
  path: string
  /**
   * The `TypeIdentifier.typeName` property is a string that represents the name
   * of a type. It is used as a part of an identifier to uniquely identify the
   * type in the program.
   */
  typeName: string
}

/**
 * Checks whether an identifier is a VariableIdentifier. Returns true if the
 * identifier's type is 'var', indicating that it represents a variable, and
 * false otherwise.
 */
export function isVariableIdentifier(
  identifier: Identifier,
): identifier is VariableIdentifier {
  return identifier.type === 'var'
}

/**
 * Checks if the given identifier is a function identifier. Returns true if the
 * identifier is a function identifier, false otherwise.
 */
export function isFunctionIdentifier(
  identifier: Identifier,
): identifier is FunctionIdentifier {
  return identifier.type === 'func'
}

/**
 * Checks if an identifier is a ClassIdentifier, which is an object that
 * represents a class declaration in TypeScript code. Returns a boolean value
 * indicating whether the identifier is a ClassIdentifier or not.
 */
export function isClassIdentifier(
  identifier: Identifier,
): identifier is ClassIdentifier {
  return identifier.type === 'class'
}

/**
 * Checks if an identifier is a method identifier, which is identified by a
 * string starting with "method:" and containing the name of the owner (class or
 * interface) and the name of the method separated by a dot.
 */
export function isMethodIdentifier(
  identifier: Identifier,
): identifier is MethodIdentifier {
  return identifier.type === 'method'
}

/**
 * Creates a raw identifier string based on the given TypeScript node. The raw
 * identifier string is used to identify the node in a unique way and can be
 * parsed into an Identifier object using the parseIdentifier() function. The
 * type of the returned raw identifier depends on the type of the node, such as
 * "func" for functions, "class" for classes, "prop" for properties, etc. If the
 * node type cannot be identified or is not supported, an error is thrown.
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
 * Finds the name of a given TypeScript declaration node using the
 * "getNameOfDeclaration" function from the TypeScript library.
 */
export function findNodeName(node: ts.Declaration) {
  return ts.getNameOfDeclaration(node)
}

/**
 * Formats an {@link Identifier} into a {@link RawIdentifier}, which is a string
 * representation of the identifier that can be used for storage or comparison
 * purposes. The format of the raw identifier depends on the type of identifier,
 * and follows the conventions defined in the type definitions for {@link
 * RawIdentifier}, {@link VariableIdentifier}, {@link FunctionIdentifier},
 * {@link ClassIdentifier}, {@link InterfaceIdentifier}, {@link
 * MethodIdentifier}, {@link PropertyIdentifier}, and {@link TypeIdentifier}.
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
 * Parses a given raw identifier string and returns an object representing the
 * identifier type and its properties. The identifier can represent a variable,
 * function, class, interface, method, property, or type.
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
 * Checks if a given string is a valid {@link RawIdentifier}, which is a string
 * that represents an identifier in a specific format. The function returns true
 * if the string starts with one of the valid prefixes ('var:', 'func:',
 * 'class:', 'iface:', 'method:', 'prop:', or 'type:'), and false otherwise.
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
