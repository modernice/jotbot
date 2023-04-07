import { useLogger } from '@modernice/openai'
import { execaSync } from 'execa'
import { fileURLToPath } from 'url'

export function patchFile(
  filepath: string,
  symbol: string,
  documentation: string,
) {
  const { log, error } = useLogger()

  const mainPath = fileURLToPath(new URL('../go/cmd/main.go', import.meta.url))

  const { stdout, stderr } = execaSync(
    'go',
    [
      'run',
      mainPath,
      '-f',
      filepath,
      '-t',
      symbol,
      '-d',
      JSON.stringify(documentation),
    ],
    { cwd: fileURLToPath(new URL('../go', import.meta.url)) },
  )

  if (stderr) {
    error(stderr)
  }

  if (stdout) {
    log(stdout)
  }
}
