# gostats

`gostats` is a CLI tool for measuring the proportion of test code in your codebase. It uses flexible file matching and .gitignore support to count both total LOC and target LOC (typically test files), then calculates their percentage share.

## Features

Line counting with newline edge case handling

- Configurable file pattern matching via .gsrc file
- Honors .gitignore and skips ignored files
- Minimal output format (suitable for CI)

## Installation

Download for Mac:

```sh
curl -L -o gostats.tar.gz https://github.com/bacebu4/gostats/releases/download/v0.0.1/gostats-v0.0.1-darwin-arm64.tar.gz && tar -xzf gostats.tar.gz && mv gostats /usr/local/bin/
```

## Configuration

Create a .`gsrc` file in your home directory (`~/.gsrc`) with the following format:

```json
{
  "targetPatterns": ["*test*.ts", "*Test*.ts", "*e2e*.ts"],
  "totalPatterns": ["*.ts"]
}
```

# Usage

```bash
gostats
```

Example output:

```
Sum target: 3401 LOC
Sum total: 4020 LOC
Percentage: 84.63%
```
