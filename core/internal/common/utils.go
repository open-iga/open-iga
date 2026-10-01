package common

import (
	"context"
	"log/slog"
	"reflect"
)

func StructToMap(strukt interface{}) map[string]interface{} {
	value := reflect.ValueOf(strukt)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	out := make(map[string]interface{}, value.NumField())

	for i := 0; i < value.NumField(); i++ {
		out[value.Type().Field(i).Name] = value.Field(i).Interface()
	}

	return out
}

func WithErrorLogged(ctx context.Context, logger *slog.Logger, cls func(ctx context.Context) error) {
	if err := cls(ctx); err != nil {
		logger.Error(err.Error())
	}
}
