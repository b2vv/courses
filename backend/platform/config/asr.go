package config

import "fmt"

// ASRMultilingualDefault is the initial mode for newly created audio tracks.
func ASRMultilingualDefault(getenv Getenv) (bool, error) {
	switch getenv("ASR_MULTILINGUAL_DEFAULT") {
	case "", "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, fmt.Errorf("ASR_MULTILINGUAL_DEFAULT must be true or false")
	}
}
