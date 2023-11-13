/**
 * Outputs the provided message to the standard output stream without adding a
 * newline character at the end. The message is expected to be a string. This
 * function does not return any value, indicating that its primary purpose is a
 * side effect (i.e., outputting text).
 */
export function print(msg: string): void {
  process.stdout.write(msg)
}
