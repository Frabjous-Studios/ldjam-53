package internal

func randMapValue[K comparable, V any](m map[K]V) V {
	var zero V
	for _, v := range m { // uses fact that map range loops are random
		return v
	}
	return zero
}

func randMapKey[K comparable, V any](m map[K]V) K {
	var zero K
	for k := range m { // uses fact that map range loops are random
		return k
	}
	return zero
}
