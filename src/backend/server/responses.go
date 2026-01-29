package server

import (
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func deferErrLog(err error) {
	if err != nil {
		slog.Error("defer", slog.String("error", err.Error()))
	}
}

const (
	headerContentType        = "Content-Type"
	headerContentDisposition = "Content-Disposition"
	applicationJSON          = "application/jsonl; charset=utf-8"
)

// Jsonify - отправка json response
func Jsonify(w http.ResponseWriter, i any, code int) {
	var err error
	w.Header().Add(headerContentType, applicationJSON)
	w.WriteHeader(code)
	switch I := i.(type) {
	default:
		err = json.NewEncoder(w).Encode(i)
	case string:
		_, err = w.Write([]byte(I))
	case []byte:
		_, err = w.Write(I)
	case io.WriterTo:
		_, err = I.WriteTo(w)
	case error:
		log.Println("ERROR:", I.Error())
		err = json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{Error: I.Error()})
	}
	deferErrLog(err)
}

// JSONLoad - преобразование json в структуру
func JSONLoad[T any](rc io.ReadCloser) (i T, err error) {
	defer func() {
		if err == nil {
			err = rc.Close()
		}
	}()
	err = json.NewDecoder(rc).Decode(&i)
	return
}

// FileSystemSPA - обертка над fs.FileSystem для SPA приложения
func FileSystemSPA(dirname string) http.Handler {
	return http.FileServerFS(spaFS{os.DirFS(dirname)})
}

type spaFS struct {
	hfs fs.FS
}

func (sfs spaFS) Open(name string) (fs.File, error) {
	if _, err := fs.Stat(sfs.hfs, name); os.IsNotExist(err) {
		return sfs.hfs.Open("index.html")
	}
	return sfs.hfs.Open(name)
}
func staticHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		h.ServeHTTP(w, r)
	})
}
