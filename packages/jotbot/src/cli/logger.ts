import type { ConsolaInstance } from 'consola'
import { createConsola } from 'consola'

/**
 * Creates a new logger instance using {@link createConsola}. The logger writes
 * to the provided output stream, defaults to `process.stderr`, and can be
 * enabled or disabled using the `enabled` option. Returns a {@link
 * ConsolaInstance} object.
 */
export function createLogger(
  out: NodeJS.WriteStream = process.stderr,
  options?: { enabled?: boolean },
): ConsolaInstance {
  const enabled = options?.enabled ?? true

  return createConsola({
    stdout: out,
    stderr: out,
    level: enabled ? 3 : 0,
  })
}
