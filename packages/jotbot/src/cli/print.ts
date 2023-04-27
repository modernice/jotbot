import type { ConsolaInstance } from 'consola'
import { createLogger } from './logger'

export const out: ConsolaInstance = createLogger(process.stdout)
