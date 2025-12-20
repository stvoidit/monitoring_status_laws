package regulation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

const badDateTime = "2006-01-02T15:04:05"

type RegulDate time.Time

func (rd *RegulDate) UnmarshalJSON(data []byte) error {
	var cleanText = strings.Trim(string(data), `"`)
	t, err := time.ParseInLocation(badDateTime, cleanText, time.Local)
	*rd = RegulDate(t)
	return err
}
func (rd RegulDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(rd))
}

type FieldAPIDescription string

func (f *FieldAPIDescription) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		token, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		switch v := token.(type) {
		case json.Delim, nil:
			continue
		case string:
			if v == "description" {
				var s string
				if err := dec.Decode(&s); err != nil {
					return err
				}
				*f = FieldAPIDescription(s)
				break
			}
		}
	}
	return nil
}

func (f FieldAPIDescription) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

type FieldAPIValues string

func (f *FieldAPIValues) UnmarshalJSON(data []byte) error {
	var s []string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*f = FieldAPIValues(strings.Join(s, ";"))
	return nil
}

func (f FieldAPIValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

type CardInfo struct {
	Title  string `json:"title"`
	Values []struct {
		Description string         `json:"description"`
		Values      FieldAPIValues `json:"values"`
	} `json:"values"`
}

func (ci CardInfo) applyToDoc(doc *RegulDraftLawV2) {
	for _, v := range ci.Values {
		switch v.Description {
		case "Дата создания":
			t, _ := time.ParseInLocation(badDateTime, string(v.Values), time.Local)
			doc.Date = t
		case "Орган государственной власти":
			doc.Department = string(v.Values)
		case "Сотрудник, ответственный за разработку проекта":
			doc.Responsible = string(v.Values)
		case "Вид":
			doc.Kind = string(v.Values)
		case "Процедура":
			doc.Procedure = FieldAPIDescription(v.Values)
		}
	}
}

type Stage struct {
	Stage             string     `json:"stage"`
	IsCurrent         bool       `json:"isCurrent"`
	Start             *RegulDate `json:"date"`
	Finish            *RegulDate `json:"endDate"`
	ResultDescription string     `json:"resultDescription"`
}

type RegulDraftLawV2 struct {
	ID          uint64              `json:"id"`
	ProjectID   string              `json:"projectId"`
	Title       string              `json:"title"`
	Procedure   FieldAPIDescription `json:"procedure"`
	Date        time.Time           `json:"date"`
	Department  string              `json:"department"`
	Responsible string              `json:"responsible"`
	Kind        string              `json:"kind"`
	Stages      struct {
		StageHistoryElements []Stage `json:"stageHistoryElements"`
		CurrentStages        []Stage `json:"currentStages"`
	} `json:"stages"`
}
