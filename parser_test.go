package gsheets

import (
	"context"
	"fmt"
	"iter"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/sheets/v4"
)

func TestParseSheetIntoStructs(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		t.Run("no service", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructs[allT](Config{}, withFetch(errorFetcher))
			assert.ErrorIs(t, err, ErrNoService)
		})

		t.Run("no spreadsheet id", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructs[allT](Config{Service: _svc}, withFetch(errorFetcher))
			assert.ErrorIs(t, err, ErrNoSpreadSheetID)
		})

		t.Run("error from fetch method", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructs[allT](
				Config{Service: _svc},
				WithSpreadsheetID("foobar"),
				withFetch(errorFetcher),
			)
			assert.ErrorIs(t, err, fetcherError)
		})

		t.Run("unsupported type", func(t *testing.T) {
			t.Parallel()

			type invalidT struct {
				Valid   int
				Invalid complex64
			}

			_, err := ParseSheetIntoStructs[invalidT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "invalidTS", cfg.sheetName)
					return &sheets.ValueRange{
						Values: [][]any{
							{"Valid", "Invalid"},
							{"1", "2.432"},
						},
					}, nil
				}),
			)
			assert.ErrorIs(t, err, ErrUnsupportedType)
		})

		t.Run("field not found in struct", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructs[boolsT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				WithTagName("sheets"),
				WithAllowSkipFields(true),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "boolsTS", cfg.sheetName)
					return stringsFetcher(cfg)
				}),
			)
			assert.ErrorIs(t, err, ErrFieldNotFoundInStruct)
		})

		t.Run("field not found in sheet", func(t *testing.T) {
			t.Parallel()

			type invalidT struct {
				Bools   boolsT
				Strings stringsT
			}

			_, err := ParseSheetIntoStructs[invalidT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				WithTagName("sheets"),
				WithAllowSkipColumns(true),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "invalidTS", cfg.sheetName)
					return stringsFetcher(cfg)
				}),
			)
			assert.ErrorIs(t, err, ErrFieldNotFoundInSheet)
		})

		t.Run("no mappings", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructs[boolsT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				WithTagName("sheets"),
				WithAllowSkipFields(true),
				WithAllowSkipColumns(true),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "boolsTS", cfg.sheetName)
					return stringsFetcher(cfg)
				}),
			)
			assert.ErrorIs(t, err, ErrNoMapping)
		})

		t.Run("parsing error", func(t *testing.T) {
			cfg := MakeConfig(_svc, "invalid", WithSheetName("invalid"), withFetch(parseValidationFetcher))

			t.Run("bool", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[boolsT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[boolsT](t, results, reflect.Bool)
			})

			t.Run("int", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[intsT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[intsT](t, results, reflect.Int)
			})

			t.Run("int8", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[int8sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[int8sT](t, results, reflect.Int8)
			})

			t.Run("int16", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[int16sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[int16sT](t, results, reflect.Int16)
			})

			t.Run("int32", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[int32sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[int32sT](t, results, reflect.Int32)
			})

			t.Run("int64", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[int64sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[int64sT](t, results, reflect.Int64)
			})

			t.Run("uint", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[uintsT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[uintsT](t, results, reflect.Uint)
			})

			t.Run("uint8", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[uint8sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[uint8sT](t, results, reflect.Uint8)
			})

			t.Run("uint16", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[uint16sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[uint16sT](t, results, reflect.Uint16)
			})

			t.Run("uint32", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[uint32sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[uint32sT](t, results, reflect.Uint32)
			})

			t.Run("uint64", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[uint64sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[uint64sT](t, results, reflect.Uint64)
			})

			t.Run("float32", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[float32sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[float32sT](t, results, reflect.Float32)
			})

			t.Run("float64", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[float64sT](cfg)
				require.NoError(t, err)

				assertParseErrorsIter[float64sT](t, results, reflect.Float64)
			})

			t.Run("time.Time", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[timesT](cfg)
				require.NoError(t, err)

				cells := []string{"A", "B"}
				fields := []string{"Value", "Ptr"}
				typeName := getTypeName[timesT]()

				var i int
				for r, item := range results {
					var mappingErr *MappingError
					require.ErrorAs(t, item.Err, &mappingErr)
					assert.Equal(t, "invalid", mappingErr.Sheet)
					assert.Equal(t, cells[i]+fmt.Sprint(r), mappingErr.Cell)
					assert.Equal(t, typeName+"."+fields[i], mappingErr.Field)

					var dateErr *InvalidDateTimeFormatError
					require.ErrorAs(t, item.Err, &dateErr)
					assert.Equal(t, dateTimeFormats[:], dateErr.Formats)
					assert.Equal(t, "invalid", dateErr.CV)
					i++
				}
			})
		})
	})

	cfg := MakeConfig(_svc, "test-workbook", WithSheetName("test-sheet"), WithDatetimeFormats("2.1.2006"))
	t.Run("nested structures", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), "test", "test")
		results, err := ParseSheetIntoStructs[nestedT](cfg,
			WithContext(ctx),
			withFetch(func(cfg Config) (*sheets.ValueRange, error) {
				assert.Equal(t, "test-workbook", cfg.spreadsheetID)
				assert.Equal(t, "test-sheet", cfg.sheetName)
				assert.Same(t, ctx, cfg.Context())
				return &sheets.ValueRange{
					Values: [][]any{
						{"f1", "f2", "f3", "f4"},
						{"1", "2"},
						{"1337", "4711", "1887"},
						{"", "", "", "42"},
					},
				}, nil
			}),
		)
		require.NoError(t, err)

		records := make([]nestedT, 0, 3)
		for _, item := range results {
			require.NoError(t, item.Err)
			records = append(records, item.Val)
		}

		// this tests, if the routine avoids unnecessary allocations of embedded struct pointers
		assert.Len(t, records, 3)
		assert.Equal(t, []nestedT{
			{L1: &nestedL1{F1: 1, L2: nestedL2{F2: 2}}},
			{L1: &nestedL1{F1: 1337, L2: nestedL2{F2: 4711, L3: &nestedL3{F3: 1887}}}},
			{F4: 42},
		}, records)
	})

	t.Run("all types", func(t *testing.T) {
		for _, tt := range allTypesTests() {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructs[allT](cfg,
					WithTagName("sheets"),
					WithAllowSkipColumns(true),
					WithAllowSkipFields(true),
					withFetch(func(cfg Config) (*sheets.ValueRange, error) {
						assert.Equal(t, "test-workbook", cfg.spreadsheetID)
						assert.Equal(t, "test-sheet", cfg.sheetName)
						return tt.fetch(cfg)
					}),
				)
				require.NoError(t, err)

				records := make([]allT, 0, len(tt.want))
				for _, item := range results {
					require.NoError(t, item.Err)
					records = append(records, item.Val)
				}

				assert.Len(t, records, len(tt.want))
				assert.Equal(t, tt.want, records)
			})
		}
	})

	t.Run("stop iter loop", func(t *testing.T) {
		t.Parallel()

		results, err := ParseSheetIntoStructs[stringsT](cfg,
			WithTagName("sheets"),
			withFetch(stringsFetcher),
		)
		require.NoError(t, err)

		records := make([]stringsT, 0, 2)
		var i int
		for _, item := range results {
			records = append(records, item.Val)
			if i++; i == 2 {
				break
			}
		}

		assert.Len(t, records, 2)
	})
}

func TestParseSheetIntoStructSlice(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		t.Run("no service", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructSlice[allT](Config{}, withFetch(errorFetcher))
			assert.ErrorIs(t, err, ErrNoService)
		})

		t.Run("no spreadsheet id", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructSlice[allT](Config{Service: _svc}, withFetch(errorFetcher))
			assert.ErrorIs(t, err, ErrNoSpreadSheetID)
		})

		t.Run("error from fetch method", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructSlice[allT](
				Config{Service: _svc},
				WithSpreadsheetID("foobar"),
				withFetch(errorFetcher),
			)
			assert.ErrorIs(t, err, fetcherError)
		})

		t.Run("unsupported type", func(t *testing.T) {
			t.Parallel()

			type invalidT struct {
				Valid   int
				Invalid complex64
			}

			_, err := ParseSheetIntoStructSlice[invalidT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "invalidTS", cfg.sheetName)
					return &sheets.ValueRange{
						Values: [][]any{
							{"Valid", "Invalid"},
							{"1", "2.432"},
						},
					}, nil
				}),
			)
			assert.ErrorIs(t, err, ErrUnsupportedType)
		})

		t.Run("field not found in struct", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructSlice[boolsT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				WithTagName("sheets"),
				WithAllowSkipFields(true),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "boolsTS", cfg.sheetName)
					return stringsFetcher(cfg)
				}),
			)
			assert.ErrorIs(t, err, ErrFieldNotFoundInStruct)
		})

		t.Run("field not found in sheet", func(t *testing.T) {
			t.Parallel()

			type invalidT struct {
				Bools   boolsT
				Strings stringsT
			}

			_, err := ParseSheetIntoStructSlice[invalidT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				WithTagName("sheets"),
				WithAllowSkipColumns(true),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "invalidTS", cfg.sheetName)
					return stringsFetcher(cfg)
				}),
			)
			assert.ErrorIs(t, err, ErrFieldNotFoundInSheet)
		})

		t.Run("no mappings", func(t *testing.T) {
			t.Parallel()

			_, err := ParseSheetIntoStructSlice[boolsT](
				Config{Service: _svc},
				WithSpreadsheetID("invalid"),
				WithTagName("sheets"),
				WithAllowSkipFields(true),
				WithAllowSkipColumns(true),
				withFetch(func(cfg Config) (*sheets.ValueRange, error) {
					assert.Equal(t, "invalid", cfg.spreadsheetID)
					assert.Equal(t, "boolsTS", cfg.sheetName)
					return stringsFetcher(cfg)
				}),
			)
			assert.ErrorIs(t, err, ErrNoMapping)
		})

		t.Run("parsing error", func(t *testing.T) {
			cfg := MakeConfig(_svc, "invalid", WithSheetName("invalid"), withFetch(parseValidationFetcher))

			t.Run("bool", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[boolsT](cfg)
				assertParseSliceError[boolsT](t, err, reflect.Bool)
				assert.Nil(t, results)
			})

			t.Run("int", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[intsT](cfg)
				assertParseSliceError[intsT](t, err, reflect.Int)
				assert.Nil(t, results)
			})

			t.Run("int8", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[int8sT](cfg)
				assertParseSliceError[int8sT](t, err, reflect.Int8)
				assert.Nil(t, results)
			})

			t.Run("int16", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[int16sT](cfg)
				assertParseSliceError[int16sT](t, err, reflect.Int16)
				assert.Nil(t, results)
			})

			t.Run("int32", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[int32sT](cfg)
				assertParseSliceError[int32sT](t, err, reflect.Int32)
				assert.Nil(t, results)
			})

			t.Run("int64", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[int64sT](cfg)
				assertParseSliceError[int64sT](t, err, reflect.Int64)
				assert.Nil(t, results)
			})

			t.Run("uint", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[uintsT](cfg)
				assertParseSliceError[uintsT](t, err, reflect.Uint)
				assert.Nil(t, results)
			})

			t.Run("uint8", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[uint8sT](cfg)
				assertParseSliceError[uint8sT](t, err, reflect.Uint8)
				assert.Nil(t, results)
			})

			t.Run("uint16", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[uint16sT](cfg)
				assertParseSliceError[uint16sT](t, err, reflect.Uint16)
				assert.Nil(t, results)
			})

			t.Run("uint32", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[uint32sT](cfg)
				assertParseSliceError[uint32sT](t, err, reflect.Uint32)
				assert.Nil(t, results)
			})

			t.Run("uint64", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[uint64sT](cfg)
				assertParseSliceError[uint64sT](t, err, reflect.Uint64)
				assert.Nil(t, results)
			})

			t.Run("float32", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[float32sT](cfg)
				assertParseSliceError[float32sT](t, err, reflect.Float32)
				assert.Nil(t, results)
			})

			t.Run("float64", func(t *testing.T) {
				t.Parallel()

				results, err := ParseSheetIntoStructSlice[float64sT](cfg)
				assertParseSliceError[float64sT](t, err, reflect.Float64)
				assert.Nil(t, results)
			})

			t.Run("time.Time", func(t *testing.T) {
				t.Parallel()

				typeName := getTypeName[timesT]()
				results, err := ParseSheetIntoStructSlice[timesT](cfg)

				var mappingErr *MappingError
				require.ErrorAs(t, err, &mappingErr)
				assert.Equal(t, "invalid", mappingErr.Sheet)
				assert.Equal(t, "A2", mappingErr.Cell)
				assert.Equal(t, typeName+".Value", mappingErr.Field)

				var dateErr *InvalidDateTimeFormatError
				require.ErrorAs(t, err, &dateErr)
				assert.Equal(t, dateTimeFormats[:], dateErr.Formats)
				assert.Equal(t, "invalid", dateErr.CV)

				assert.Nil(t, results)
			})
		})
	})

	cfg := MakeConfig(_svc, "test-workbook", WithSheetName("test-sheet"), WithDatetimeFormats("2.1.2006"))
	t.Run("nested structures", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), "test", "test")
		records, err := ParseSheetIntoStructSlice[nestedT](cfg,
			WithContext(ctx),
			withFetch(func(cfg Config) (*sheets.ValueRange, error) {
				assert.Equal(t, "test-workbook", cfg.spreadsheetID)
				assert.Equal(t, "test-sheet", cfg.sheetName)
				assert.Same(t, ctx, cfg.Context())
				return &sheets.ValueRange{
					Values: [][]any{
						{"f1", "f2", "f3", "f4"},
						{"1", "2"},
						{"1337", "4711", "1887"},
						{"", "", "", "42"},
					},
				}, nil
			}),
		)
		require.NoError(t, err)

		// this tests, if the routine avoids unnecessary allocations of embedded struct pointers
		assert.Len(t, records, 3)
		assert.Equal(t, []nestedT{
			{L1: &nestedL1{F1: 1, L2: nestedL2{F2: 2}}},
			{L1: &nestedL1{F1: 1337, L2: nestedL2{F2: 4711, L3: &nestedL3{F3: 1887}}}},
			{F4: 42},
		}, records)
	})

	t.Run("all types", func(t *testing.T) {
		for _, tt := range allTypesTests() {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				records, err := ParseSheetIntoStructSlice[allT](cfg,
					WithTagName("sheets"),
					WithAllowSkipColumns(true),
					WithAllowSkipFields(true),
					withFetch(func(cfg Config) (*sheets.ValueRange, error) {
						assert.Equal(t, "test-workbook", cfg.spreadsheetID)
						assert.Equal(t, "test-sheet", cfg.sheetName)
						return tt.fetch(cfg)
					}),
				)
				require.NoError(t, err)

				assert.Len(t, records, len(tt.want))
				assert.Equal(t, tt.want, records)
			})
		}
	})
}

type allTypesTT struct {
	name  string
	fetch fetchFN
	want  []allT
}

func allTypesTests() []allTypesTT {
	return []allTypesTT{
		{
			name:  "bools",
			fetch: boolsFetcher,
			want: []allT{
				{Bools: &boolsT{Value: true, Ptr: ptrTo(true)}},
				{Bools: &boolsT{Value: false}},
				{Bools: &boolsT{Ptr: ptrTo(false)}},
				{Bools: nil},
			},
		},
		{
			name:  "strings",
			fetch: stringsFetcher,
			want: []allT{
				{Strings: &stringsT{Value: "foo", Ptr: ptrTo("bar")}},
				{Strings: &stringsT{Value: "baz"}},
				{Strings: &stringsT{Ptr: ptrTo("boo")}},
				{Strings: nil},
			},
		},
		{
			name:  "ints",
			fetch: makeIntsFetcher(""), // int
			want: []allT{
				{Ints: &intsT{Value: 1337, Ptr: ptrTo(4711)}},
				{Ints: &intsT{Value: 1887}},
				{Ints: &intsT{Ptr: ptrTo(-6996)}},
				{Ints: nil},
			},
		},
		{
			name:  "int8s",
			fetch: makeIntsFetcher("8"), // int
			want: []allT{
				{Int8s: &int8sT{Value: math.MinInt8, Ptr: ptrTo[int8](math.MaxInt8)}},
				{Int8s: &int8sT{Value: math.MaxInt8 / 4}},
				{Int8s: &int8sT{Ptr: ptrTo[int8](math.MinInt8 / 2)}},
				{Int8s: nil},
			},
		},
		{
			name:  "int16s",
			fetch: makeIntsFetcher("16"), // int
			want: []allT{
				{Int16s: &int16sT{Value: math.MinInt16, Ptr: ptrTo[int16](math.MaxInt16)}},
				{Int16s: &int16sT{Value: math.MaxInt16 / 4}},
				{Int16s: &int16sT{Ptr: ptrTo[int16](math.MinInt16 / 2)}},
				{Int16s: nil},
			},
		},
		{
			name:  "int32s",
			fetch: makeIntsFetcher("32"), // int
			want: []allT{
				{Int32s: &int32sT{Value: math.MinInt32, Ptr: ptrTo[int32](math.MaxInt32)}},
				{Int32s: &int32sT{Value: math.MaxInt32 / 4}},
				{Int32s: &int32sT{Ptr: ptrTo[int32](math.MinInt32 / 2)}},
				{Int32s: nil},
			},
		},
		{
			name:  "int64s",
			fetch: makeIntsFetcher("64"), // int
			want: []allT{
				{Int64s: &int64sT{Value: math.MinInt64, Ptr: ptrTo[int64](math.MaxInt64)}},
				{Int64s: &int64sT{Value: math.MaxInt64 / 4}},
				{Int64s: &int64sT{Ptr: ptrTo[int64](math.MinInt64 / 2)}},
				{Int64s: nil},
			},
		},
		{
			name:  "uints",
			fetch: makeUintsFetcher(""), // int
			want: []allT{
				{Uints: &uintsT{Value: 1337, Ptr: ptrTo[uint](4711)}},
				{Uints: &uintsT{Value: 1887}},
				{Uints: &uintsT{Ptr: ptrTo[uint](6996)}},
				{Uints: nil},
			},
		},
		{
			name:  "uint8s",
			fetch: makeUintsFetcher("8"), // int
			want: []allT{
				{Uint8s: &uint8sT{Value: math.MaxUint8 / 2, Ptr: ptrTo[uint8](math.MaxUint8)}},
				{Uint8s: &uint8sT{Value: math.MaxUint8 / 8}},
				{Uint8s: &uint8sT{Ptr: ptrTo[uint8](math.MaxUint8 / 4)}},
				{Uint8s: nil},
			},
		},
		{
			name:  "uint16s",
			fetch: makeUintsFetcher("16"), // int
			want: []allT{
				{Uint16s: &uint16sT{Value: math.MaxUint16 / 2, Ptr: ptrTo[uint16](math.MaxUint16)}},
				{Uint16s: &uint16sT{Value: math.MaxUint16 / 8}},
				{Uint16s: &uint16sT{Ptr: ptrTo[uint16](math.MaxUint16 / 4)}},
				{Uint16s: nil},
			},
		},
		{
			name:  "uint32s",
			fetch: makeUintsFetcher("32"), // int
			want: []allT{
				{Uint32s: &uint32sT{Value: math.MaxUint32 / 2, Ptr: ptrTo[uint32](math.MaxUint32)}},
				{Uint32s: &uint32sT{Value: math.MaxUint32 / 8}},
				{Uint32s: &uint32sT{Ptr: ptrTo[uint32](math.MaxUint32 / 4)}},
				{Uint32s: nil},
			},
		},
		{
			name:  "uint64s",
			fetch: makeUintsFetcher("64"), // int
			want: []allT{
				{Uint64s: &uint64sT{Value: math.MaxUint64 / 2, Ptr: ptrTo[uint64](math.MaxUint64)}},
				{Uint64s: &uint64sT{Value: math.MaxUint64 / 8}},
				{Uint64s: &uint64sT{Ptr: ptrTo[uint64](math.MaxUint64 / 4)}},
				{Uint64s: nil},
			},
		},
		{
			name:  "float32s",
			fetch: makeFloatsFetcher("32"),
			want: []allT{
				{Float32s: &float32sT{Value: math.SmallestNonzeroFloat32, Ptr: ptrTo[float32](math.MaxFloat32)}},
				{Float32s: &float32sT{Value: math.MaxFloat32 / 2}},
				{Float32s: &float32sT{Ptr: ptrTo[float32](math.SmallestNonzeroFloat32 * 4)}},
				{Float32s: nil},
			},
		},
		{
			name:  "float64s",
			fetch: makeFloatsFetcher("64"),
			want: []allT{
				{Float64s: &float64sT{Value: math.SmallestNonzeroFloat64, Ptr: ptrTo(math.MaxFloat64)}},
				{Float64s: &float64sT{Value: math.MaxFloat64 / 2}},
				{Float64s: &float64sT{Ptr: ptrTo(math.SmallestNonzeroFloat64 * 4)}},
				{Float64s: nil},
			},
		},
		{
			name:  "times",
			fetch: timesFetcher,
			want: []allT{
				{Times: &timesT{
					Value: time.Date(2024, time.December, 19, 17, 35, 8, 0, time.UTC),
					Ptr:   ptrTo(time.Date(2020, time.November, 11, 11, 11, 11, 0, time.UTC)),
				}},
				{Times: &timesT{Value: time.Date(2021, time.March, 9, 0, 0, 0, 0, time.UTC)}},
				{Times: &timesT{Ptr: ptrTo(time.Date(2016, time.January, 6, 0, 0, 0, 0, time.UTC))}},
				{Times: nil},
			},
		},
	}
}

func assertParseErrorsIter[T any](t *testing.T, results iter.Seq2[int, Result[T]], typ reflect.Kind) {
	cells := []string{"A", "B"}
	fields := []string{"Value", "Ptr"}
	typeName := getTypeName[T]()

	var i int
	for r, item := range results {
		var mappingErr *MappingError
		require.ErrorAs(t, item.Err, &mappingErr)
		assert.Equal(t, "invalid", mappingErr.Sheet)
		assert.Equal(t, cells[i]+fmt.Sprint(r), mappingErr.Cell)
		assert.Equal(t, typeName+"."+fields[i], mappingErr.Field)

		var convertErr *ConvertError
		require.ErrorAs(t, item.Err, &convertErr)
		assert.Equal(t, typ, convertErr.Typ)
		assert.Equal(t, "invalid", convertErr.CV)
		i++
	}
}

func assertParseSliceError[T any](t *testing.T, err error, typ reflect.Kind) {
	typeName := getTypeName[T]()

	var mappingErr *MappingError
	require.ErrorAs(t, err, &mappingErr)
	assert.Equal(t, "invalid", mappingErr.Sheet)
	assert.Equal(t, "A2", mappingErr.Cell)
	assert.Equal(t, typeName+".Value", mappingErr.Field)

	var convertErr *ConvertError
	require.ErrorAs(t, err, &convertErr)
	assert.Equal(t, typ, convertErr.Typ)
	assert.Equal(t, "invalid", convertErr.CV)
}

type allT struct {
	Bools    *boolsT
	Strings  *stringsT
	Ints     *intsT
	Int8s    *int8sT
	Int16s   *int16sT
	Int32s   *int32sT
	Int64s   *int64sT
	Uints    *uintsT
	Uint8s   *uint8sT
	Uint16s  *uint16sT
	Uint32s  *uint32sT
	Uint64s  *uint64sT
	Float32s *float32sT
	Float64s *float64sT
	Times    *timesT
}

type nestedT struct {
	L1   *nestedL1
	Done <-chan struct{} `gsheets:"-"` // <- must be ignored, would cause ErrUnsupportedType otherwise
	F4   int             `gsheets:"f4"`
}

type nestedL1 struct {
	F1 int `gsheets:"f1"`
	L2 nestedL2
}

type nestedL2 struct {
	F2 int `gsheets:"f2"`
	L3 *nestedL3
}
type nestedL3 struct {
	F3 int `gsheets:"f3"`
}

type boolsT struct {
	Value bool  `sheets:"boolsT_value"`
	Ptr   *bool `sheets:"boolsT_ptr"`
}

type stringsT struct {
	Value string  `sheets:"stringsT_value"`
	Ptr   *string `sheets:"stringsT_ptr"`
}

type intsT struct {
	Value int  `sheets:"intsT_value"`
	Ptr   *int `sheets:"intsT_ptr"`
}
type int8sT struct {
	Value int8  `sheets:"int8sT_value"`
	Ptr   *int8 `sheets:"int8sT_ptr"`
}

type int16sT struct {
	Value int16  `sheets:"int16sT_value"`
	Ptr   *int16 `sheets:"int16sT_ptr"`
}
type int32sT struct {
	Value int32  `sheets:"int32sT_value"`
	Ptr   *int32 `sheets:"int32sT_ptr"`
}

type int64sT struct {
	Value int64  `sheets:"int64sT_value"`
	Ptr   *int64 `sheets:"int64sT_ptr"`
}
type uintsT struct {
	Value uint  `sheets:"uintsT_value"`
	Ptr   *uint `sheets:"uintsT_ptr"`
}
type uint8sT struct {
	Value uint8  `sheets:"uint8sT_value"`
	Ptr   *uint8 `sheets:"uint8sT_ptr"`
}

type uint16sT struct {
	Value uint16  `sheets:"uint16sT_value"`
	Ptr   *uint16 `sheets:"uint16sT_ptr"`
}
type uint32sT struct {
	Value uint32  `sheets:"uint32sT_value"`
	Ptr   *uint32 `sheets:"uint32sT_ptr"`
}

type uint64sT struct {
	Value uint64  `sheets:"uint64sT_value"`
	Ptr   *uint64 `sheets:"uint64sT_ptr"`
}
type float32sT struct {
	Value float32  `sheets:"float32sT_value"`
	Ptr   *float32 `sheets:"float32sT_ptr"`
}

type float64sT struct {
	Value float64  `sheets:"float64sT_value"`
	Ptr   *float64 `sheets:"float64sT_ptr"`
}

type timesT struct {
	Value time.Time  `sheets:"timesT_value"`
	Ptr   *time.Time `sheets:"timesT_ptr"`
}

// _svc is an empty Google Sheets service
var _svc = &sheets.Service{}

var fetcherError = fmt.Errorf("this must never be called")

func errorFetcher(Config) (*sheets.ValueRange, error) {
	return nil, fetcherError
}

func boolsFetcher(Config) (*sheets.ValueRange, error) {
	return &sheets.ValueRange{
		Values: [][]any{
			{"boolsT_value", "boolsT_ptr"},
			{"true", "1"},
			{"False"},
			{"", "0"},
			{"", ""},
		},
	}, nil
}

func stringsFetcher(Config) (*sheets.ValueRange, error) {
	return &sheets.ValueRange{
		Values: [][]any{
			{"stringsT_value", "stringsT_ptr"},
			{"foo", "bar"},
			{"baz"},
			{"", "boo"},
			{"", ""},
		},
	}, nil
}

func timesFetcher(Config) (*sheets.ValueRange, error) {
	return &sheets.ValueRange{
		Values: [][]any{
			{"timesT_value", "timesT_ptr"},
			{
				time.Date(2024, time.December, 19, 17, 35, 8, 0, time.UTC).Format(time.DateTime),
				time.Date(2020, time.November, 11, 11, 11, 11, 0, time.UTC).Format(time.DateTime),
			},
			{time.Date(2021, time.March, 9, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)},
			{"", time.Date(2016, time.January, 6, 0, 0, 0, 0, time.UTC).Format("2.1.2006")},
			{"", ""},
		},
	}, nil
}

func makeIntsFetcher(bitSize string) fetchFN {
	prefix := fmt.Sprintf("int%ssT_", bitSize)

	values := map[string][4]string{
		"": {"1337", "4711", "1887", "-6996"},
		"8": {
			fmt.Sprint(math.MinInt8),
			fmt.Sprint(math.MaxInt8),
			fmt.Sprint(math.MaxInt8 / 4),
			fmt.Sprint(math.MinInt8 / 2),
		},
		"16": {
			fmt.Sprint(math.MinInt16),
			fmt.Sprint(math.MaxInt16),
			fmt.Sprint(math.MaxInt16 / 4),
			fmt.Sprint(math.MinInt16 / 2),
		},
		"32": {
			fmt.Sprint(math.MinInt32),
			fmt.Sprint(math.MaxInt32),
			fmt.Sprint(math.MaxInt32 / 4),
			fmt.Sprint(math.MinInt32 / 2),
		},
		"64": {
			fmt.Sprint(int64(math.MinInt64)),
			fmt.Sprint(int64(math.MaxInt64)),
			fmt.Sprint(int64(math.MaxInt64) / 4),
			fmt.Sprint(int64(math.MinInt64) / 2),
		},
	}

	vals := values[bitSize]

	return func(Config) (*sheets.ValueRange, error) {
		return &sheets.ValueRange{
			Values: [][]any{
				{prefix + "value", prefix + "ptr"},
				{vals[0], vals[1]},
				{vals[2]},
				{"", vals[3]},
				{"", ""},
			},
		}, nil
	}
}

func makeUintsFetcher(bitSize string) fetchFN {
	prefix := fmt.Sprintf("uint%ssT_", bitSize)

	values := map[string][4]string{
		"": {"1337", "4711", "1887", "6996"},
		"8": {
			fmt.Sprint(math.MaxUint8 / 2),
			fmt.Sprint(math.MaxUint8),
			fmt.Sprint(math.MaxUint8 / 8),
			fmt.Sprint(math.MaxUint8 / 4),
		},
		"16": {
			fmt.Sprint(math.MaxUint16 / 2),
			fmt.Sprint(math.MaxUint16),
			fmt.Sprint(math.MaxUint16 / 8),
			fmt.Sprint(math.MaxUint16 / 4),
		},
		"32": {
			fmt.Sprint(uint32(math.MaxUint32) / 2),
			fmt.Sprint(uint32(math.MaxUint32)),
			fmt.Sprint(uint32(math.MaxUint32) / 8),
			fmt.Sprint(uint32(math.MaxUint32) / 4),
		},
		"64": {
			fmt.Sprint(uint64(math.MaxUint64) / 2),
			fmt.Sprint(uint64(math.MaxUint64)),
			fmt.Sprint(uint64(math.MaxUint64) / 8),
			fmt.Sprint(uint64(math.MaxUint64) / 4),
		},
	}

	vals := values[bitSize]

	return func(Config) (*sheets.ValueRange, error) {
		return &sheets.ValueRange{
			Values: [][]any{
				{prefix + "value", prefix + "ptr"},
				{vals[0], vals[1]},
				{vals[2]},
				{"", vals[3]},
				{"", ""},
			},
		}, nil
	}
}

func makeFloatsFetcher(bitSize string) fetchFN {
	prefix := fmt.Sprintf("float%ssT_", bitSize)

	values := map[string][4]string{
		"32": {
			fmt.Sprint(float32(math.SmallestNonzeroFloat32)),
			fmt.Sprint(float32(math.MaxFloat32)),
			fmt.Sprint(float32(math.MaxFloat32) / 2),
			fmt.Sprint(float32(math.SmallestNonzeroFloat32) * 4),
		},
		"64": {
			fmt.Sprint(math.SmallestNonzeroFloat64),
			fmt.Sprint(math.MaxFloat64),
			fmt.Sprint(math.MaxFloat64 / 2),
			fmt.Sprint(math.SmallestNonzeroFloat64 * 4),
		},
	}

	vals := values[bitSize]

	return func(Config) (*sheets.ValueRange, error) {
		return &sheets.ValueRange{
			Values: [][]any{
				{prefix + "value", prefix + "ptr"},
				{vals[0], vals[1]},
				{vals[2]},
				{"", vals[3]},
				{"", ""},
			},
		}, nil
	}
}

func parseValidationFetcher(Config) (*sheets.ValueRange, error) {
	return &sheets.ValueRange{
		Values: [][]any{
			{"Value", "Ptr"},
			{"invalid", ""},
			{"", "invalid"},
		},
	}, nil
}

func ptrTo[T any](v T) *T {
	return &v
}

func getTypeName[T any]() string {
	t := reflect.TypeFor[T]()
	name := t.PkgPath()
	if name != "" {
		name += "/"
	}
	name += t.Name()

	return name
}
