/**
 * Outputs the given message as a string to the standard output stream.
 * @param msg - The message to be printed.
 */
export function print(msg: string): void {
  process.stdout.write(msg)
}
