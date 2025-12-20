package lawsparser

import (
	"fmt"
	"net/url"
)

func BuildDownloadURL(source, fileID string) string {
	var downloadHREF = map[string]*url.URL{
		"regulation.gov.ru": {
			Scheme: "https",
			Host:   "regulation.gov.ru",
			Path:   fmt.Sprintf("/api/public/Files/GetFile/%s", fileID)},
		"sozd.duma.gov.ru": {
			Scheme: "https",
			Host:   "sozd.duma.gov.ru",
			Path:   fmt.Sprintf("/download/%s", fileID)},
	}
	sourceURL, ok := downloadHREF[source]
	if !ok {
		return ""
	}
	return sourceURL.String()
}
