package store

import (
	"fmt"
	"monitoring_draft_laws/internals/lawsparser"
	"time"
)

// User - пользователь
type User struct {
	ID        uint64 `json:"id"`            // ID
	FIO       string `json:"fio,omitempty"` // ФИО полностью
	Shortname string `json:"shortname"`     // ФИО коротко
}

// MegaplanUser - расширенный профиль пользователя
type MegaplanUser struct {
	User                    // встраивание от User
	DepartmentID    uint64  `json:"department_id"`    // ID отдела
	DepartmentLabel string  `json:"department_label"` // Название отдела
	Position        *string `json:"position"`         // Должность
	IsAdmin         bool    `json:"is_admin"`         // Является админом
	IsResponsible   bool    `json:"is_responsible"`   // Является ответственным
}

// JournalRow - запись лога
type JournalRow struct {
	Created time.Time `json:"created"`
	User    User      `json:"user"`
	Changes map[string]struct {
		After  any `json:"after"`
		Before any `json:"before"`
	} `json:"changes"`
}

type AdditionFile struct {
	ID         string         `json:"id"`
	DocumentID string         `json:"document_id"`
	MetaInfo   map[string]any `json:"meta_info"`
	BLOB       []byte         `json:"-"`
}

func (af AdditionFile) ToAPIFormat() map[string]any {
	return map[string]any{
		"id":      af.ID,
		"name":    af.MetaInfo["filename"],
		"size":    af.MetaInfo["size"],
		"headers": af.MetaInfo["headers"],
		"status":  "success",
		"url":     fmt.Sprintf("/api/file?id=%s", af.ID),
	}
}

type AdditionFileList []AdditionFile

func (afs AdditionFileList) ToAPIFormat() []map[string]any {
	var files = make([]map[string]any, len(afs))
	for i, af := range afs {
		files[i] = af.ToAPIFormat()
	}
	return files
}

type StagesFiles map[string][]lawsparser.File

func (sf StagesFiles) Flat() (flat map[string]lawsparser.File) {
	flat = make(map[string]lawsparser.File)
	for _, values := range sf {
		for _, value := range values {
			// flat[value.ID] = value
			flat[value.HREF] = value
		}
	}
	return
}
func (fs StagesFiles) CompareFiles(target StagesFiles) (newFiles []lawsparser.File) {
	newFiles = make([]lawsparser.File, 0)
	flatOrigins := fs.Flat()
	flatTarget := target.Flat()
	for key := range flatTarget {
		if _, ok := flatOrigins[key]; !ok {
			newFiles = append(newFiles, flatTarget[key])
		}
	}
	return newFiles
}
