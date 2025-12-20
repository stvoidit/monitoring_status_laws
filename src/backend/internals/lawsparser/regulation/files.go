package regulation

import (
	"encoding/json"
	"io"
	"net/url"
)

const apiGetFile = "/api/public/Files/GetFile"

type ObjectNodeAPI struct {
	Type        string           `json:"type,omitempty"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Values      []map[string]any `json:"values"`
}

type ObjectFileNode struct {
	FileId      string    `json:"fileId"`
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Varsion     any       `json:"varsion"`
	Date        RegulDate `json:"date"`
}

func (ofn ObjectFileNode) ConvertToFileType() File {
	var href = url.URL{
		Scheme:   "https",
		Host:     hostName,
		Path:     apiGetFile,
		RawQuery: (url.Values{"fileId": {ofn.FileId}}).Encode(),
	}
	return File{
		ID:   ofn.FileId,
		HREF: href.String(),
		Name: ofn.Description,
	}
}

func (ofn *ObjectFileNode) FromAny(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, ofn)
}

func FilesNodesFromAny(v []any) ([]ObjectFileNode, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var files []ObjectFileNode
	if err := json.Unmarshal(b, &files); err != nil {
		return nil, err
	}
	return files, nil
}

func assertingFileNode(value any) (map[string][]ObjectFileNode, error) {
	objMap, ok := value.(map[string]any)
	if !ok {
		return nil, nil
	}
	objType1, ok1 := objMap["type"].(string)
	if !ok1 || objType1 != "File" {
		return nil, nil
	}
	anyFiles, ok2 := objMap["values"].([]any)
	if !ok2 || anyFiles == nil {
		return nil, nil
	}
	var description string
	if desc, ok := objMap["description"].(string); ok {
		description = desc
	}
	files, err := FilesNodesFromAny(anyFiles)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	var m = map[string][]ObjectFileNode{
		description: files,
	}
	return m, nil
}

func ParseFilesV2(rc io.ReadCloser) (map[string][]File, error) {
	var m = make(map[string][]File)
	if rc == nil {
		return m, nil
	}
	var r ObjectNodeAPI
	if err := json.NewDecoder(rc).Decode(&r); err != nil {
		return nil, err
	}
	if err := rc.Close(); err != nil {
		return nil, err
	}
	for _, value := range r.Values {
		filesMap, err := assertingFileNode(value)
		if err != nil {
			return nil, err
		}
		for k, vs := range filesMap {
			for _, v := range vs {
				m[k] = append(m[k], v.ConvertToFileType())
			}
		}
	}
	return m, nil
}
