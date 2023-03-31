package config

import (
	"strconv"
)

// Bool looks up a boolean value and returns false when not found or failure.
func Bool(section string, k string) bool {
	return BoolOpt(section, k, false)
}

// BoolOpt looks up a boolean value and returns the given
// fallback if not found or when failure.
func BoolOpt(section string, k string, fallback bool) bool {
	b, e := db().queryOne(section, k)
	if e != nil {
		return fallback
	}

	v, e := strconv.ParseBool(string(b))
	if e != nil {
		return fallback
	}

	return v
}

// Float looks up a float value and returns 0.0 when not found or failure.
func Float(section string, k string) float64 {
	return FloatOpt(section, k, 0.0)
}

// FloatOpt looks up a float value and returns the given
// fallback if not found or when failure.
func FloatOpt(section string, k string, fallback float64) float64 {
	b, e := db().queryOne(section, k)
	if e != nil {
		return fallback
	}

	v, e := strconv.ParseFloat(string(b), 64)
	if e != nil {
		return fallback
	}

	return v
}

// Int looks up an integer value and returns 0 when not found or failure.
func Int(section string, k string) int {
	return IntOpt(section, k, 0)
}

// IntOpt looks up an integer value and returns the given
// fallback if not found or when failure.
func IntOpt(section string, k string, fallback int) int {
	b, e := db().queryOne(section, k)
	if e != nil {
		return fallback
	}

	v, e := strconv.ParseInt(string(b), 10, 64)
	if e != nil {
		return fallback
	}

	return int(v)
}

// String looks up a string value and returns "" when not found or failure.
func String(section string, k string) string {
	return StringOpt(section, k, "")
}

// StringOpt looks up a string value and returns the given
// fallback if not found or when failure.
func StringOpt(section string, k string, fallback string) string {
	b, e := db().queryOne(section, k)
	if e != nil {
		return fallback
	}

	return string(b)
}

// SetBool saves a boolean value for the given key.
func SetBool(section string, k string, v bool) {
	_ = db().upsert(section, k, []byte(strconv.FormatBool(v)))
}

// SetFloat saves a float value for the given key.
func SetFloat(section string, k string, v float64) {
	_ = db().upsert(section, k, []byte(strconv.FormatFloat(v, 'f', -1, 64)))
}

// SetInt saves an integer value for the given key.
func SetInt(section string, k string, v int) {
	_ = db().upsert(section, k, []byte(strconv.FormatInt(int64(v), 10)))
}

// SetString saves a string value for the given key.
func SetString(section string, k string, v string) {
	_ = db().upsert(section, k, []byte(v))
}

// Remove saves a string value for the given key.
// Does nothing if key not found.
func Remove(section string, k string) {
	_ = db().deleteOne(section, k)
}

// Watch adds a listener for the given key.
func Watch(section, key string, fn func(v []byte)) {

}
