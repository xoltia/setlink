package sliceutil

func Map[T any, U any](slice []T, f func(T) U) []U {
	out := make([]U, len(slice))
	for i, v := range slice {
		out[i] = f(v)
	}
	return out
}

func Find[T any](slice []T, f func(T) bool) *T {
	for _, v := range slice {
		if f(v) {
			return &v
		}
	}
	return nil
}
