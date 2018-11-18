# Gorack

[![Go Report Card](https://goreportcard.com/badge/github.com/pachev/gorack)][1]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)][6]


An API for calculating racking weight for barbell. The api is available here [https://gorack.pachevjoseph.com/v1][2]

An example application using this API can be found [here][3]

**Featuers:**

* Easily Calculate needed weight for any exercise
* Easily Define bar weight along with available plates
* Sensible defaults are provided with common weights along with Olympic barbell weight (45lb)
* Simple to use without any UI needed: [`v1/rack?weigh=335`][4] in your browser will return instant results 

## Getting Started

### Installation

A Makefile is included with this repository. 
```bash
$ make all
```

The command above will fetch sources and run the application on http://localhost:8080/v1

TODO: Enhance README

## License
MIT

[1]: https://goreportcard.com/report/github.com/pachev/gorack 
[2]: https://gorack.pachevjoseph.com/v1
[3]: Nothing
[4]: https://gorack.pachevjoseph.com/v1/rack?weight=335
[5]: https://golang.org/doc/install
[6]: https://opensource.org/licenses/MIT