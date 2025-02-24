# GONX

A terminal user interface for managing Nx workspaces.

## Features

- Bundle analyser
- Build analyser
- Lint analyser

## Installation

Make sure to have Go installed on your machine. You can install it from [here](https://golang.org/doc/install).

Add the Go bin directory to your PATH (in the `.zshrc` or `.bashrc` file):

```zshrc
export PATH=$PATH:$(go env GOPATH)/bin
```

Source the file:
```bash
source ~/.zshrc
```

Install the application:

```bash
go install github.com/ionut-t/gonx@latest
```

Install from a specific branch:

```bash
go install github.com/ionut-t/gonx@branch-name
```

## Usage

Navigate to your Nx workspace directory and run:

```bash
gonx
```

## Development

1. Clone the repository:
```bash
git clone https://github.com/ionut-t/gonx.git
cd gonx
```
2. Install dependencies:

```bash
go mod tidy
```

3. Build and install the application:

```bash
go build && go install
```

## License

MIT

