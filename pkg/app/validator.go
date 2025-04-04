package app

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	v10 "github.com/go-playground/validator/v10"
)

type Validator struct{}

func (Validator) Validate(i interface{}) error {
	err := v10.New().Struct(i)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(v10.ValidationErrors)
	if !ok {
		return err
	}

	errs := []string{}
	for _, err := range validationErrors {
		key := buildPath(reflect.TypeOf(i).Elem(), prepareNamespace(err.Namespace()))
		e := fmt.Sprintf("%s=%s", key, err.Tag())
		if err.Param() != "" {
			e += "=" + err.Param()
		}
		errs = append(errs, e)
	}
	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, ","))
}

func prepareNamespace(namespace string) []string {
	namespace = strings.SplitN(namespace, ".", 2)[1]
	namespace = strings.ReplaceAll(
		strings.ReplaceAll(namespace, "[", "."),
		"]",
		"",
	)

	return strings.Split(namespace, ".")
}

// Build path of error with json tags
func buildPath(objectType reflect.Type, namespace []string) string {
	field := namespace[0]
	_, err := strconv.Atoi(field)
	if err == nil {
		if len(namespace) > 1 {
			return field + "." + buildPath(objectType.Elem(), namespace[1:])
		}
		return field
	}
	var f reflect.StructField
	if objectType.Kind() == reflect.Ptr {
		f, _ = objectType.Elem().FieldByName(field)
	} else {
		f, _ = objectType.FieldByName(field)
	}
	tag := getJSONTag(f.Tag)
	path := tag
	if len(namespace) > 1 {
		path += "." + buildPath(f.Type, namespace[1:])
	}

	return path
}

func getJSONTag(tag reflect.StructTag) string {
	return strings.Split(tag.Get("json"), ",")[0]
}
