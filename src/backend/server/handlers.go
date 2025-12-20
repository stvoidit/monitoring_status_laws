package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"monitoring_draft_laws/internals/lawsparser"
	"monitoring_draft_laws/store"
	"monitoring_draft_laws/utils"
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

// InitHandler - точка входа для SPA - авторизация, получение конфигурации с бэка
func (app *Application) InitHandler(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Set("Cache-Control", "no-store, private")
	type applicationInfo struct {
		Domain        string       `json:"megaplanDomain"`
		UUID          string       `json:"appUUID"`
		IsAdmin       bool         `json:"isAdmin"`
		IsResponsible bool         `json:"isResponsible"`
		TgBotName     string       `json:"tg_bot_name"`
		NType         uint8        `json:"ntype"` // тип уведомлений
		Token         any          `json:"token,omitempty"`
		User          *UserProfile `json:"user"`
	}
	var ctx = r.Context()
	profile, err := app.readContextValue(ctx)
	if err != nil || profile == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	isadmin, isresponsible := app.db.GetUserRoles(ctx, profile.UserID)
	isadmin = true //! just local dev
	Jsonify(w, applicationInfo{
		Domain:        app.config.Megaplan.Domain,
		UUID:          app.config.Megaplan.UUID,
		IsAdmin:       isadmin,
		IsResponsible: isresponsible,
		TgBotName:     app.config.TGBOT.Name,
		NType:         app.db.GetNotificationType(ctx, profile.UserID),
		User:          profile},
		http.StatusOK)
}

// AddDocument - добавление документа
func (app *Application) AddDocument(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	fd, err := JSONLoad[lawsparser.FormatDocument](r.Body)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	if !fd.IsDraft {
		if err := fd.UpdateDocument(ctx); err != nil {
			Jsonify(w, err, http.StatusBadRequest)
			return
		}
	} else {
		fd.SourceHost = "draft"
		fd.DocumentID = ""
	}
	if err := app.db.InsertDocument(ctx, fd); err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	Jsonify(w, fd, http.StatusCreated)
}

// GetDocuments - получить документы
func (app *Application) GetDocuments(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	user, err := app.readContextValue(ctx)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	data, err := app.db.SelectDocuments(ctx, user.UserID)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
	} else {
		Jsonify(w, &data, http.StatusOK)
	}
}

// GetDocument - получить документ
func (app *Application) GetDocument(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	user, err := app.readContextValue(ctx)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	var documentID = r.URL.Query().Get("id")
	if doc, err := app.db.SelectDocument(ctx, documentID, user.UserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			Jsonify(w, err, http.StatusNotFound)
		} else {
			Jsonify(w, err, http.StatusInternalServerError)
		}
	} else {
		Jsonify(w, &doc, http.StatusOK)
	}
}

// DeleteDocument - удалить документ
func (app *Application) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	var documentID = r.URL.Query().Get("id")
	if err := app.db.DeleteDocument(r.Context(), documentID); err != nil {
		Jsonify(w, err, http.StatusBadGateway)
	} else {
		Jsonify(w, "OK", http.StatusOK)
	}
}

// UpdateDocument - обновить поля документа с формы редактирования
func (app *Application) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	user, err := app.readContextValue(ctx)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	fd, err := JSONLoad[lawsparser.FormatDocument](r.Body)
	if err != nil {
		Jsonify(w, err, http.StatusBadGateway)
		return
	}
	if err := app.db.UpdateDocument(ctx, fd, user.UserID); err != nil {
		Jsonify(w, err, http.StatusBadGateway)
		return
	}
	if fd, err = app.db.SelectDocument(ctx, fd.DocumentID, user.UserID); err != nil {
		Jsonify(w, err, http.StatusBadGateway)
	} else {
		Jsonify(w, &fd, http.StatusCreated)
	}
}

// PatchDraftDocument - установить для черновика оригинальный ID документа
func (app *Application) PatchDraftDocument(w http.ResponseWriter, r *http.Request) {
	type PatchRequest struct {
		TMPID  string `json:"tmpID"`
		ID     string `json:"id"`
		Source string `json:"source"`
	}
	var ctx = r.Context()
	user, err := app.readContextValue(ctx)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	pr, err := JSONLoad[PatchRequest](r.Body)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	fd, err := app.db.SelectDocument(ctx, pr.TMPID, user.UserID)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	fd.SourceHost = pr.Source
	fd.DocumentID = pr.ID
	if err := fd.UpdateDocument(ctx); err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	if err := app.db.PatchDraftDocument(ctx, pr.TMPID, fd, user.UserID); err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	Jsonify(w, fd.DocumentID, http.StatusCreated)
}

// DownloadDocuments - скачать документы
func (app *Application) DownloadDocuments(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Headers []string   `json:"headers"`
		Data    [][]string `json:"data"`
	}
	xlsxReq, err := JSONLoad[body](r.Body)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	var f = utils.CreateXLSX(xlsxReq.Headers, xlsxReq.Data)
	defer deferErrLog(f.Close())
	w.WriteHeader(http.StatusOK)
	_, err = f.WriteTo(w)
	deferErrLog(err)
}

// FetchUpdate - обновить документ из источника
func (app *Application) FetchUpdate(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	fd, err := JSONLoad[lawsparser.FormatDocument](r.Body)
	if err != nil {
		Jsonify(w, err, 513)
		return
	}
	doc, err := app.db.SelectDocument(ctx, fd.DocumentID, 0)
	if err != nil {
		Jsonify(w, err, 513)
		return
	}
	if err := doc.UpdateDocument(ctx); err != nil {
		Jsonify(w, err, 513)
		return
	}
	_fd, newStatus, newFiles, err := app.db.UpdateDocumenScheduler(ctx, doc)
	if err != nil {
		Jsonify(w, err, 513)
		return
	}
	if _fd != nil && (len(newFiles) > 0 || newStatus != "") {
		// нотификация в ТГ
		if chatsID, err := app.db.SelectChatsTG(ctx, _fd.DocumentID); err == nil {
			deferErrLog(app.Notify(ctx, _fd, chatsID, newStatus, newFiles))
			slog.Info("NOTIFY:", slog.String("DocumentID", _fd.DocumentID), slog.Any("chatsID", chatsID))
		} else {
			slog.Error("ERROR: notify", slog.String("error", err.Error()))
		}
	}
	user, err := app.readContextValue(ctx)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	if doc, err := app.db.SelectDocument(ctx, fd.DocumentID, user.UserID); err != nil {
		Jsonify(w, err, 513)
		return
	} else {
		Jsonify(w, doc, http.StatusCreated)
	}
}

// FetchUpdateAll - обновить документы из источника
func (app *Application) FetchUpdateAll() http.HandlerFunc {
	var onupdate = false
	return func(w http.ResponseWriter, r *http.Request) {
		if onupdate {
			Jsonify(w,
				map[string]string{"message": "Документы уже в процессе обновления"},
				http.StatusOK)
			return
		}
		go func() {
			onupdate = true
			defer func() {
				onupdate = false
				slog.Info("toggle FetchUpdateAll", slog.Bool("onupdate", onupdate))
			}()
			app.defaultJobFunc(context.Background())
		}()
		Jsonify(w,
			map[string]string{"message": "Запущен процесс обновления документов"},
			http.StatusCreated)
	}
}

// FavoriteHandle - toggle для метки избранное
func (app *Application) FavoriteHandle(w http.ResponseWriter, r *http.Request) {
	type requestFavorite struct {
		ProjectID  string `json:"project_id"`
		IsFavorite bool   `json:"is_favorite"`
	}
	var ctx = r.Context()
	profile, err := app.readContextValue(ctx)
	if err != nil || profile == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rf, err := JSONLoad[requestFavorite](r.Body)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	if err := app.db.ToggleFavorite(ctx, rf.ProjectID, profile.UserID, rf.IsFavorite); err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	Jsonify(w, "ok", http.StatusOK)
}

// RolesSettings - управление доступом
func (app *Application) RolesSettings(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	switch r.Method {
	case http.MethodGet:
		if users, err := app.db.SelectUsers(ctx); err != nil {
			Jsonify(w, err, http.StatusBadRequest)
		} else {
			Jsonify(w, &users, http.StatusOK)
		}
	case http.MethodPost:
		mu, err := JSONLoad[store.MegaplanUser](r.Body)
		if err != nil {
			Jsonify(w, err, http.StatusBadRequest)
			return
		}
		if err := app.db.ChangeRole(ctx, mu); err != nil {
			Jsonify(w, err, http.StatusBadRequest)
		} else {
			Jsonify(w, nil, http.StatusCreated)
		}
	}
}

// // TelegramBotWebHook - хэндлер обработки вебхука для телеграм бота
// func (app *Application) TelegramBotWebHook(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	switch r.Method {
// 	case http.MethodPost:
// 		update, err := app.tgbot.HandleUpdate(r)
// 		if err != nil {
// 			log.Println("ERROR TelegramBotWebHook.HandleUpdate:", err)
// 			Jsonify(w, err, http.StatusInternalServerError)
// 			return
// 		}

// 		switch {
// 		case update.Message != nil:
// 			log.Println("TelegramBotWebHook:", update.Message.Chat.ID, update.Message.Command(), update.Message.Text)
// 			switch update.Message.Command() {
// 			case "start":
// 				decodeString, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.Replace(update.Message.Text, "/start", "", 1)))
// 				log.Println("DecodeString:", string(decodeString))
// 				userID, err := strconv.ParseUint(string(decodeString), 10, 64)
// 				if err != nil {
// 					log.Println("ERROR ParseUint.UserID.TgBot", update.Message.Chat.ID, err)
// 					app.sendBotMessage(update.Message.Chat.ID, "Извините, ваш ID не распознан. Пожалуйста, воспользуйтесь кнопкой телеграм бота на сайте")
// 					break
// 				}
// 				username, err := app.db.CheckUserForBot(ctx, userID)
// 				if err != nil {
// 					log.Println("ERROR CheckUserForBot", update.Message.Chat.ID, userID, err)
// 					app.sendBotMessage(update.Message.Chat.ID, "Извините, ваш ID не распознан. Пожалуйста, воспользуйтесь кнопкой телеграм бота на сайте")
// 					break
// 				}
// 				if err := app.db.SaveChatID(ctx, update.Message.Chat.ID, userID, true); err == nil {
// 					app.sendBotMessage(update.Message.Chat.ID, fmt.Sprintf("Здравствуйте, %s\nУведомления по изменениям статуса документов включены.", username))
// 				} else {
// 					log.Println("ERROR SaveChatID:", update.Message.Chat.ID, err)
// 				}
// 			case "stop":
// 				if err := app.db.SaveChatID(ctx, update.Message.Chat.ID, 0, false); err != nil {
// 					log.Println("ERROR SaveChatID:", err)
// 				}
// 			}
// 			log.Println("command:", update.Message.Command())
// 		case update.MyChatMember != nil:
// 			if update.MyChatMember.NewChatMember.Status == "kicked" {
// 				if err := app.db.SaveChatID(ctx, update.MyChatMember.Chat.ID, 0, false); err != nil {
// 					log.Println("ERROR SaveChatID:", update.Message.Chat.ID, err)
// 				}
// 			}
// 		}
// 	default:
// 		Jsonify(w, map[string]string{"now": time.Now().GoString()}, http.StatusOK)
// 	}
// }

// ChangeNType - изменить тип уведомлений бота
func (app *Application) ChangeNType(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	profile, err := app.readContextValue(ctx)
	if err != nil || profile == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	type reqNType struct {
		NType uint8 `json:"ntype"`
	}
	rnt, err := JSONLoad[reqNType](r.Body)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	if err := app.db.ChangeNType(ctx, profile.UserID, rnt.NType); err != nil {
		Jsonify(w, err, http.StatusBadRequest)
	} else {
		Jsonify(w, "ok", http.StatusCreated)
	}
}

// GetJournal - журнал изменений
func (app *Application) GetJournal(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	profile, err := app.readContextValue(ctx)
	if err != nil || profile == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if data, err := app.db.SelectJournal(ctx, r.URL.Query().Get("id")); err != nil {
		Jsonify(w, err, http.StatusBadRequest)
	} else {
		Jsonify(w, data, http.StatusOK)
	}
}

func (app *Application) DocumentFiles(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	profile, err := app.readContextValue(ctx)
	if err != nil || profile == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		var did = r.URL.Query().Get("id")
		if err := app.db.DeleteFile(ctx, did); err != nil {
			Jsonify(w, err, http.StatusInternalServerError)
			return
		}
		Jsonify(w, "ok", http.StatusOK)
	case http.MethodPost:
		if err := r.ParseMultipartForm(1048576); err != nil {
			Jsonify(w, err, http.StatusInternalServerError)
			return
		}
		documentID := r.PostForm.Get("documentID")
		f, fh, err := r.FormFile("file")
		if err != nil {
			Jsonify(w, err, http.StatusInternalServerError)
			return
		}

		metaInfo := map[string]any{
			"uploader": profile,
			"time":     time.Now().String(),
			"filename": fh.Filename,
			"size":     fh.Size,
			"headers":  fh.Header,
		}
		uf := store.AdditionFile{
			DocumentID: documentID,
			MetaInfo:   metaInfo,
		}
		if fileInfo, err := app.db.InsertNewFile(ctx, f, uf); err != nil {
			Jsonify(w, err, http.StatusInternalServerError)
		} else {
			Jsonify(w, fileInfo, http.StatusCreated)
		}
	case http.MethodGet:
		var documentID = r.URL.Query().Get("id")
		switch r.URL.Path {
		case "/api/files":
			files, err := app.db.GetDocumentFilesMeta(ctx, documentID)
			if err != nil {
				Jsonify(w, err, http.StatusBadRequest)
				return
			}
			Jsonify(w, files.ToAPIFormat(), http.StatusOK)
		case "/api/file":
			var did = r.URL.Query().Get("id")
			if err := app.db.DownloadFile(ctx, w, did); err != nil {
				slog.Error("download file", slog.String("error", err.Error()))
			}
		}
	}
}

func (app *Application) ArchiveStatus(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	profile, err := app.readContextValue(ctx)
	if err != nil || profile == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	type body struct {
		ID string `json:"id"`
	}
	requestBody, err := JSONLoad[body](r.Body)
	if err != nil {
		Jsonify(w, err, 513)
		return
	}
	if err := app.db.ChangeArchiveStatus(ctx, requestBody.ID, profile.UserID); err != nil {
		Jsonify(w, err, 513)
	} else {
		Jsonify(w, "ok", http.StatusCreated)
	}
}

func (app *Application) ProxyDownload(w http.ResponseWriter, r *http.Request) {
	var query = r.URL.Query()
	var source = query.Get("source")
	var fileID = query.Get("id")
	var filename = query.Get("name")
	var availableHosts = []string{
		"regulation.gov.ru",
		"sozd.duma.gov.ru",
	}
	if !slices.Contains(availableHosts, source) {
		err := fmt.Errorf("недопустимый хост: %s", source)
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	var sourceURL = lawsparser.BuildDownloadURL(source, fileID)
	slog.Debug("ProxyDownload",
		slog.String("source", source),
		slog.String("fileID", fileID),
		slog.String("sourceURL", sourceURL),
	)
	var ctx = r.Context()
	proxyReq, err := http.NewRequestWithContext(ctx, "GET", sourceURL, nil)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	proxyReq.Header.Add("User-Agent", r.Header.Get("User-Agent"))
	response, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		Jsonify(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close()
	clientHeaders := w.Header()
	clientHeaders.Add(headerContentType, response.Header.Get(headerContentType))
	if value := response.Header.Get(headerContentDisposition); value != "" {
		clientHeaders.Add(headerContentDisposition, value)
	} else {
		value := fmt.Sprintf(`attachment; filename="%s"`, filename)
		clientHeaders.Add(headerContentDisposition, value)
	}
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, response.Body)
	deferErrLog(err)
}
