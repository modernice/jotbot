# opendocs - AI Powered Code Documentation Generator

`opendocs` is an AI-powered Go library and CLI tool that generates code
documentation for your Go projects (more languages to come).

opendocs leverages the power of OpenAI's GPT-3 model to generate documentation
for functions and types that lack proper documentation. The tool is configurable
and offers options to filter files, limit the number of generated documentations,
override or clear existing documentation, and commit changes to Git.

The documentation of this repository was entirely generated using opendocs, so
feel free to check out the source code to see the results. The documentation was
generated using the `gpt-3.5-turbo` model.

## Features

- Generate documentation for Go projects using OpenAI's GPT-3 model
- Apply generated documentation patches to the codebase
- Filter target files using glob patterns
- Limit the number of files and documentations generated
- Override or clear existing documentation
- Dry run mode to preview changes without applying them
- Commit changes directly to a Git branch

## Installation

To use opendocs as a library, run:

```bash
go get -u github.com/modernice/opendocs
```

To use opendocs as a CLI tool, run:

```bash
go install github.com/modernice/opendocs/cmd/opendocs@latest
```

## Usage

You can use opendocs directly from the command line to generate missing
documentation for your Go files:

```bash
opendocs generate --key <OPENAI_API_KEY>
```

### CLI options

- `--key`: OpenAI API key (required)
- `--root`: Root directory of the repository (default: ".")
- `-f`, `--filter:` Glob pattern(s) to filter files
- `--commit`: Commit changes to Git (default: true)
- `--branch`: Branch name to commit changes to (default: "opendocs-patch")
- `--limit:` Limit the number of documentations to generate (default: 0, no limit)
- `--file-limit`: Limit the number of files to generate documentations for (default: 0, no limit)
- `--dry`: Print the changes without applying them (default: false)
- `--model`: OpenAI model to use (default: "gpt-3.5-turbo")
- `-o, `--override`: Override existing documentation (default: false)
- `-c, `--clear`: Clear existing documentation (default: false)
- `-v, `--verbose`: Enable verbose logging (default: false)

## License

[MIT](./LICENSE)
