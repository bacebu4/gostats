# gostats

`gostats` is a CLI tool for measuring the proportion of test code in your codebase. It uses flexible file matching and `.gitignore` support to count both total LOC and target LOC (typically test files), then calculates their percentage share.

## Features

Line counting with newline edge case handling

- Configurable file pattern matching via `.gostats.json` file
- Honors `.gitignore` and skips ignored files
- Minimal output format (suitable for CI)

## Installation

Download the latest release from the [Releases](https://github.com/bacebu4/gostats/releases) page.

### macOS (ARM64)

```sh
curl -L -o gostats.tar.gz https://github.com/bacebu4/gostats/releases/download/v0.0.1/gostats-v0.0.1-darwin-arm64.tar.gz && tar -xzf gostats.tar.gz && mv gostats /usr/local/bin/
```

### Linux (x86_64)

```sh
curl -L -o gostats.tar.gz https://github.com/bacebu4/gostats/releases/download/v0.0.1/gostats-v0.0.1-linux-amd64.tar.gz && tar -xzf gostats.tar.gz && mv gostats /usr/local/bin/
```

## Configuration

Create a `.gostats.json` file in your home directory (`~/.gostats.json`) with the following format:

```json
{
  "targetPatterns": ["src/**/*test*.ts", "src/**/*test*.ts", "src/**/*Test*.ts", "src/**/*e2e*.ts", "src/**/*generate*.ts"],
  "totalPatterns": ["src/**/*.ts"]
}
```

Both targetPatterns and totalPatterns are glob patterns using the syntax from [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar).

> [!NOTE]
> Only `.gitignore` from the working directory is considered; nested and global `.gitignore` files are not supported.

## Usage

From your project root:

```bash
gostats
```

Example output:

```
Sum target: 3401 LOC
Sum total: 4020 LOC
Percentage: 84.63%
```

## Roadmap

- Ability to pass config via CLI flags (override `.gostats.json`)