package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"monitoring_draft_laws/config"
	"monitoring_draft_laws/internals/lawsparser"
	"monitoring_draft_laws/internals/scheduler"
	"monitoring_draft_laws/store"
	"monitoring_draft_laws/utils"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stvoidit/megaplan"
)

// Application - сборник всего необходимого в одну структуру
type Application struct {
	srv        *http.Server             // http.Server
	db         *store.DB                // подключение к БД
	scheduler  *scheduler.JobsScheduler // планировщик задач
	tgbot      *tgbotapi.BotAPI         // телеграм бот
	config     *config.Config           // конфиг
	sg         *utils.Cryptographer     // шифрование
	mpapi      *megaplan.API
	contextKey contextKey
}

func (app *Application) GetStore() *store.DB {
	return app.db
}

// NewApplication - фабрика
func NewApplication(config *config.Config) *Application {
	slog.Info(fmt.Sprintf(`use database: '%s'`, config.DB.Dbname))
	db, err := store.NewDB(config.ConnStringDB())
	if err != nil {
		panic(err)
	}
	mpapi := megaplan.NewAPI(config.Megaplan.UUID, config.Megaplan.Token, config.Megaplan.Domain)
	mpapi.EnableCompression(true)
	var app = &Application{
		contextKey: "profile",
		mpapi:      mpapi,
		sg:         utils.NewCryptographer(config.Megaplan.Token),
		db:         db,
		config:     config}
	if err := app.initTgBot(); err != nil {
		panic(err)
	}
	return app
}

// ListenAndServe - http.ListenAndServe + graceful shutdown
func (app *Application) ListenAndServe(ctx context.Context) {
	r := mux.NewRouter()
	r.Use(
		CompressHandler,
		app.CORSHandler,
	)
	app.setHandlers(r)
	// app.initTGbot(r)
	app.setStatic(r)
	var protocols http.Protocols
	protocols.SetHTTP1(true)
	protocols.SetHTTP2(true)
	protocols.SetUnencryptedHTTP2(true)
	app.srv = &http.Server{
		Addr:                         app.config.Srv.Address,
		Handler:                      handlers.LoggingHandler(os.Stdout, r),
		ErrorLog:                     log.Default(),
		WriteTimeout:                 time.Minute,
		ReadTimeout:                  time.Minute,
		Protocols:                    &protocols,
		DisableGeneralOptionsHandler: false,
		HTTP2: &http.HTTP2Config{
			MaxConcurrentStreams:         100,
			SendPingTimeout:              time.Second * 15,
			PermitProhibitedCipherSuites: true,
		},
	}
	ctx = app.shutdown(ctx)
	app.scheduler = scheduler.NewJobsScheduler(ctx)
	app.srv.RegisterOnShutdown(app.db.Close)
	go func() {
		if err := app.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ListenAndServe", slog.String("error", err.Error()))
		}
	}()
	// go func() {
	// 	if app.tgbot != nil && !app.config.Debug {
	// 		time.Sleep(time.Second * 3)
	// 		var whendpoint = (url.URL{
	// 			Scheme: "https",
	// 			Host:   app.config.TGBOT.Domain,
	// 			Path:   "/tgbot/hook",
	// 		})
	// 		wh, err := tgbotapi.NewWebhook(whendpoint.String())
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		resp, err := app.tgbot.Request(wh)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		log.Printf("%+v\n", resp)
	// 		info, err := app.tgbot.GetWebhookInfo()
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		log.Printf("%+v\n", info)
	// 		log.Println("webhook set:", info.IsSet())
	// 	}
	// }()

	deferErrLog(app.startTgPolling(ctx))
	app.loadJobs()

	log.Printf("http://%s\n", app.srv.Addr)
	<-ctx.Done()
}

func (app *Application) setHandlers(r *mux.Router) {
	sr := r.PathPrefix("/api/").Subrouter()
	sr.Use(app.megaplanverify())
	sr.HandleFunc("/init", app.InitHandler)
	sr.HandleFunc("/documents", app.AddDocument).Methods(http.MethodPost)
	sr.HandleFunc("/documents", app.GetDocuments).Methods(http.MethodGet)
	sr.HandleFunc("/documents/fetch", app.FetchUpdateAll()).Methods(http.MethodPost)
	sr.HandleFunc("/documents/download", app.DownloadDocuments).Methods(http.MethodPost)
	sr.HandleFunc("/document", app.GetDocument).Methods(http.MethodGet)
	sr.HandleFunc("/document", app.UpdateDocument).Methods(http.MethodPut)
	sr.HandleFunc("/document", app.PatchDraftDocument).Methods(http.MethodPatch)
	sr.HandleFunc("/document", app.DeleteDocument).Methods(http.MethodDelete)
	sr.HandleFunc("/document/fetch", app.FetchUpdate).Methods(http.MethodPost)
	sr.HandleFunc("/roles_settings", app.RolesSettings).Methods(http.MethodGet, http.MethodPost)
	sr.HandleFunc("/favorite", app.FavoriteHandle).Methods(http.MethodPut)
	sr.HandleFunc("/ntype", app.ChangeNType).Methods(http.MethodPost)
	sr.HandleFunc("/changeslogs", app.GetJournal).Methods(http.MethodGet)
	sr.HandleFunc("/files", app.DocumentFiles).Methods(http.MethodGet, http.MethodPost)
	sr.HandleFunc("/file", app.DocumentFiles).Methods(http.MethodGet, http.MethodDelete)
	sr.HandleFunc("/archive", app.ArchiveStatus).Methods(http.MethodPatch)
	sr.HandleFunc("/proxy_download", app.ProxyDownload).Methods(http.MethodGet)
}

// func (app *Application) initTGbot(r *mux.Router) {
// 	if app.tgbot == nil {
// 		return
// 	}
// 	br := r.PathPrefix("/tgbot/").Subrouter()
// 	br.HandleFunc("/hook", app.TelegramBotWebHook).Methods(http.MethodPost, http.MethodGet)
// }

func (app *Application) setStatic(router *mux.Router) {
	const staticDir = "/www/data/static"
	router.PathPrefix("/").Handler(staticHandler(FileSystemSPA(staticDir)))
}

func (app *Application) shutdown(ctx context.Context) context.Context {
	srvCtx, cancel := context.WithCancel(ctx)
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancel()
		<-sigChan
		close(sigChan)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		err := app.srv.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			log.Println("shutdown error:", err)
			return
		}
		log.Println("server shutdown")
	}()
	return srvCtx
}

func (app *Application) CheckDocuments(ctx context.Context) {
	app.defaultJobFunc(ctx)
}

// defaultJonFunc - стандартная функции для планировщика
func (app *Application) defaultJobFunc(ctx context.Context) {
	docs, err := app.db.SelectDocuments(ctx, 0)
	if err != nil {
		slog.Error("SelectDocuments", slog.String("error", err.Error()))
		return
	}
	var count uint64
	slices.SortFunc(docs, func(a, b lawsparser.FormatDocument) int {
		if a.SourceHost[0] > b.SourceHost[0] {
			return -1
		} else if a.SourceHost[0] == b.SourceHost[0] {
			return 0
		} else {
			return 1
		}
	})

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			slog.Warn("defaultJobFunc.context.Done", slog.String("error", ctx.Err().Error()))
			return
		default:
			if doc.IsNotActual() {
				continue
			}
			var start = time.Now()
			if err := doc.UpdateDocument(ctx); err != nil {
				slog.Error("UpdateDocument",
					slog.String("error", err.Error()),
					slog.String("id", doc.DocumentID),
					slog.String("host", doc.SourceHost),
				)
				time.Sleep(time.Millisecond * 200)
				continue
			}
			_fd, newStatus, newFiles, err := app.db.UpdateDocumenScheduler(ctx, doc)
			if err != nil {
				slog.Error("UpdateDocumenScheduler",
					slog.String("error", err.Error()),
					slog.String("id", doc.DocumentID),
					slog.String("host", doc.SourceHost),
				)
				continue
			}
			if _fd == nil {
				slog.Info("UpdateDocumenScheduler",
					slog.String("id", doc.DocumentID),
					slog.String("host", doc.SourceHost),
					slog.String("cause", "no have changes, skip"))
				continue
			}
			slog.Info("UpdateDocumenScheduler",
				slog.String("status", "updated"),
				slog.String("id", _fd.DocumentID),
				slog.String("host", _fd.SourceHost),
				slog.String("duration", time.Since(start).String()),
				slog.String("newStatus", newStatus),
				slog.Any("newFiles", newFiles),
			)
			// нотификация в тг
			chatsID, err := app.db.SelectChatsTG(ctx, _fd.DocumentID)
			if err != nil {
				slog.Error("SelectChatsTG",
					slog.String("error", err.Error()),
					slog.String("id", _fd.DocumentID),
					slog.String("host", _fd.SourceHost),
				)
				continue
			}
			if err := app.Notify(ctx, _fd, chatsID, newStatus, newFiles); err != nil {
				slog.Error("notify",
					slog.String("error", err.Error()),
					slog.String("id", _fd.DocumentID),
					slog.String("host", _fd.SourceHost),
					slog.Any("chats", chatsID),
				)
			} else {
				slog.Info("notify",
					slog.String("id", _fd.DocumentID),
					slog.String("host", _fd.SourceHost),
					slog.Any("chats", chatsID),
				)
			}
			count++

			time.Sleep(time.Millisecond * 200)
		}

	}
	slog.Info("defaultJobFunc",
		slog.String("status", "complete"),
		slog.Uint64("count", count))
}

// loadJobs - загрузка документов для постановки на автопроверку
func (app *Application) loadJobs() {
	app.scheduler.AddJob("fetchUpdateAll_9", scheduler.NewJob(9, 0, app.defaultJobFunc))
	app.scheduler.AddJob("fetchUpdateAll_10", scheduler.NewJob(10, 0, app.defaultJobFunc))
	app.scheduler.AddJob("fetchUpdateAll_12", scheduler.NewJob(12, 0, app.defaultJobFunc))
	app.scheduler.AddJob("fetchUpdateAll_15", scheduler.NewJob(15, 0, app.defaultJobFunc))
	app.scheduler.AddJob("fetchUpdateAll_18", scheduler.NewJob(18, 0, app.defaultJobFunc))
	slog.Info("loadJobs: scheduler started")
}
