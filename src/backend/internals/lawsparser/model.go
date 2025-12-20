package lawsparser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/url"
	"strings"
	"time"
)

// ParsedDocument - распаршенный документ
type ParsedDocument interface {
	ToJSON(io.Writer) error
	SourceName() string
	DocumentLabel() string
	DocumentID() string
	DocumentNumber() string
	DepartmentName() string
	CreatedDate() time.Time
	KindType() string
	CurrentStage() string
	CurrentStatus() string
	Journal() []map[string]string
	GetFiles() map[string][]map[string]string
}

type File struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	HREF string `json:"href"`
}

func (f File) QueryParams(source string) string {
	return (url.Values{
		"id":     []string{f.ID},
		"source": []string{source},
		"name":   []string{f.Name},
	}).Encode()
}

// FormatDocument - единый формат для разных источников
type FormatDocument struct {
	DocumentID      string              `json:"id"`            // ID
	DocumentNumber  string              `json:"project"`       // номер документа
	Label           string              `json:"label"`         // Наименование
	ShortLabel      string              `json:"short_label"`   //* Краткое наименование
	SourceHost      string              `json:"source"`        // Источник
	Date            time.Time           `json:"date"`          // Дата создания
	Department      string              `json:"department"`    // Разработчик
	KindType        string              `json:"kind"`          // Вид проекта НПА
	ScopeRegulation string              `json:"scope"`         //* Область регулирования
	CurrentStage    string              `json:"current_stage"` // текущий этап
	CurrentStatus   string              `json:"status"`        // текущий статус
	Updated         time.Time           `json:"updated"`       // обновлено
	IsCancelled     bool                `json:"is_cancelled"`  // в статусе отмены
	Description     string              `json:"desc"`          //* Краткое содержание
	Note            string              `json:"note"`          //* Примечания
	TaxType         string              `json:"tax_type"`      //* Вид налога (сбора)
	IsFavorite      bool                `json:"is_favorite"`   //* В избранном пользователя
	TaskLink        string              `json:"task_id"`       //* Ссылка на задачу в ССР
	NumberEDO       string              `json:"number_edo"`    //* Номер в ЭДО
	Priority        uint8               `json:"priority"`      //* Приоритет
	ActualStatus    string              `json:"actual_status"` //* Актуализированный статус
	IsDraft         bool                `json:"is_draft"`      //* Черновик
	IsLaw           bool                `json:"is_law"`        // Закон или Законопроект
	Archive         bool                `json:"archive"`       // проект в архиве
	Journal         []map[string]string `json:"journal"`       // Журнал
	Files           map[string][]File   `json:"files"`         // файлы из документа
}

func (fd FormatDocument) FilesCount() (count int64) {
	for _, v := range fd.Files {
		count += int64(len(v))
	}
	return
}

func (fd FormatDocument) PPrint(w io.Writer) {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	if err := e.Encode(&fd); err != nil {
		panic(err)
	}
}

func (fd FormatDocument) IsNotActual() bool {
	return fd.IsDraft || fd.IsCancelled ||
		strings.HasPrefix(fd.CurrentStage, "8.2") ||
		strings.EqualFold(fd.CurrentStage, "Принятие")
}

// OriginalHREF - оригинальная ссылка на документ
func (fd FormatDocument) OriginalHREF() string {
	switch fd.SourceHost {
	case "regulation.gov.ru":
		return fmt.Sprintf("https://regulation.gov.ru/projects/%s", fd.DocumentID)
	case "sozd.duma.gov.ru":
		return fmt.Sprintf("https://sozd.duma.gov.ru/bill/%s", fd.DocumentID)
	default:
		return fmt.Sprintf("N/A domain: %s", fd.DocumentID)
	}
}

// UpdateDocument - обновить документ
func (fd *FormatDocument) UpdateDocument(ctx context.Context) error {
	doc, err := FetchDocument(ctx, fd.SourceHost, fd.DocumentID)
	if err != nil {
		return err
	}
	var files = make(map[string][]File)
	maps.Copy(files, doc.Files)
	// if doc.FilesCount() >= fd.FilesCount() {
	// 	maps.Copy(files, doc.Files)
	// } else {
	// 	maps.Copy(files, fd.Files)
	// }
	// копирование кастомных полей

	doc.Description,
		doc.Note,
		doc.TaxType,
		doc.ScopeRegulation,
		doc.ShortLabel,
		doc.TaskLink,
		doc.NumberEDO,
		doc.Priority,
		doc.ActualStatus,
		doc.IsDraft,
		doc.IsFavorite,
		doc.Archive =
		fd.Description,
		fd.Note,
		fd.TaxType,
		fd.ScopeRegulation,
		fd.ShortLabel,
		fd.TaskLink,
		fd.NumberEDO,
		fd.Priority,
		fd.ActualStatus,
		fd.IsDraft,
		fd.IsFavorite,
		fd.Archive
	*fd = *doc
	fd.Files = files
	return nil
}

// Только для полей источника
func toFormatDocument(pd ParsedDocument) *FormatDocument {
	var ifilesStages = pd.GetFiles()
	var filesStages = make(map[string][]File, 1)
	for key, values := range ifilesStages {
		files := make([]File, len(values))
		for i, value := range values {
			files[i] = File{
				ID:   value["id"],
				Name: value["name"],
				HREF: value["href"],
			}
		}
		filesStages[key] = files
	}

	return &FormatDocument{
		DocumentID:     pd.DocumentID(),
		DocumentNumber: pd.DocumentNumber(),
		Label:          pd.DocumentLabel(),
		SourceHost:     pd.SourceName(),
		Department:     pd.DepartmentName(),
		Date:           pd.CreatedDate(),
		KindType:       pd.KindType(),
		CurrentStage:   pd.CurrentStage(),
		CurrentStatus:  pd.CurrentStatus(),
		Journal:        pd.Journal(),
		Updated:        time.Now().Local(),
		Files:          filesStages}
}

func (fd FormatDocument) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", fd.DocumentID),
		slog.String("host", fd.SourceHost),
		slog.String("label", fd.Label),
		slog.String("status", fd.CurrentStatus),
		slog.String("stage", fd.CurrentStage),
		slog.Bool("archive", fd.Archive),
		slog.Time("date", fd.Date),
		slog.Time("updated", fd.Updated),
		slog.Any("journal", fd.Journal),
	)
}
