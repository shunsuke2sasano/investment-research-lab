package handlers

func isVersionOne(v any) bool {
	switch t := v.(type) {
	case float64:
		return t == 1
	case float32:
		return t == 1
	case int:
		return t == 1
	case int64:
		return t == 1
	case int32:
		return t == 1
	case int16:
		return t == 1
	case int8:
		return t == 1
	case uint:
		return t == 1
	case uint64:
		return t == 1
	case uint32:
		return t == 1
	case uint16:
		return t == 1
	case uint8:
		return t == 1
	default:
		return false
	}
}
