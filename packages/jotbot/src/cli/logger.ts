import type { ConsolaInstance } from 'consola'
import { createConsola } from 'consola'

/**
 * Creates a new {@link ConsolaInstance} for logging purposes, using the
 * provided output stream and options. It defaults to using `process.stderr` for
 * output if no stream is specified. The verbosity of the logger can be
 * controlled through the `enabled` option in the settings; when enabled, it
 * sets a default logging level, otherwise it silences the logger.
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
