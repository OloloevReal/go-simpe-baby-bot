package goenvs

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Parse(env interface{}) error {
	if env == nil {
		return errors.New("input env data is nil")
	}
	fnErr := func(env string) error {
		return fmt.Errorf("failed to lookup env variable %s", env)
	}

	s := reflect.ValueOf(env).Elem()
	//tOf := reflect.TypeOf(env)
	tOf := reflect.Indirect(s).Type()
	for i := 0; i < tOf.NumField(); i++ {
		tag := tOf.Field(i).Tag.Get("env")
		name := tOf.Field(i).Name
		if tag == "" {
			continue
		}
		value, ok := os.LookupEnv(tag)
		if !ok {
			return fnErr(tag)
		}
		if name == "NotifyType" {
			_ = name
		}
		f := s.FieldByName(name)
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueInt64, err := strconv.ParseInt(value, 10, f.Type().Bits())
			if err != nil {
				return err
			}
			f.SetInt(valueInt64)
		case reflect.Bool:
			valueBool, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			f.SetBool(valueBool)
		case reflect.String:
			f.SetString(value)
		case reflect.Slice:
			{
				//typ := tOf.Field(i).Type.Elem()
				typ := f.Type().Elem().Kind()
				if typ == reflect.String {
					values := strings.Split(value, ",")
					for _, v := range values {
						v = strings.Trim(v, " ")
						f.Set(reflect.Append(f, reflect.ValueOf(v)))
					}
				} else {
					return fmt.Errorf("unsupported type of slice %s=%s", name, typ.String())
				}
			}
		}
	}

	return nil
}
