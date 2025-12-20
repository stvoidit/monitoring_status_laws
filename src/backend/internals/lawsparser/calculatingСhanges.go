/**

Вычисления изменения в полях через рефлексию - заготовка для унификации в будущем

**/

package lawsparser

import (
	"log"
	"reflect"
	"strings"
	"time"
)

var typeFD = reflect.TypeOf(FormatDocument{})

type FieldValuesDiffs struct {
	Before any `json:"before"`
	After  any `json:"after"`
}

var structFields = [...]string{
	"ActualStatus",
	"ShortLabel",
	"Description",
	"TaxType",
	"ScopeRegulation",
	"Priority",
	"TaskLink",
	"NumberEDO",
	"Note",
	"IsDraft",
	"Archive",
}

func CalculatingСhanges(oldDoc, newDoc FormatDocument) (diffsValues map[string]FieldValuesDiffs) {
	diffsValues = make(map[string]FieldValuesDiffs, len(structFields))
	var (
		oldElem = reflect.ValueOf(oldDoc)
		newElem = reflect.ValueOf(newDoc)
	)
	for _, fieldName := range structFields {
		structField, ok := typeFD.FieldByName(fieldName)
		if !ok {
			log.Println("ERROR reflect field not found:", fieldName)
			continue
		}
		jsonTagName := structField.Tag.Get("json")
		oldFieldValue := oldElem.FieldByName(fieldName)
		newFieldValue := newElem.FieldByName(fieldName)
		switch structField.Type.Kind() {
		case reflect.String:
			if o, n := oldFieldValue.String(), newFieldValue.String(); !strings.EqualFold(n, o) {
				diffsValues[jsonTagName] = FieldValuesDiffs{
					Before: o,
					After:  n,
				}
			}
		case reflect.Uint8:
			if o, n := oldFieldValue.Uint(), newFieldValue.Uint(); o != n {
				diffsValues[jsonTagName] = FieldValuesDiffs{
					Before: o,
					After:  n,
				}
			}
		case reflect.Bool:
			if o, n := oldFieldValue.Bool(), newFieldValue.Bool(); o != n {
				diffsValues[jsonTagName] = FieldValuesDiffs{
					Before: o,
					After:  n,
				}
			}
		case reflect.Ptr:
			oldTime := oldFieldValue.Interface().(*time.Time)
			newTime := newFieldValue.Interface().(*time.Time)
			if oldTime != nil && newTime != nil && oldTime.Equal(*newTime) {
				continue
			} else if oldTime == nil && newTime == nil {
				continue
			}
			diffsValues[jsonTagName] = FieldValuesDiffs{
				Before: oldTime,
				After:  newTime,
			}
		}
	}
	return
}
