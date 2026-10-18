package util

import "encoding/json"

func MustUnmarshal[T any](buff []byte, t T) T {
	if err := json.Unmarshal(buff, t); err != nil {
		panic(err)
	}
	return t
}
