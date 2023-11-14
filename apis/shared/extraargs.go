package shared

import "fmt"

// ExtraArgs is extra arguments to add to the cli
// if kv is true, arguments will bey key=val format
type ExtraArgs map[string]string

func (extra ExtraArgs) Encode(kv bool) (args []string) {

	for key, val := range extra {
		// for toggles
		if val == "" {
			args = append(args, key)
			continue
		}

		if kv {
			args = append(args, fmt.Sprintf("%s=%s", key, val))
		} else {
			args = append(args, key, val)
		}
	}

	return
}
