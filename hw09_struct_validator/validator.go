package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotStruct = errors.New("input is not a struct")
	ErrLen       = errors.New("invalid length")
	ErrRegexp    = errors.New("does not match regexp")
	ErrIn        = errors.New("not in allowed set")
	ErrMin       = errors.New("less than min")
	ErrMax       = errors.New("greater than max")
)

type ValidationError struct {
	Field string
	Err   error
}

func (e ValidationError) Error() string { return e.Field + ": " + e.Err.Error() }

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	s := make([]string, len(e))
	for i, v := range e {
		s[i] = v.Error()
	}
	return strings.Join(s, "\n")
}

// intRules и strRules: правила для int и string, возвращают nil или ошибку валидации;
// ошибка парсинга параметра (неверный тэг) — программная, возвращается как есть.
var (
	intRules = map[string]func(x int64, param string) error{
		"min": func(x int64, p string) error {
			m, err := strconv.ParseInt(p, 10, 64)
			if err != nil || x >= m {
				return err
			}
			return ErrMin
		},
		"max": func(x int64, p string) error {
			m, err := strconv.ParseInt(p, 10, 64)
			if err != nil || x <= m {
				return err
			}
			return ErrMax
		},
		"in": func(x int64, p string) error {
			for _, s := range strings.Split(p, ",") {
				m, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return err
				}
				if m == x {
					return nil
				}
			}
			return ErrIn
		},
	}

	strRules = map[string]func(s, param string) error{
		"len": func(s, p string) error {
			n, err := strconv.Atoi(p)
			if err != nil || len(s) == n {
				return err
			}
			return ErrLen
		},
		"regexp": func(s, p string) error {
			re, err := regexp.Compile(p)
			if err != nil {
				return err
			}
			if re.MatchString(s) {
				return nil
			}
			return ErrRegexp
		},
		"in": func(s, p string) error {
			for _, v := range strings.Split(p, ",") {
				if v == s {
					return nil
				}
			}
			return ErrIn
		},
	}
)

// Validate проверяет экспортируемые поля структуры по тэгу validate.
func Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	var errs ValidationErrors
	for i, n := 0, rv.NumField(); i < n; i++ {
		f, fv := rv.Type().Field(i), rv.Field(i)
		tag, ok := f.Tag.Lookup("validate")
		if !ok || !f.IsExported() {
			continue
		}
		if err := validateField(f.Name, tag, fv, &errs); err != nil {
			return err
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// validateField применяет правила (через "|") к значению; для слайса — к каждому элементу.
// Ошибки валидации накапливаются, программные возвращаются сразу.
func validateField(field, tag string, v reflect.Value, errs *ValidationErrors) error {
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if err := validateField(field, tag, v.Index(i), errs); err != nil {
				return err
			}
		}
		return nil
	}
	for _, rule := range strings.Split(tag, "|") {
		name, param, _ := strings.Cut(rule, ":")
		err := check(name, param, v)
		if err == nil {
			continue
		}
		if !isRuleErr(err) {
			return err
		}
		*errs = append(*errs, ValidationError{Field: field, Err: err})
	}
	return nil
}

func check(name, param string, v reflect.Value) error {
	switch k := v.Kind(); {
	case k >= reflect.Int && k <= reflect.Int64:
		if r, ok := intRules[name]; ok {
			return r(v.Int(), param)
		}
	case k == reflect.String:
		if r, ok := strRules[name]; ok {
			return r(v.String(), param)
		}
	}
	return fmt.Errorf("unsupported rule %q for type %s", name, v.Kind())
}

func isRuleErr(err error) bool {
	switch {
	case errors.Is(err, ErrLen), errors.Is(err, ErrRegexp), errors.Is(err, ErrIn),
		errors.Is(err, ErrMin), errors.Is(err, ErrMax):
		return true
	}
	return false
}
