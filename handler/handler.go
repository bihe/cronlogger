package handler

import (
	"cronlogger/handler/html"
	"cronlogger/store"
	"fmt"
	"log/slog"
	"net/http"
)

// CronLogHandler is used to visualize the content of
// the cronlogger store via HTML templates
type CronLogHandler struct {
	store   store.OpResultStore
	logger  *slog.Logger
	version string
}

// New returns a new instance of the CronLogHandler
func New(store store.OpResultStore, logger *slog.Logger, version string) *CronLogHandler {
	return &CronLogHandler{
		store:   store,
		logger:  logger,
		version: version,
	}
}

// StartPage is the first page and displays items of the cronlogger store
func (c *CronLogHandler) StartPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.logger.Info("serving the StartPage")

		items, err := c.store.GetAll()
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not get items from store; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		html.Layout(html.StartPage(items), c.version).Render(r.Context(), w)
	}
}
