package utils

// Contains 函数判断一个 int 是否在 []int 切片中
func Contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
