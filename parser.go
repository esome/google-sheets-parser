package gsheets

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"reflect"
	"strings"

	"google.golang.org/api/sheets/v4"
)

var (
	ErrNoService             = errors.New("gsheets: no Google API service  provided")
	ErrNoSpreadSheetID       = errors.New("gsheets: no spreadsheet id provided")
	ErrUnsupportedType       = errors.New("gsheets: unsupported type")
	ErrFieldNotFoundInSheet  = errors.New("gsheets: field not found in sheet")
	ErrFieldNotFoundInStruct = errors.New("gsheets: field not found in struct")
	ErrInvalidDateTimeFormat = errors.New("gsheets: invalid datetime format")
)

// Result is a struct that holds the result of a parsing operation.
// It contains the parsed value or an error if any.
type Result[T any] struct {
	Val T
	Err error
}

// ParseSheetIntoStructs parses a sheet page and returns an iterator over the parsing Result.
// If an error occurs during validation or when fetching data, the function will return an error.
// Parsing errors are returned as part of the Result, and can therefore be handled by the caller.
func ParseSheetIntoStructs[T any](ctx context.Context, cfg Config, opts ...ConfigOption) (iter.Seq2[int, Result[T]], error) {
	results, _, err := parseSheet[T](ctx, cfg, opts)
	return results, err
}

// ParseSheetIntoStructSlice parses a sheet page and returns a slice of structs with the give type.
// If an error occurs during parsing, the function will return an error.
func ParseSheetIntoStructSlice[T any](ctx context.Context, cfg Config, opts ...ConfigOption) ([]T, error) {
	results, rows, err := parseSheet[T](ctx, cfg, opts)
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

func parseSheet[T any](ctx context.Context, cfg Config, opts []ConfigOption) (iter.Seq2[int, Result[T]], int, error) {
	refT := reflect.TypeFor[T]()
	cfg, err := cfg.init(refT, opts)
	if err != nil {
		log.Println("gsheets: Warning: Using cfg that are not built")
	}

	resp, err := cfg.fetch(ctx, cfg)
	if err != nil {
		return nil, 0, err
	}

	mappings, err := createMappings(refT, resp.Values[0], cfg)
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
					val, nonEmpty, err := mapping.convert(row[mapping.colIndex].(string), cfg.datetimeFormats)
					if err != nil {
						err = fmt.Errorf("%s: %s%d: %w", cfg.sheetName, columnName(mapping.colIndex), rowIdx, err)
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

type fetchFN func(ctx context.Context, cfg Config) (*sheets.ValueRange, error)

func fetchViaGoogleAPI(ctx context.Context, cfg Config) (*sheets.ValueRange, error) {
	return cfg.Service.Spreadsheets.Values.Get(cfg.spreadsheetID, cfg.sheetName).
		Context(ctx).
		MajorDimension("ROWS").
		Do()
}
