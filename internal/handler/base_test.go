package handler

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

// 这还是切片的数据结构
//type slice struct {
//	array unsafe.Pointer // 这个是底层的数组指针地址，包含了一个指向一个数组的指针，数据实际上存储在这个指针指向的数组上，占用 8 bytes
//	len   int  // 元素的个数 占用8 bytes
//	cap   int // 容量长度 同时也是底层数组 array 的长度， 8 bytes
//}

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

// 切片共享存储空间
func TestSlice1(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9, 11}
	fmt.Printf("arr-->> %v %p %d %s \n", arr, &arr, cap(arr), reflect.TypeOf(arr).Kind())
	//把数组arr赋值给切片s1,此时是属于浅拷贝 此时s1的值指向arr的指针而且s1的数组内容是arr的内容,s1 和 arr 所指向的数组都是相同，就是两个切片享受同一个存储空间
	s1 := arr[:]
	//var s1 = make([]int, len(arr))
	// 如果使用copy 此时属于深拷贝 s1 就会有相对独立的内存空间，数组空间也是独立，对他的操作不会影响arr
	//i := copy(s1, arr)
	//fmt.Printf("i -->>> %d \n", i)

	// 如果此时arr添加一个元素后，arr 底层的数组会扩容到原来的两倍
	//arr = append(arr, 0)
	fmt.Printf("s1 -->>>  %p %v %d\n", &s1, s1, cap(s1))
	fmt.Printf("arr -->>  %p  %v  %d\n", &arr, arr, cap(arr))
	//删除s1的中的3 5 s1的结果为[1 7 9 11]，因为此时s1 和 arr 的数组是相同的地址，此时修改了s1以后会影响到arr
	s1 = append(s1[:1], s1[3:]...)
	//打印S1的内存地址
	// 此时s1 的元素个数为4,容量长度是6  s1的值为[1 7 9 11]
	fmt.Printf("s1 -->>> %p %d %d\n", &s1, cap(s1), len(s1))
	// 此时arr 的元素个数为6,容量长度是6 arr的值为[1 7 9 11 9 11]
	fmt.Printf("arr-->>  %p  %v  %v %d %d\n", &arr, arr, s1, cap(arr), len(arr))
}

// 切片删除元素
func TestSliceDelete(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	var x int
	// 删除最后一个元素
	x, slice1 = slice1[len(slice1)-1], slice1[:len(slice1)-1]
	fmt.Printf("%p %d %v %d %d \n", slice1, x, slice1, len(slice1), cap(slice1))
	// 5 [1 2 3 4] 4 5

	// 删除第2个元素
	slice1 = append(slice1[:2], slice1[3:]...)
	fmt.Printf("%p  %v %d %d \n", slice1, slice1, len(slice1), cap(slice1))
	// [1 2 4] 3 5
}

// 测试切片是并发线程不安全的
func TestSlice2(t *testing.T) {
	a := make([]int, 0)
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			a = append(a, i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf(" len = %d\n", len(a))
	// 此时输出的值会少于 10000
}

// 使用锁 来使切片并发操作线程安全
func TestSlice3(t *testing.T) {
	var lock sync.Mutex //互斥锁
	a := make([]int, 0)
	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func(i int) {
			lock.Lock()
			defer lock.Unlock()
			a = append(a, i)
			defer wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf(" len = %d\n", len(a))
	// 此时 输出的长度 一定会是 1000000，耗时平均在0.45s
}

// 使用channel 来使切片并发操作线程安全
func TestSlice4(t *testing.T) {
	buffer := make(chan int)
	a := make([]int, 0)
	// 消费者
	go func() {
		for v := range buffer {
			a = append(a, v)
		}
	}()
	// 生产者
	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			buffer <- i
		}(i)
	}
	wg.Wait()
	fmt.Printf(" len = %d\n", len(a))
	// 此时 输出的长度 一定会是 1000000，耗时平均在0.82s
}
