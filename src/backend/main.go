package main

import (
	"context"
	"flag"
	"log/slog"
	"monitoring_draft_laws/config"
	"monitoring_draft_laws/server"
	"os"
	"path/filepath"
	_ "time/tzdata" // tzdata
)

var (
	logLevel       = new(slog.LevelVar)
	loglevel       string
	configFilename string
	checkDocs      bool
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					source, ok := a.Value.Any().(*slog.Source)
					if !ok {
						return a
					}
					source.File = filepath.Base(source.File)
				}
				return a
			}})))
	flag.StringVar(&configFilename, "config", "config.json", "--config config.json")
	flag.StringVar(&loglevel, "loglevel", "INFO", "--loglevel debug")
	flag.BoolVar(&checkDocs, "checkdocs", false, "--checkdocs")
	flag.Parse()
	if err := logLevel.UnmarshalText([]byte(loglevel)); err != nil {
		panic(err)
	}
}

func main() {
	cnf, err := config.LoadConfigFromFile(configFilename)
	if err != nil {
		panic(err)
	}
	if cnf.Debug {
		slog.Debug("config", slog.Any("props", cnf))
	}
	ctx := context.Background()
	app := server.NewApplication(cnf)
	switch {
	case checkDocs:
		app.CheckDocuments(ctx)
		return
	default:
		app.ListenAndServe(ctx)
	}
}
