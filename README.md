# JotBot - AI-powered code documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/modernice/jotbot.svg)](https://pkg.go.dev/github.com/modernice/jotbot)
[![Test](https://github.com/modernice/jotbot/actions/workflows/test.yml/badge.svg)](https://github.com/modernice/jotbot/actions/workflows/test.yml)
[![jotbot-es](https://github.com/modernice/jotbot/actions/workflows/jotbot-es.yml/badge.svg)](https://github.com/modernice/jotbot/actions/workflows/jotbot-es.yml)

`JotBot` auto-generates missing documentation for your Go and TypeScript files.
The documentation of this repository was entirely generated by JotBot, so feel
free to check out the results in the source code (Go code is currently missing
documentation).

~~The documentation for this repository was created using the `gpt-3.5-turbo` model.~~
In my personal tests, I found that `text-davinci-003` consistently produces
better results. However, due to its 10x higher cost, the default setting remains
gpt-3.5-turbo.

I'm in the process of re-generating the documentation using GPT-4. Will update when done.

## Features

- Generate documentation for Go and TypeScript codebases
- Customize glob patterns for included and excluded files
- Filter code symbols by matching regular expressions
- Limit the number of files to generate documentation for
- Run in dry mode to preview changes without applying them
- Control the AI model and token limits used for generating documentation
- Optionally commit changes to a Git branch

### To-Do

- [ ] Configurable OpenAI settings (temperature, top_p etc.)
- [ ] _Any ideas?_ [open an issue](//github.com/modernice/jotbot/issues) or [start a discussion](//github.com/modernice/jotbot/discussions)

## Installation

### Via `go install`

If you have Go installed, you can simply install JotBot using `go install`:

```
go install github.com/modernice/jotbot/cmd/jotbot@main
```

### Standalone binary

> TODO: Setup GitHub Action (GoReleaser)

### TypeScript support

~~To enable TypeScript (and JavaScript) support, you also need to install the
`jotbot-es` npm package:~~

> TODO: Upload jotbot-es to npm

```
npm install -g jotbot-es
pnpm install -g jotbot-es
```

## Usage

To generate missing documentation for your codebase, run the following command:

```
jotbot generate [options]
```

By default, this command will find all Go and TypeScript (and JavaScript) files
in the current and nested directories and generate documentation for them.
Excluded from the search are by default:

```ts
[
	"**/.*/**",
	"**/dist/**",
	"**/node_modules/**",
	"**/vendor/**",
	"**/testdata/**",
	"**/test/**",
	"**/tests/**"
]
```

## CLI options

| Option             | Alias | Description                                                   | Default          |
|--------------------|-------|---------------------------------------------------------------|------------------|
| `--root`           |       | Root directory of the repository                              | "."              |
| `--include`        | `-i`  | Glob pattern(s) to include files                              |                  |
| `--exclude`        | `-e`  | Glob pattern(s) to exclude files                              |                  |
| `--exclude-internal` | `-E` | Exclude 'internal' directories (Go-specific)                  | true             |
| `--match`          | `-m`  | Regular expression(s) to match identifiers                    |                  |
| `--symbol`         | `-s`  | Symbol(s) to search for in code (TS/JS-specific)              |                  |
| `--branch`         |       | Branch name to commit changes to (leave empty to not commit)  |                  |
| `--limit`          |       | Limit the number of files to generate documentation for       | 0                |
| `--dry`            |      | Print the changes without applying them                       | false            |
| `--model`          |       | OpenAI model used to generate documentation                   | "gpt-3.5-turbo"   |
| `--maxTokens`      |       | Maximum number of tokens to generate for a single documentation | 512              |
| `--parallel`       | `-p`  | Number of files to handle concurrently                        | 4                |
| `--workers`        |       | Number of workers to use per file                             | 2                |
| `--override`       | `-o`  | Override existing documentation                               |                  |
| `--key`            |       | OpenAI API key                                                |                  |
| `--verbose`        | `-v`  | Enable verbose logging                                        | false            |

For a full list of options and their descriptions, run:

```
jotbot --help
```

## License

[MIT](./LICENSE)
