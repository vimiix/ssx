package slice

// Distinct returns the unique vals of a slice
func Distinct[T comparable](arrs []T) []T {
	m := make(map[T]int)
	order := 0
	for idx := range arrs {
		if _, exist := m[arrs[idx]]; !exist {
			m[arrs[idx]] = order
			order++
		}
	}
	res := make([]T, len(m))
	for k, v := range m {
		res[v] = k
	}
	return res
}

// Union returns a slice that contains the unique values of all the input slices
func Union[T comparable](arrs ...[]T) []T {
	m := make(map[T]int)
	order := 0
	for idx1 := range arrs {
		for idx2 := range arrs[idx1] {
			if _, exist := m[arrs[idx1][idx2]]; !exist {
				m[arrs[idx1][idx2]] = order
				order++
			}
		}
	}

	ret := make([]T, len(m))
	for k, v := range m {
		ret[v] = k
	}

	return ret
}

// Intersect returns a slice of values that are present in all the input slices
func Intersect[T comparable](arrs ...[]T) []T {
	m := make(map[T]int)
	var order []T
	for idx1 := range arrs {
		tmpArr := Distinct(arrs[idx1])
		for idx2 := range tmpArr {
			count, ok := m[tmpArr[idx2]]
			if !ok {
				order = append(order, tmpArr[idx2])
				m[tmpArr[idx2]] = 1
			} else {
				m[tmpArr[idx2]] = count + 1
			}
		}
	}

	var (
		ret     []T
		lenArrs = len(arrs)
	)
	for idx := range order {
		if m[order[idx]] == lenArrs {
			ret = append(ret, order[idx])
		}
	}

	return ret
}

// Difference returns a slice of values that are only present in one of the input slices
func Difference[T comparable](arrs ...[]T) []T {
	m := make(map[T]int)
	var order []T
	for idx1 := range arrs {
		tmpArr := Distinct(arrs[idx1])
		for idx2 := range tmpArr {
			count, ok := m[tmpArr[idx2]]
			if !ok {
				order = append(order, tmpArr[idx2])
				m[tmpArr[idx2]] = 1
			} else {
				m[tmpArr[idx2]] = count + 1
			}
		}
	}

	var (
		ret []T
	)
	for idx := range order {
		if m[order[idx]] == 1 {
			ret = append(ret, order[idx])
		}
	}

	return ret
}

// Delete deletes the element from the slice
func Delete[T comparable](slice []T, elems ...T) []T {
	for _, val := range elems {
		for idx, elem := range slice {
			if val == elem {
				slice = append(slice[:idx], slice[idx+1:]...)
			}
		}
	}
	return slice
}
