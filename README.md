# JotBot - AI-powered code documentation

`JotBot` auto-generates missing documentation for your Go repositories
(more languages to come). The documentation of this repository was entirely
generated by JotBot, so feel free to check out the results in the source code.

## Features

- Generate documentation for Go projects using OpenAI's GPT models
- Apply generated documentation patches to the codebase
- Filter target files using glob patterns
- Limit the number of files and documentations generated
- Optionally override existing documentation
- Optionally send source code without comments to GPT (when generating new comments)
- Dry run mode to preview changes without applying them
- Commit changes in a separate a Git branch

### To-Do

- [ ] Proof-read generated documentation 😅
- [ ] Test with GPT-4 (need access)
- [ ] Add support for other languages
- [ ] Configurable OpenAI settings (temperature, top_p etc.)
- [ ] rename `filter` option to `include`
- [ ] `exclude` option to exclude files from documentation generation
- [ ] GitHub Action to generate documentation on push
- [ ] ...
- [ ] _any ideas?_ [open an issue](./issues) or [start a discussion](./discussions)

## Installation

```bash
go install github.com/modernice/jotbot/cmd/jotbot@main
```

## Usage

Call `jotbot` from the command line to generate missing documentation for your
Go files:

```bash
jotbot generate --key <OPENAI_API_KEY>
```

### Options

| Option | Description | Default |
| --- | --- | --- |
| `--key` | OpenAI API key (required) | - |
| `--root` | Root directory to start searching for files | `.` |
| `-f, --filter` | Glob pattern(s) to filter files | - |
| `--commit` | Commit changes to Git | `true` |
| `--branch` | Branch name to commit changes to | `"jotbot-patch"` |
| `--limit` | Limit the number of documentations to generate | `0` (no limit) |
| `--file`-limit | Limit the number of files to generate documentations for | `0` (no limit) |
| `--dry` | Print the changes without applying them | `false` |
| `--model` | OpenAI model to use | `"gpt-3.5-turbo"` |
| `-o, --override` | Override existing documentation | `false` |
| `-c, --clear` | Clear existing documentation for GPT | `false` |
| `-v, --verbose` | Enable verbose logging | `false` |

## License

[MIT](./LICENSE)
