package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path"

	"github.com/invopop/datauri"
)

var (
	performDecode bool
	asciiEncoding bool
	mimetype      string
)

func init() {
	const decodeUsage = "decode data instead of encoding"
	flag.BoolVar(&performDecode, "decode", false, decodeUsage)
	flag.BoolVar(&performDecode, "d", false, decodeUsage)

	const mimetypeUsage = "force the mimetype of the data to encode to this value"
	flag.StringVar(&mimetype, "mimetype", "", mimetypeUsage)
	flag.StringVar(&mimetype, "m", "", mimetypeUsage)

	const asciiUsage = "encode data using ascii instead of base64"
	flag.BoolVar(&asciiEncoding, "ascii", false, asciiUsage)
	flag.BoolVar(&asciiEncoding, "a", false, asciiUsage)

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`datauri - Encode or decode datauri data and print to standard output

Usage: datauri [OPTION]... [FILE]

  datauri encodes or decodes FILE or standard input if FILE is - or omitted, and prints to standard output.
  Unless -mimetype is used, when FILE is specified, datauri will attempt to detect its mimetype using Go's mime.TypeByExtension (http://golang.org/pkg/mime/#TypeByExtension). If this fails or data is read from STDIN, the mimetype will default to application/octet-stream.

Options:
`)
		flag.PrintDefaults()
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	var (
		in               io.Reader
		out              = os.Stdout
		encoding         = datauri.EncodingBase64
		detectedMimetype string
	)
	switch n := flag.NArg(); n {
	case 0:
		in = os.Stdin
	case 1:
		if flag.Arg(0) == "-" {
			in = os.Stdin
			break
		}
		if f, err := os.Open(flag.Arg(0)); err != nil {
			log.Fatal(err)
		} else {
			in = f
			defer f.Close() //nolint:errcheck
		}
		ext := path.Ext(flag.Arg(0))
		detectedMimetype = mime.TypeByExtension(ext)
	}

	switch {
	case mimetype == "" && detectedMimetype == "":
		mimetype = "application/octet-stream"
	case mimetype == "" && detectedMimetype != "":
		mimetype = detectedMimetype
	}

	if performDecode {
		if err := decode(in, out); err != nil {
			log.Fatal(err)
		}
	} else {
		if asciiEncoding {
			encoding = datauri.EncodingASCII
		}
		if err := encode(in, out, encoding, mimetype); err != nil {
			log.Fatal(err)
		}
	}
}

func decode(in io.Reader, out io.Writer) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	du, err := datauri.Decode(in)
	if err != nil {
		return
	}

	_, err = out.Write(du.Data)
	return
}

func encode(in io.Reader, out io.Writer, encoding string, mediatype string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("%v", e)
			}
			return
		}
	}()
	b, err := io.ReadAll(in)
	if err != nil {
		return
	}

	du := datauri.New(b, mediatype)
	du.Encoding = encoding

	_, err = du.WriteTo(out)
	return
}
