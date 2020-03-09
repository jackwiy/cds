package exportentities

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/ovh/cds/sdk"
)

// GetFormatFromPath return a format.
func GetFormatFromPath(f string) (Format, error) {
	s := strings.ToLower(f)
	s = strings.TrimSpace(s)
	switch s {
	case "yaml", "yml", ".yaml", ".yml":
		return FormatYAML, nil
	case "json", ".json":
		return FormatJSON, nil
	default:
		return UnknownFormat, sdk.WithStack(ErrUnsupportedFormat)
	}
}

// GetFormatFromContentType return a format.
func GetFormatFromContentType(ct string) (Format, error) {
	switch ct {
	case "application/x-yaml", "text/x-yaml":
		return FormatYAML, nil
	case "application/json":
		return FormatJSON, nil
	default:
		return UnknownFormat, sdk.WithStack(ErrUnsupportedFormat)
	}
}

// GetFormatStr return a format in string.
func GetFormatStr(f Format) (string, error) {
	switch f {
	case FormatJSON:
		return "json", nil
	case FormatYAML:
		return "yaml", nil
	default:
		return "", sdk.WithStack(ErrUnsupportedFormat)
	}
}

// GetContentType returns the content type for a content type.
func GetContentType(f Format) string {
	switch f {
	case FormatYAML:
		return "application/x-yaml"
	case FormatJSON:
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// Unmarshal supports JSON, YAML
func Unmarshal(btes []byte, f Format, i interface{}) error {
	var err error
	switch f {
	case FormatJSON:
		err = json.Unmarshal(btes, i)
	case FormatYAML:
		err = yaml.Unmarshal(btes, i)
	default:
		err = ErrUnsupportedFormat
	}
	return sdk.WithStack(err)
}

// UnmarshalStrict supports JSON, YAML
func UnmarshalStrict(btes []byte, f Format, i interface{}) error {
	var err error
	switch f {
	case FormatJSON:
		err = json.Unmarshal(btes, i)
	case FormatYAML:
		err = yaml.UnmarshalStrict(btes, i)
	default:
		err = ErrUnsupportedFormat
	}
	return sdk.WithStack(err)
}

//Marshal supports JSON, YAML
func Marshal(i interface{}, f Format) ([]byte, error) {
	var btes []byte
	var errMarshal error
	switch f {
	case FormatJSON:
		btes, errMarshal = json.Marshal(i)
	case FormatYAML:
		btes, errMarshal = yaml.Marshal(i)
	}
	return btes, sdk.WithStack(errMarshal)
}

// ReadFile reads the file and return the content, the format and eventually an error
func ReadFile(filename string) ([]byte, Format, error) {
	format := FormatYAML
	if strings.HasSuffix(filename, ".json") {
		format = FormatJSON
	}

	btes, err := ioutil.ReadFile(filename)
	return btes, format, err
}

//OpenFile opens a file
func OpenFile(filename string) (io.ReadCloser, error) {
	r, err := os.Open(filename)
	return r, err
}

// ReadURL reads the file given by an URL
func ReadURL(u string, f string) ([]byte, Format, error) {
	format, err := GetFormatFromPath(f)
	if err != nil {
		return nil, format, err
	}
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	response, err := netClient.Get(u)
	if err != nil {
		return nil, format, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, format, err
	}
	defer response.Body.Close()

	return body, format, nil
}

// OpenURL opens an URL
func OpenURL(u string) (io.ReadCloser, error) {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := netClient.Get(u)
	if err != nil {
		return nil, sdk.WithStack(err)
	}
	return response.Body, err
}

// OpenPath opens an URL or a file
func OpenPath(path string) (io.ReadCloser, Format, error) {
	format, err := GetFormatFromPath(path)
	if err != nil {
		return nil, format, err
	}

	var contentFile io.ReadCloser
	if sdk.IsURL(path) {
		var err error
		contentFile, err = OpenURL(path)
		if err != nil {
			return nil, format, err
		}
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, format, sdk.WithStack(err)
		}
		contentFile = f
	}

	return contentFile, format, nil
}
