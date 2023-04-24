import ts from 'typescript'
import type { Identifier, RawIdentifier } from './identifier'
import { parseIdentifier } from './identifier'

export function hasComments(node: ts.Node) {
  return !(
    (ts.getLeadingCommentRanges(node.getFullText(), 0) ?? []).length === 0
  )
}

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

export function isWithinFunction(node: ts.Node): boolean {
  if (!node.parent) {
    return false
  }
  if (isFunction(node.parent)) {
    return true
  }

  return isWithinFunction(node.parent)
}

export function isVariable(node: ts.Node): node is ts.VariableDeclaration {
  return ts.isVariableDeclaration(node)
}

export function isExportedVariable(
  node: ts.Node,
): node is ts.VariableDeclaration {
  if (!isVariable(node)) {
    return false
  }
  return isExported(parseVariable(node).statement)
}

export type SupportedFunction = ts.FunctionDeclaration | ts.FunctionExpression

export function isFunction(node: ts.Node): node is SupportedFunction {
  return ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node)
}

export function isExportedFunction(node: ts.Node): node is SupportedFunction {
  return isFunction(node) && isExported(node)
}

export function isExportedClass(
  node: ts.Node,
): node is ts.ClassLikeDeclaration {
  return isClass(node) && isExported(node)
}

export function isExportedInterface(
  node: ts.Node,
): node is ts.InterfaceDeclaration {
  return isInterface(node) && isExported(node)
}

export function isInterface(node: ts.Node): node is ts.InterfaceDeclaration {
  return ts.isInterfaceDeclaration(node)
}

export type SupportedMethod = ts.MethodDeclaration | ts.MethodSignature

export function isMethod(node: ts.Node): node is SupportedMethod {
  return ts.isMethodDeclaration(node) || ts.isMethodSignature(node)
}

export function isPrivate<Node extends ts.Node>(node: Node): boolean {
  if (ts.isPropertyDeclaration(node) && node.name.getText().startsWith('#')) {
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

export function getClassOfMethod(node: SupportedMethod) {
  return node.parent && isClass(node.parent) ? node.parent : null
}

export function isMethodOfExportedInterface(
  node: ts.Node,
): node is SupportedMethod {
  if (!isMethod(node)) {
    return false
  }

  const interfaceNode = getInterfaceOfMethod(node)

  return !!interfaceNode && isExported(interfaceNode)
}

export type SupportedProperty = ts.PropertyDeclaration | ts.PropertySignature

export function isProperty(node: ts.Node): node is SupportedProperty {
  return ts.isPropertyDeclaration(node) || ts.isPropertySignature(node)
}

export function isPublicPropertyOfExportedOwner(
  node: ts.Node,
): node is SupportedProperty {
  return (
    isProperty(node) &&
    !isPrivate(node) &&
    (isClass(node.parent) || isInterface(node.parent)) &&
    isExported(node.parent)
  )
}

export function getInterfaceOfMethod(node: SupportedMethod) {
  return node.parent && isInterface(node.parent) ? node.parent : null
}

export function isClass(node: ts.Node): node is ts.ClassLikeDeclaration {
  return ts.isClassLike(node)
}

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
  }
}

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

export function getFunctionName(func: SupportedFunction) {
  if (ts.isFunctionDeclaration(func)) {
    return func.name?.getText()
  }

  if (ts.isFunctionExpression(func)) {
    return func.name?.getText()
  }
}

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

export interface CommentableVariable {
  statement: ts.VariableStatement
  declaration: ts.VariableDeclaration
}

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

export function getVariableName(
  node: ts.VariableDeclaration | ts.VariableStatement,
) {
  return parseVariable(node).declaration.name.getText()
}

export function getOwnerName(node: SupportedProperty | SupportedMethod) {
  if (!node.parent) {
    return
  }

  if (isClass(node.parent)) {
    return getClassName(node.parent)
  }

  if (isInterface(node.parent)) {
    return getInterfaceName(node.parent)
  }
}

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

export function getInterfaceName(node: ts.InterfaceDeclaration): string {
  return node.name.getText()
}

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

export function getClassNameOfMethod(method: SupportedMethod) {
  if (ts.isClassLike(method.parent)) {
    return getClassName(method.parent)
  }

  if (ts.isObjectLiteralExpression(method.parent)) {
    // anonymous classes not supported
  }

  return undefined
}

export function getMethodName(method: SupportedMethod) {
  return method.name.getText()
}

export function getInterfaceNameOfMethod(method: SupportedMethod) {
  if (ts.isInterfaceDeclaration(method.parent))
    return getInterfaceName(method.parent)

  return undefined
}

export function getPropertyName(property: SupportedProperty) {
  return property.name.getText()
}

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
