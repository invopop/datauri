# Dataurl

[![Go Reference](https://pkg.go.dev/badge/github.com/invopop/dataurl.svg)](https://pkg.go.dev/github.com/invopop/dataurl)

Data URL Schemes for Go

[![CI 🏗](https://github.com/invopop/dataurl/actions/workflows/ci.yml/badge.svg)](https://github.com/invopop/dataurl/actions/workflows/ci.yml)

This package parses and generates Data URL Schemes for the Go language,
according to [RFC 2397](http://tools.ietf.org/html/rfc2397).

Data URLs are small chunks of data commonly used in browsers to display inline data,
typically like small images, or when you use the FileReader API of the browser.

## Command

Use the [`dataurl`](./cmd/dataurl) command to encode/decode dataurl streams.

Install it with `go install github.com/ananthb/dataurl/cmd/dataurl@latest`.

## [LICENSE](LICENSE)

Forked from [RealImage/dataurl](https://github.com/RealImage/dataurl), which in turn is forked from [vincent-petithory/dataurl](https://github.com/vincent-petithory/dataurl)
with contributions from [MagicalTux/dataurl](https://github.com/MagicalTux/dataurl/tree/fix-issue-5).

Dataurl is available under the terms of the MIT license.
