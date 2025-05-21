package datauri

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

type dataURITest struct {
	InputRawDataURI string
	ExpectedItems   []item
	ExpectedDataURI DataURI
}

func genTestTable() []dataURITest {
	return []dataURITest{
		{
			`data:;base64,aGV5YQ==`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemParamSemicolon, ";"},
				{itemBase64Enc, "base64"},
				{itemDataComma, ","},
				{itemData, "aGV5YQ=="},
				{itemEOF, ""},
			},
			DataURI{
				defaultMediaType(),
				EncodingBase64,
				[]byte("heya"),
			},
		},
		{
			`data:text/plain;base64,aGV5YQ==`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemMediaType, "text"},
				{itemMediaSep, "/"},
				{itemMediaSubType, "plain"},
				{itemParamSemicolon, ";"},
				{itemBase64Enc, "base64"},
				{itemDataComma, ","},
				{itemData, "aGV5YQ=="},
				{itemEOF, ""},
			},
			DataURI{
				MediaType{
					"text",
					"plain",
					map[string]string{},
				},
				EncodingBase64,
				[]byte("heya"),
			},
		},
		{
			`data:text/plain;charset=utf-8;base64,aGV5YQ==`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemMediaType, "text"},
				{itemMediaSep, "/"},
				{itemMediaSubType, "plain"},
				{itemParamSemicolon, ";"},
				{itemParamAttr, "charset"},
				{itemParamEqual, "="},
				{itemParamVal, "utf-8"},
				{itemParamSemicolon, ";"},
				{itemBase64Enc, "base64"},
				{itemDataComma, ","},
				{itemData, "aGV5YQ=="},
				{itemEOF, ""},
			},
			DataURI{
				MediaType{
					"text",
					"plain",
					map[string]string{
						"charset": "utf-8",
					},
				},
				EncodingBase64,
				[]byte("heya"),
			},
		},
		{
			`data:text/plain;charset=utf-8;foo=bar;base64,aGV5YQ==`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemMediaType, "text"},
				{itemMediaSep, "/"},
				{itemMediaSubType, "plain"},
				{itemParamSemicolon, ";"},
				{itemParamAttr, "charset"},
				{itemParamEqual, "="},
				{itemParamVal, "utf-8"},
				{itemParamSemicolon, ";"},
				{itemParamAttr, "foo"},
				{itemParamEqual, "="},
				{itemParamVal, "bar"},
				{itemParamSemicolon, ";"},
				{itemBase64Enc, "base64"},
				{itemDataComma, ","},
				{itemData, "aGV5YQ=="},
				{itemEOF, ""},
			},
			DataURI{
				MediaType{
					"text",
					"plain",
					map[string]string{
						"charset": "utf-8",
						"foo":     "bar",
					},
				},
				EncodingBase64,
				[]byte("heya"),
			},
		},
		{
			`data:application/json;charset=utf-8;foo="b\"<@>\"r";style=unformatted%20json;base64,eyJtc2ciOiAiaGV5YSJ9`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemMediaType, "application"},
				{itemMediaSep, "/"},
				{itemMediaSubType, "json"},
				{itemParamSemicolon, ";"},
				{itemParamAttr, "charset"},
				{itemParamEqual, "="},
				{itemParamVal, "utf-8"},
				{itemParamSemicolon, ";"},
				{itemParamAttr, "foo"},
				{itemParamEqual, "="},
				{itemLeftStringQuote, "\""},
				{itemParamVal, `b\"<@>\"r`},
				{itemRightStringQuote, "\""},
				{itemParamSemicolon, ";"},
				{itemParamAttr, "style"},
				{itemParamEqual, "="},
				{itemParamVal, "unformatted%20json"},
				{itemParamSemicolon, ";"},
				{itemBase64Enc, "base64"},
				{itemDataComma, ","},
				{itemData, "eyJtc2ciOiAiaGV5YSJ9"},
				{itemEOF, ""},
			},
			DataURI{
				MediaType{
					"application",
					"json",
					map[string]string{
						"charset": "utf-8",
						"foo":     `b"<@>"r`,
						"style":   "unformatted json",
					},
				},
				EncodingBase64,
				[]byte(`{"msg": "heya"}`),
			},
		},
		{
			`data:xxx;base64,aGV5YQ==`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemError, "invalid character for media type"},
			},
			DataURI{},
		},
		{
			`data:,`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemDataComma, ","},
				{itemEOF, ""},
			},
			DataURI{
				defaultMediaType(),
				EncodingASCII,
				[]byte(""),
			},
		},
		{
			`data:,A%20brief%20note`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemDataComma, ","},
				{itemData, "A%20brief%20note"},
				{itemEOF, ""},
			},
			DataURI{
				defaultMediaType(),
				EncodingASCII,
				[]byte("A brief note"),
			},
		},
		{
			`data:image/svg+xml-im.a.fake;base64,cGllLXN0b2NrX1RoaXJ0eQ==`,
			[]item{
				{itemDataPrefix, dataPrefix},
				{itemMediaType, "image"},
				{itemMediaSep, "/"},
				{itemMediaSubType, "svg+xml-im.a.fake"},
				{itemParamSemicolon, ";"},
				{itemBase64Enc, "base64"},
				{itemDataComma, ","},
				{itemData, "cGllLXN0b2NrX1RoaXJ0eQ=="},
				{itemEOF, ""},
			},
			DataURI{
				MediaType{
					"image",
					"svg+xml-im.a.fake",
					map[string]string{},
				},
				EncodingBase64,
				[]byte("pie-stock_Thirty"),
			},
		},
	}
}

func expectItems(expected, actual []item) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := range expected {
		if expected[i].t != actual[i].t {
			return false
		}
		if expected[i].val != actual[i].val {
			return false
		}
	}
	return true
}

func equal(du1, du2 *DataURI) (bool, error) {
	if !reflect.DeepEqual(du1.MediaType, du2.MediaType) {
		return false, nil
	}
	if du1.Encoding != du2.Encoding {
		return false, nil
	}

	if du1.Data == nil || du2.Data == nil {
		return false, fmt.Errorf("nil Data")
	}

	if !bytes.Equal(du1.Data, du2.Data) {
		return false, nil
	}
	return true, nil
}

func TestLexDataURIs(t *testing.T) {
	for _, test := range genTestTable() {
		l := lex(test.InputRawDataURI)
		var items []item
		for item := range l.items {
			items = append(items, item)
		}
		if !expectItems(test.ExpectedItems, items) {
			t.Errorf("Expected %v, got %v", test.ExpectedItems, items)
		}
	}
}

func testDataURIs(t *testing.T, factory func(string) (*DataURI, error)) {
	for _, test := range genTestTable() {
		var expectedItemError string
		for _, item := range test.ExpectedItems {
			if item.t == itemError {
				expectedItemError = item.String()
				break
			}
		}
		dataURI, err := factory(test.InputRawDataURI)
		if expectedItemError == "" && err != nil {
			t.Error(err)
			continue
		} else if expectedItemError != "" && err == nil {
			t.Errorf("Expected error \"%s\", got nil", expectedItemError)
			continue
		} else if expectedItemError != "" && err != nil {
			if err.Error() != expectedItemError {
				t.Errorf("Expected error \"%s\", got \"%s\"", expectedItemError, err.Error())
			}
			continue
		}

		if ok, err := equal(dataURI, &test.ExpectedDataURI); err != nil {
			t.Error(err)
		} else if !ok {
			t.Errorf("Expected %v, got %v", test.ExpectedDataURI, *dataURI)
		}
	}
}

func TestDataURIsWithDecode(t *testing.T) {
	testDataURIs(t, func(s string) (*DataURI, error) {
		return Decode(strings.NewReader(s))
	})
}

func TestDataURIsWithDecodeString(t *testing.T) {
	testDataURIs(t, func(s string) (*DataURI, error) {
		return DecodeString(s)
	})
}

func TestDataURIsWithUnmarshalText(t *testing.T) {
	testDataURIs(t, func(s string) (*DataURI, error) {
		d := &DataURI{}
		err := d.UnmarshalText([]byte(s))
		return d, err
	})
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		s           string
		roundTripOk bool
	}{
		{`data:text/plain;charset=utf-8;foo=bar;base64,aGV5YQ==`, true},
		{`data:;charset=utf-8;foo=bar;base64,aGV5YQ==`, false},
		{`data:text/plain;charset=utf-8;foo="bar";base64,aGV5YQ==`, false},
		{`data:text/plain;charset=utf-8;foo="bar",A%20brief%20note`, false},
		{`data:text/plain;charset=utf-8;foo=bar,A%20brief%20note`, true},
	}
	for _, test := range tests {
		dataURI, err := DecodeString(test.s)
		if err != nil {
			t.Error(err)
			continue
		}
		dus := dataURI.String()
		if test.roundTripOk && dus != test.s {
			t.Errorf("Expected %s, got %s", test.s, dus)
		} else if !test.roundTripOk && dus == test.s {
			t.Errorf("Found %s, expected something else", test.s)
		}

		txt, err := dataURI.MarshalText()
		if err != nil {
			t.Error(err)
			continue
		}
		if test.roundTripOk && string(txt) != test.s {
			t.Errorf("MarshalText roundtrip: got '%s', want '%s'", txt, test.s)
		} else if !test.roundTripOk && string(txt) == test.s {
			t.Errorf("MarshalText roundtrip: got '%s', want something else", txt)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		Data            []byte
		MediaType       string
		ParamPairs      []string
		WillPanic       bool
		ExpectedDataURI *DataURI
	}{
		{
			[]byte(`{"msg": "heya"}`),
			"application/json",
			[]string{},
			false,
			&DataURI{
				MediaType{
					"application",
					"json",
					map[string]string{},
				},
				EncodingBase64,
				[]byte(`{"msg": "heya"}`),
			},
		},
		{
			[]byte(``),
			"application//json",
			[]string{},
			true,
			nil,
		},
		{
			[]byte(``),
			"",
			[]string{},
			true,
			nil,
		},
		{
			[]byte(`{"msg": "heya"}`),
			"text/plain",
			[]string{"charset", "utf-8"},
			false,
			&DataURI{
				MediaType{
					"text",
					"plain",
					map[string]string{
						"charset": "utf-8",
					},
				},
				EncodingBase64,
				[]byte(`{"msg": "heya"}`),
			},
		},
		{
			[]byte(`{"msg": "heya"}`),
			"text/plain",
			[]string{"charset", "utf-8", "name"},
			true,
			nil,
		},
	}
	for _, test := range tests {
		var dataURI *DataURI
		func() {
			defer func() {
				if test.WillPanic {
					if e := recover(); e == nil {
						t.Error("Expected panic didn't happen")
					}
				} else {
					if e := recover(); e != nil {
						t.Errorf("Unexpected panic: %v", e)
					}
				}
			}()
			dataURI = New(test.Data, test.MediaType, test.ParamPairs...)
		}()
		if test.WillPanic {
			if dataURI != nil {
				t.Error("Expected nil DataURI")
			}
		} else {
			if ok, err := equal(dataURI, test.ExpectedDataURI); err != nil {
				t.Error(err)
			} else if !ok {
				t.Errorf("Expected %v, got %v", test.ExpectedDataURI, *dataURI)
			}
		}
	}
}

var golangFavicon = strings.ReplaceAll(
	`AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAD///8AVE44//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb/
/uF2/1ROOP////8A////AFROOP/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+
4Xb//uF2//7hdv9UTjj/////AP///wBUTjj//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7h
dv/+4Xb//uF2//7hdv/+4Xb/VE44/////wD///8AVE44//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2
//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2/1ROOP////8A////AFROOP/+4Xb//uF2//7hdv/+4Xb/
/uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv9UTjj/////AP///wBUTjj//uF2//7hdv/+
4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb/VE44/////wD///8AVE44//7h
dv/+4Xb//uF2//7hdv/+4Xb/z7t5/8Kyev/+4Xb//993///dd///3Xf//uF2/1ROOP////8A////
AFROOP/+4Xb//uF2//7hdv//4Hn/dIzD//v8///7/P//dIzD//7hdv//3Xf//913//7hdv9UTjj/
////AP///wBUTjj//uF2///fd//+4Xb//uF2/6ajif90jMP/dIzD/46Zpv/+4Xb//+F1///feP/+
4Xb/VE44/////wD///8AVE44//7hdv/z1XT////////////Is3L/HyAj/x8gI//Is3L/////////
///z1XT//uF2/1ROOP////8A19nd/1ROOP/+4Xb/5+HS//v+//8RExf/Liwn//7hdv/+4Xb/5+HS
//v8//8RExf/Liwn//7hdv9UTjj/19nd/1ROOP94aDT/yKdO/+fh0v//////ERMX/y4sJ//+4Xb/
/uF2/+fh0v//////ERMX/y4sJ//Ip07/dWU3/1ROOP9UTjj/yKdO/6qSSP/Is3L/9fb7//f6///I
s3L//uF2//7hdv/Is3L////////////Is3L/qpJI/8inTv9UTjj/19nd/1ROOP97c07/qpJI/8in
Tv/Ip07//uF2//7hdv/+4Xb//uF2/8zBlv/Kv4//pZJU/3tzTv9UTjj/19nd/////wD///8A4eLl
/6CcjP97c07/e3NO/1dOMf9BOiX/TkUn/2VXLf97c07/e3NO/6CcjP/h4uX/////AP///wD///8A
////AP///wD///8A////AP///wDq6/H/3N/j/9fZ3f/q6/H/////AP///wD///8A////AP///wD/
//8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAA==`,
	"\n",
	"",
)

func TestEncodeBytes(t *testing.T) {
	mustDecode := func(s string) []byte {
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			panic(err)
		}
		return data
	}
	tests := []struct {
		Data           []byte
		ExpectedString string
	}{
		{
			[]byte(`A brief note`),
			"data:text/plain;charset=utf-8;base64,QSBicmllZiBub3Rl",
		},
		{
			[]byte{0xA, 0xFF, 0x99, 0x34, 0x56, 0x34, 0x00},
			`data:application/octet-stream;base64,Cv+ZNFY0AA==`,
		},
		{
			mustDecode(golangFavicon),
			`data:image/x-icon;base64,` + golangFavicon,
		},
	}
	for _, test := range tests {
		str := EncodeBytes(test.Data)
		if str != test.ExpectedString {
			t.Errorf("Expected %s, got %s", test.ExpectedString, str)
		}
	}
}

func BenchmarkLex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range genTestTable() {
			l := lex(test.InputRawDataURI)
			for range l.items {
			}
		}
	}
}

const rep = `^data:(?P<mediatype>\w+/[\w\+\-\.]+)?(?P<parameter>(?:;[\w\-]+="?[\w\-\\<>@,";:%]*"?)+)?(?P<base64>;base64)?,(?P<data>.*)$`

func TestRegexp(t *testing.T) {
	re, err := regexp.Compile(rep)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range genTestTable() {
		shouldMatch := true
		for _, item := range test.ExpectedItems {
			if item.t == itemError {
				shouldMatch = false
				break
			}
		}
		// just test it matches, do not parse
		if re.MatchString(test.InputRawDataURI) && !shouldMatch {
			t.Error("doesn't match", test.InputRawDataURI)
		} else if !re.MatchString(test.InputRawDataURI) && shouldMatch {
			t.Error("match", test.InputRawDataURI)
		}
	}
}

func BenchmarkRegexp(b *testing.B) {
	re, err := regexp.Compile(rep)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		for _, test := range genTestTable() {
			_ = re.FindStringSubmatch(test.InputRawDataURI)
		}
	}
}

func ExampleDecodeString() {
	dataURI, err := DecodeString(`data:text/plain;charset=utf-8;base64,aGV5YQ==`)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s, %s", dataURI.ContentType(), string(dataURI.Data))
	// Output: text/plain, heya
}

func ExampleDecode() {
	r, err := http.NewRequest(
		"POST",
		"/",
		strings.NewReader(
			`data:image/vnd.microsoft.icon;name=golang%20favicon;base64,`+golangFavicon,
		),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	var dataURI *DataURI
	h := func(_ http.ResponseWriter, r *http.Request) {
		var err error
		dataURI, err = Decode(r.Body)
		defer r.Body.Close() //nolint:errcheck
		if err != nil {
			fmt.Println(err)
		}
	}
	w := httptest.NewRecorder()
	h(w, r)
	fmt.Printf("%s: %s", dataURI.Params["name"], dataURI.ContentType())
	// Output: golang favicon: image/vnd.microsoft.icon
}
