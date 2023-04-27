import { createLogger } from './logger'

const { log: print } = createLogger(process.stdout)

export { print }
