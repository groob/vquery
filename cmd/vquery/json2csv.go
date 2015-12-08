package main

import (
	"fmt"
	"math"
	"strings"
)

type stringArray []string

func (a *stringArray) Set(s string) error {
	for _, ss := range strings.Split(s, ",") {
		*a = append(*a, ss)
	}
	return nil
}

func (a *stringArray) String() string {
	return fmt.Sprint(*a)
}

func getValue(data map[string]interface{}, keyparts []string) string {
	if len(keyparts) > 1 {
		subdata, _ := data[keyparts[0]].(map[string]interface{})
		return getValue(subdata, keyparts[1:])
	} else if v, ok := data[keyparts[0]]; ok {
		switch v.(type) {
		case nil:
			return ""
		case float64:
			f, _ := v.(float64)
			if math.Mod(f, 1.0) == 0.0 {
				return fmt.Sprintf("%d", int(f))
			}
			return fmt.Sprintf("%f", f)
		default:
			return fmt.Sprintf("%+v", v)
		}
	}

	return ""
}
