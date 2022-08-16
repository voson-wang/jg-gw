package util

func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for _, d := range a {
		if !Contain(d, b) {
			return false
		}
	}

	for _, c := range b {
		if !Contain(c, a) {
			return false
		}
	}

	return true
}

func Contain[T comparable](a T, b []T) bool {
	for _, s := range b {
		if s == a {
			return true
		}
	}
	return false
}
