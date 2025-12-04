package handler

import (
	"cronlogger/handler/html"
	"cronlogger/store"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
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

const defaultPageSize = 20

// RedirectStart redirects the root path to the startpage
func (c *CronLogHandler) RedirectStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/StartPage", http.StatusFound)
	}
}

// StartPage is the first page and displays items of the cronlogger store
func (c *CronLogHandler) StartPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.logger.Info("serving the StartPage")

		var skip int64
		result, err := c.store.GetPagedItems(defaultPageSize, int(skip), nil, nil)
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not get items from store; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		totalPages, currentPage := getPageInfo(result, int64(skip))
		skip = skip + defaultPageSize

		html.Layout(html.StartPage(result, defaultPageSize, totalPages, currentPage, skip), c.version).Render(r.Context(), w)
	}
}

const skipParamName = "skip"
const dateFromParamName = "from"
const dateUntilParamName = "until"
const dateFormat = "2006-01-02"

// TableResult is used via htmx and only provides the table results
func (c *CronLogHandler) TableResult() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not parse provided formdate; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		skipParam := r.FormValue(skipParamName)
		fromParam := r.FormValue(dateFromParamName)
		untilParam := r.FormValue(dateUntilParamName)

		var (
			skip  int64
			from  *time.Time
			until *time.Time
		)
		if skipParam != "" {
			s, err := strconv.Atoi(skipParam)
			if err != nil {
				c.logger.Warn(fmt.Sprintf("could not parse skip param: '%s'; %v", skipParam, err))
				skip = 0
			}
			skip = int64(s)
			if skip < 0 {
				skip = 0
			}
		}

		if fromParam != "" {
			from = parseDate(fromParam)
		}
		if untilParam != "" {
			until = parseDate(untilParam)
		}

		result, err := c.store.GetPagedItems(defaultPageSize, int(skip), getStartDate(from), getEndDate(until))
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not get items from store; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		skip = skip + defaultPageSize
		totalPages, currentPage := getPageInfo(result, int64(skip))
		html.TableResult(result, defaultPageSize, totalPages, currentPage, skip, formatDate(from), formatDate(until)).Render(r.Context(), w)
	}
}

func getPageInfo(result store.PagedOpResults, skip int64) (totalPages, currentPage int64) {
	totalPages = result.TotalCount / defaultPageSize
	modPageSize := result.TotalCount % defaultPageSize
	if modPageSize != 0 {
		totalPages += 1
	}
	currentPage = skip / defaultPageSize
	if currentPage > 0 {
		currentPage -= 1
	}
	return
}

func parseDate(input string) *time.Time {
	t, err := time.Parse(dateFormat, input)
	if err != nil {
		return nil
	}
	return &t
}

func getEndDate(date *time.Time) *time.Time {
	if date == nil {
		return nil
	}

	d := time.Date(date.Year(), date.Month(), date.Day(), 23, 23, 59, 0, time.UTC)
	return &d
}

func getStartDate(date *time.Time) *time.Time {
	if date == nil {
		return nil
	}

	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 1, 0, time.UTC)
	return &d
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	var (
		mPrefix = "0"
		dPrefix = "0"
	)
	if t.Month() > 9 {
		mPrefix = ""
	}
	if t.Day() > 9 {
		dPrefix = ""
	}

	return fmt.Sprintf("%d-%s%d-%s%d", t.Year(), mPrefix, t.Month(), dPrefix, t.Day())
}
