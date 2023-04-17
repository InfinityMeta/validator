package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	// TODO: implement this
	var valRes string

	if len(v) == 1 {
		return v[0].Err.Error()
	}

	for _, e := range v {
		valRes = valRes + e.Err.Error() + "\n"
	}
	return valRes
}

func validateLen(v reflect.Value, lenStr string) ValidationError {

	var valErr ValidationError

	length, err := strconv.Atoi(lenStr)

	if err != nil {
		valErr.Err = ErrInvalidValidatorSyntax
		return valErr
	}

	switch v.Kind() {

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			valErr = validateLen(v.Index(i), lenStr)
			if valErr.Err != nil {
				return valErr
			}
		}

	case reflect.String:
		if len(v.String()) != length {
			valErr.Err = errors.New("length of string is not equal")
			return valErr
		}
	}

	return valErr

}

func validateMin(v reflect.Value, minStr string) ValidationError {

	var valErr ValidationError

	min, err := strconv.ParseInt(minStr, 10, 64)

	if err != nil {
		valErr.Err = ErrInvalidValidatorSyntax
		return valErr
	}

	switch v.Kind() {

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			valErr = validateMin(v.Index(i), minStr)
			if valErr.Err != nil {
				return valErr
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < min {
			valErr.Err = errors.New("value is less than allowed")
			return valErr
		}

	case reflect.String:
		if len(v.String()) == 0 {
			valErr.Err = errors.New("empty text")
			return valErr
		}

		if len(v.String()) < int(min) {
			valErr.Err = errors.New("len of string is less than allowed")
			return valErr
		}
	}

	return valErr

}

func validateMax(v reflect.Value, maxStr string) ValidationError {

	var valErr ValidationError

	max, err := strconv.ParseInt(maxStr, 10, 64)

	if err != nil {
		valErr.Err = ErrInvalidValidatorSyntax
		return valErr
	}

	switch v.Kind() {

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			valErr = validateMax(v.Index(i), maxStr)
			if valErr.Err != nil {
				return valErr
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > max {
			valErr.Err = errors.New("value is bigger than allowed")
			return valErr
		}

	case reflect.String:
		if len(v.String()) == 0 {
			valErr.Err = errors.New("empty text")
			return valErr
		}

		if len(v.String()) > int(max) {
			valErr.Err = errors.New("len of string is bigger than allowed")
			return valErr
		}
	}

	return valErr

}

func validateIn(v reflect.Value, inStr string) ValidationError {

	var valErr ValidationError

	var flag bool

	if inStr == "" {
		valErr.Err = ErrInvalidValidatorSyntax
		return valErr
	}

	searchSpace := strings.Split(inStr, ",")

	switch v.Kind() {

	case reflect.Slice:

		for i := 0; i < v.Len(); i++ {
			valErr = validateIn(v.Index(i), inStr)
			if valErr.Err != nil {
				return valErr
			}
			flag = true
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vInt := v.Int()
		for i := range searchSpace {
			val, err := strconv.ParseInt(searchSpace[i], 10, 64)
			if err != nil {
				valErr.Err = ErrInvalidValidatorSyntax
				return valErr
			}
			if vInt == val {
				flag = true
			}
		}

	case reflect.String:
		vString := v.String()
		for i := range searchSpace {
			if vString == searchSpace[i] {
				flag = true
			}
		}
	}

	if !flag {
		valErr.Err = errors.New("value not in a valid set")
		return valErr
	}

	return valErr

}

func Validate(v any) error {

	var valErrs ValidationErrors
	var valErr ValidationError

	sv := reflect.Indirect(reflect.ValueOf(v))

	if sv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	for i := 0; i < sv.NumField(); i++ {

		typeField := sv.Type().Field(i)

		tag := typeField.Tag.Get("validate")

		if tag == "" {
			continue
		}

		if !typeField.IsExported() {
			valErr.Err = ErrValidateForUnexportedFields
			valErrs = append(valErrs, valErr)
			return valErrs
		}

		parts := strings.Split(tag, ":")

		key, value := parts[0], parts[1]

		switch key {

		case "len":
			valErr = validateLen(sv.Field(i), value)
		case "min":
			valErr = validateMin(sv.Field(i), value)
		case "max":
			valErr = validateMax(sv.Field(i), value)
		case "in":
			valErr = validateIn(sv.Field(i), value)
		}

		if valErr.Err != nil {

			if valErr.Err != ErrInvalidValidatorSyntax {
				valErr.Err = fmt.Errorf("validation error: field %s: %w", typeField.Name, valErr.Err)
			}

			valErrs = append(valErrs, valErr)
		}

	}

	if len(valErrs) > 0 {
		return valErrs
	}

	return nil

}
