package config

import (
	"errors"
	"fmt"
	"os"
)

func ValidateRequiredEnv(keys ...string) error {
	missing := make([]string, 0)
	for _, key := range keys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

func RequireEnv(keys ...string) {
	if err := ValidateRequiredEnv(keys...); err != nil {
		panic(err)
	}
}

func optionalEnv(key string) error {
	if os.Getenv(key) == "" {
		return errors.New(key + " is empty")
	}
	return nil
}
