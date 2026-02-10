/*
Package api exposes the main API engine. All HTTP APIs are handled here.
*/
package api

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sapienzaapps/wasatext/service/database"
	"github.com/sirupsen/logrus"
)

// Config is used to provide dependencies and configuration to the New function.
type Config struct {
	Logger   logrus.FieldLogger
	Database database.AppDatabase
}

// Router is the package API interface representing an API handler builder
type Router interface {
	Handler() http.Handler
	Close() error
}

// New returns a new Router instance
func New(cfg Config) (Router, error) {
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	return &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}, nil
}

type _router struct {
	router     *httprouter.Router
	baseLogger logrus.FieldLogger
	db         database.AppDatabase
}

func (rt *_router) Close() error {
	return nil
}
