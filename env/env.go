package env

import "os"

// Optional attempt to get an optional environment variable or choose default value
func Optional(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
