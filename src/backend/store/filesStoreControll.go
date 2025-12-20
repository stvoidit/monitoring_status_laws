package store

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (db *DB) InsertNewFile(ctx context.Context, rc io.ReadCloser, af AdditionFile) (fileInfo map[string]string, err error) {
	defer rc.Close()
	const q = `
INSERT
    INTO
    monitoring_draft_laws.additional_files
(
        document_id
        , meta_info
        , object_id
    )
VALUES(
    $1
    , $2::jsonb
    , $3
)
RETURNING id`
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	lobs := tx.LargeObjects()
	oid, err := lobs.Create(ctx, 0)
	if err != nil {
		return nil, err
	}

	var fileID string
	err = db.pool.QueryRow(ctx, q, af.DocumentID, af.MetaInfo, oid).Scan(&fileID)
	if err != nil {
		return nil, err
	}

	obj, err := lobs.Open(ctx, oid, pgx.LargeObjectModeWrite)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(obj, rc)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	return map[string]string{
		"id":  fileID,
		"url": fmt.Sprintf("/api/file?id=%s", fileID),
	}, nil
}
func (db *DB) GetDocumentFilesMeta(ctx context.Context, did string) (AdditionFileList, error) {
	const q = `
SELECT
    id
    , document_id
    , meta_info
FROM
    monitoring_draft_laws.additional_files
WHERE
    document_id = $1
ORDER BY created ASC`
	rows, err := db.pool.Query(ctx, q, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files = make([]AdditionFile, 0)
	for rows.Next() {
		var ad AdditionFile
		if err := rows.Scan(&ad.ID, &ad.DocumentID, &ad.MetaInfo); err != nil {
			return nil, err
		}
		files = append(files, ad)
	}
	return files, err
}

func (db *DB) DownloadFile(ctx context.Context, w http.ResponseWriter, fileUUID string) (err error) {
	const q = `
SELECT
	meta_info -> 'headers' AS headers
    , object_id
FROM
    monitoring_draft_laws.additional_files
WHERE
    id = $1::uuid
	`
	var oid uint32
	var h http.Header
	if err := db.pool.QueryRow(ctx, q, fileUUID).Scan(&h, &oid); err != nil {
		return err
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	lobs := tx.LargeObjects()
	obj, err := lobs.Open(ctx, oid, pgx.LargeObjectModeRead)
	if err != nil {
		return err
	}
	for k, v := range h {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, obj)
	return
}
func (db *DB) DeleteFile(ctx context.Context, did string) error {
	const q = `
DELETE FROM monitoring_draft_laws.additional_files
WHERE id=$1::uuid`
	_, err := db.pool.Exec(ctx, q, did)
	return err
}
