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
 * RawIdentifier represents a unique identifier for a symbol in a TypeScript
 * codebase. It can be one of the following types: RawGlobalIdentifier,
 * RawMethodIdentifier, or RawPropertyIdentifier. A RawGlobalIdentifier is a
 * string that contains a global symbol and its name. A RawMethodIdentifier is a
 * string that represents a method with its owner and method name. A
 * RawPropertyIdentifier is a string that represents a property with its owner
 * and property name. RawIdentifiers are used to identify and distinguish
 * different symbols in the codebase, such as variables, functions, classes,
 * interfaces, methods, properties, and types.
 */
export type RawIdentifier<Symbols extends SymbolType = SymbolType> =
  | RawGlobalIdentifier<Symbols & GlobalSymbol>
  | RawMethodIdentifier
  | RawPropertyIdentifier

/**
 * RawGlobalIdentifier represents a unique identifier for global symbols, such
 * as exported functions, classes, interfaces, and types. It is a string
 * template that combines the symbol type and the symbol's name, separated by a
 * colon (e.g., "class:ExampleClass"). This identifier is used to differentiate
 * global symbols in the codebase and can be parsed into more specific
 * Identifier types for further processing.
 */
export type RawGlobalIdentifier<Symbols extends GlobalSymbol = GlobalSymbol> =
  `${Symbols}:${string}`

/**
 * RawMethodIdentifier represents a unique identifier for a method within a
 * TypeScript codebase. It follows the format
 * "method:${ownerName}.${methodName}", where ${ownerName} is the name of the
 * class or interface that owns the method, and ${methodName} is the name of the
 * method itself. This identifier is used for tracking and referencing methods
 * across different parts of the codebase.
 */
export type RawMethodIdentifier = `method:${string}.${string}`

/**
 * RawPropertyIdentifier is a type representing the unique identifier for a
 * property within an owner object in the format
 * "prop:{ownerName}.{propertyName}". It is used to differentiate between
 * properties of different owner objects and to facilitate the parsing and
 * formatting of property identifiers.
 */
export type RawPropertyIdentifier = `prop:${string}.${string}`

/**
 * Identifier represents a unique identifier for a TypeScript entity, such as a
 * variable, function, class, interface, method, property, or type. It provides
 * a way to reference these entities in the codebase and can be used for tasks
 * like code navigation, refactoring, or analysis. Identifier types include
 * {@link VariableIdentifier}, {@link FunctionIdentifier}, {@link
 * ClassIdentifier}, {@link InterfaceIdentifier}, {@link MethodIdentifier},
 * {@link PropertyIdentifier}, and {@link TypeIdentifier}. Each of these types
 * contains a `type` field indicating its specific identifier type and a `path`
 * field representing the unique path of the entity within the codebase.
 *
 * Functions are provided to create, parse, and format identifiers, as well as
 * to check if a given value is an identifier or a specific type of identifier.
 * These functions include {@link createRawIdentifier}, {@link parseIdentifier},
 * {@link formatIdentifier}, {@link isRawIdentifier}, and various type-checking
 * functions such as {@link isVariableIdentifier}, {@link isFunctionIdentifier},
 * and so on.
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
 * VariableIdentifier is an interface that represents a variable identifier
 * within a TypeScript program. It contains three properties: `type`, `path`,
 * and `variableName`. The `type` property has a value of 'var', indicating that
 * the identifier is for a variable. The `path` property represents the location
 * of the variable in the program, while the `variableName` property holds the
 * name of the variable. This interface is used to uniquely identify variables
 * and provide information about their location and name within a TypeScript
 * program.
 */
export interface VariableIdentifier {
  /**
   * The `VariableIdentifier.type` property is a string that represents the type
   * of identifier for a variable. It is set to `'var'` to indicate that the
   * identifier is for a variable. This property is used in conjunction with
   * other properties like `path` and `variableName` to uniquely identify a
   * variable in the codebase. It helps in distinguishing variable identifiers
   * from other types of identifiers such as functions, classes, interfaces,
   * methods, properties, and types.
   */
  type: 'var'
  /**
   * The `path` property of the `VariableIdentifier` object represents the
   * unique identifier for a variable within the code. It is a string that helps
   * to locate and differentiate the variable from other elements in the
   * codebase. This property is used in various operations like parsing,
   * formatting, and comparing identifiers.
   */
  path: string
  /**
   * The property "variableName" of the VariableIdentifier interface represents
   * the name of a variable. It is used to uniquely identify a variable within a
   * given scope or context. This property is a string, and it is an essential
   * part of the VariableIdentifier object, which provides information about the
   * variable's type, path, and name.
   */
  variableName: string
}

/**
 * FunctionIdentifier represents a unique identifier for a function within a
 * module. It contains the following properties:
 *
 * - `type`: A string with the value 'func', indicating that the identifier
 * represents a function.
 * - `path`: A string representing the path to the module containing the
 * function.
 * - `functionName`: A string representing the name of the function.
 *
 * FunctionIdentifier is used to uniquely identify and reference functions
 * across different modules in a codebase.
 */
export interface FunctionIdentifier {
  /**
   * The `FunctionIdentifier.type` property is a string with the value `'func'`.
   * It is used to identify that the associated `Identifier` object represents a
   * function in the TypeScript code. This property helps differentiate between
   * various types of identifiers, such as variables, classes, methods,
   * properties, and interfaces.
   */
  type: 'func'
  /**
   * The "path" property in the "FunctionIdentifier" interface represents the
   * unique identifier for a specific function. It is a string that helps to
   * distinguish this function from other functions, classes, interfaces,
   * methods, properties, and types within the same module or package. This
   * allows for more efficient referencing and searching of functions in large
   * codebases.
   */
  path: string
  /**
   * The `FunctionIdentifier.functionName` property represents the name of a
   * function in a {@link FunctionIdentifier} object. This property is used to
   * uniquely identify a function within a TypeScript project.
   */
  functionName: string
}

/**
 * ClassIdentifier represents the unique identifier for a class in a TypeScript
 * program. It contains information about the class's type, path, and class
 * name. The `type` property is set to `'class'`, the `path` property contains
 * the file path where the class is defined, and the `className` property holds
 * the name of the class. This identifier is used to track and reference classes
 * throughout a TypeScript codebase.
 */
export interface ClassIdentifier {
  /**
   * The `ClassIdentifier.type` property is a string with the value `'class'`.
   * It indicates that the associated `Identifier` object represents a class in
   * the TypeScript code. This property is used to differentiate between various
   * types of identifiers, such as variables, functions, classes, interfaces,
   * methods, properties, and types.
   */
  type: 'class'
  /**
   * The `path` property in the `ClassIdentifier` interface represents the
   * unique identifier for a class within a TypeScript program. It contains the
   * class name as a string and is used to uniquely identify and reference the
   * class throughout the codebase. This property is also present in other
   * identifier interfaces such as `VariableIdentifier`, `FunctionIdentifier`,
   * `InterfaceIdentifier`, `MethodIdentifier`, `PropertyIdentifier`, and
   * `TypeIdentifier`.
   */
  path: string
  /**
   * The `ClassIdentifier.className` property represents the name of the class
   * in a `ClassIdentifier` object. It is a string value that uniquely
   * identifies the class within its scope. This property is used to store and
   * retrieve information about a specific class, such as its methods,
   * properties, and other related metadata.
   */
  className: string
}

/**
 * InterfaceIdentifier represents an identifier for an exported interface in a
 * TypeScript module. It contains the following properties:
 *
 * - type: A string with the value 'iface', indicating that this identifier
 * represents an interface.
 * - path: A string representing the path of the TypeScript module where the
 * interface is defined.
 * - interfaceName: A string representing the name of the interface.
 */
export interface InterfaceIdentifier {
  /**
   * The `InterfaceIdentifier.type` property is a string representing the type
   * of an interface identifier. It is used to distinguish between different
   * types of identifiers in a TypeScript program. The value for this property
   * is always set to `'iface'` for interface identifiers. This property, along
   * with other properties like `path` and `interfaceName`, together form the
   * complete information about an interface identifier.
   */
  type: 'iface'
  /**
   * The `path` property of the `InterfaceIdentifier` object represents the
   * location of the interface within the source code. It is a string that
   * uniquely identifies the interface in the codebase.
   */
  path: string
  /**
   * Property "InterfaceIdentifier.interfaceName" represents the name of the
   * interface for an {@link InterfaceIdentifier} object. This property is a
   * string value that uniquely identifies the interface within its
   * corresponding TypeScript source file.
   */
  interfaceName: string
}

/**
 * MethodIdentifier represents a method within a TypeScript program. It contains
 * information about the method's type, path, owner name, and method name. The
 * type is always set to 'method', the path is a string representation of the
 * method's location within the program, the owner name refers to the class or
 * interface that the method belongs to, and the method name is the actual name
 * of the method. This identifier can be used to reference a specific method in
 * a TypeScript program for various purposes such as documentation or code
 * analysis.
 */
export interface MethodIdentifier {
  /**
   * The `MethodIdentifier.type` property is a string with the value `'method'`.
   * It represents the type of identifier for a method within an object, class,
   * or interface. This property is used to distinguish method identifiers from
   * other types of identifiers, such as variables, functions, classes,
   * interfaces, properties, and types. When working with an `Identifier`
   * object, you can use the `MethodIdentifier.type` property to determine if
   * the identifier is a method identifier by checking if its value is
   * `'method'`.
   */
  type: 'method'
  /**
   * The `MethodIdentifier.path` property represents the path of a method
   * identifier. It is a string that consists of the owner name and the method
   * name, separated by a dot. The owner name can be a class or an interface,
   * while the method name is the name of the method itself. This property is
   * used to uniquely identify a method within a given context.
   */
  path: string
  /**
   * The `ownerName` property of a {@link MethodIdentifier} object represents
   * the name of the class, interface, or type alias that owns the method. This
   * property is used to uniquely identify the method within the context of its
   * owner.
   */
  ownerName: string
  /**
   * The `MethodIdentifier.methodName` property represents the name of the
   * method in a `MethodIdentifier` object. It is a string value that
   * corresponds to the method's name within its owner (e.g., class or
   * interface). This property is used to uniquely identify a method when
   * working with TypeScript symbols and nodes.
   */
  methodName: string
}

/**
 * PropertyIdentifier represents a property of an exported owner (e.g., class,
 * interface) in a TypeScript module. It contains the following fields:
 *
 * - `type`: A string with the value 'prop', representing that this identifier
 * is for a property.
 * - `path`: A string representing the path to the module containing the
 * property.
 * - `ownerName`: A string representing the name of the exported owner (e.g.,
 * class or interface) that contains the property.
 * - `propertyName`: A string representing the name of the property itself.
 *
 * This identifier is used to uniquely identify and reference properties in
 * TypeScript code.
 */
export interface PropertyIdentifier {
  /**
   * The `PropertyIdentifier.type` property is a string that represents the type
   * of an identifier object for a property. Its value is always `'prop'`. The
   * `PropertyIdentifier` object also contains other properties such as `path`,
   * `ownerName`, and `propertyName` that provide additional information about
   * the property identifier. This is useful for differentiating between various
   * types of identifiers in the codebase, such as variables, functions,
   * classes, interfaces, methods, and types.
   */
  type: 'prop'
  /**
   * The `PropertyIdentifier.path` property represents the file path of the
   * TypeScript source file where a specific property is defined. It is part of
   * the `PropertyIdentifier` object, which includes information about the
   * property's type, owner name, and property name. The `path` property helps
   * to locate the exact source file containing the property definition for
   * further analysis or processing.
   */
  path: string
  /**
   * The `PropertyIdentifier.ownerName` property represents the name of the
   * owner (class, interface, or type alias) of a property in a TypeScript
   * program. It is a string that stores the name of the owner where the
   * property is defined. This information can be used to identify and reference
   * properties within their respective owners in the TypeScript code.
   */
  ownerName: string
  /**
   * The `PropertyIdentifier.propertyName` property represents the name of a
   * property within an object or class. It is used in conjunction with the
   * `ownerName` property to uniquely identify a specific property in the
   * context of a larger codebase.
   */
  propertyName: string
}

/**
 * TypeIdentifier represents a type identifier, which is used to uniquely
 * identify a TypeScript type. It contains the following properties:
 *
 * - `type`: A string with the value "type".
 * - `path`: A string representing the path of the type.
 * - `typeName`: A string representing the name of the type.
 *
 * TypeIdentifier is part of the Identifier union type, which also includes
 * VariableIdentifier, FunctionIdentifier, ClassIdentifier, InterfaceIdentifier,
 * MethodIdentifier, and PropertyIdentifier. Each of these represents a
 * different kind of TypeScript identifier.
 */
export interface TypeIdentifier {
  /**
   * The `TypeIdentifier.type` property is a string that indicates the type of
   * identifier for a TypeScript type. It has a value of `'type'` and is used to
   * differentiate between various identifier types, such as variable, function,
   * class, interface, method, or property identifiers. This property is part of
   * the `TypeIdentifier` interface, which also includes the `path` and
   * `typeName` properties for representing the location and name of the
   * TypeScript type respectively.
   */
  type: 'type'
  /**
   * The `TypeIdentifier.path` property represents the path of the type
   * identifier. It is a string that uniquely identifies a specific type within
   * the codebase. This property is used to reference and access the type when
   * needed, such as during code generation or analysis. The
   * `TypeIdentifier.path` is part of the `TypeIdentifier` object, which also
   * includes a `type` property (with value `'type'`) and a `typeName` property
   * containing the name of the type.
   */
  path: string
  /**
   * The `TypeIdentifier.typeName` property represents the name of a TypeScript
   * type. It is a string value that is part of the `TypeIdentifier` object,
   * which is used to uniquely identify a TypeScript type within the codebase.
   * This property is essential when working with TypeScript types, as it allows
   * you to reference and manipulate them in various ways, such as creating new
   * instances or checking compatibility between different types.
   */
  typeName: string
}

/**
 * The `isVariableIdentifier()` function is a type guard that checks if the
 * given `identifier` is of the `VariableIdentifier` type. It returns `true` if
 * the `type` property of the input `identifier` is `'var'`, and `false`
 * otherwise.
 */
export function isVariableIdentifier(
  identifier: Identifier,
): identifier is VariableIdentifier {
  return identifier.type === 'var'
}

/**
 * The `isFunctionIdentifier()` function is a type guard that checks if the
 * given `identifier` is of the `FunctionIdentifier` type. It returns `true` if
 * the `identifier` has a `type` property with the value `'func'`, and `false`
 * otherwise.
 */
export function isFunctionIdentifier(
  identifier: Identifier,
): identifier is FunctionIdentifier {
  return identifier.type === 'func'
}

/**
 * The isClassIdentifier() function is a type guard that checks if the given
 * identifier is of the ClassIdentifier type. It takes an Identifier object as
 * input and returns a boolean value indicating whether the input identifier
 * represents a class or not. The function checks the type property of the input
 * identifier, and if its value is 'class', it returns true; otherwise, it
 * returns false.
 */
export function isClassIdentifier(
  identifier: Identifier,
): identifier is ClassIdentifier {
  return identifier.type === 'class'
}

/**
 * The `isMethodIdentifier()` function is a type guard that checks if a given
 * identifier is of type `MethodIdentifier`. It takes an `Identifier` object as
 * its input and returns a boolean value indicating whether the input identifier
 * is a method identifier or not. The function checks the `type` property of the
 * input identifier and returns `true` if it is equal to `'method'`, otherwise,
 * it returns `false`.
 */
export function isMethodIdentifier(
  identifier: Identifier,
): identifier is MethodIdentifier {
  return identifier.type === 'method'
}

/**
 * The `createRawIdentifier()` function takes a TypeScript node as input and
 * returns a raw identifier, which is a string representation of the symbol
 * associated with the input node. The returned raw identifier is in one of the
 * following formats: `func:${functionName}`, `var:${variableName}`,
 * `class:${className}`, `iface:${interfaceName}`,
 * `method:${ownerName}.${methodName}`, `prop:${ownerName}.${propertyName}`, or
 * `type:${typeName}`. If the input node does not have a valid symbol, the
 * function returns `null`.
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
 * The `findNodeName()` function takes a TypeScript `ts.Declaration` node as its
 * input and returns the name of the declaration. This is useful for extracting
 * the identifier name from a given declaration node in a TypeScript Abstract
 * Syntax Tree (AST).
 */
export function findNodeName(node: ts.Declaration) {
  return ts.getNameOfDeclaration(node)
}

/**
 * The `formatIdentifier()` function takes an `Identifier` object as input and
 * returns a `RawIdentifier` string representation of the input. It converts
 * various types of identifiers, such as variable, function, class, interface,
 * method, property, and type identifiers, into their corresponding raw
 * identifier string formats. The returned raw identifier string can be used for
 * further processing or comparison purposes.
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
 * The `parseIdentifier()` function takes a RawIdentifier as input and returns
 * an Identifier object. It converts the raw identifier string into a structured
 * object with properties such as type, path, and name (e.g., className,
 * methodName, propertyName) based on the SymbolType of the input identifier.
 * The supported SymbolTypes are 'var', 'func', 'class', 'iface', 'method',
 * 'prop', and 'type'. If the input identifier is invalid or has an unsupported
 * SymbolType, an error will be thrown.
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
 * The `isRawIdentifier()` function is used to determine if a given string is a
 * valid RawIdentifier. It checks if the string starts with one of the following
 * prefixes: 'var:', 'func:', 'class:', 'iface:', 'method:', 'prop:', or
 * 'type:'. If the string starts with any of these prefixes, the function
 * returns true, indicating that the string is a valid RawIdentifier. Otherwise,
 * it returns false.
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
