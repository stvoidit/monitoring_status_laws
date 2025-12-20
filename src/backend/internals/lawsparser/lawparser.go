package lawsparser

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"monitoring_draft_laws/internals/lawsparser/regulation"
	"monitoring_draft_laws/internals/lawsparser/sozdduma"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

const userAgent = `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 YaBrowser/25.4.0.0 Safari/537.36`

type innerClient struct {
	c *http.Client
}

func (ic innerClient) Do(req *http.Request) (*http.Response, error) {
	if ug := req.Header.Get("User-Agent"); ug == "" {
		req.Header.Add("User-Agent", userAgent)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	start := time.Now()
	response, err := ic.c.Do(req)
	if err != nil {
		return nil, err
	}
	slog.Debug("response",
		slog.String("status", response.Status),
		slog.String("duration", time.Since(start).String()))
	if err := unzipResponse(response); err != nil {
		return nil, err
	}
	return response, nil
}

var ErrUnknownCompressionMethod = errors.New("unknown compression method")

// unzipResponse - распаковка сжатого ответа
func unzipResponse(response *http.Response) (err error) {
	if response.Uncompressed {
		return nil
	}
	switch response.Header.Get("Content-Encoding") {
	case "":
		return nil
	case "gzip":
		b, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		if err := response.Body.Close(); err != nil {
			return err
		}
		gz, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return err
		}
		response.Body = io.NopCloser(gz)
		response.Header.Del("Content-Encoding")
		response.Uncompressed = true
		return nil
	default:
		return ErrUnknownCompressionMethod
	}
}

func init() {
	jar, _ := cookiejar.New(nil)
	var c = &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			DisableKeepAlives: false,
		},
	}
	var ic = &innerClient{c: c}
	regulation.SetClientHTTP(ic)
	sozdduma.SetClientHTTP(ic)
}

// ErrNotRecognizedSource - ошибка, не распознан источник
var ErrNotRecognizedSource = errors.New("источник не распознан")

// FetchDocument - ...
func FetchDocument(ctx context.Context, host string, id string) (*FormatDocument, error) {
	var pd ParsedDocument
	var err error
	switch {
	case strings.Contains(host, "regulation.gov.ru"):
		pd, err = regulation.GetDocument(ctx, id)
	case strings.Contains(host, "sozd.duma.gov.ru"):
		pd, err = sozdduma.GetDocument(ctx, id)
	default:
		return nil, ErrNotRecognizedSource
	}
	if err != nil {
		return nil, err
	}
	return toFormatDocument(pd), nil
}

func DownloadHTML(ctx context.Context, host, id string) error {
	var err error
	var rc io.ReadCloser
	switch {
	case strings.Contains(host, "regulation.gov.ru"):
		rc, err = regulation.FetachHTML(ctx, id)
	case strings.Contains(host, "sozd.duma.gov.ru"):
		rc, err = sozdduma.FetachHTML(ctx, id)
	default:
		err = ErrNotRecognizedSource
	}
	if err != nil {
		return err
	}
	f, err := os.Create(fmt.Sprintf(`%s.html`, id))
	if err != nil {
		return err
	}
	defer rc.Close()
	if _, err := f.ReadFrom(rc); err != nil {
		return err
	}
	return f.Close()
}

func ParseDocument(host string, rc io.ReadCloser) (doc any, err error) {
	defer rc.Close()
	switch {
	case strings.Contains(host, "regulation.gov.ru"):
		doc, err = regulation.ParseDocument(rc)
	case strings.Contains(host, "sozd.duma.gov.ru"):
		doc, err = sozdduma.ParseDocument(rc)
	default:
		err = ErrNotRecognizedSource
	}
	return
}
