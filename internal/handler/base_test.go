package handler

import (
	"fmt"
	"reflect"
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
	fmt.Printf("arr--->>>> %v %s\n", arr, reflect.TypeOf(arr).Kind())
	fmt.Printf("arr--->>>> %v\n", arr2)
}

func TestSlice1(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9, 11}
	fmt.Printf("arr-->> %v %p %d %s \n", arr, &arr, cap(arr), reflect.TypeOf(arr).Kind())
	//把数组arr赋值给切片s1,此时s1的值指向arr的指针而且s1的数组内容是arr的内容,s1 和 arr 所指向的数组都是相同
	s1 := arr[:]
	//var s1 = make([]int, len(arr))
	//i := copy(s1, arr)
	//fmt.Printf("i -->>> %d \n", i)

	arr = append(arr, 0)
	fmt.Printf("s1 -->>>  %p %v %d\n", &s1, s1, cap(s1))
	fmt.Printf("arr-->>  %p  %v  %d\n", &arr, arr, cap(arr))
	//删除s1的中的3 5 s1的结果为[1 7 9 11]，因为此时s1 和 arr 的数组是相同的地址，此时修改了s1以后会影响到arr
	s1 = append(s1[:1], s1[3:]...)
	fmt.Println(s1)
	fmt.Println(arr)
	//打印S1的内存地址
	// 此时s1 的元素个数为4,容量长度是6
	fmt.Printf("s1 -->>> %p %d %d\n", &s1, cap(s1), len(s1))
	// 此时arr 的元素个数为6,容量长度是6
	fmt.Printf("arr-->>  %p  %v  %v %d %d\n", &arr, arr, s1, cap(arr), len(arr))

	//s1删除元素的值为[1 7 9 11]
	fmt.Println(s1)
	//s1指向的数组arr的值为[1 7 9 11 9 11]
	fmt.Println(arr) //[1 7 9 11 9 11]
}
