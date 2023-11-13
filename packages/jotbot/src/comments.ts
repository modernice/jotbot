import ts from 'typescript'
import type { RawIdentifier } from './identifier'
import { parseIdentifier } from './identifier'
import { findNode } from './nodes'
import { printComment } from './print'

export type Comments = Record<RawIdentifier, string>

/**
 * Retrieves a collection of parsed identifiers from the provided comments
 * mapping. Each identifier in the returned array represents a parsed version of
 * the raw identifiers used as keys in the comments object. The parsing process
 * is handled by the {@link parseIdentifier} function, ensuring that each
 * identifier adheres to a standardized format suitable for further processing
 * or analysis. This function is typically used to extract and manipulate
 * identifiers from source code comments for tasks such as documentation
 * generation or code analysis.
 */
export function getIdentifiers(comments: Comments) {
  return getRawIdentifiers(comments).map(parseIdentifier)
}

/**
 * Retrieves an array of keys from the provided {@link Comments} object, each
 * key representing a raw identifier within the comments. This array is
 * specifically typed to match the keys of the {@link Comments} type.
 */
export function getRawIdentifiers(comments: Comments) {
  return Object.keys(comments) as Array<keyof Comments>
}

/**
 * Updates the comments for identifiers within a TypeScript source file based on
 * the provided comments map. It iterates over each identifier from the map and
 * attempts to find and update the corresponding comment in the source file's
 * syntax tree. If an identifier cannot be found, it throws an error indicating
 * the missing identifier and file name. The function uses {@link formatComment}
 * to ensure comments are correctly formatted before being attached to their
 * respective nodes. Returns nothing but can throw an error for unresolved
 * identifiers.
 */
export function patchComments(
  file: ts.SourceFile,
  comments: Record<string, string>,
) {
  const fileName = file.fileName
  const identifiers = getRawIdentifiers(comments)

  for (const identifier of identifiers) {
    const comment = formatComment(comments[identifier])

    if (!updateComment(file, identifier, comment)) {
      throw new Error(`Could not find identifier ${identifier} in ${fileName}`)
    }
  }
}

function updateComment(
  tree: ts.Node,
  identifier: RawIdentifier,
  comment: string,
) {
  const node = findNode(tree, parseIdentifier(identifier))
  if (!node) {
    return false
  }

  updateNodeComments(node.commentTarget, [comment])
  return true
}

/**
 * Transforms a given string into a formatted comment block, adhering to
 * specified maximum line length and optional enclosure within comment syntax.
 * If no maximum line length is provided, it defaults to 80 characters. The
 * resulting string is structured with asterisk-prefixed lines and, if the
 * `enclose` option is set to true, the entire block is wrapped within comment
 * delimiters. This function assumes that the input string uses newlines to
 * separate paragraphs and maintains paragraph separation in the output. Returns
 * the formatted comment as a {@link string}.
 */
export function formatComment(
  comment: string,
  options?: {
    maxLen?: number

    enclose?: boolean
  },
): string {
  const lines = splitIntoLines(comment, (options?.maxLen ?? 80) - 3)
  let formatted = `*\n${lines.map((line) => ` * ${line}`).join('\n')}\n `

  if (options?.enclose) {
    formatted = `/*${formatted}*/`
  }

  return formatted
}

/**
 * Retrieves the comments associated with a given TypeScript {@link ts.Node},
 * combining both synthetic leading comments and the text representation of the
 * node's comments. It returns an object containing two properties: `comments`,
 * an array of synthetic comment objects, and `text`, a string that represents
 * the formatted comment text for the node.
 */
export function getNodeComments(node: ts.Node) {
  const comments = ts.getSyntheticLeadingComments(node) ?? []
  const text = printComment(node)

  return {
    comments,
    text,
  }
}

/**
 * Determines the line and character position within the source file where a new
 * comment should be inserted for the specified {@link ts.Node}. Returns an
 * object containing line and character information.
 */
export function getInsertPosition(node: ts.Node) {
  return ts.getLineAndCharacterOfPosition(node.getSourceFile(), node.getStart())
}

/**
 * Attaches an array of formatted comment strings to a specified {@link ts.Node}
 * as synthetic leading comments. Each comment is treated as a multi-line
 * comment trivia and appended to any existing leading comments of the node. If
 * the node already has synthetic leading comments, the new comments are added
 * after the existing ones. This function ensures that each comment has a
 * trailing newline for proper formatting in the syntax tree. If multiple
 * comments are provided, they are processed in order and each is appended in
 * sequence.
 */
export function updateNodeComments(node: ts.Node, comments: string[]) {
  ts.setSyntheticLeadingComments(node, [
    ...(ts.getSyntheticLeadingComments(node) ?? []),
    ...comments.map(
      (comment, i) =>
        ({
          kind: ts.SyntaxKind.MultiLineCommentTrivia,
          text: comment.trim() + (i === comments.length - 1 ? '\n ' : ''),
          hasTrailingNewLine: true,
          pos: -1,
          end: -1,
        } satisfies ts.SynthesizedComment),
    ),
  ])
}

function splitIntoLines(str: string, maxLen: number): string[] {
  const out: string[] = []

  const paras: string[] = str.split('\n\n')
  for (let i = 0; i < paras.length; i++) {
    const para: string = paras[i]
    const lines: string[] = splitParagraph(para, maxLen)
    out.push(...lines)
    if (i < paras.length - 1) {
      out.push('')
    }
  }

  return out
}

function splitParagraph(str: string, maxLen: number): string[] {
  const words: string[] = str.split(' ')

  const lines: string[] = []
  let line = ''
  for (const word of words) {
    if (line.length + word.length > maxLen) {
      lines.push(line)
      line = ''
    }
    line += `${word} `
  }
  lines.push(line.trim())

  return lines
}
