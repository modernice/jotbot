# JotBot Utilities for TypeScript and JavaScript

JotBot is a CLI tool that provides utilities for TypeScript and JavaScript.
It can find uncommented symbols in the source code, extract node positions,
and minify the code to fit into OpenAI's model contexts.

## Features

- Find uncommented symbols in TypeScript and JavaScript source code
- Extract node positions from source code
- Minify TypeScript and JavaScript code

## Installation

To add TypeScript support to JotBot, install the `jotbot-ts` package globally:

```
npm i -g jotbot-ts
pnpm i -g jotbot-ts
```

## Usage

`jotbot-ts` should normally not be called directly. Instead, use the main
[JotBot](https://github.com/modernice/jotbot) CLI that utilizes `jotbot-ts`
under the hood.

```
jotbot-ts [command] [options]
```

### Commands

#### `find`

Find uncommented symbols in TypeScript and JavaScript source code.

Example:
```
jotbot-ts find 'export const foo = "foo"; export function bar() {}' --json
```

Ouput:
```
[
	"var:foo",
	"func:bar"
]
```

Options:
- `-p, --path <file>`: Path to the TypeScript or JavaScript file (instead of code)
- `-s, --symbols <symbols>`: Symbols to search for (comma-separated)
- `--documented`: Also find documented symbols
- `-v, --verbose`: Verbose output
- `-f, --format <format>`: Configure formatting
- `--json`: Output findings as JSON (same as `--format json`)

#### `pos`

Extract node position from TypeScript and JavaScript source code.

Example:
```
jotbot-ts pos var:foo 'export const foo = "foo"; export const bar = "bar"'
```

Ouput:
```
{
	"line": 0,
	"character": 26
}
```

Options:
- `-p, --path <file>`: Path to the TypeScript or JavaScript file (instead of code)
- `-v, --verbose`: Verbose output

#### `minify`

Minify TypeScript and JavaScript source code.

Example:
```
jotbot-ts minify [code]
```

Options:
- `-p, --path <file>`: Path to the TypeScript or JavaScript file (instead of code)
- `-m, --model <model>`: Use max tokens of OpenAI model
- `-t, --tokens <number>`: Maximum output tokens (overrides `--model`)
- `-s, --steps <steps>`: Minification steps (comma-separated)
- `--print-tokens`: Return input and output tokens (requires `--format="json"`)
- `-v, --verbose`: Verbose output
- `-f, --format <format>`: Configure formatting
- `--json`: Output findings as JSON (same as `--format json`)

## License

JotBot is released under the MIT License.
