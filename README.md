# datauri

[![Go Reference](https://pkg.go.dev/badge/github.com/invopop/datauri.svg)](https://pkg.go.dev/github.com/invopop/datauri)

Data URI Schemes for Go

[![CI 🏗](https://github.com/invopop/datauri/actions/workflows/ci.yml/badge.svg)](https://github.com/invopop/datauri/actions/workflows/ci.yml)

This package parses and generates Data URI Schemes for the Go language,
according to [RFC 2397](http://tools.ietf.org/html/rfc2397).

Data URIs are small chunks of data commonly used in browsers to display inline data,
typically like small images, or when you use the FileReader API of the browser.

## Command

Use the [`datauri`](./cmd/datauri) command to encode/decode data URI streams.

Install it with `go install github.com/invopop/datauri/cmd/datauri@latest`.

## [LICENSE](LICENSE)

Forked from [RealImage/dataurl](https://github.com/RealImage/dataurl), which in turn is forked from [vincent-petithory/dataurl](https://github.com/vincent-petithory/dataurl)
with contributions from [MagicalTux/dataurl](https://github.com/MagicalTux/dataurl/tree/fix-issue-5).

Datauri is available under the terms of the MIT license.
