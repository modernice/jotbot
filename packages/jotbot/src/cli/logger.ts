import { createConsola } from 'consola'

export function createLogger(
  out: NodeJS.WriteStream = process.stderr,
  options?: { enabled?: boolean },
) {
  const enabled = options?.enabled ?? true

  return createConsola({
    stdout: out,
    stderr: out,
    level: enabled ? 3 : 0,
  })
}