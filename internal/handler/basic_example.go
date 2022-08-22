package handler

import (
	"fmt"
	"reflect"
)

type BasicExample struct {
}
type Person struct {
	Name string
	Age  int
}

//DeepEqual 判断两个对象是否相等， 和 值引用和 指针引用的区别
func (c *BasicExample) DeepEqual() {
	p := Person{
		Name: "zhansan",
		Age:  18,
	}
	p1 := Person{
		Name: "zhansan",
		Age:  18,
	}
	p3 := Person{
		Name: "zhansan",
		Age:  10,
	}
	// 此时p2 是copy p这个对象，并且从自己创建了一个内存，所以p2所有的操作不会影响到p
	p2 := p
	p2.Age = 28
	// 此时p4 是 引用了p这个对象内存,只是复制了一个新的指针地址，现在对p4的操作就是相对在p上操作一样。
	p4 := &p
	(*p4).Age = 38

	// 判断两个对象是否相等
	res := reflect.DeepEqual(p, p1)
	res1 := reflect.DeepEqual(p, p2)
	res2 := reflect.DeepEqual(p, p3)
	res3 := reflect.DeepEqual(p, *p4)
	fmt.Printf("对象是否相等--->>>> \n %v \n %v\n %v\n %v\n", res, res1, res2, res3)
	fmt.Printf("对象是否相等--->>>> \n %v \n %v\n %v\n %v\n %v\n", p, p1, p2, p3, *p4)
}
