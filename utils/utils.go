package utils

import "strings"

func MapString(src map[string]any, key string) (string, bool) {
	if val, ok := src[key]; ok {
		if v, ok := val.(string); ok {
			return v, true
		}
	}
	return "", false
}
func MapStringArray(src map[string]any, key string) ([]string, bool) {
	if val, ok := src[key]; ok {
		switch v := val.(type) {
		case string:
			return []string{v}, true
		case []string:
			return v, true
		}
	}
	return []string{}, false
}

func Trim(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), `"`), `"`)
}

// Read a space seperated line of key value pairs seperated by "=", and return a map
func ReadMap(line string, delim byte) map[string]any {
	out := make(map[string]any)
	needEqual := true
	quoted := false
	key := ""
	tmp := strings.Builder{}
	for i := 0; i < len(line); i++ {
		if needEqual {
			if line[i] == '=' {
				if tmp.Len() == 0 {
					continue
				}
				key = tmp.String()
				tmp.Reset()
				needEqual = false
			} else {
				tmp.WriteByte(line[i])
			}
		} else {
			if !quoted && line[i] == delim {
				needEqual = true
				addUpdateKey(out, key, tmp.String())
				tmp.Reset()
				key = ""
				continue
			}

			tmp.WriteByte(line[i])
			if line[i] == '"' {
				quoted = !quoted
			}
		}
	}
	if key != "" {
		addUpdateKey(out, key, tmp.String())
	}
	return out
}

func addUpdateKey(m map[string]any, key string, val string) {
	if old, ok := m[key]; ok {
		var arr []string
		switch v := old.(type) {
		case []string:
			arr = append(v, Trim(val))
		case string:
			arr = []string{v, Trim(val)}
		default:
			panic("invalid data type")
		}
		m[key] = arr
	} else {
		m[key] = Trim(val)
	}
}

func ToStringMap(in map[string]any) map[string]string {
	out := make(map[string]string)
	for key, val := range in {
		switch v := val.(type) {
		case string:
			out[key] = v
		case []string:
			if len(v) > 0 {
				out[key] = v[0]
			}
		}
	}
	return out
}
