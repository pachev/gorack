# Gorack

[![Go Report Card](https://goreportcard.com/badge/github.com/pachev/gorack)][1]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)][6]


An API for calculating racking weight for barbell. The api is available here [https://gorack.pachevjoseph.com/v1/api][2]

An example application using this API can be found [here][3]

**Features:**

* Easily Calculate needed weight for any exercise
* Easily Define bar weight along with available plates
* Sensible defaults are provided with common weights along with Olympic barbell weight (45lb)
* Simple to use without any UI needed: [`v1/api/rack?weigh=335`][4] in your browser will return instant results 

_Note: returned values should be considered as pairs e.g. (`fortyFives: 2` means two forty-five plates on each side)_

## Getting Started

### Installation

A Makefile is included with this repository. 
```bash
$ make all
```

The command above will fetch sources and run the application on http://localhost:8080/v1/api

## Advanced Requests

If you don't like the defaults provided from the `GET` call, you can make a `POST` request with the available weights 
that you have (the application assumes inputs as pairs). 

__Example: I'd like to achieve 285 lb with a standard olympic bar with limited plates:__
```bash
curl -X POST \
  https://gorack.pachevjoseph.com/v1/api/rack \
  -H 'content-type: application/json' \
  -d '{
	"fortyFives": 1,
	"thirtyFives": 1,
	"twentyFives": 1,
	"tens": 1,
	"fives": 2,
	"twoDotFives": 1,
	"desiredWeight": 285
}'
```

In the request above, "barWeight" was not included as was therefore defaulted to 45lb. The returned payload with look like below (returned values are in pairs):

```json
{
    "fortyFives": 1,
    "thirtyFives": 1,
    "twentyFives": 1,
    "tens": 1,
    "fives": 1,
    "desiredWeight": 285,
    "achievedWeight": 285,
    "message": "You got this!"
}
```

## License
MIT

TODO: Enhance README

[1]: https://goreportcard.com/report/github.com/pachev/gorack
[2]: https://gorack.pachevjoseph.com/v1/api/rack
[3]: Nothing
[4]: https://gorack.pachevjoseph.com/v1/api/rack?weight=335
[5]: https://golang.org/doc/install
[6]: https://opensource.org/licenses/MIT