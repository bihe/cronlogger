package handler

import (
	"cronlogger"
	"cronlogger/handler/html"
	"cronlogger/store"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CronLogHandler is used to visualize the content of
// the cronlogger store via HTML templates
type CronLogHandler struct {
	store   store.OpResultStore
	logger  *slog.Logger
	version string
	config  cronlogger.AppConfig
}

// New returns a new instance of the CronLogHandler
func New(store store.OpResultStore, logger *slog.Logger, version string, config cronlogger.AppConfig) *CronLogHandler {
	return &CronLogHandler{
		store:   store,
		logger:  logger,
		version: version,
		config:  config,
	}
}

const defaultPageSize = 20

// RedirectStart redirects the root path to the startpage
func (c *CronLogHandler) RedirectStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/cronlogger/StartPage", http.StatusFound)
	}
}

// StartPage is the first page and displays items of the cronlogger store
func (c *CronLogHandler) StartPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.logger.Info("serving the StartPage")

		var skip int64
		result, err := c.store.GetPagedItems(defaultPageSize, int(skip), nil, nil, "")
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not get items from store; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		apps, err := c.store.GetAvailApps()
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not get available apps from store; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get available apps from store; %v", err))).Render(r.Context(), w)
			return
		}

		totalPages, currentPage := getPaginationInfo(result, int64(skip))
		skip = skip + defaultPageSize

		html.Layout(html.StartPage(result, c.config, apps, defaultPageSize, totalPages, currentPage, skip), c.version).Render(r.Context(), w)
	}
}

const skipParamName = "skip"
const dateFromParamName = "from"
const dateUntilParamName = "until"
const applicationParamName = "application"
const dateFormat = "2006-01-02"

// TableResult is used via htmx and only provides the table results
func (c *CronLogHandler) TableResult() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not parse provided formdata; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		skipParam := r.FormValue(skipParamName)
		fromParam := r.FormValue(dateFromParamName)
		untilParam := r.FormValue(dateUntilParamName)
		appParam := r.FormValue(applicationParamName)

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

		result, err := c.store.GetPagedItems(defaultPageSize, int(skip), getStartDate(from), getEndDate(until), appParam)
		if err != nil {
			c.logger.Error(fmt.Sprintf("could not get items from store; %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			html.ErrorPageLayout(html.ErrorApplication("/", r, fmt.Sprintf("could not get items from store; %v", err))).Render(r.Context(), w)
			return
		}

		skip = skip + defaultPageSize
		totalPages, currentPage := getPaginationInfo(result, int64(skip))
		html.TableResult(result, c.config, defaultPageSize, totalPages, currentPage, skip, formatDate(from), formatDate(until), appParam).Render(r.Context(), w)
	}
}

// define a header for htmx to trigger events
// https://htmx.org/headers/hx-trigger/
const htmxHeaderTrigger = "HX-Trigger"

/*
https://htmx.org/headers/hx-trigger/

Targeting Other Elements
You can trigger events on other target elements by adding a target argument to the JSON object.

HX-Trigger: {"showMessage":{"target" : "#otherElement"}}
*/

type elementTarget struct {
	Target string `json:"target,omitempty"`
}

// this event triggers an action on a DOM element
// the provided id tells htmx to execute the trigger on the specific element
type showOutputTrigger struct {
	ShowOutput elementTarget `json:"showOutput,omitempty"`
}

// ToggleOutputDetail triggers the custom event to show the output details
// this logic only sends an HTTP header with the target-id of the DOM
// element which should react
func (c *CronLogHandler) ToggleOutputDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := r.PathValue("id")
		if idParam == "" {
			c.logger.Error("no id param supplied")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		triggerEvent := showOutputTrigger{
			ShowOutput: elementTarget{
				Target: fmt.Sprintf("#item-output-%s", idParam),
			},
		}

		triggerJson := Json(triggerEvent)
		w.Header().Add(htmxHeaderTrigger, triggerJson)
	}
}

// OutputDetail provides the specific output of an execution
func (c *CronLogHandler) OutputDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := r.PathValue("id")
		if idParam == "" {
			c.logger.Error("no id param supplied")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		showParam := r.PathValue("show")
		var toggle bool
		toggle = true
		if strings.ToUpper(showParam) == "TRUE" {
			toggle = false
		}

		item, err := c.store.GetById(idParam)
		if err != nil {
			c.logger.Error("could not get item by id '%s'; %v", idParam, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		html.OutputDetails(item, toggle).Render(r.Context(), w)
	}
}

func getPaginationInfo(result store.PagedOpResults, skip int64) (totalPages, currentPage int64) {
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

// Json serialized the given data
func Json[T any](data T) string {
	payload, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshall data; %v", err))
	}
	return string(payload)
}
