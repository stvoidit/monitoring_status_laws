package store

import (
	"context"
	"errors"
	"log/slog"
	"monitoring_draft_laws/internals/lawsparser"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	errDuplicate = `ERROR: duplicate key value violates unique constraint "documents_data_pk" (SQLSTATE 23505)`
	// ErrDuplicateKey - дубликат документа
	ErrDuplicateKey = errors.New(`дубликат документа, уже добавлен в реестр`)
	// ErrInvalideDocument - ошибка парсинга документа или ввода данных нового документа
	ErrInvalideDocument = errors.New(`документ не найден или введены некорректные данные`)
)

// InsertDocument - добавить документ
func (db *DB) InsertDocument(ctx context.Context, fd lawsparser.FormatDocument) error {
	if fd.SourceHost == "regulation.gov.ru" && !fd.IsDraft {
		journal, err := db.checkJournalForRegulation(ctx, &fd)
		if err != nil {
			return err
		}
		fd.Journal = journal
	}
	if fd.IsDraft {
		fd.Journal = make([]map[string]string, 0)
	}
	const q = `
	INSERT INTO monitoring_draft_laws.documents_data
	(projectid, document_values, host, journal)
	VALUES(
		CASE WHEN $1 = '' THEN uuid_generate_v4()::text ELSE $1 END
		, $2::jsonb
		, $3
		, $4::jsonb)`
	_, err := db.pool.Exec(ctx, q, fd.DocumentID, fd, fd.SourceHost, fd.Journal)
	if err != nil && strings.Compare(err.Error(), errDuplicate) == 0 {
		err = ErrDuplicateKey
	}
	return err
}

// UpdateDocument - обновить поля документа
func (db *DB) UpdateDocument(ctx context.Context, fd lawsparser.FormatDocument, userID uint64) error {
	oldDoc, err := db.SelectDocument(ctx, fd.DocumentID, userID)
	if err != nil {
		return err
	}
	diffs := lawsparser.CalculatingСhanges(oldDoc, fd)
	const q = `
	UPDATE monitoring_draft_laws.documents_data
	SET
		document_values=$2::jsonb
		, updated=now()
	WHERE projectid=$1::text`
	if _, err := db.pool.Exec(ctx, q, fd.DocumentID, fd); err != nil {
		return err
	}
	db.writeLogDiffsChanges(ctx, fd.DocumentID, userID, diffs)
	return nil
}

// writeLogDiffsChanges - запись изменений в лог
func (db *DB) writeLogDiffsChanges(ctx context.Context, documentID string, userID uint64, diffs map[string]lawsparser.FieldValuesDiffs) {
	if len(diffs) == 0 {
		return
	}
	const q = `
	INSERT INTO monitoring_draft_laws.logs
	(doc_id, user_id, "values")
	VALUES($1::text, $2::int, $3::jsonb)`
	if _, err := db.pool.Exec(ctx, q, documentID, userID, diffs); err != nil {
		slog.Error("writeLogDiffsChanges", slog.String("error", err.Error()))
	}
}

// checkJournalForRegulation - формирование журнала для regulation.gov.ru
func (db *DB) checkJournalForRegulation(ctx context.Context, fd *lawsparser.FormatDocument) (j []map[string]string, err error) {
	const q = `SELECT dd.journal FROM monitoring_draft_laws.documents_data dd WHERE dd.projectid = $1 LIMIT 1`
	j = make([]map[string]string, 0)
	err = db.pool.QueryRow(ctx, q, fd.DocumentID).Scan(&j)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return j, err
	}
	if len(j) == 0 {
		j = append(j, map[string]string{
			"date":     time.Now().Format("02.01.2006"),
			"header":   fd.CurrentStage,
			"decision": fd.CurrentStatus})
		return j, nil
	}
	var lastRow = j[len(j)-1]
	if lastRow["header"] != fd.CurrentStage || lastRow["decision"] != fd.CurrentStatus {
		j = append(j, map[string]string{
			"date":     time.Now().Format("02.01.2006"),
			"header":   fd.CurrentStage,
			"decision": fd.CurrentStatus})
	}
	return j, nil
}

func (db *DB) CheckChanged(ctx context.Context, projectID string) (
	count int64,
	status string,
	files StagesFiles,
	err error) {
	files = make(StagesFiles)
	const q = `
	WITH unnest_files AS (
		SELECT
			dd.projectid
			, jrec."key" AS "label"
			, jsonb_agg(file) AS files
		FROM
			monitoring_draft_laws.documents_data AS dd
		LEFT JOIN LATERAL jsonb_each(dd.document_values -> 'files') AS jrec ON
				TRUE
		LEFT JOIN LATERAL jsonb_array_elements(jrec.value) AS file ON
				TRUE
		WHERE
			file ->> 'name' ~* 'текст.*втор'
			OR
			file->> 'name' ~* 'текст.*треть'
			OR
			dd.host = 'regulation.gov.ru'
			-- OR
			-- jrec."key" ~* 'дополнительные документы .* тексту'
			-- OR
			-- jrec."key" ~* 'доработанный по итогам .* текст.*'
		GROUP BY
			dd.projectid
			, "label"
	)
	SELECT
		dd.projectid
		, dd.document_values ->> 'status' AS status
		, sum( jsonb_array_length(COALESCE(uf.files, '[]'::jsonb))) AS files_count
		, jsonb_object_agg(COALESCE(uf."label", ''), COALESCE(uf.files, '[]'::jsonb)) AS stages_files
	FROM
		monitoring_draft_laws.documents_data AS dd
	LEFT JOIN unnest_files AS uf ON
		uf.projectid = dd.projectid
	WHERE
		dd.projectid = @projectID::text
	GROUP BY
		dd.projectid
		, "status"
	LIMIT 1`

	err = db.pool.QueryRow(ctx, q,
		pgx.NamedArgs{"projectID": projectID}).
		Scan(&projectID, &status, &count, &files)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	return
}

// UpdateDocumenScheduler - обновление через планировщик, установка времени обновления
func (db *DB) UpdateDocumenScheduler(ctx context.Context, fd lawsparser.FormatDocument) (
	_fd *lawsparser.FormatDocument,
	newStatus string,
	newFiles []lawsparser.File,
	err error) {
	if fd.SourceHost == "regulation.gov.ru" {
		fd.Journal, err = db.checkJournalForRegulation(ctx, &fd)
		if err != nil {
			return &fd, "", nil, err
		}
	}

	var oldStatus string
	var files StagesFiles
	_, oldStatus, files, err = db.CheckChanged(ctx, fd.DocumentID)
	if err != nil {
		return nil, "", nil, err
	}
	var statusIsChanged = oldStatus != fd.CurrentStatus

	fd.Updated = time.Now().Local()
	const q = `
	UPDATE monitoring_draft_laws.documents_data
	SET document_values=$2::jsonb, journal=$3::jsonb, updated=$4
	WHERE projectid=$1`
	if _, err := db.pool.Exec(ctx, q, fd.DocumentID, fd, fd.Journal, fd.Updated); err != nil {
		return nil, "", nil, err
	}
	newFiles = files.CompareFiles(fd.Files)
	// fmt.Printf("%v\n%q\n%q\n\n%+v\n\n", statusIsChanged, oldStatus, fd.CurrentStatus, newFiles)
	if statusIsChanged || len(newFiles) > 0 {
		if _fd, err := db.SelectDocument(ctx, fd.DocumentID, 0); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, "", nil, err
		} else {
			return &_fd, _fd.CurrentStatus, newFiles, nil
		}
	}
	return nil, "", nil, nil
}

// SelectDocuments - выборка документов
func (db *DB) SelectDocuments(ctx context.Context, userID uint64) ([]lawsparser.FormatDocument, error) {
	const q = `
	SELECT
		dd.document_values || jsonb_build_object(
			'journal'
			, dd.journal
			, 'id'
			, dd.projectid
			, 'archive'
			, dd.archive
			, 'updated'
			, dd.updated
		)
		, COALESCE(
			(
				dd.document_values ->> 'status'
			) = ANY(cp.cancellation_phrases)
			, FALSE
		) AS is_cancelled
		, fav.user_id IS NOT NULL AS is_favorite
		, COALESCE(
			sil.is_law
			, FALSE
		) AS is_law
	FROM
		monitoring_draft_laws.documents_data dd
	LEFT JOIN monitoring_draft_laws.cancellation_policy cp ON
		cp.host = dd.host
	LEFT JOIN monitoring_draft_laws.status_is_law AS sil ON
		sil.projectid = dd.projectid
	LEFT JOIN monitoring_draft_laws.favorites AS fav ON
		fav.project_id = dd.projectid
		AND fav.user_id = @userID
	ORDER BY
		dd.created DESC`
	rows, err := db.pool.Query(ctx, q, pgx.NamedArgs{"userID": userID})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make([]lawsparser.FormatDocument, 0, 50)
	for rows.Next() {
		var fd lawsparser.FormatDocument
		if err := rows.Scan(
			&fd,
			&fd.IsCancelled,
			&fd.IsFavorite,
			&fd.IsLaw); err != nil {
			return nil, err
		}
		data = append(data, fd)
	}
	return data, rows.Err()
}

// SelectDocument - выбор документа
func (db *DB) SelectDocument(ctx context.Context, documentID string, userID uint64) (fd lawsparser.FormatDocument, err error) {
	const q = `
	SELECT
		dd.document_values || jsonb_build_object('journal', dd.journal, 'id', dd.projectid, 'archive', dd.archive, 'updated', dd.updated)
		, COALESCE((dd.document_values ->> 'status') = ANY(cp.cancellation_phrases), false) is_cancelled
		, fav.is_favorite
		, COALESCE(sil.is_law, false) AS is_law
	FROM
		monitoring_draft_laws.documents_data dd
	LEFT JOIN monitoring_draft_laws.cancellation_policy cp ON
		cp.host = dd.host
	LEFT JOIN monitoring_draft_laws.status_is_law AS sil ON
		sil.projectid = dd.projectid
	INNER JOIN LATERAL (
		SELECT
			EXISTS (
				SELECT
					1
				FROM
					monitoring_draft_laws.favorites AS f
				WHERE
					f.project_id = dd.projectid
					AND f.user_id = $2::int4
			) AS is_favorite
	) AS fav ON
	TRUE
	WHERE dd.projectid = $1
	LIMIT 1`
	err = db.pool.QueryRow(ctx, q, documentID, userID).Scan(
		&fd,
		&fd.IsCancelled,
		&fd.IsFavorite,
		&fd.IsLaw)
	return
}

// DeleteDocument - удалить документ
func (db *DB) DeleteDocument(ctx context.Context, documentID string) error {
	const q = `DELETE FROM monitoring_draft_laws.documents_data WHERE projectid=$1`
	_, err := db.pool.Exec(ctx, q, documentID)
	return err
}

// PatchDraftDocument - установка оригинальной ссылки для черновика дкоумента
func (db *DB) PatchDraftDocument(ctx context.Context, tmpID string, fd lawsparser.FormatDocument, userID uint64) error {
	if fd.SourceHost == "regulation.gov.ru" {
		journal, err := db.checkJournalForRegulation(ctx, &fd)
		if err != nil {
			return err
		}
		fd.Journal = journal
	}
	fd.IsDraft = false
	oldDoc, err := db.SelectDocument(ctx, tmpID, userID)
	if err != nil {
		return err
	}
	diffs := lawsparser.CalculatingСhanges(oldDoc, fd)
	const q = `
	UPDATE
		monitoring_draft_laws.documents_data
	SET
		projectid = $2::text
		, document_values = $4::jsonb
		, host = $3::text
		, updated = now()
		, journal = $5::jsonb
	WHERE
		projectid = $1::text`
	if _, err := db.pool.Exec(ctx, q, tmpID, fd.DocumentID, fd.SourceHost, fd, fd.Journal); err != nil {
		return err
	}
	db.writeLogDiffsChanges(ctx, fd.DocumentID, userID, diffs)
	return nil
}

// SelectUsers - все пользователи из мегаплана
func (db *DB) SelectUsers(ctx context.Context) ([]MegaplanUser, error) {
	const q = `
	SELECT
		up.userid
		, up.shortname
		, up.fio
		, up.depid
		, up.department
		, up."position"
		, COALESCE(rs.is_admin, FALSE)
		, COALESCE(rs.is_responsible, FALSE)
	FROM
		public.user_profiles up
	LEFT JOIN monitoring_draft_laws.roles_settings rs ON
		rs.user_id = up.userid`
	rows, err := db.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users = make([]MegaplanUser, 0)
	for rows.Next() {
		var u MegaplanUser
		if err := rows.Scan(
			&u.ID,
			&u.Shortname,
			&u.FIO,
			&u.DepartmentID,
			&u.DepartmentLabel,
			&u.Position,
			&u.IsAdmin,
			&u.IsResponsible); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ChangeRole - изменение роли пользователя
func (db *DB) ChangeRole(ctx context.Context, mu MegaplanUser) (err error) {
	var exists bool
	if err := db.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM monitoring_draft_laws.roles_settings rs WHERE rs.user_id = $1)`, mu.ID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		_, err = db.pool.Exec(ctx, `UPDATE monitoring_draft_laws.roles_settings SET is_admin=$2, is_responsible=$3 WHERE user_id=$1`, mu.ID, mu.IsAdmin, mu.IsResponsible)
	} else {
		_, err = db.pool.Exec(ctx, `INSERT INTO monitoring_draft_laws.roles_settings (user_id, is_admin, is_responsible) VALUES($1, $2, $3)`, mu.ID, mu.IsAdmin, mu.IsResponsible)
	}
	return err
}

// GetUserRoles - роли пользователя
func (db *DB) GetUserRoles(ctx context.Context, userID uint64) (isAdmin, isResponsible bool) {
	const q = `SELECT rs.is_admin, rs.is_responsible FROM monitoring_draft_laws.roles_settings rs WHERE rs.user_id = $1::int LIMIT 1`
	if err := db.pool.QueryRow(ctx, q, userID).Scan(&isAdmin, &isResponsible); err != nil {
		return false, false
	}
	return
}

// GetNotificationType - получить тип уведомления
func (db *DB) GetNotificationType(ctx context.Context, userID uint64) (ntype uint8) {
	const q = `SELECT tc.notification_type FROM monitoring_draft_laws.tg_chats AS tc WHERE tc.user_id = $1::int LIMIT 1`
	if err := db.pool.QueryRow(ctx, q, userID).Scan(&ntype); err != nil {
		return 0
	}
	return
}

func (db *DB) DeleteChatID(ctx context.Context, chatID int64) (err error) {
	const delete = `DELETE FROM monitoring_draft_laws.tg_chats WHERE chat_id=$1::int8`
	_, err = db.pool.Exec(ctx, delete, chatID)
	return
}

// SaveChatID - сохранить чат для уведомлений
func (db *DB) SaveChatID(ctx context.Context, chatID int64, userID uint64, enabled bool) (err error) {
	const insert = `
	INSERT INTO monitoring_draft_laws.tg_chats (chat_id, user_id, enabled)
	VALUES($1::int8, $2::int4, $3::bool)
	ON CONFLICT (chat_id)
	DO UPDATE SET
		enabled = EXCLUDED.enabled,
		user_id = EXCLUDED.user_id`
	if enabled {
		_, err = db.pool.Exec(ctx, insert, chatID, userID, enabled)
	} else {
		err = db.DeleteChatID(ctx, chatID)
	}
	return
}

// SelectChatsTG - получить список ID чатов для нотификации
func (db *DB) SelectChatsTG(ctx context.Context, documentID string) (chatsID []int64, err error) {
	chatsID = make([]int64, 0)
	const q = `
	WITH subq AS (
		SELECT
			tc.chat_id
		FROM
			monitoring_draft_laws.tg_chats AS tc
		WHERE
			tc.notification_type = 0
	UNION
		SELECT
			tc.chat_id
		FROM
			monitoring_draft_laws.tg_chats AS tc
		WHERE
			tc.notification_type = 1
			AND EXISTS (
				SELECT
					1
				FROM
					monitoring_draft_laws.favorites AS f
				WHERE
					f.project_id = $1::TEXT
					AND
				f.user_id = tc.user_id
			)
	UNION
		SELECT
			tc.chat_id
		FROM
			monitoring_draft_laws.tg_chats AS tc
		WHERE
			tc.notification_type = 2
			AND EXISTS (
				SELECT
					1
				FROM
					monitoring_draft_laws.documents_data AS dd
				WHERE
					dd.projectid = $2::TEXT
					AND (
						dd.document_values ->> 'priority'
					)::int <= 2
			)
	)
	SELECT COALESCE(array_agg(subq.chat_id), '{}'::int8[]) FROM subq`
	err = db.pool.QueryRow(ctx, q, documentID, documentID).Scan(&chatsID)
	return
}

// ToggleFavorite - метка избранных
func (db *DB) ToggleFavorite(ctx context.Context, projectID string, userID uint64, isFavorite bool) error {
	if isFavorite {
		if _, err := db.pool.Exec(ctx,
			`INSERT INTO monitoring_draft_laws.favorites (user_id, project_id) VALUES($1::int4, $2::text) ON CONFLICT (user_id, project_id) DO NOTHING`,
			userID,
			projectID); err != nil {
			return err
		}
	} else {
		if _, err := db.pool.Exec(ctx,
			`DELETE FROM monitoring_draft_laws.favorites WHERE user_id=$1::int4 AND project_id=$2::text`,
			userID,
			projectID); err != nil {
			return err
		}
	}
	return nil
}

// CheckUserForBot - проверка существования пользователя в мегаплане
func (db *DB) CheckUserForBot(ctx context.Context, userID uint64) (username string, err error) {
	const q = `SELECT shortname FROM public.user_profiles WHERE userid = $1::int LIMIT 1`
	err = db.pool.QueryRow(ctx, q, userID).Scan(&username)
	return
}

// ChangeNType - Изменить тип уведомлений
func (db *DB) ChangeNType(ctx context.Context, userID uint64, ntype uint8) error {
	const q = `
	UPDATE monitoring_draft_laws.tg_chats
	SET notification_type=$2::int
	WHERE user_id=$1::int`
	cmd, err := db.pool.Exec(ctx, q, userID, ntype)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("Ваш чат для уведомлений не найден. Пожалуйста, активируйте бота по ссылке на главной странице.")
	}
	return nil
}

// SelectJournal - журнал изменений
func (db *DB) SelectJournal(ctx context.Context, documentID string) ([]JournalRow, error) {
	const q = `
	SELECT
		l.created
		, l.user_id
		, COALESCE(concat_ws(' ', u."LastName", u."FirstName", u."MiddleName") , 'user_id:'||l.user_id::text)
		, l."values"
	FROM
		monitoring_draft_laws.logs AS l
	LEFT JOIN public.users AS u ON
		u.id = l.user_id
	WHERE
		l.doc_id = $1::text
	ORDER BY
		l.created ASC`
	rows, err := db.pool.Query(ctx, q, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var data = make([]JournalRow, 0)
	for rows.Next() {
		var j JournalRow
		if err := rows.Scan(
			&j.Created,
			&j.User.ID,
			&j.User.Shortname,
			&j.Changes,
		); err != nil {
			return nil, err
		}
		data = append(data, j)
	}
	return data, nil
}

func (db *DB) ChangeArchiveStatus(ctx context.Context, did string, userID uint64) error {
	oldDoc, err := db.SelectDocument(ctx, did, userID)
	if err != nil {
		return err
	}
	const q = `
	UPDATE
		monitoring_draft_laws.documents_data
	SET
		archive = NOT archive
	WHERE
		projectid = $1`
	if _, err := db.pool.Exec(ctx, q, did); err != nil {
		return err
	}
	currentDocument, err := db.SelectDocument(ctx, did, userID)
	if err != nil {
		return err
	}
	db.writeLogDiffsChanges(ctx, did, userID, lawsparser.CalculatingСhanges(oldDoc, currentDocument))
	return nil
}
