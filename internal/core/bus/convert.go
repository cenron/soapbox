package bus

import "encoding/json"

// Convert does a JSON roundtrip to convert an any-typed bus payload
// into a local struct type. This is necessary because Go uses nominal typing —
// a struct from package A cannot be type-asserted to a structurally identical
// struct from package B. The bus passes values as any, so consumers need this
// to decode payloads into their local mirror types.
func Convert[T any](payload any) (T, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		var zero T
		return zero, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
