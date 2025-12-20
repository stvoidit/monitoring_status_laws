package sozdduma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	httpClient HTTPClient = http.DefaultClient
)

const baseURL = `https://sozd.duma.gov.ru`

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// SetClientHTTP - ...
func SetClientHTTP(c HTTPClient) { httpClient = c }

// DumaDrftLaw - ...
type DumaDrftLaw struct {
	ProjectID    string            `json:"projectId"`
	Number       string            `json:"number"`
	Label        string            `json:"label"`
	PassportData map[string]string `json:"passportData"`
	Stages       []Stage           `json:"stages"`
}

// ToJSON - ...
func (ddl *DumaDrftLaw) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	return e.Encode(ddl)
}

// SourceName - хост-источник
func (ddl DumaDrftLaw) SourceName() string { return "sozd.duma.gov.ru" }

// DocumentLabel - наименование
func (ddl DumaDrftLaw) DocumentLabel() string { return ddl.Label }

// DocumentID - ID
func (ddl DumaDrftLaw) DocumentID() string { return ddl.ProjectID }

// DocumentNumber - номер документа
func (ddl DumaDrftLaw) DocumentNumber() string { return ddl.ProjectID }

// DepartmentName - департамент
func (ddl DumaDrftLaw) DepartmentName() string {
	return ddl.PassportData["Ответственный комитет"]
}

// CreatedDate - дата создания
func (ddl DumaDrftLaw) CreatedDate() time.Time {
	if len(ddl.Stages) < 1 {
		return time.Time{}
	}
	date, err := time.ParseInLocation("02.01.2006 15:04", ddl.Stages[0].Date, time.Local)
	if err != nil {
		return time.Time{}
	}
	return date
}

// KindType - тип
func (ddl DumaDrftLaw) KindType() string {
	return ddl.PassportData["Форма законопроекта"]
}

// CurrentStage - текущий этап
func (ddl DumaDrftLaw) CurrentStage() string {
	if len(ddl.Stages) < 1 {
		return ""
	}
	return ddl.Stages[len(ddl.Stages)-1].Header
}

// CurrentStatus - текущий статус
func (ddl DumaDrftLaw) CurrentStatus() string {
	if len(ddl.Stages) < 1 {
		return ""
	}
	return ddl.Stages[len(ddl.Stages)-1].Decision
}

// Journal - журнал изменений
func (ddl DumaDrftLaw) Journal() (j []map[string]string) {
	j = make([]map[string]string, 0)
	for _, stage := range ddl.Stages {
		j = append(j, map[string]string{
			"date":     stage.Date,
			"header":   stage.Header,
			"decision": stage.Decision})
	}
	return
}

func filterFiles(files []File) []File {
	var (
		rgx1 = regexp.MustCompile(`(?i)текст.*втор`)
		rgx2 = regexp.MustCompile(`(?i)текст.*треть`)
	)
	var filtered = make([]File, 0, len(files))
	for _, value := range files {
		if rgx1.MatchString(value.Name) || rgx2.MatchString(value.Name) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (ddl DumaDrftLaw) GetFiles() map[string][]map[string]string {
	var IFilesMap = make(map[string][]map[string]string, 1)
	for _, stage := range ddl.Stages {
		filtered := filterFiles(stage.Files)
		if len(filtered) == 0 {
			continue
		}
		var files = make([]map[string]string, len(filtered))
		for i, v := range filtered {
			files[i] = map[string]string{
				"id":   v.ID,
				"name": v.Name,
				"href": v.HREF,
			}
		}
		IFilesMap[stage.Header] = files
	}
	return IFilesMap
}

type File struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	HREF string `json:"href"`
}

// Stage - ..
type Stage struct {
	Header   string `json:"header"`
	Date     string `json:"date"`
	Decision string `json:"decision"`
	Files    []File `json:"files"`
}

func FetachHTML(ctx context.Context, id string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = fmt.Sprintf("/bill/%s", id)
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return res.Body, errors.New("документ не найден")
	}
	return res.Body, err
}

func ParseDocument(rc io.ReadCloser) (*DumaDrftLaw, error) {
	defer rc.Close()
	root, err := goquery.NewDocumentFromReader(rc)
	if err != nil {
		return nil, err
	}
	var doc = new(DumaDrftLaw)
	doc.PassportData = make(map[string]string)
	doc.Number = root.Find("span#number_oz_id").Text()
	doc.Label = root.Find("span#oz_name").Text()
	root.Find("div#opc_hild").Find("tr").Each(func(_ int, tr *goquery.Selection) {
		tr.Find("td").Each(func(n int, td *goquery.Selection) {
			if n == 0 {
				var field, value string
				field = strings.TrimSpace(td.Text())
				if field == "Пакет документов при внесении" {
					value, _ = td.Next().Find("a").Attr("href")
				} else {
					value = strings.TrimSpace(td.Next().Text())
				}
				doc.PassportData[field] = value
			}
		})
	})
	root.Find("div#bh_hron").Find("div.ch-item").Each(func(_ int, stage *goquery.Selection) {
		stageHeader := trim(stage.Find(".ch-item-header").Text())
		stageDate := trim(stage.Find(".hron_date").Text())
		eventDiv := stage.Find("div.ch-item-event")
		eventDiv.Find(".hron_date").Remove()
		eventDiv.Find(".flr_mr").Remove()
		decision := trim(eventDiv.Text())
		doc.Stages = append(doc.Stages, Stage{
			Header:   stageHeader,
			Date:     stageDate,
			Decision: decision,
			Files:    make([]File, 0)})
	})
	var stageFilesMap = make(map[string][]File, 1)
	var eventnum = ""
	root.Find("div#bh_histras").Find(".oz_event").Each(func(_ int, event *goquery.Selection) {
		event.Find("a.a_event_files").Each(func(_ int, stage *goquery.Selection) {
			if num, ok := stage.Parent().Parent().Parent().Attr("data-eventnum"); ok {
				eventnum = num
			}
			fileName := strings.TrimSpace(stage.Find("div.doc_wrap").Text())
			fileHref, _ := stage.Attr("href")
			fileID := fileHref[strings.LastIndex(fileHref, "/")+1:]
			stageFilesMap[eventnum] = append(stageFilesMap[eventnum], File{
				ID:   fileID,
				Name: fileName,
				HREF: fmt.Sprintf(`%s%s`, baseURL, fileHref),
			})
		})
	})
	for i, stage := range doc.Stages {
		for key, values := range stageFilesMap {
			if strings.HasPrefix(stage.Header, key) {
				doc.Stages[i].Files = values
			}
		}
	}
	doc.ProjectID = strings.TrimPrefix(doc.Number, "№ ")
	return doc, nil

}

func trim(s string) string {
	var stringParts = strings.Fields(strings.ReplaceAll(s, "\n", ""))
	for i := range stringParts {
		stringParts[i] = strings.TrimSpace(stringParts[i])
	}
	return strings.Join(stringParts, " ")
}

// GetDocument - ...
func GetDocument(ctx context.Context, src string) (*DumaDrftLaw, error) {
	rc, err := FetachHTML(ctx, src)
	if err != nil {
		return nil, err
	}
	return ParseDocument(rc)
}
