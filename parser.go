package googlesheetsparser

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"reflect"
	"strings"

	"github.com/gertd/go-pluralize"
	"google.golang.org/api/sheets/v4"
)

var pluralizeClient = pluralize.NewClient()

var (
	ErrNoSpreadSheetID       = errors.New("no spreadsheet id provided")
	ErrNoSheetName           = errors.New("no sheet name provided")
	ErrUnsupportedType       = errors.New("unsupported type")
	ErrFieldNotFoundInSheet  = errors.New("field not found in sheet")
	ErrFieldNotFoundInStruct = errors.New("field not found in struct")
	ErrInvalidDateTimeFormat = errors.New("invalid datetime format")
)

var dateTimeFormats = [...]string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -0700",
}

type Options struct {
	Service          *sheets.Service
	SpreadsheetID    string
	SheetName        string
	DatetimeFormats  []string
	AllowSkipFields  bool
	AllowSkipColumns bool

	fetch fetchFN
	built bool
}

func (o Options) Build() Options {
	if o.built {
		return o
	}
	o.built = true
	o.fetch = fetchViaGoogleAPI

	o.DatetimeFormats = append(o.DatetimeFormats, dateTimeFormats[:]...)

	return o
}

// Result is a struct that holds the result of a parsing operation.
// It contains the parsed value or an error if any.
type Result[T any] struct {
	Val T
	Err error
}

// ParseSheetIntoStructs parses a sheet page and calls the callback for each object.
func ParseSheetIntoStructs[T any](ctx context.Context, options Options) (iter.Seq2[int, Result[T]], error) {
	results, _, err := parseSheet[T](ctx, options)
	return results, err
}

// ParseSheetIntoStructSlice parses a sheet page and returns a slice of structs with the give type.
func ParseSheetIntoStructSlice[T any](ctx context.Context, options Options) ([]T, error) {
	results, rows, err := parseSheet[T](ctx, options)
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, rows)
	for _, item := range results {
		if item.Err != nil {
			return nil, item.Err
		}
		items = append(items, item.Val)
	}

	return items, ctx.Err()
}

// ParseSheetIntoStructs parses a sheet page and calls the callback for each object.
func parseSheet[T any](ctx context.Context, options Options) (iter.Seq2[int, Result[T]], int, error) {
	if !options.built {
		log.Println("googlesheetsparser: Warning: Using options that are not built")
	}

	// Set Params
	refT := reflect.TypeFor[T]()
	if options.SheetName != "" {
		options.SheetName = pluralizeClient.Plural(refT.Name())
	}

	// Validate Params
	if options.SpreadsheetID == "" {
		return nil, 0, ErrNoSpreadSheetID
	}
	if options.SheetName == "" {
		return nil, 0, ErrNoSheetName
	}

	resp, err := options.fetch(ctx, options)
	if err != nil {
		return nil, 0, err
	}

	mappings, err := createMappings(refT, resp.Values[0], options)
	if err != nil {
		return nil, 0, err
	}

	fillEmptyValues(resp)

	return func(yield func(int, Result[T]) bool) {
	rows:
		for i, row := range resp.Values[1:] {
			select {
			case <-ctx.Done():
				return
			default:
				rowIdx := i + 2 // 1-based index + first row is captions
				var item T
				refItem := reflect.ValueOf(&item).Elem()
				for _, mapping := range mappings {
					val, nonEmpty, err := mapping.convert(row[mapping.colIndex].(string), options.DatetimeFormats)
					if err != nil {
						err = fmt.Errorf("%s: %s%d: %w", options.SheetName, columnName(mapping.colIndex), rowIdx, err)
						if !yield(rowIdx, Result[T]{Err: err}) {
							return
						}

						continue rows
					}

					if !nonEmpty {
						continue
					}

					if mapping.initEmbedPtr != nil {
						mapping.initEmbedPtr(refItem)
					}

					refItem.FieldByIndex(mapping.field.Index).Set(val)
				}

				if !yield(rowIdx, Result[T]{Val: item}) {
					return
				}
			}
		}
	}, len(resp.Values[1:]), nil
}

func fillEmptyValues(data *sheets.ValueRange) {
	var maxWidth int
	for _, row := range data.Values {
		if len(row) > maxWidth {
			maxWidth = len(row)
		}
	}

	for rowIdx, row := range data.Values {
		for colIdx := len(row); colIdx < maxWidth; colIdx++ {
			data.Values[rowIdx] = append(data.Values[rowIdx], "")
		}
	}
}

func columnName(index int) string {
	index += 1
	var res string
	for index > 0 {
		index--
		res = string(rune(index%26+97)) + res
		index /= 26
	}
	return strings.ToUpper(res)
}

type fetchFN func(ctx context.Context, cfg Options) (*sheets.ValueRange, error)

func fetchViaGoogleAPI(ctx context.Context, cfg Options) (*sheets.ValueRange, error) {
	return cfg.Service.Spreadsheets.Values.Get(cfg.SpreadsheetID, cfg.SheetName).
		Context(ctx).
		MajorDimension("ROWS").
		Do()
}
