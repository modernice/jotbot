import ts from 'typescript'

const printer = ts.createPrinter({
  newLine: ts.NewLineKind.LineFeed,
  removeComments: false,
  omitTrailingSemicolon: true,
})

/**
 * The `printNode()` function takes a TypeScript {@link ts.Node} as input and
 * returns a string representation of the node, preserving its original
 * formatting. If the node does not have an associated source file, an error is
 * thrown.
 */
export function printNode(node: ts.Node) {
  const file = node.getSourceFile()
  if (!file) {
    throw new Error('Node has no source file')
  }
  return printer.printNode(ts.EmitHint.Unspecified, node, file)
}

/**
 * The `printComment()` function retrieves the leading comments associated with
 * a given TypeScript AST node and returns them as a formatted string. If the
 * node has synthetic leading comments, the function will print those comments;
 * otherwise, it will print the comments from the node's source file.
 */
export function printComment(node: ts.Node) {
  const synthetic = ts.getSyntheticLeadingComments(node)
  if (synthetic != null) {
    return printSyntheticComments(synthetic)
  }
  return printSourceFileComments(node)
}

/**
 * The `printSyntheticComments()` function takes a TypeScript node or an array
 * of synthesized comments as its input and returns a formatted string
 * containing the comments. If the input is a node, it retrieves the synthetic
 * leading comments associated with the node. The function then concatenates the
 * text of each comment with line breaks and wraps the resulting string in
 * comment delimiters (i.e., `/*` and `*\/`).
 */
export function printSyntheticComments(
  nodeOrComments: ts.Node | ts.SynthesizedComment[],
) {
  const comments = Array.isArray(nodeOrComments)
    ? nodeOrComments
    : ts.getSyntheticLeadingComments(nodeOrComments) ?? []
  const body = comments.map((comment) => comment.text).join('\n')

  return `/*${body}*/`
}

/**
 * The `printSourceFileComments()` function is used to extract and return the
 * leading comments associated with a given TypeScript node. It retrieves the
 * source file containing the node and its leading comment ranges, then returns
 * a concatenated string of all the leading comments. If the node does not have
 * any source file or leading comments, an error will be thrown or an empty
 * string will be returned, respectively.
 */
export function printSourceFileComments(node: ts.Node) {
  const file = node.getSourceFile()
  if (!file) {
    throw new Error('Node has no source file')
  }
  const comments = ts.getLeadingCommentRanges(file.text, node.pos)
  if (!comments) {
    return ''
  }
  return comments
    .map((comment) => file.text.slice(comment.pos, comment.end))
    .join('\n')
}
