# Gorack

[![Go Report Card](https://goreportcard.com/badge/github.com/pachev/gorack)](https://goreportcard.com/report/github.com/pachev/gorack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go API for calculating optimal weight plate combinations for barbell exercises. Available at [https://gorack.pachevjoseph.com/v1/api](https://gorack.pachevjoseph.com/v1/api)

Try the web interface at [https://pachev.github.io/gorack/](https://pachev.github.io/gorack/)

## Features

* Calculate optimal plate combinations for any desired barbell weight
* Customize bar weight and available plates
* Get sensible defaults with common weights and standard Olympic barbell (45lb)
* Quick REST API access: [`/v1/api/rack?weight=335`](https://gorack.pachevjoseph.com/v1/api/rack?weight=335) returns instant results
* Smart optimization algorithm for finding the best plate combinations

> **Note**: All plate values in responses represent **pairs** (e.g., `"fortyFives": 2` means two 45lb plates on **each side** of the barbell)

## Getting Started

### Prerequisites

* Go 1.22 or later
* [mise](https://github.com/jdx/mise) (optional but recommended for development)

### Installation

#### Using mise (recommended)

The project includes a `mise.toml` configuration file for easy setup:

```bash
# Install dependencies and tools defined in mise.toml
mise install

# Build the project
mise run build

# Run the server
mise run run
```

#### Manual installation

```bash
# Get dependencies
go mod tidy

# Build the binary
go build -o ./tmp/gorack .

# Run the server
./tmp/gorack
```

The API will be available at http://localhost:8080/v1/api

### Environment Variables

* `API_PORT`: Set custom port (default: 8080)
  ```bash
  API_PORT=9000 mise run run
  ```

## API Usage

### Simple GET Request

For quick calculations with default plate availability:

```
GET /v1/api/rack?weight=225
```

Response:
```json
{
  "barWeight": 45,
  "fortyFives": 2,
  "desiredWeight": 225,
  "achievedWeight": 225,
  "message": "You got this!"
}
```

### Customized POST Request

For calculating with specific plate availability:

```bash
curl -X POST \
  https://gorack.pachevjoseph.com/v1/api/rack \
  -H 'content-type: application/json' \
  -d '{
    "barWeight": 35,
    "fortyFives": 1,
    "thirtyFives": 2,
    "twentyFives": 2,
    "tens": 3,
    "fives": 2,
    "twoDotFives": 2,
    "desiredWeight": 255
}'
```

Response:
```json
{
  "barWeight": 35,
  "fortyFives": 1,
  "thirtyFives": 1,
  "twentyFives": 1,
  "tens": 1,
  "fives": 1,
  "desiredWeight": 255,
  "achievedWeight": 255,
  "message": "You got this!"
}
```

## Available Plate Types

The API supports the following plate types (values represent pairs):

| JSON Parameter | Weight (lbs per plate) | Description |
|----------------|------------------------|-------------|
| `hundreds` | 100 | 100lb plates |
| `fortyFives` | 45 | 45lb plates |
| `thirtyFives` | 35 | 35lb plates |
| `twentyFives` | 25 | 25lb plates |
| `tens` | 10 | 10lb plates |
| `fives` | 5 | 5lb plates |
| `twoDotFives` | 2.5 | 2.5lb plates |
| `oneDotTwoFives` | 1.25 | 1.25lb plates |

## Development

The project uses mise for streamlined development workflows:

```bash
# Live reload during development
mise run watch

# Run Go mod tidy
mise run tidy

# Clean build artifacts
mise run clean
```

## How It Works

Gorack uses a greedy algorithm to calculate the optimal plate combination:

1. Start with the heaviest available plate
2. Add pairs of plates to the bar, always selecting the heaviest available option
3. Continue until reaching the desired weight or running out of suitable plates
4. Return the achieved weight and plate configuration

## License

MIT
