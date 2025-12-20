package cache

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

type CacheRule struct {
	Path string
	TTL  time.Duration
}

func (cr CacheRule) Validate(path string) bool {
	return strings.EqualFold(cr.Path, path)
}

func NewCacheMiddleware(rules ...CacheRule) func(http.Handler) http.Handler {
	var cs = newCacheStore()
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for i := range rules {
				if rules[i].Validate(r.URL.Path) {
					if cr := cs.Get(r.URL.Path); cr != nil {
						if err := cr.WriteResponse(w); err != nil {
							slog.Error("cache.NewCacheMiddleware.WriteResponse",
								slog.String("url", r.URL.Path),
								slog.String("error", err.Error()))
						} else {
							slog.Debug("from cache", slog.String("url", r.URL.Path))
						}
						return
					}
					rec := httptest.NewRecorder()
					h.ServeHTTP(rec, r)
					result := rec.Result()
					res := cacheResponse{
						header:     result.Header.Clone(),
						statusCode: result.StatusCode,
						value:      rec.Body.Bytes(),
						TTL:        rules[i].TTL,
					}
					if result.StatusCode == http.StatusOK {
						cs.Set(r.URL.Path, res)
					}
					if err := res.WriteResponse(w); err != nil {
						slog.Error("cache.WriteResponse", slog.String("error", err.Error()))
					}
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}
