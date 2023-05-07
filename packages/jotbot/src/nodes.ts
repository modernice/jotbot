import ts from 'typescript'
import type { Identifier, RawIdentifier } from './identifier'
import { parseIdentifier } from './identifier'

/**
 * Checks if a given TypeScript node has comments. Returns true if there are
 * comments, false otherwise.
 */
export function hasComments(node: ts.Node) {
  return !(
    (ts.getLeadingCommentRanges(node.getFullText(), 0) ?? []).length === 0
  )
}

/**
 * Checks whether a given TypeScript node is exported from its module or not. If
 * the node has an "export" modifier or is a member of an exported parent node,
 * it is considered to be exported. Otherwise, it is not considered to be
 * exported.
 */
export function isExported(
  node: ts.Node & { modifiers?: ts.NodeArray<ts.ModifierLike> },
): boolean {
  if (node.kind === ts.SyntaxKind.ExportKeyword) {
    return false
  }

  if (node.modifiers?.some((mod) => mod.kind === ts.SyntaxKind.ExportKeyword)) {
    return true
  }

  if (node.parent && !ts.isSourceFile(node.parent) && !isWithinFunction(node)) {
    return isExported(node.parent)
  }

  return false
}

/**
 * The `isWithinFunction()` function determines if a given TypeScript node is
 * located within a function. It returns `true` if the node is inside a
 * function, and `false` otherwise. This function is used to help identify if a
 * node is exported or not by checking its position in the code structure.
 */
export function isWithinFunction(node: ts.Node): boolean {
  if (!node.parent) {
    return false
  }
  if (isFunction(node.parent)) {
    return true
  }
  return isWithinFunction(node.parent)
}

/**
 * The `isExportedVariable()` function checks if a given TypeScript node is an
 * exported variable statement. It returns true if the node is a variable
 * statement and is exported, otherwise it returns false.
 */
export function isExportedVariable(
  node: ts.Node,
): node is ts.VariableStatement {
  if (!ts.isVariableStatement(node)) {
    return false
  }
  return isExported(node)
}

/**
 * The `SupportedFunction` type represents a TypeScript function that can be
 * either a `FunctionDeclaration` or a `FunctionExpression`. This type is used
 * in various utility functions to check, find, and manipulate supported
 * functions in the TypeScript Abstract Syntax Tree (AST). Functions that work
 * with `SupportedFunction` include `isFunction`, `isExportedFunction`,
 * `findFunction`, `getFunctionName`, among others. These functions help in
 * determining if a given node is a supported function, if it is exported,
 * finding a supported function by name, and getting the name of a supported
 * function.
 */
export type SupportedFunction = ts.FunctionDeclaration | ts.FunctionExpression

/**
 * The `isFunction()` function checks if a given TypeScript node is a function
 * declaration or a function expression. It returns true if the node is either a
 * function declaration or a function expression, and false otherwise. The
 * function takes a TypeScript node as input and returns a boolean value
 * indicating whether the input node is of type SupportedFunction, which
 * includes both function declarations and function expressions.
 */
export function isFunction(node: ts.Node): node is SupportedFunction {
  return ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node)
}

/**
 * Checks whether a given TypeScript node is an exported function. It determines
 * this by checking if the node is a function declaration or expression and if
 * it is exported using the `isExported()` function.
 */
export function isExportedFunction(node: ts.Node): node is SupportedFunction {
  return isFunction(node) && isExported(node)
}

/**
 * Checks if a given node is an exported class declaration. Returns true if the
 * node is a class declaration that has been exported, false otherwise.
 */
export function isExportedClass(
  node: ts.Node,
): node is ts.ClassLikeDeclaration {
  return isClass(node) && isExported(node)
}

/**
 * Checks if a given TypeScript node is an exported interface. Returns true if
 * the node is an interface declaration and has been marked for export either
 * via the `export` keyword or by being a member of an exported parent node.
 * Otherwise, returns false.
 */
export function isExportedInterface(
  node: ts.Node,
): node is ts.InterfaceDeclaration {
  return isInterface(node) && isExported(node)
}

/**
 * The "isInterface()" function checks if a given node is an interface
 * declaration.
 * It returns true if the node is an interface declaration, and false otherwise.
 */
export function isInterface(node: ts.Node): node is ts.InterfaceDeclaration {
  return ts.isInterfaceDeclaration(node)
}

/**
 * Checks if a given TypeScript node is a type alias declaration. If the node is
 * a type alias declaration, returns true; otherwise, returns false. A type
 * alias declaration defines a name for a specific type, which can then be used
 * in other parts of the code.
 */
export function isTypeAlias(node: ts.Node): node is ts.TypeAliasDeclaration {
  return ts.isTypeAliasDeclaration(node)
}

/**
 * The `SupportedMethod` type represents a TypeScript method node that can be
 * either a `MethodDeclaration` or a `MethodSignature`. This type is used in
 * various utility functions to check, find, and manipulate methods in the
 * TypeScript Abstract Syntax Tree (AST). Functions like `isMethod`,
 * `isPublicMethodOfExportedClass`, `getClassOfMethod`,
 * `isMethodOfExportedInterface`, and `isMethodOfExportedTypeAlias` use this
 * type to perform their respective operations on method nodes.
 */
export type SupportedMethod = ts.MethodDeclaration | ts.MethodSignature

/**
 * The `isMethod()` function checks if a given TypeScript node is a method. It
 * returns true if the node is either a MethodDeclaration or a MethodSignature,
 * and false otherwise.
 */
export function isMethod(node: ts.Node): node is SupportedMethod {
  return ts.isMethodDeclaration(node) || ts.isMethodSignature(node)
}

/**
 * The `isPrivate()` function checks if a given TypeScript node is private or
 * not. It returns true if the node is a property declaration or method
 * declaration with a name starting with '#' or has a 'private' or 'protected'
 * modifier. Otherwise, it returns false.
 */
export function isPrivate<Node extends ts.Node>(node: Node): boolean {
  if (
    (ts.isPropertyDeclaration(node) || ts.isMethodDeclaration(node)) &&
    node.name.getText().startsWith('#')
  ) {
    return true
  }

  let modifiers: ts.Modifier[] = []

  if ('modifiers' in node) {
    modifiers = [...((node?.modifiers as ts.NodeArray<ts.Modifier>) ?? [])]
  }

  return modifiers.some(
    (mod) =>
      mod.kind === ts.SyntaxKind.PrivateKeyword ||
      mod.kind === ts.SyntaxKind.ProtectedKeyword,
  )
}

/**
 * The `isPublicMethodOfExportedClass()` function checks if a given TypeScript
 * node is a public method of an exported class. It takes a `ts.Node` as an
 * argument and returns `true` if the node is a public method of an exported
 * class, and `false` otherwise. The method is considered public if it does not
 * have a private or protected modifier, and the class is considered exported if
 * it has been marked with the `export` keyword or is a member of an exported
 * parent node.
 */
export function isPublicMethodOfExportedClass(
  node: ts.Node,
): node is SupportedMethod {
  if (!isMethod(node) || isPrivate(node)) {
    return false
  }

  const classNode = getClassOfMethod(node)
  if (!classNode) {
    return false
  }

  return isExported(classNode)
}

/**
 * The "getClassOfMethod()" function takes a SupportedMethod node as an argument
 * and returns the parent ClassLikeDeclaration node if it exists, otherwise it
 * returns null.
 */
export function getClassOfMethod(node: SupportedMethod) {
  return node.parent && isClass(node.parent) ? node.parent : null
}

/**
 * The `isMethodOfExportedInterface()` function checks if a given TypeScript
 * node is a method of an exported interface. It returns true if the node is a
 * method (either a method declaration or a method signature) and belongs to an
 * interface that has been marked for export either via the `export` keyword or
 * by being a member of an exported parent node. Otherwise, it returns false.
 */
export function isMethodOfExportedInterface(
  node: ts.Node,
): node is SupportedMethod {
  if (!isMethod(node)) {
    return false
  }

  const interfaceNode = getInterfaceOfMethod(node)

  return !!interfaceNode && isExported(interfaceNode)
}

/**
 * The `isMethodOfExportedTypeAlias()` function checks if a given TypeScript
 * node is a method of an exported type alias. It takes a TypeScript node as
 * input and returns true if the node is a method (either a method declaration
 * or a method signature) that belongs to an exported type alias, otherwise, it
 * returns false. The function first checks if the node is a method, then it
 * searches for the parent type alias node that must be exported. If found and
 * the parent type alias is exported, it returns true; otherwise, it returns
 * false.
 */
export function isMethodOfExportedTypeAlias(
  node: ts.Node,
): node is SupportedMethod {
  if (!isMethod(node)) {
    return false
  }

  const mustBeExported = findParentThatMustBeExported(node)
  if (
    !mustBeExported ||
    mustBeExported.kind !== ts.SyntaxKind.TypeAliasDeclaration
  ) {
    return false
  }

  return !!mustBeExported && isExported(mustBeExported)
}

/**
 * The `SupportedProperty` type represents a TypeScript node that is either a
 * `PropertyDeclaration` or a `PropertySignature`. This type is used in various
 * utility functions to check for specific properties, determine if they are
 * public or private, and find their owner nodes (such as class or interface
 * declarations). Functions like `isProperty`,
 * `isPublicPropertyOfExportedOwner`, and `findProperty` work with the
 * `SupportedProperty` type to provide information about property nodes in a
 * TypeScript Abstract Syntax Tree (AST).
 */
export type SupportedProperty = ts.PropertyDeclaration | ts.PropertySignature

/**
 * The `isProperty()` function checks if a given TypeScript node is a supported
 * property. It returns true if the node is a PropertyDeclaration or
 * PropertySignature, and false otherwise. Supported properties are defined by
 * the `SupportedProperty` type, which includes `ts.PropertyDeclaration` and
 * `ts.PropertySignature`.
 */
export function isProperty(node: ts.Node): node is SupportedProperty {
  return ts.isPropertyDeclaration(node) || ts.isPropertySignature(node)
}

/**
 * The `isPublicPropertyOfExportedOwner()` function checks whether a given
 * TypeScript node is a public property of an exported class or interface. It
 * returns true if the node is a property, is not private, and belongs to an
 * exported owner (class or interface). Otherwise, it returns false.
 */
export function isPublicPropertyOfExportedOwner(
  node: ts.Node,
): node is SupportedProperty {
  if (!isProperty(node) || isPrivate(node)) {
    return false
  }

  const mustBeExported = findParentThatMustBeExported(node)

  return !!mustBeExported && isExported(mustBeExported)
}

function findParentThatMustBeExported(node: ts.Node): ts.Node | null {
  if (isClass(node) || isInterface(node) || isTypeAlias(node)) {
    return node
  }

  if (!node.parent) {
    return null
  }

  if (
    ts.isTypeLiteralNode(node.parent) &&
    ts.isTypeAliasDeclaration(node.parent.parent)
  ) {
    return node.parent.parent || null
  }

  return findParentThatMustBeExported(node.parent)
}

/**
 * The `isLiteralObject()` function checks if a given TypeScript node is an
 * object literal expression. It returns `true` if the node is an object literal
 * expression, and `false` otherwise.
 */
export function isLiteralObject(
  node: ts.Node,
): node is ts.ObjectLiteralExpression {
  return ts.isObjectLiteralExpression(node)
}

/**
 * The `isExportedType()` function checks if a given TypeScript node is an
 * exported type alias. It returns true if the node is a type alias declaration
 * and is exported using the `isExported()` function. Otherwise, it returns
 * false.
 */
export function isExportedType(node: ts.Node): node is ts.TypeAliasDeclaration {
  return ts.isTypeAliasDeclaration(node) && isExported(node)
}

/**
 * Returns the interface declaration that a given method declaration or method
 * signature belongs to, if any.
 */
export function getInterfaceOfMethod(node: SupportedMethod) {
  return node.parent && isInterface(node.parent) ? node.parent : null
}

/**
 * Function "isClass()" checks if a given node is a class declaration. It
 * returns true if the node is a class declaration, and false otherwise.
 */
export function isClass(node: ts.Node): node is ts.ClassLikeDeclaration {
  return ts.isClassLike(node)
}

/**
 * Finds a node in the TypeScript AST based on an identifier, and returns the
 * node along with its comment target. The function supports finding variables,
 * functions, classes, interfaces, methods, properties, and type aliases.
 */
export function findNode(
  root: ts.Node,
  ident: Identifier | RawIdentifier,
): {
  declaration: ts.Node
  commentTarget: ts.Node
} | null {
  const identifier = typeof ident === 'string' ? parseIdentifier(ident) : ident

  switch (identifier.type) {
    case 'var':
      return findVariable(root, identifier.variableName)
    case 'func':
      return findFunction(root, identifier.functionName)
    case 'class':
      return findClass(root, identifier.className)
    case 'iface':
      return findInterface(root, identifier.interfaceName)
    case 'method':
      return findMethod(root, identifier.ownerName, identifier.methodName)
    case 'prop':
      return findProperty(root, identifier.ownerName, identifier.propertyName)
    case 'type':
      return findType(root, identifier.typeName)
  }
}

/**
 * Finds a variable declaration by name in a TypeScript AST node. If the
 * variable is found and exported, returns an object containing the declaration
 * and the statement that contains the comment. Otherwise, recursively searches
 * through child nodes until a match is found or all nodes have been traversed.
 */
export function findVariable(
  node: ts.Node,
  variableName: string,
): {
  declaration: ts.VariableDeclaration
  commentTarget: ts.VariableStatement
} | null {
  const traverse = () =>
    ts.forEachChild(node, (node) => findVariable(node, variableName)) ?? null

  if (!ts.isVariableStatement(node)) {
    return traverse()
  }

  if (!isExported(node)) {
    return traverse()
  }

  for (const declaration of node.declarationList.declarations) {
    if (getVariableName(declaration) === variableName) {
      return {
        declaration,
        commentTarget: node,
      }
    }
  }

  return traverse()
}

/**
 * Finds a supported function declaration or expression in a TypeScript AST node
 * tree by name and returns an object containing the declaration and the comment
 * target. The function must be exported to be considered. If no matching
 * function is found, it returns null.
 */
export function findFunction(
  node: ts.Node,
  functionName: string,
): {
  declaration: SupportedFunction
  commentTarget: SupportedFunction
} | null {
  const traverse = () =>
    ts.forEachChild(node, (node) => findFunction(node, functionName)) ?? null

  if (!isFunction(node)) {
    return traverse()
  }

  if (!isExported(node)) {
    return traverse()
  }

  const name = getFunctionName(node)
  if (name !== functionName) {
    return traverse()
  }

  return {
    declaration: node,
    commentTarget: node,
  }
}

/**
 * Returns the name of a supported function, which can be either a
 * FunctionDeclaration or a FunctionExpression.
 */
export function getFunctionName(func: SupportedFunction) {
  if (ts.isFunctionDeclaration(func)) {
    return func.name?.getText()
  }

  if (ts.isFunctionExpression(func)) {
    return func.name?.getText()
  }
}

/**
 * Finds and returns the declaration and comment target of a class with the
 * given name within the provided TypeScript node.
 */
export function findClass(
  node: ts.Node,
  className: string,
): {
  declaration: ts.ClassLikeDeclaration
  commentTarget: ts.ClassLikeDeclaration
} | null {
  if (isClass(node) && isExported(node) && getClassName(node) === className) {
    return {
      declaration: node,
      commentTarget: node,
    }
  }

  return ts.forEachChild(node, (node) => findClass(node, className)) ?? null
}

/**
 * The `iface:CommentableVariable` is an interface that represents a variable
 * with a statement and a declaration. It is used in the `parseVariable()`
 * function to return an object containing the statement and declaration of a
 * given TypeScript node. This interface helps in identifying variables with
 * comments when working with TypeScript Abstract Syntax Trees (ASTs) and is
 * useful when searching for specific variables, functions, classes, interfaces,
 * methods, properties, or type aliases within the AST.
 */
export interface CommentableVariable {
  /**
   * The `CommentableVariable.statement` property represents the statement
   * containing the comment for a variable declaration in a TypeScript Abstract
   * Syntax Tree (AST). It is part of the `CommentableVariable` object, which
   * also includes the `declaration` property for the variable declaration
   * itself. This property is used to associate comments with their
   * corresponding variable declarations in the AST.
   */
  statement: ts.VariableStatement
  /**
   * The "CommentableVariable.declaration" property represents the variable
   * declaration of a commentable variable in a TypeScript node. It is part of
   * an object returned by functions such as `findVariable()`, which searches
   * for a variable declaration by name in a TypeScript Abstract Syntax Tree
   * (AST) node. The object also includes a "commentTarget" property, which
   * refers to the statement that contains the comment for the variable
   * declaration. The "declaration" property is useful for accessing information
   * about the variable, such as its name or type, within the context of the
   * TypeScript AST.
   */
  declaration: ts.VariableDeclaration
}

/**
 * The `parseVariable()` function takes a TypeScript node of type
 * `ts.VariableDeclaration` or `ts.VariableStatement` as its argument and
 * returns an object containing the `statement` and `declaration` properties.
 * The `statement` property is of type `ts.VariableStatement`, while the
 * `declaration` property is of type `ts.VariableDeclaration`. This function is
 * useful for extracting the variable statement and declaration from a given
 * node, making it easier to work with variables in the TypeScript Abstract
 * Syntax Tree (AST).
 */
export function parseVariable(
  node: ts.VariableDeclaration | ts.VariableStatement,
) {
  if (ts.isVariableDeclaration(node)) {
    return {
      statement: node.parent.parent,
      declaration: node,
    }
  }

  if (!node.declarationList.declarations?.[0])
    throw new Error('Variable statement has no declarations')

  return {
    statement: node,
    declaration: node.declarationList.declarations?.[0] ?? undefined,
  }
}

/** Returns the name of a variable declaration or statement. */
export function getVariableName(
  node: ts.VariableDeclaration | ts.VariableStatement,
) {
  return parseVariable(node).declaration.name.getText()
}

/**
 * Returns the name of the owner (class, interface, or type alias) of a given
 * supported property or method node.
 */
export function getOwnerName(node: SupportedProperty | SupportedMethod) {
  if (!node.parent) {
    return
  }

  if (isClass(node.parent)) {
    return getClassName(node.parent)
  }

  if (isInterface(node.parent)) {
    return node.parent.name.getText()
  }

  if (ts.isTypeLiteralNode(node.parent) && isTypeAlias(node.parent.parent)) {
    return node.parent.parent.name.getText()
  }
}

/**
 * Returns the name of a class declaration node as a string. If the class has a
 * name, it returns the name. Otherwise, it looks for the parent node and
 * returns its name if it is a variable declaration or property assignment.
 */
export function getClassName(node: ts.ClassLikeDeclaration) {
  const name = node.name?.getText()
  if (name) {
    return name
  }

  if (!node.parent) {
    return
  }

  if (ts.isVariableDeclaration(node.parent)) {
    return node.parent.name.getText()
  }

  if (ts.isPropertyAssignment(node.parent)) {
    return node.parent.name.getText()
  }
}

// export function getInterfaceName(node: ts.InterfaceDeclaration): string {
//   return node.name.getText()
// }

// export function getTypeAliasName(node: ts.TypeAliasDeclaration): string {
//   return node.name.getText()
// }

export function findInterface(
  node: ts.Node,
  interfaceName: string,
): {
  declaration: ts.InterfaceDeclaration
  commentTarget: ts.InterfaceDeclaration
} | null {
  if (
    ts.isInterfaceDeclaration(node) &&
    node.name.getText() === interfaceName
  ) {
    return {
      declaration: node,
      commentTarget: node,
    }
  }

  return (
    ts.forEachChild(node, (node) => findInterface(node, interfaceName)) ?? null
  )
}

/**
 * Finds a method with the given owner name and method name in the provided
 * TypeScript node. Returns an object containing the declaration and comment
 * target of the method if found, or null otherwise.
 */
export function findMethod(
  node: ts.Node,
  ownerName: string,
  methodName: string,
): {
  declaration: SupportedMethod
  commentTarget: SupportedMethod
} | null {
  const owner = findClass(node, ownerName) ?? findInterface(node, ownerName)
  if (!owner) {
    return null
  }

  const { declaration } = owner

  const members = declaration.members as ts.NodeArray<ts.Node>

  const method = members.find(
    (member) => isMethod(member) && member.name.getText() === methodName,
  )

  if (!method || !isMethod(method)) {
    return null
  }

  if (isPrivate(method)) {
    return null
  }

  return {
    declaration: method,
    commentTarget: method,
  }
}

/**
 * findProperty() is a function that finds and returns a property declaration
 * node within a given owner node based on the name of the property. It searches
 * for the owner node (a class or an interface) with the given name, then looks
 * for a property with the given name within its members. If found, it returns
 * an object containing the property declaration node and its comment target
 * node. If not found or the property is private, it returns null.
 */
export function findProperty(
  node: ts.Node,
  ownerName: string,
  propertyName: string,
) {
  const owner = findClass(node, ownerName) ?? findInterface(node, ownerName)
  if (!owner) {
    return null
  }

  const { declaration } = owner

  const members = declaration.members as ts.NodeArray<ts.Node>

  const prop = members.find(
    (member) => isProperty(member) && member.name.getText() === propertyName,
  )

  if (!prop) {
    return null
  }

  if (isPrivate(prop)) {
    return null
  }

  return {
    declaration: prop,
    commentTarget: prop,
  }
}

/**
 * Finds a type alias declaration node with the given name in the TypeScript
 * AST. Returns an object containing the declaration node and its corresponding
 * comment target node, or null if no matching node is found.
 */
export function findType(
  node: ts.Node,
  typeName: string,
): {
  declaration: ts.TypeAliasDeclaration
  commentTarget: ts.TypeAliasDeclaration
} | null {
  if (ts.isTypeAliasDeclaration(node) && node.name.getText() === typeName) {
    return {
      declaration: node,
      commentTarget: node,
    }
  }

  return ts.forEachChild(node, (node) => findType(node, typeName)) ?? null
}
