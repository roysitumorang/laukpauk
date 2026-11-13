package sanitizers

import "reflect"

var (
	mimeTypes = map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
	}
)

// keepFormFields clears every field a form submission could not set
// (tagged form:"-"), so a JSON body fills only the fields a form could.
// Fields tagged input:"json" are kept: they are JSON-only inputs.
func keepFormFields(v any) {
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		if !field.IsExported() || field.Tag.Get("input") == "json" {
			continue
		}
		if field.Tag.Get("form") == "-" {
			rv.Field(i).SetZero()
		}
	}
}
