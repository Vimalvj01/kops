package fi

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func String(s string) *string {
	return &s
}

func Bool(v bool) *bool {
	return &v
}

func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func Int(v int) *int {
	return &v
}

func IntValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func Int64(v int64) *int64 {
	return &v
}

func Int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func Uint64Value(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

func DebugPrint(o interface{}) string {
	if o == nil {
		return "<nil>"
	}
	if rh, ok := o.(*ResourceHolder); ok {
		if rh == nil {
			// Avoid go nil vs interface problems
			return "<nil>"
		}

		if rh.Resource == nil {
			return fmt.Sprintf("unknown resource %q", rh.Name)
		}
	}
	if resource, ok := o.(Resource); ok {
		if resource == nil {
			// Avoid go nil vs interface problems
			return "<nil>"
		}

		s, err := ResourceAsString(resource)
		if err != nil {
			return fmt.Sprintf("error converting resource to string: %v", err)
		}
		if len(s) >= 256 {
			s = s[:256] + "... (truncated)"
		}
		return s
	}

	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "<nil>"
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return "<?>"
	}
	o = v.Interface()
	if stringer, ok := o.(fmt.Stringer); ok {
		if stringer == nil {
			// Avoid go nil vs interface problems
			return "<nil>"
		}
		return stringer.String()
	}

	return fmt.Sprint(o)
}

func DebugAsJsonString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error marshalling: %v", err)
	}
	return string(data)
}

func DebugAsJsonStringIndent(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshalling: %v", err)
	}
	return string(data)
}
