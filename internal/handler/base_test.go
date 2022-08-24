package handler

import (
	"fmt"
	"testing"
)

// TestSlice 切片的常用操作
func TestSlice(t *testing.T) {
	var arr = []int{0, 1}
	var arr1 = []int{10, 11}
	// 这是一次性添加多个元素
	arr = append(arr, 2, 3)
	// 这是添加单个元素
	arr = append(arr, 4)
	// 这是一次性添加多个元素-- 添加一个切片对象
	arr = append(arr, arr1...)

	//使用其他的切片构建新的切片，
	arr2 := append(arr[:2], arr1[0:1]...)
	fmt.Printf("arr--->>>> %v\n", arr)
	fmt.Printf("arr--->>>> %v\n", arr2)
}
