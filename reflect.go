package googlesheetsparser

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

type convertFunc func(string, []string) (reflect.Value, bool, error)
type mapping struct {
	field        reflect.StructField
	convert      convertFunc
	initEmbedPtr func(reflect.Value)
	colIndex     int
	colName      string
	err          error
}

func createMappings(t reflect.Type, captions []any, opts Options) ([]*mapping, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, t.Kind().String())
	}

	// first we determine the column names and their corresponding fields
	colNames := make(map[string]int, len(captions))
	for colIdx, cellIf := range captions {
		cell := cellIf.(string)
		if cell == "" {
			break
		}
		colNames[cell] = colIdx
	}

	// then we read the tags and create the mappings
	fields := readTags("sheets", t, nil, nil)

	// next we set the column index for each mapping
	mapped := make([]*mapping, 0, len(fields))
	for _, m := range fields {
		if idx, ok := colNames[m.colName]; ok {
			m.colIndex = idx
			if m.err != nil {
				return nil, m.err
			}
			mapped = append(mapped, m)
			delete(colNames, m.colName)
			continue
		}
		if !opts.AllowSkipFields {
			return nil, fmt.Errorf("%w: %q", ErrFieldNotFoundInSheet, m.colName)
		}
	}

	// finally we check if there are any columns left, and raise an error if it is not allowed to skip them
	if len(colNames) > 0 && !opts.AllowSkipColumns {
		errs := make([]error, 0, len(colNames))
		// todo: sort by column index
		for colName := range colNames {
			errs = append(errs, fmt.Errorf("%w: %q", ErrFieldNotFoundInStruct, colName))
		}
		return nil, errors.Join(errs...)
	}

	return mapped, nil
}

func readTags(tagName string, t reflect.Type, index []int, parentInit func(reflect.Value)) []*mapping {
	out := make([]*mapping, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		f.Index = append(slices.Clone(index), f.Index...)
		m := mapping{
			field:        f,
			colName:      f.Name,
			colIndex:     -1,
			initEmbedPtr: parentInit,
		}

		if v, ok := f.Tag.Lookup(tagName); ok {
			name := v
			if s := strings.Split(v, ","); len(s) > 1 {
				name = s[0]
			}

			if name == "-" {
				continue
			}

			m.colName = name
		}

		switch field, isPointer := indirect(f.Type); field.Kind() {
		case reflect.Struct:
			if field == timeType {
				if isPointer {
					m.convert = convertTimeP
					break
				}
				m.convert = convertTime
				break
			}
			initEmbedPtr := parentInit
			if isPointer {
				initEmbedPtr = func(ref reflect.Value) {
					if parentInit != nil {
						parentInit(ref)
					}
					f := ref.FieldByIndex(f.Index)
					if f.IsNil() {
						f.Set(reflect.New(field))
					}
				}
			}

			out = append(out, readTags(tagName, field, f.Index, initEmbedPtr)...)
			continue
		case reflect.String:
			if isPointer {
				m.convert = convertStringP
				break
			}
			m.convert = convertString
		case reflect.Int:
			if isPointer {
				m.convert = convertIntP
				break
			}
			m.convert = convertInt
		case reflect.Int8:
			if isPointer {
				m.convert = makeConvertIntxP[int8](8)
				break
			}
			m.convert = makeConvertIntx[int8](8)
		case reflect.Int16:
			if isPointer {
				m.convert = makeConvertIntxP[int16](16)
				break
			}
			m.convert = makeConvertIntx[int16](16)
		case reflect.Int32:
			if isPointer {
				m.convert = makeConvertIntxP[int32](32)
				break
			}
			m.convert = makeConvertIntx[int32](32)
		case reflect.Int64:
			if isPointer {
				m.convert = makeConvertIntxP[int64](64)
				break
			}
			m.convert = makeConvertIntx[int64](64)
		case reflect.Uint:
			if isPointer {
				m.convert = makeConvertUintP[uint](0)
				break
			}
			m.convert = makeConvertUint[uint](0)
		case reflect.Uint8:
			if isPointer {
				m.convert = makeConvertUintP[uint8](8)
				break
			}
			m.convert = makeConvertUint[uint8](8)
		case reflect.Uint16:
			if isPointer {
				m.convert = makeConvertUintP[uint16](16)
				break
			}
			m.convert = makeConvertUint[uint16](16)
		case reflect.Uint32:
			if isPointer {
				m.convert = makeConvertUintP[uint32](32)
				break
			}
			m.convert = makeConvertUint[uint32](32)
		case reflect.Uint64:
			if isPointer {
				m.convert = makeConvertUintP[uint64](64)
				break
			}
			m.convert = makeConvertUint[uint64](64)
		case reflect.Float32:
			if isPointer {
				m.convert = makeConvertFloatP[float32](32)
				break
			}
			m.convert = makeConvertFloat[float32](32)
		case reflect.Float64:
			if isPointer {
				m.convert = makeConvertFloatP[float64](64)
				break
			}
			m.convert = makeConvertFloat[float64](64)
		case reflect.Bool:
			if isPointer {
				m.convert = convertBoolP
				break
			}
			m.convert = convertBool

			reflect.New(field).Elem().IsNil()

		default:
			m.err = fmt.Errorf("%w: field %q of type %q is unsupported", ErrUnsupportedType, f.Name, field.Kind().String())
		}

		m.convert = wrapEmpty(f.Type, m.convert)
		out = append(out, &m)
	}

	return out
}

func indirect(t reflect.Type) (reflect.Type, bool) {
	isPointer := false
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		isPointer = true
	}
	return t, isPointer
}

var timeType = reflect.TypeFor[time.Time]()

var errVal reflect.Value

func wrapEmpty(p reflect.Type, f convertFunc) convertFunc {
	zeroVal := reflect.Zero(p)
	return func(cv string, dateTimeValues []string) (reflect.Value, bool, error) {
		if cv == "" {
			return zeroVal, false, nil
		}
		return f(cv, dateTimeValues)
	}
}

func convertString(cv string, _ []string) (reflect.Value, bool, error) {
	return reflect.ValueOf(cv), true, nil
}

func convertStringP(cv string, _ []string) (reflect.Value, bool, error) {
	return reflect.ValueOf(&cv), true, nil
}

func convertInt(cv string, _ []string) (reflect.Value, bool, error) {
	i, err := strconv.Atoi(cv)
	if err != nil {
		return errVal, false, err
	}
	return reflect.ValueOf(i), true, nil
}

func convertIntP(cv string, _ []string) (reflect.Value, bool, error) {
	i, err := strconv.Atoi(cv)
	if err != nil {
		return errVal, false, err
	}
	return reflect.ValueOf(&i), true, nil
}

func makeConvertIntx[T int8 | int16 | int32 | int64](bitSize int) convertFunc {
	var emptyT T
	return func(cv string, _ []string) (reflect.Value, bool, error) {
		i, err := strconv.ParseInt(cv, 10, bitSize)
		if err != nil {
			return errVal, false, err
		}
		v := T(i)
		return reflect.ValueOf(v), v != emptyT, nil
	}
}

func makeConvertIntxP[T int8 | int16 | int32 | int64](bitSize int) convertFunc {
	return func(cv string, _ []string) (reflect.Value, bool, error) {
		i, err := strconv.ParseInt(cv, 10, bitSize)
		if err != nil {
			return errVal, false, err
		}
		v := T(i)
		return reflect.ValueOf(&v), true, nil
	}
}

func makeConvertUint[T uint | uint8 | uint16 | uint32 | uint64](bitSize int) convertFunc {
	var emptyT T
	return func(cv string, _ []string) (reflect.Value, bool, error) {
		i, err := strconv.ParseUint(cv, 10, bitSize)
		if err != nil {
			return errVal, false, err
		}
		v := T(i)
		return reflect.ValueOf(v), v != emptyT, nil
	}
}

func makeConvertUintP[T uint | uint8 | uint16 | uint32 | uint64](bitSize int) convertFunc {
	return func(cv string, _ []string) (reflect.Value, bool, error) {
		i, err := strconv.ParseUint(cv, 10, bitSize)
		if err != nil {
			return errVal, false, err
		}
		v := T(i)
		return reflect.ValueOf(&v), true, nil
	}
}

func makeConvertFloat[T float32 | float64](bitSize int) convertFunc {
	var emptyT T
	return func(cv string, _ []string) (reflect.Value, bool, error) {
		f, err := strconv.ParseFloat(cv, bitSize)
		if err != nil {
			return errVal, false, err
		}
		v := T(f)
		return reflect.ValueOf(v), v != emptyT, nil
	}
}

func makeConvertFloatP[T float32 | float64](bitSize int) convertFunc {
	return func(cv string, _ []string) (reflect.Value, bool, error) {
		f, err := strconv.ParseFloat(cv, bitSize)
		if err != nil {
			return errVal, false, err
		}
		v := T(f)
		return reflect.ValueOf(&v), true, nil
	}
}

func convertBool(cv string, _ []string) (reflect.Value, bool, error) {
	b, err := strconv.ParseBool(cv)
	if err != nil {
		return errVal, false, err
	}
	return reflect.ValueOf(b), b, nil
}

func convertBoolP(cv string, _ []string) (reflect.Value, bool, error) {
	b, err := strconv.ParseBool(cv)
	if err != nil {
		return errVal, false, err
	}
	return reflect.ValueOf(&b), true, nil
}

func parseTime(cv string, dateTimeFormats []string) (time.Time, error) {
	for _, dateTimeFormat := range dateTimeFormats {
		t, err := time.Parse(dateTimeFormat, cv)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidDateTimeFormat, cv)
}

func convertTime(cv string, dateTimeFormats []string) (reflect.Value, bool, error) {
	t, err := parseTime(cv, dateTimeFormats)
	if err != nil {
		return errVal, false, err
	}
	return reflect.ValueOf(t), !t.IsZero(), nil
}

func convertTimeP(cv string, dateTimeFormats []string) (reflect.Value, bool, error) {
	t, err := parseTime(cv, dateTimeFormats)
	if err != nil {
		return errVal, false, err
	}
	return reflect.ValueOf(&t), !t.IsZero(), nil
}
