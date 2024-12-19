package gsheets

import "errors"

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
	// ErrInvalidDateTimeFormat is returned when a datetime format is invalid, or not configured.
	ErrInvalidDateTimeFormat = errors.New("gsheets: invalid datetime format")
)
