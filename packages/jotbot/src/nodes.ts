import ts from 'typescript'
import type { Identifier, RawIdentifier } from './identifier'
import { parseIdentifier } from './identifier'

/**
 * Determines whether the provided {@link ts.Node} has any associated leading
 * comments. Returns `true` if comments are present; otherwise, returns `false`.
 */
export function hasComments(node: ts.Node) {
  return !(
    (ts.getLeadingCommentRanges(node.getFullText(), 0) ?? []).length === 0
  )
}

/**
 * Determines whether a given TypeScript {@link ts.Node} is part of an exported
 * declaration. This includes any node that is directly exported or contained
 * within an exported node, but not nodes explicitly marked with the 'export'
 * keyword themselves. This function recursively checks the export status by
 * examining the node's modifiers and traversing up its parent nodes if
 * necessary, until it reaches the top-level {@link ts.SourceFile} or determines
 * the export status. It returns `true` if the node is part of an exported
 * declaration, otherwise `false`.
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
 * Determines whether the given node is contained within a function. It
 * recursively traverses the parent nodes until it finds a function or reaches
 * the root of the syntax tree. Returns `true` if any parent node is a function
 * as determined by {@link isFunction}, otherwise returns `false`.
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
 * Determines if a given {@link ts.Node} represents a variable statement that
 * has been exported. The check is performed by first asserting whether the node
 * is a variable statement, followed by verifying its export status. Returns
 * `true` if the node is a variable statement and is marked as exported;
 * otherwise, returns `false`. This function uses type guards to refine the type
 * of the node if it satisfies the conditions for being an exported variable.
 */
export function isExportedVariable(
  node: ts.Node,
): node is ts.VariableStatement {
  if (!ts.isVariableStatement(node)) {
    return false
  }
  return isExported(node)
}

export type SupportedFunction = ts.FunctionDeclaration | ts.FunctionExpression

/**
 * Determines whether a given {@link ts.Node} is a function, specifically if
 * it's either a {@link ts.FunctionDeclaration} or a {@link
 * ts.FunctionExpression}. Returns true if the node is a function, and false
 * otherwise.
 */
export function isFunction(node: ts.Node): node is SupportedFunction {
  return ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node)
}

/**
 * Determines if a given {@link ts.Node} represents a function that is exported.
 * The node must satisfy the conditions of being a function as defined by {@link
 * isFunction} and being exported as defined by {@link isExported} to be
 * considered an exported function. This function acts as a type guard,
 * narrowing the type of the input node to {@link SupportedFunction} when it
 * returns true.
 */
export function isExportedFunction(node: ts.Node): node is SupportedFunction {
  return isFunction(node) && isExported(node)
}

/**
 * Determines if the provided TypeScript node represents a class that has been
 * exported. This function checks if a given node is a class-like declaration,
 * and if so, verifies whether it is marked with export modifiers directly or
 * indirectly through module exports. It returns true if the class is exported
 * and false otherwise. The result is type-guarded, ensuring that the node can
 * be safely treated as a {@link ts.ClassLikeDeclaration} when the function
 * returns true.
 */
export function isExportedClass(
  node: ts.Node,
): node is ts.ClassLikeDeclaration {
  return isClass(node) && isExported(node)
}

/**
 * Determines whether a given TypeScript node is an interface declaration that
 * has been exported. The function checks if the node represents an interface
 * and, if so, whether it is marked with the export modifier either directly or
 * through module or namespace export patterns. Returns `true` if the node is an
 * exported interface, otherwise `false`. The return type includes a type
 * predicate that refines the input type to {@link ts.InterfaceDeclaration} when
 * the function returns `true`.
 */
export function isExportedInterface(
  node: ts.Node,
): node is ts.InterfaceDeclaration {
  return isInterface(node) && isExported(node)
}

/**
 * Determines whether a given TypeScript node is an {@link
 * ts.InterfaceDeclaration}. Returns `true` if the node is an interface
 * declaration, otherwise `false`.
 */
export function isInterface(node: ts.Node): node is ts.InterfaceDeclaration {
  return ts.isInterfaceDeclaration(node)
}

/**
 * Determines whether a given {@link ts.Node} is a {@link
 * ts.TypeAliasDeclaration}. Returns `true` if the node is a type alias
 * declaration, otherwise `false`.
 */
export function isTypeAlias(node: ts.Node): node is ts.TypeAliasDeclaration {
  return ts.isTypeAliasDeclaration(node)
}

export type SupportedMethod = ts.MethodDeclaration | ts.MethodSignature

/**
 * Determines whether a given TypeScript AST node represents either a method
 * declaration or a method signature. If the node is a method, it narrows the
 * type to {@link SupportedMethod}, which includes both {@link
 * ts.MethodDeclaration} and {@link ts.MethodSignature}. This type predicate can
 * be used to assert the specific kind of method node being dealt with in
 * TypeScript compiler API transformations or analyses.
 */
export function isMethod(node: ts.Node): node is SupportedMethod {
  return ts.isMethodDeclaration(node) || ts.isMethodSignature(node)
}

/**
 * Determines whether a given TypeScript node is private within its context.
 * This includes nodes that are explicitly marked with the `private` or
 * `protected` keyword, as well as properties and methods denoted with the
 * ECMAScript private field syntax (prefixed with `#`). It returns `true` if the
 * node is private or protected, otherwise returns `false`. The node is checked
 * for relevant modifiers, and in the case of properties and methods, their
 * naming conventions to infer privacy.
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
 * Determines whether a given {@link ts.Node} is a {@link SupportedMethod} that
 * is both public and a member of an exported class. It verifies that the node
 * represents a method, is not private or protected, and belongs to a class that
 * is accessible from outside the module. If these conditions are met, it
 * returns `true`, otherwise `false`.
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
 * Returns the class declaration that a method is a part of, if it exists. If
 * the method is not contained within a class, or if the node provided is not a
 * method, this function returns `null`. This function assumes that the input
 * node is already confirmed to be a {@link SupportedMethod}, which includes
 * both method declarations and method signatures.
 */
export function getClassOfMethod(node: SupportedMethod) {
  return node.parent && isClass(node.parent) ? node.parent : null
}

/**
 * Determines whether a given node represents a method that belongs to an
 * exported interface. This function asserts that the node is of type {@link
 * SupportedMethod}, which includes method declarations and method signatures.
 * It ensures the method is part of an interface structure and that the
 * interface is marked for export, making it part of the module's public API
 * surface.
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
 * Determines if the given node represents a method that is part of a type alias
 * declaration which is exported. If the node is identified as a method and it
 * exists within an exported type alias, the function will return `true`,
 * otherwise `false`. The returned value also asserts the node as a {@link
 * SupportedMethod}, narrowing down its type. This utility is useful for
 * ensuring that methods from exported type aliases are correctly processed or
 * manipulated in accordance with their visibility and usage within a TypeScript
 * module.
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

export type SupportedProperty = ts.PropertyDeclaration | ts.PropertySignature

/**
 * Determines if the given node is a property, either as a declaration or a
 * signature. This is applicable to nodes representing properties in classes,
 * interfaces, and type literals. Returns `true` if the node is a {@link
 * SupportedProperty}, which includes {@link ts.PropertyDeclaration} or {@link
 * ts.PropertySignature}, and `false` otherwise.
 */
export function isProperty(node: ts.Node): node is SupportedProperty {
  return ts.isPropertyDeclaration(node) || ts.isPropertySignature(node)
}

/**
 * Determines if a given node is a {@link SupportedProperty} that is public and
 * belongs to an owner such as a {@link ts.ClassLikeDeclaration}, {@link
 * ts.InterfaceDeclaration}, or {@link ts.TypeAliasDeclaration} which is
 * exported. This function returns `true` if the property is not private and its
 * owner is part of the module's public API, otherwise it returns `false`.
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
 * Determines whether a given TypeScript AST node is an {@link
 * ts.ObjectLiteralExpression}. It returns true if the node is an object literal
 * expression, otherwise false. This type predicate can be used to assert the
 * specific type of a node within the TypeScript compiler API.
 */
export function isLiteralObject(
  node: ts.Node,
): node is ts.ObjectLiteralExpression {
  return ts.isObjectLiteralExpression(node)
}

/**
 * Determines whether a given {@link ts.Node} represents a type alias that is
 * exported. If the node is a type alias and is marked with an export modifier,
 * it returns true, indicating that the type alias is part of the module's
 * public API. Otherwise, it returns false. This function can be useful when
 * generating documentation or analyzing the structure of TypeScript code to
 * understand which types are part of the public interface of a module.
 */
export function isExportedType(node: ts.Node): node is ts.TypeAliasDeclaration {
  return ts.isTypeAliasDeclaration(node) && isExported(node)
}

/**
 * Retrieves the parent {@link ts.InterfaceDeclaration} of a given method if the
 * method is part of an interface. If the method is not contained within an
 * interface, returns `null`. This function is applicable to methods represented
 * by either {@link ts.MethodDeclaration} or {@link ts.MethodSignature}.
 */
export function getInterfaceOfMethod(node: SupportedMethod) {
  return node.parent && isInterface(node.parent) ? node.parent : null
}

/**
 * Determines whether the provided node is a class-like structure. This includes
 * classes defined with the `class` keyword or through variable assignments
 * where the variable's value is a class expression. If true, the function
 * narrows the type to {@link ts.ClassLikeDeclaration}.
 */
export function isClass(node: ts.Node): node is ts.ClassLikeDeclaration {
  return ts.isClassLike(node)
}

/**
 * Locates the declaration and comment target for a specified identifier within
 * the provided abstract syntax tree node. It accepts an identifier which can be
 * a variable, function, class, interface, method, property, or type alias name
 * represented either as a structured {@link Identifier} object or as a raw
 * string {@link RawIdentifier}. If the matching node is found, it returns an
 * object containing the `declaration` and `commentTarget` nodes. Otherwise, it
 * returns `null`. The search includes various checks to ensure the correct type
 * of node is found according to the identifier's type specification.
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
 * Locates a variable declaration within the AST, matching the specified
 * variable name, and returns an object containing the declaration node and its
 * parent statement if it is exported. If no matching exported variable is
 * found, it recursively searches child nodes or returns null. This function is
 * designed to be used in scenarios such as documentation generation or analysis
 * tools where it's necessary to find and possibly comment on exported variables
 * given their name. The return type is an object with properties `declaration`
 * of type {@link ts.VariableDeclaration} and `commentTarget` of type {@link
 * ts.VariableStatement}, or `null` if the variable cannot be found.
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
 * Finds a function declaration in the given TypeScript AST node that matches
 * the specified function name. If the function is found and is exported, it
 * returns an object containing both the declaration and a reference to the node
 * that should be targeted for comments. If no matching exported function is
 * found, it returns `null`. This process includes recursively searching child
 * nodes if necessary. The function must be one of the supported function types,
 * namely a {@link ts.FunctionDeclaration} or a {@link ts.FunctionExpression}.
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
 * Retrieves the name of a given function if available. This function supports
 * both {@link ts.FunctionDeclaration} and {@link ts.FunctionExpression}.
 * Returns the name as a string, or undefined if the function does not have an
 * explicit name.
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
 * Locates a class declaration with the specified name within the given
 * TypeScript syntax tree node. If a matching exported class is found, it
 * returns an object containing both the class declaration and itself as the
 * target for comments. If no matching class is found, it returns null. This
 * function is useful for tools that need to associate metadata or documentation
 * with specific TypeScript nodes. It ensures that only exported classes are
 * considered, which typically correspond to the API surface of a module or
 * package.
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
 * Represents a variable within TypeScript code that is capable of being
 * annotated with comments. The `CommentableVariable` type provides access to
 * both the `statement` part of the variable, which is the entire variable
 * statement, and the `declaration` part, which specifically refers to the
 * variable declaration itself. This dual reference allows for precise targeting
 * when adding or retrieving comments related to a particular variable within
 * the source code.
 */
export interface CommentableVariable {
  statement: ts.VariableStatement
  declaration: ts.VariableDeclaration
}

/**
 * Parses the given variable declaration or statement and returns an object
 * containing both the variable declaration and its associated statement. This
 * function ensures that the variable structure is properly accessed regardless
 * of whether a {@link ts.VariableDeclaration} or {@link ts.VariableStatement}
 * is provided. If a variable statement with no declarations is passed, it
 * throws an error indicating that the expected variable declarations are
 * missing.
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

/**
 * Retrieves the name of a variable from a given TypeScript VariableDeclaration
 * or VariableStatement. If the input is a VariableDeclaration, it extracts the
 * variable's identifier name directly. If the input is a VariableStatement, it
 * identifies the first VariableDeclaration within it and then extracts the name
 * from that declaration. This function is useful for obtaining the string
 * representation of a variable's name for further processing or analysis in
 * code transformation and introspection tasks.
 */
export function getVariableName(
  node: ts.VariableDeclaration | ts.VariableStatement,
) {
  return parseVariable(node).declaration.name.getText()
}

/**
 * Retrieves the name of the owner class, interface, or type alias for a given
 * method or property. If the method or property is not part of a class,
 * interface, or type alias, or if the parent node cannot be determined, the
 * function will return undefined. This function is useful when trying to
 * ascertain the context in which a method or property is defined within the
 * broader structure of TypeScript code. It specifically handles cases where
 * methods and properties are nested within classes, interfaces, and type
 * aliases that are exported from their respective modules.
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
 * Retrieves the name of a given class-like declaration. If the class is
 * declared with a name, it returns the textual representation of that name. For
 * anonymous classes, it attempts to infer the name based on its variable or
 * property assignment. If no clear name can be determined, the result is
 * undefined. This function is useful when needing to reference or document the
 * class by its name in situations where a class may not have been explicitly
 * named or when it's being exported as part of a larger expression.
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

/**
 * Finds the specified interface within the provided TypeScript AST node. If the
 * interface with the given name exists, returns an object containing both the
 * interface declaration and the node to which comments can be applied. If no
 * matching interface is found, returns null. This function is useful for tools
 * that need to locate and possibly annotate or modify TypeScript interfaces
 * based on their names within a codebase.
 */
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
 * Locates a method within a given TypeScript AST node that matches the
 * specified owner and method names. If found, it returns an object containing
 * the method declaration and the appropriate comment target for documentation
 * purposes, provided the method is public and belongs to either an exported
 * class or interface. Otherwise, returns `null`. The method must not be private
 * or protected to be considered a valid match. This function can be useful when
 * generating documentation or performing code analysis tasks.
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
 * Locates a property with the specified name within the given owner's scope and
 * returns its declaration and the node where comments should be attached. The
 * search is conducted recursively through the AST starting from the provided
 * root node. If the property is found and it is not private, its declaration
 * and comment target are returned; otherwise, null is returned. The owner can
 * be either a class or an interface as determined by the owner name parameter.
 * This function ensures that only publicly accessible properties of exported
 * classes or interfaces are considered in the search.
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
 * Locates a type alias by its name within the given TypeScript AST node and
 * returns its declaration along with the node that should be targeted for
 * comments. If no matching type alias is found, `null` is returned. The
 * function ensures that the identified type alias has the specified name and
 * that it is part of the AST structure rooted at the provided node. The result
 * includes both the declaration of the type alias itself and the node which
 * should receive associated documentation comments, typically this would be the
 * same as the declaration node. This function is useful for tools that need to
 * associate comments with specific type aliases in a TypeScript codebase.
 * Returns an object containing the `declaration` and `commentTarget` if a match
 * is found, otherwise `null`.
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
