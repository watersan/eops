package opsv2tmpl

import (
	"fmt"
	"html/template"
	"strconv"
)

func Addition(a, b interface{}) (float64, error) {
	fa, fb, err := conv(a, b)
	return fa + fb, err
}
func Subtract(a, b interface{}) (float64, error) {
	fa, fb, err := conv(a, b)
	return fa - fb, err
}
func Multiply(a, b interface{}) (float64, error) {
	fa, fb, err := conv(a, b)
	return fa * fb, err
}
func Divide(a, b interface{}) (float64, error) {
	fa, fb, err := conv(a, b)
	return fa / fb, err
}
func Mod(a, b interface{}) (float64, error) {
	fa, fb, err := conv(a, b)
	return float64(int(fa) % int(fb)), err
}

func conv(a, b interface{}) (float64, float64, error) {
	var fa, fb float64
	var err error
	switch v := a.(type) {
	case int:
		fa = float64(v)
	case float64:
		fa = v
	case string:
		var ia int
		if ia, err = strconv.Atoi(v); err != nil {
			return fa, fb, err
		}
		fa = float64(ia)
	default:
		err = fmt.Errorf("a: %v is not number", a)
	}
	switch v := b.(type) {
	case int:
		fb = float64(v)
	case float64:
		fb = v
	case string:
		var ib int
		if ib, err = strconv.Atoi(v); err != nil {
			return fa, fb, err
		}
		fb = float64(ib)
	default:
		err = fmt.Errorf("b: %v is not number", b)
	}
	return fa, fb, err
}

var FuncMap = template.FuncMap{
	"Addition": Addition,
	"Subtract": Subtract,
	"Multiply": Multiply,
	"Divide":   Divide,
	"Mod":      Mod,
}
