package regulation

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

var (
	httpClient HTTPClient = http.DefaultClient
)

const hostName = "regulation.gov.ru"
const baseURL = "https://regulation.gov.ru"
const apiGetCardInfo = "/api/public/PublicProjects/GetCardInfo/%s"
const apiGetProjectStageInfo = "/api/public/PublicProjects/GetProjectStageInfo/%s/Undefined"
const apiGetProjectStages = "/api/public/PublicProjects/GetProjectStages/%s"

// const downloadFileMask = `https://regulation.gov.ru/Files/GetFile?fileid=%s`
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// SetClientHTTP - ...
func SetClientHTTP(c HTTPClient) { httpClient = c }

type File struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	HREF string `json:"href"`
}

// Projects - ...
type Projects struct {
	XMLName       xml.Name      `xml:"projects"`
	Text          string        `xml:",chardata"`
	Total         string        `xml:"total,attr"`
	RegulDraftLaw RegulDraftLaw `xml:"project"`
}

// RegulDraftLaw - ...
type RegulDraftLaw struct {
	ID        int       `xml:"id,attr" json:"id"`          // ID
	Title     string    `xml:"title" json:"title"`         // Наименование
	ProjectID string    `xml:"projectId" json:"projectId"` // ID проекта
	Date      time.Time `xml:"date" json:"date"`           // Дата
	Stage     struct {
		Text string `xml:",chardata" json:"title"` // Название
		ID   int    `xml:"id,attr" json:"id"`      // ID
	} `xml:"stage" json:"stage"` // Этап
	Status struct {
		Text string `xml:",chardata" json:"title"` // Название
		ID   int    `xml:"id,attr" json:"id"`      // ID
	} `xml:"status" json:"status"` // Статус
	Department struct {
		Text string `xml:",chardata" json:"title"` // Название
		ID   int    `xml:"id,attr" json:"id"`      // ID
	} `xml:"department" json:"department"` // Разработчик
	Procedure struct {
		Text string `xml:",chardata" json:"title"` // Название
		ID   int    `xml:"id,attr" json:"id"`      // ID
	} `xml:"procedure" json:"procedure"` // Процедура
	RegulatoryImpact struct {
		Text string `xml:",chardata" json:"title"` // Название
		ID   int    `xml:"id,attr" json:"id"`      // ID
	} `xml:"regulatoryImpact" json:"regulatoryImpact"`
	ProcedureResult struct {
		Text string `xml:",chardata" json:"title"` // Название
		ID   int    `xml:"id,attr" json:"id"`      // ID
	} `xml:"procedureResult" json:"procedureResult"`
	Responsible string `xml:"responsible" json:"responsible"` // Ответственный сотрудник
	// ParallelStageStartDiscussion string `xml:"parallelStageStartDiscussion" json:"parallelStageStartDiscussion"` // Дата начала независимой антикоррупционной экспертизы
	// ParallelStageEndDiscussion   string `xml:"parallelStageEndDiscussion" json:"parallelStageEndDiscussion"`     // Дата окончания независимой антикоррупционной экспертизы
	// StartDiscussion              string `xml:"startDiscussion" json:"startDiscussion"`                           // Дата начала общественного обсуждения
	// EndDiscussion                string `xml:"endDiscussion" json:"endDiscussion"`                               // Дата окончания общественного обсуждения
	// DiscussionDays               int    `xml:"discussionDays" json:"discussionDays"`                             // Длительность общественного обсуждения
	Kind string `xml:"kind" json:"kind"` // Вид
	// RegulatorScissors            bool   `xml:"regulatorScissors" json:"regulatorScissors"`                       // Обязательные требования (Регуляторная гильотина)
	// PublishDate                  string            `xml:"publishDate" json:"publishDate"`
	// NextStageDuration int               `xml:"nextStageDuration" json:"nextStageDuration"`
	Files map[string][]File `xml:"-" json:"files"`
}

// ToJSON - ...
func (rdl *RegulDraftLaw) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	return e.Encode(rdl)
}

// SourceName - хост-источник
func (rdl RegulDraftLaw) SourceName() string { return "regulation.gov.ru" }

// DocumentLabel - наименование
func (rdl RegulDraftLaw) DocumentLabel() string { return rdl.Title }

// DocumentID - ID
func (rdl RegulDraftLaw) DocumentID() string { return strconv.Itoa(rdl.ID) }

// DocumentNumber - номер документа
func (rdl RegulDraftLaw) DocumentNumber() string { return rdl.ProjectID }

// DepartmentName - департамент
func (rdl RegulDraftLaw) DepartmentName() string { return rdl.Department.Text }

// CreatedDate - дата создания
func (rdl RegulDraftLaw) CreatedDate() time.Time {
	// date, err := time.ParseInLocation("2006-01-02T15:04:05", rdl.Date, time.Local)
	// if err != nil {
	// 	return time.Time{}
	// }
	// return date
	return rdl.Date.Local()
}

// KindType - тип
func (rdl RegulDraftLaw) KindType() string { return rdl.Kind }

// CurrentStage - текущий этап
func (rdl RegulDraftLaw) CurrentStage() string { return rdl.Stage.Text }

// CurrentStatus - текущий статус
func (rdl RegulDraftLaw) CurrentStatus() string { return rdl.Status.Text }

func (rdl RegulDraftLaw) GetFiles() map[string][]map[string]string {
	var IFilesMap = make(map[string][]map[string]string, 1)
	for k, vs := range rdl.Files {
		var files = make([]map[string]string, len(vs))
		for i, v := range vs {
			files[i] = map[string]string{
				"id":   v.ID,
				"name": v.Name,
				"href": v.HREF,
			}
		}
		IFilesMap[k] = files
	}
	return IFilesMap
}

// Journal - журнал изменений
func (rdl RegulDraftLaw) Journal() (j []map[string]string) {
	return make([]map[string]string, 0)
}

func FetchFilesText(ctx context.Context, id string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = fmt.Sprintf("/api/public/PublicProjects/GetProjectStageInfo/%s/Text", id)
	req.Header.Add("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 400 {
		return nil, nil
	}
	if res.StatusCode != 200 {
		return res.Body, errors.New("документ не найден")
	}
	return res.Body, nil
}
func FetchFilesProcedure(ctx context.Context, id string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = fmt.Sprintf("/api/public/PublicProjects/GetProjectStageInfo/%s/Procedure", id)
	req.Header.Add("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 400 {
		return nil, nil
	}
	if res.StatusCode != 200 {
		return res.Body, errors.New("документ не найден")
	}
	return res.Body, nil
}

// // Fetch - in dev
// func Fetch(ctx context.Context, id string) (RegulDraftLawV2, error) {
// 	var doc RegulDraftLawV2
// 	{
// 		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
// 		if err != nil {
// 			return doc, err
// 		}
// 		req.URL.Path = fmt.Sprintf(apiGetCardInfo, id)
// 		res, err := httpClient.Do(req)
// 		if err != nil {
// 			return doc, err
// 		}
// 		defer res.Body.Close()
// 		if res.StatusCode != 200 {
// 			return doc, errors.New("документ не найден")
// 		}
// 		if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
// 			return doc, err
// 		}
// 	}
// 	{
// 		var cardInfo CardInfo
// 		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
// 		if err != nil {
// 			return doc, err
// 		}
// 		req.URL.Path = fmt.Sprintf(apiGetProjectStageInfo, id)
// 		res, err := httpClient.Do(req)
// 		if err != nil {
// 			return doc, err
// 		}
// 		defer res.Body.Close()
// 		if res.StatusCode != 200 {
// 			return doc, errors.New("документ не найден")
// 		}
// 		if err := json.NewDecoder(res.Body).Decode(&cardInfo); err != nil {
// 			return doc, err
// 		}
// 		cardInfo.applyToDoc(&doc)
// 	}
// 	return doc, nil
// }

func FetachHTML(ctx context.Context, id string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = "/api/npalist"
	req.URL.RawQuery = (url.Values{"id": {id}}).Encode()
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return res.Body, errors.New("документ не найден")
	}
	return res.Body, err
}

func ParseDocument(rc io.ReadCloser) (*Projects, error) {
	var doc = new(Projects)
	if err := xml.NewDecoder(rc).Decode(doc); err != nil {
		return nil, err
	}
	return doc, rc.Close()
}

var (
	rgx1 = regexp.MustCompile(`(?i)Дополнительные документы .* итог.*`)
	rgx2 = regexp.MustCompile(`(?i)Доработанный .* итог.*текст`)
)

func filterFiles(mapFiles map[string][]File) map[string][]File {
	var filtered = make(map[string][]File, 1)
	for key, value := range mapFiles {
		if rgx1.MatchString(key) || rgx2.MatchString(key) {
			filtered[key] = value
		}
	}
	return filtered
}

// GetDocument - ...
func GetDocument(ctx context.Context, id string) (*RegulDraftLaw, error) {
	rc, err := FetachHTML(ctx, id)
	if err != nil {
		return nil, err
	}
	proj, err := ParseDocument(rc)
	if err != nil {
		return nil, err
	}
	var files = make(map[string][]File, 0)
	{
		rc2, err := FetchFilesText(ctx, id)
		if err != nil {
			return nil, err
		}
		filesText, err := ParseFilesV2(rc2)
		if err != nil {
			return nil, err
		}
		for k, v := range filesText {
			files[k] = append(files[k], v...)
		}
	}
	doc := &proj.RegulDraftLaw
	doc.Files = filterFiles(files)
	return doc, nil
}
