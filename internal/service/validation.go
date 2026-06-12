package service

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func ValidateData(entity any) error {
	v := reflect.ValueOf(entity)

	if v.Kind() != reflect.Pointer {
		return errors.New("need struct pointer")
	}

	unreqFieldsID := verifyReqFields(v)

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return errors.New("incorrect data type")
	}

	t := v.Type()

	for i, j := 0, 0; i < v.NumField(); i++ {
		if unreqFieldsID[j] == i {
			j++
			continue
		}

		field := v.Field(i)
		fieldName := t.Field(i).Name

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.String {
			var fieldValue string

			if fieldName == "PasswordHash" {
				continue
			}

			fieldValue = strings.ToLower(strings.TrimSpace(field.String()))

			if field.String() == "" {
				return fmt.Errorf("%s is required", fieldName)
			}

			field.SetString(fieldValue)
		}
	}

	return nil
}

func verifyReqFields(v reflect.Value) (ids []int) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		switch field.Kind() {
		case reflect.Pointer, reflect.Slice:
			if field.IsNil() {
				ids = append(ids, i)
			}
		}
	}

	return
}
