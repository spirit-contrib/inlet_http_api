package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//some function copied from https://github.com/spf13/hugo

var funcMap template.FuncMap

func init() {
	funcMap = template.FuncMap{
		"getJSON":      GetJSON,
		"getenv":       func(varName string) string { return os.Getenv(varName) },
		"replaceMaps":  ReplaceMaps,
		"replace":      Replace,
		"regexCompile": regexp.Compile,
		"substr":       Substr,
		"split":        Split,
		"in":           In,
		"toStr":        ToStr,
		"dateFormat":   DateFormat,
		"now":          time.Now,
		"localtime":    time.Now().Local,
		"newDict":      NewDict,
		"scanf":        fmt.Sscanf,
	}
}

type RenderDict map[string]interface{}

func NewDict() RenderDict {
	return RenderDict{}
}

func (p RenderDict) Put(key string, val interface{}) bool {
	p[key] = val
	return true
}

func (p RenderDict) Get(key string) (val interface{}) {
	val, _ = p[key]
	return
}

func (p RenderDict) Exist(key string) (exist bool) {
	_, exist = p[key]
	return
}

func (p RenderDict) Del(key string) bool {
	if _, exist := p[key]; exist {
		delete(p, key)
		return true
	}
	return false
}

func ToStr(v interface{}) (str string) {
	return fmt.Sprintf("%v", v)
}

func GetJSON(v interface{}) (jsonStr string, err error) {
	var data []byte

	if data, err = json.Marshal(v); err != nil {
		return
	}

	jsonStr = string(data)

	return
}

func ReplaceMaps(v interface{}, key string, expr string, repl string) (ret interface{}, err error) {
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Slice:
		{
			items, _ := v.([]interface{})
			mapItems := []map[string]interface{}{}
			for _, item := range items {
				if mapItem, ok := item.(map[string]interface{}); ok {
					mapItems = append(mapItems, mapItem)
				} else {
					err = fmt.Errorf("the map's values type is not interface{}")
					return
				}
			}
			return replaceMapSlice(mapItems, key, expr, repl)
		}
	case reflect.Map:
		{
			if m, ok := v.(map[string]interface{}); ok {
				return replaceMap(m, key, expr, repl)
			} else {
				err = fmt.Errorf("the map's values type is not interface{}")
				return
			}
		}
	default:
		{
			err = fmt.Errorf("unsupport Kind of %v", kind)
			return
		}
	}
	return
}

func Replace(v interface{}, expr string, repl string) (ret interface{}, err error) {
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Slice:
		{
			strs, _ := v.([]interface{})
			strItems := []string{}
			for _, strV := range strs {
				if str, ok := strV.(string); ok {
					strItems = append(strItems, str)
				} else {
					err = fmt.Errorf("the map's values type is not string")
					return
				}
			}
			return replaceStringSlice(strItems, expr, repl)
		}
	case reflect.String:
		{
			str, _ := v.(string)
			return replaceString(str, expr, repl)
		}
	default:
		{
			err = fmt.Errorf("unsupport Kind of %v", kind)
			return
		}
	}
}

func replaceStringSlice(strs []string, expr string, repl string) (ret []string, err error) {
	if strs != nil {
		rets := make([]string, 0)
		for _, src := range strs {
			if rStr, e := replaceString(src, expr, repl); e != nil {
				err = e
				return
			} else {
				rets = append(rets, rStr)
			}
		}
		ret = rets
	}
	return
}

func replaceString(src string, expr string, repl string) (ret string, err error) {
	var reg *regexp.Regexp
	if reg, err = regexp.Compile(expr); err != nil {
		return
	}

	ret = reg.ReplaceAllString(src, repl)

	return
}

func replaceMapSlice(data []map[string]interface{}, key string, expr string, repl string) (ret []map[string]interface{}, err error) {
	if data != nil {
		retMap := make([]map[string]interface{}, 0)
		for _, item := range data {
			if rMap, e := replaceMap(item, key, expr, repl); e != nil {
				err = e
				return
			} else {
				retMap = append(retMap, rMap)
			}
		}
		ret = retMap
	}
	return
}

func replaceMap(data map[string]interface{}, key string, expr string, repl string) (ret map[string]interface{}, err error) {

	src := ""

	if val, exist := data[key]; !exist {
		return
	} else if strVal, ok := val.(string); ok {
		src = strVal
	} else {
		err = fmt.Errorf("value of %s's type is not string", key)
		return
	}

	retval := ""
	if retval, err = replaceString(src, expr, repl); err != nil {
		return
	}

	ret = make(map[string]interface{})

	for k, v := range data {
		ret[k] = v
	}

	ret[key] = retval

	return
}

func DateFormat(layout string, v interface{}) (string, error) {
	t, err := toTimeE(v)

	if err != nil {
		return "", err
	}
	return t.Format(layout), nil
}

func Split(a interface{}, delimiter string) ([]string, error) {
	aStr := ""
	if str, ok := a.(string); !ok {
		return nil, fmt.Errorf("type of a is not string")
	} else {
		aStr = str
	}

	return strings.Split(aStr, delimiter), nil
}

func Substr(a interface{}, nums ...interface{}) (string, error) {

	aStr := ""
	if str, ok := a.(string); !ok {
		return "", fmt.Errorf("type of a is not string")
	} else {
		aStr = str
	}

	var start, length int
	toInt := func(v interface{}, message string) (int, error) {
		switch i := v.(type) {
		case int:
			return i, nil
		case int8:
			return int(i), nil
		case int16:
			return int(i), nil
		case int32:
			return int(i), nil
		case int64:
			return int(i), nil
		default:
			return 0, fmt.Errorf(message)
		}
	}

	var err error
	switch len(nums) {
	case 0:
		return "", fmt.Errorf("too less arguments")
	case 1:
		if start, err = toInt(nums[0], "start argument must be integer"); err != nil {
			return "", err
		}
		length = len(aStr)
	case 2:
		if start, err = toInt(nums[0], "start argument must be integer"); err != nil {
			return "", err
		}
		if length, err = toInt(nums[1], "length argument must be integer"); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("too many arguments")
	}

	if start < -len(aStr) {
		start = 0
	}
	if start > len(aStr) {
		return "", fmt.Errorf("start position out of bounds for %d-byte string", len(aStr))
	}

	var s, e int
	if start >= 0 && length >= 0 {
		s = start
		e = start + length
	} else if start < 0 && length >= 0 {
		s = len(aStr) + start - length + 1
		e = len(aStr) + start + 1
	} else if start >= 0 && length < 0 {
		s = start
		e = len(aStr) + length
	} else {
		s = len(aStr) + start
		e = len(aStr) + length
	}

	if s > e {
		return "", fmt.Errorf("calculated start position greater than end position: %d > %d", s, e)
	}
	if e > len(aStr) {
		e = len(aStr)
	}

	return aStr[s:e], nil
}

func In(l interface{}, v interface{}) bool {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < lv.Len(); i++ {
			lvv := lv.Index(i)
			switch lvv.Kind() {
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true
					}
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
			return true
		}
	}
	return false
}

func toTimeE(i interface{}) (tim time.Time, err error) {
	i = indirect(i)

	switch s := i.(type) {
	case time.Time:
		return s, nil
	case string:
		d, e := stringToDate(s)
		if e == nil {
			return d, nil
		}
		return time.Time{}, fmt.Errorf("Could not parse Date/Time format: %v\n", e)
	default:
		return time.Time{}, fmt.Errorf("Unable to Cast %#v to Time\n", i)
	}
}

func stringToDate(s string) (time.Time, error) {
	return parseDateWith(s, []string{
		time.RFC3339,
		"2006-01-02T15:04:05", // iso8601 without timezone
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04:05Z07:00",
		"02 Jan 06 15:04 MST",
		"2006-01-02",
		"02 Jan 2006",
	})
}

func parseDateWith(s string, dates []string) (d time.Time, e error) {
	for _, dateType := range dates {
		if d, e = time.Parse(dateType, s); e == nil {
			return
		}
	}
	return d, fmt.Errorf("Unable to parse date: %s", s)
}

func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
