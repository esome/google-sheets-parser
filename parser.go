package gsheets

import (
	"fmt"
	"iter"
	"reflect"
	"strings"

	"google.golang.org/api/sheets/v4"
)

// Result is a struct that holds the result of a parsing operation.
// It contains the parsed value or an error if any.
type Result[T any] struct {
	Val T
	Err error
}

// ParseSheetIntoStructs parses a sheet page and returns an iterator over the parsing Result.
// If an error occurs during validation or when fetching data, the function will return it.
// Parsing errors are returned as part of the Result, and can therefore be handled by the caller.
// The iterator will still proceed to the next row, if it isn't stopped.
func ParseSheetIntoStructs[T any](cfg Config, opts ...ConfigOption) (iter.Seq2[int, Result[T]], error) {
	results, _, err := parseSheet[T](cfg, opts)
	return results, err
}

// ParseSheetIntoStructSlice parses a sheet page and returns a slice of structs with the give type.
// If an error occurs, the function will immediately return it.
func ParseSheetIntoStructSlice[T any](cfg Config, opts ...ConfigOption) ([]T, error) {
	results, rows, err := parseSheet[T](cfg, opts)
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

	return items, cfg.Context().Err()
}

func parseSheet[T any](cfg Config, opts []ConfigOption) (iter.Seq2[int, Result[T]], int, error) {
	refT := reflect.TypeFor[T]()
	cfg, err := cfg.init(refT, opts)
	if err != nil {
		return nil, 0, err
	}

	resp, err := cfg.fetch(cfg)
	if err != nil {
		return nil, 0, err
	}

	mappings, err := createMappings(refT, resp.Values[0], cfg)
	if err != nil {
		return nil, 0, err
	}

	fillEmptyValues(resp)

	ctx := cfg.Context()
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
						err = &MappingError{
							Sheet: cfg.sheetName,
							Cell:  fmt.Sprintf("%s%d", columnName(mapping.colIndex), rowIdx),
							Field: mapping.typeName + "." + mapping.field.Name,
							err:   err,
						}
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
	}, len(resp.Values[1:]), ctx.Err()
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

type fetchFN func(cfg Config) (*sheets.ValueRange, error)

func fetchViaGoogleAPI(cfg Config) (*sheets.ValueRange, error) {
	return cfg.Service.Spreadsheets.Values.Get(cfg.spreadsheetID, cfg.sheetName).
		Context(cfg.Context()).
		MajorDimension("ROWS").
		Do()
}
