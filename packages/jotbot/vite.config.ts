import { defineConfig } from 'vite'
import dts from 'vite-plugin-dts'
import { readPackageJSON } from 'pkg-types'

export default defineConfig(async () => {
  const pkg = await readPackageJSON()

  function isExternal(dep: string) {
    return (
      ['perf_hooks'].includes(dep) ||
      dep.startsWith('node:') ||
      [
        ...Object.keys(pkg.dependencies ?? {}),
        ...Object.keys(pkg.peerDependencies ?? {}),
      ].includes(dep)
    )
  }

  return {
    plugins: [dts({ entryRoot: './src' })],

    build: {
      lib: {
        entry: './src/index.ts',
        name: 'JotBot',
        fileName: (format, name) =>
          `${name}.${format === 'es' ? 'mjs' : format}`,
        formats: ['es', 'cjs'],
      },

      rollupOptions: {
        external: isExternal,
        input: {
          index: './src/index.ts',
          cli: './src/cli/index.ts',
        },
      },
    },
  }
})
