package runaihttp

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type CsvFile struct {
	Data     [][]string
	FileName string
}

// PlainTextResponse write an interface to the http response with text/plain header an optional status code
func PlainTextResponse(i interface{}, status *int /*{{#addResponseHeaders}} headers map[string][]string,{{/addResponseHeaders}}*/, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	return TextResponse(i, status /*{{#addResponseHeaders}} headers,{{/addResponseHeaders}}*/, w)
}

// HtmlResponse write an interface to the http response with text/html header an optional status code
func HtmlResponse(i interface{}, status *int /*{{#addResponseHeaders}} headers map[string][]string,{{/addResponseHeaders}}*/, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	return TextResponse(i, status /*{{#addResponseHeaders}} headers,{{/addResponseHeaders}}*/, w)
}

// TextResponse write an interface to the http response with text/html header an optional status code
func TextResponse(i interface{}, status *int /*{{#addResponseHeaders}} headers map[string][]string,{{/addResponseHeaders}}*/, w http.ResponseWriter) error {
	/*
	   {{#addResponseHeaders}}
	   wHeader := w.Header()
	   if headers != nil {
	       for key, values := range headers {
	           for _, value := range values {
	               wHeader.Add(key, value)
	           }
	       }
	   }
	   {{/addResponseHeaders}}
	*/

	var body []byte
	switch data := i.(type) {
	case string:
		body = []byte(data)
	case []byte:
		body = data
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("invalid file type")
	}

	if status != nil {
		w.WriteHeader(*status)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_, err := w.Write(body)
	return err
}

// DownloadFileResponse write an interface to the http response with file content and an optional status code
func DownloadFileResponse(i interface{}, status *int /*{{#addResponseHeaders}} headers map[string][]string,{{/addResponseHeaders}}*/, w http.ResponseWriter) error {
	wHeader := w.Header()
	/*
	   {{#addResponseHeaders}}
	   if headers != nil {
	       for key, values := range headers {
	           for _, value := range values {
	               wHeader.Add(key, value)
	           }
	       }
	   }
	   {{/addResponseHeaders}}
	*/
	filePath, ok := i.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("invalid file type")
	}

	file, err := os.Open(filePath) //nolint:gosec // G304: caller-provided/trusted file path
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return fmt.Errorf("file not found %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error reading binary %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error reading binary %w", err)
	}

	wHeader.Set("Content-Type", http.DetectContentType(data))
	wHeader.Set("Content-Disposition", "attachment; filename="+fileInfo.Name())
	wHeader.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	if status != nil {
		w.WriteHeader(*status)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_, err = w.Write(data)
	return err
}

func CreateCsvResponse(i interface{}, status *int /*{{#addResponseHeaders}} headers map[string][]string,{{/addResponseHeaders}}*/, w http.ResponseWriter) error {
	csvFile, ok := i.(CsvFile)
	if !ok {
		return errors.New("failed creating csv file. return type is not CsvFile")
	}
	data := csvFile.Data
	fileName := csvFile.FileName
	if fileName == "" {
		fileName = "data.csv"
	}
	/*
	   {{#addResponseHeaders}}
	   wHeader := w.Header()
	   if headers != nil {
	       for key, values := range headers {
	           for _, value := range values {
	               wHeader.Add(key, value)
	           }
	       }
	   }
	   wHeader.Set("Content-Type", "text/csv")
	   wHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	   {{/addResponseHeaders}}
	   {{^addResponseHeaders}}
	*/
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	/*
	   {{/addResponseHeaders}}
	*/

	if status != nil {
		w.WriteHeader(*status)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	writer := csv.NewWriter(w)

	err := writer.WriteAll(data)
	return err
}
