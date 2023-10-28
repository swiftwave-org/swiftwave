package graphql

func DefaultInt(value *int, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	return *value
}

func DefaultInt64(value *int64, defaultValue int64) int64 {
	if value == nil {
		return defaultValue
	}
	return *value
}

func DefaultUint(value *uint, defaultValue uint) uint {
	if value == nil {
		return defaultValue
	}
	return *value
}

func DefaultString(value *string, defaultValue string) string {
	if value == nil {
		return defaultValue
	}
	return *value
}

func DefaultBool(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

func DefaultFloat64(value *float64, defaultValue float64) float64 {
	if value == nil {
		return defaultValue
	}
	return *value
}

func DefaultFloat32(value *float32, defaultValue float32) float32 {
	if value == nil {
		return defaultValue
	}
	return *value
}
