package common

import "reflect"

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
