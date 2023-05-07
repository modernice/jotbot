import type { ConsolaInstance } from 'consola'
import { createLogger } from './logger'

/**
 * The variable {@link out} is a ConsolaInstance object that is created using
 * the createLogger function and process.stdout. It represents an instance of a
 * logger that outputs messages to the standard output stream.
 */
export const out: ConsolaInstance = createLogger(process.stdout)
