package gsheets

import (
	"context"
	"reflect"

	"github.com/gertd/go-pluralize"
	"google.golang.org/api/sheets/v4"
)

// defaultTag is the default tag-name to be looked at in the structs.
const defaultTag = "gsheets"

// Config holds the configuration for the Google Sheets parser.
type Config struct {
	Service *sheets.Service
	config
}

type config struct {
	spreadsheetID    string
	sheetName        string
	tagName          string
	datetimeFormats  []string
	allowSkipFields  bool
	allowSkipColumns bool
	built            bool
	ctx              context.Context
	fetch            fetchFN
}

// MakeConfig creates a new Config with the given Google Sheets service and arbitrary options.
func MakeConfig(svc *sheets.Service, spreadsheetID string, opts ...ConfigOption) Config {
	cfg := &config{
		spreadsheetID: spreadsheetID,
		tagName:       defaultTag,
		fetch:         fetchViaGoogleAPI,
	}

	for _, modify := range opts {
		modify(cfg)
	}

	return Config{
		Service: svc,
		config:  *cfg,
	}
}

// ConfigOption is a function allows to modify a Config.
type ConfigOption func(*config)

// WithContext sets the given context for the Config.
// This context will be used for all API calls, and cancellation will be respected during iteration.
func WithContext(ctx context.Context) ConfigOption {
	return func(c *config) {
		c.ctx = ctx
	}
}

// WithSpreadsheetID sets the spreadsheet ID for the Config.
func WithSpreadsheetID(id string) ConfigOption {
	return func(c *config) {
		c.spreadsheetID = id
	}
}

// WithSheetName sets the sheet-name for the Config.
func WithSheetName(name string) ConfigOption {
	return func(c *config) {
		c.sheetName = name
	}
}

// WithTagName sets the tag-name to be looked at in the structs.
// This might come in handy if you have multiple structs with different tags,
// or another library also uses `gsheets:` as tag identifier.
func WithTagName(name string) ConfigOption {
	return func(c *config) {
		c.tagName = name
	}
}

// WithDatetimeFormats allows to define additional date-time formats to be recognized during the parsing.
func WithDatetimeFormats(formats ...string) ConfigOption {
	return func(c *config) {
		c.datetimeFormats = formats
	}
}

// WithAllowSkipFields allows to skip fields that are not found in the sheet.
// If this is set to false, an error will be raised.
func WithAllowSkipFields(allow bool) ConfigOption {
	return func(c *config) {
		c.allowSkipFields = allow
	}
}

// WithAllowSkipColumns allows to skip columns that cannot be mapped to a struct field.
// If this is set to false, an error will be raised.
func WithAllowSkipColumns(allow bool) ConfigOption {
	return func(c *config) {
		c.allowSkipColumns = allow
	}
}

// withFetch is for testing purposes only, and allows to mock the call to the Google Sheets API.
func withFetch(fetch fetchFN) ConfigOption {
	return func(c *config) {
		c.fetch = fetch
	}
}

func (c Config) init(ref reflect.Type, opts []ConfigOption) (Config, error) {
	if len(opts) > 0 {
		for _, modify := range opts {
			modify(&c.config)
		}
		c.built = false
	}

	if c.built {
		return c, nil
	}

	if c.Service == nil {
		return c, ErrNoService
	}

	if c.spreadsheetID == "" {
		return c, ErrNoSpreadSheetID
	}

	if c.fetch == nil {
		c.fetch = fetchViaGoogleAPI
	}
	if c.tagName == "" {
		c.tagName = defaultTag
	}
	if c.sheetName == "" {
		c.sheetName = pluralizeClient.Plural(ref.Name())
	}

	c.datetimeFormats = append(c.datetimeFormats, dateTimeFormats[:]...)

	c.built = true
	return c, nil
}

// Context returns the configured context, or creates a new background context
func (c *config) Context() context.Context {
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	return c.ctx
}

var pluralizeClient = pluralize.NewClient()

var dateTimeFormats = [...]string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -0700",
}
