package util

func AppendMultipleSlices[T any](slices [][]T) []T {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}

	ret := make([]T, totalLen)
	var i int
	for _, s := range slices {
		i += copy(ret[i:], s)
	}
	return ret
}

func SliceToMap[T comparable](slice []T) map[T]struct{} {
	ret := make(map[T]struct{})
	for _, s := range slice {
		ret[s] = struct{}{}
	}
	return ret
}

// Function that takes a slice of type T and returns a slice of type interface{}
// This is useful when you want to pass a slice of type T to a function that expects a slice of type interface{}
// For example:
//
//	type MyType struct {
//	    Name string
//	}
//
//	func MyFunc(slice []interface{}) {
//	    for _, item := range slice {
//	        fmt.Println(item)
//	    }
//	}
//
//	func main() {
//	    mySlice := []MyType{
//	        {Name: "Guy"},
//	        {Name: "Yael"},
//	    }
//
//	    MyFunc(util.SliceToInterface(mySlice))
//	}
func SliceToInterface[T any](slice []T) []interface{} {
	ret := make([]interface{}, len(slice))
	for i, item := range slice {
		ret[i] = item
	}
	return ret
}
