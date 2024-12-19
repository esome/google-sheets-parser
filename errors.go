package gsheets

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	// ErrNoService is returned when no Google API service is registered to the Config.
	ErrNoService = errors.New("gsheets: no Google API service registered")
	// ErrNoSpreadSheetID is returned when no spreadsheet ID is provided to the parse call.
	ErrNoSpreadSheetID = errors.New("gsheets: no spreadsheet id provided")
	// ErrUnsupportedType is returned when the type of field is not supported.
	ErrUnsupportedType = errors.New("gsheets: unsupported type")
	// ErrNoMapping is returned when not a single field mapping is found.
	ErrNoMapping = errors.New("gsheets: no mapping found")
	// ErrFieldNotFoundInSheet is returned when a field is not found in the sheet.
	ErrFieldNotFoundInSheet = errors.New("gsheets: field not found in sheet")
	// ErrFieldNotFoundInStruct is returned when a field/column is not found in the struct.
	ErrFieldNotFoundInStruct = errors.New("gsheets: field not found in struct")
)

// InvalidDateTimeFormatError is returned when an invalid datetime format is encountered.
type InvalidDateTimeFormatError struct {
	CV      string
	Formats []string
}

func (e *InvalidDateTimeFormatError) Error() string {
	return fmt.Sprintf("gsheets: invalid datetime format in value %q, recognized formats are: [\"%v\"]", e.CV, strings.Join(e.Formats, `", "`))
}

// ConvertError is returned when a conversion error occurs.
type ConvertError struct {
	Typ reflect.Kind
	CV  string
	err error
}

func (e *ConvertError) Error() string {
	return fmt.Sprintf("gsheets: conversion error, could not convert value %q into Go type %q", e.CV, e.Typ.String())
}

func (e *ConvertError) Unwrap() error {
	return e.err
}

// MappingError is returned when an error is encountered during the mapping.
type MappingError struct {
	Sheet string
	Cell  string
	Field string
	err   error
}

func (e *MappingError) Error() string {
	return fmt.Sprintf("%s\n\tsheet: %q\n\tcell: %q\n\tfield: %q", e.err, e.Sheet, e.Cell, e.Field)
}

func (e *MappingError) Unwrap() error {
	return e.err
}
