import { Command } from 'commander'
import pkg from '../../package.json'
import { withFindCmd } from './find'

export function createCLI() {
  const program = new Command('jotbot-es')
    .description(pkg.description)
    .version(pkg.version)

  return withFindCmd(program)
}
