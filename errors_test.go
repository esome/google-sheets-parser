package gsheets

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertError_Error(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &ConvertError{
		CV:  "test",
		Typ: reflect.Bool,
		err: innerErr,
	}
	assert.Equal(t, `gsheets: conversion error, could not convert value "test" into Go type "bool"`, err.Error())
}

func TestConvertError_Unwrap(t *testing.T) {
	expectedErr := errors.New("inner error")
	err := &ConvertError{err: expectedErr}
	assert.Equal(t, expectedErr, err.Unwrap())
}

func TestInvalidDateTimeFormatError_Error(t *testing.T) {
	err := &InvalidDateTimeFormatError{
		CV:      "2024-12-31",
		Formats: []string{"2.1.2006", "1/2/2006"},
	}
	assert.Equal(t, `gsheets: invalid datetime format in value "2024-12-31", recognized formats are: ["2.1.2006", "1/2/2006"]`, err.Error())
}

func TestMappingError_Error(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &MappingError{
		Sheet: "test",
		Cell:  "A1",
		Field: "Type.Field",
		err:   innerErr,
	}

	const msg = `inner error
	sheet: "test"
	cell: "A1"
	field: "Type.Field"`

	assert.Equal(t, msg, err.Error())
}

func TestMappingError_Unwrap(t *testing.T) {
	expectedErr := errors.New("inner error")
	err := &MappingError{err: expectedErr}
	assert.Equal(t, expectedErr, err.Unwrap())
}
