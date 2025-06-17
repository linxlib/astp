package types

type IElem[T any] interface {
	String() string
	Clone() T
}

func CopySlice[T IElem[T]](src []T) []T {
	if src == nil {
		return nil
	}
	result := make([]T, 0)
	for _, t := range src {
		//if t == nil {
		//	continue
		//}

		result = append(result, t.Clone())
	}
	return result
}
func CopySliceWithFilter[T IElem[T]](src []T, filter func(T) bool) []T {
	if src == nil {
		return nil
	}
	result := make([]T, 0)
	for _, t := range src {
		if filter(t) {
			result = append(result, t.Clone())
		}
	}
	return result
}
