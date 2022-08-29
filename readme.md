# golang-demo
常用的golang学习资料记录

------
## 1.基础知识链接
1. go开发版本V1.18.3 Or Higher -- IDE vscode or goland
2. [基础语法](https://www.runoob.com/go/go-tutorial.html)
3. [基础框架 iris (只使用iris的路由功能)](https://www.topgoer.com/Iris/%E8%B7%AF%E7%94%B1)
4. MySQL框架 [goqu -- mysql构建](http://doug-martin.github.io/goqu/docs/expressions.html) + [sqlx --sql执行](https://jmoiron.github.io/sqlx/) [中文介绍文档](https://www.tizi365.com/archives/100.html) 
5. Redis框架 [go-redis](https://github.com/go-redis/redis)
6. RPC框架 [hprose](https://github.com/hprose/hprose-golang/wiki)
7. [es](https://github.com/elastic/go-elasticsearch)
8. [beanstalkd](https://github.com/beanstalkd/go-beanstalk)
9. [etcd 微服务管理 类似于zookeeper](https://etcd.io/docs/v3.6/dev-internal/modules/)

------

## 2.项目结构
### internal
1. consts -- 常有的配置 和 全局静态值 常量设置 --所有常量使用大写
2. handler -- 所有的控制，处理请求的入口
3. logic -- 所有接口的业务逻辑
4. middleware -- iris 中间件的文件
5. models -- 所有的数据库表的对应模型
6. router -- 路由管理


### doraemon
1. helper --- 常有的工具类文件
2. .yaml 项目的环境配置

### main -- 项目入口

------
## 3.代码规范
1. handler 内所有控制器使用ctl_ 开头
2. model 内所有的模型使用 tbl_ 开头
3. 数据库表的设计规范
   * created/updated_at 使用bigint 存储
   * id 使用bigint --根据业务需要设计，最好使用自定义生成id，不使用自增
4. logic 内所有业务层类 使用 logic_ 开头

------
## 4.各种框架使用例子
1. iris 路由功能 --- 在router/http.go
2. mysql 使用 --- logic/logic_base.go
3. redis --- 待完善
4. goqu sql 构建 -- logic/logic_user.go
   ```Go
   
   // 构建插入数据	
   sql, _, err := G.From("user").Insert().Rows(ex).ToSQL()
   // 更新数据，
   rd := goqu.Ex{"name": "lisi"}
    sql, _, err := G.From("user").Update().Where(goqu.Ex{"id": 1}).Set(rd).ToSQL()
   
   // 根据条件查询数据
   sql, _, err := G.From("user").Select().Where(goqu.Ex{"id": 1}).Limit(10).Offset(1).ToSQL()
	
   // 查询所有数据
    sql, _, err = G.From("user").Select().ToSQL()

   //连表查询-- 并且返回求和参数
   	sql, _, err := logic.G.From(`bet`).Select(
		goqu.COUNT("bet.amount").As("total_amount"),
		goqu.COUNT("tips.amount").As("total_tips_amount"),
		goqu.COUNT("bet.win_amount").As("total_win_amount")).LeftJoin(
		goqu.T("tips"),
		goqu.On(goqu.Ex{
			"bet.game_no": goqu.I("tips.game_no"),
			"bet.desk_no": goqu.I("tips.desk_no"),
		}),
	).Where(goqu.Ex{
		"bet.created_at": goqu.Op{
			"gt": 10000,
		},
		"bet.better_id": []int64{20, 24, 26, 28, 30},
		"trade_type_id": 0,
	}).ToSQL()
   ```
5. sqlx -- 待完善
```Go

    // 根据查询到的数据解析成对应的model
	res, err := Db.Queryx(sql)
    for res.Next() {
        var p models.UserModel
        err = res.StructScan(&p)
         v = append(v, p)
     }
	 
	 // 执行sql语句 
	 _, err = Db.Exec(sql)
	 
	 // 获取单条数据 
	 err:=Db.Get(&user, sql)
    

```
6. hprose -- 待完善
7. nats
8. websocket 使用 -- handler/socket_server.go
9. context 包
10. net/http 包
11. 单元测试包 testing --base_test.go
12. channel 重点
13. file
14. IO
15. buffer
16. jwt

------

## 5.golang 常用例子介绍
1. func方法介绍

```Go

   type Person struct {
      name string
      age  uint
   }
    // 值传递 不会影响 原有初始化的结构体
    // 此种方式类似于Person对象的实例方法
    func (p Person) setName(name string) { 
        p.name = name
    }
    // 指针地址传递 会影响原有的数据
    // 此种方式类似于Person对象的实例指针对象方法
    func (p *Person) setAge(age uint) {
        p.age = age
    }
```

2. 接口 interface 的介绍
3. map 的使用介绍
4. slice 切片 --- base_test.go
```go
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

   arr --->>>> [0 1 10 3 4 10 11]
   arr2 --->>>> [0 1 10]

```
5. goroutine 介绍
6. 读取本地文件 --- consts/config.go
7. 反射
8. 值引用和指针引用的区别 。判断两个对象是否相等 ---handler/basic_exmaple.go
```Go
      // 注意：Go语言中所有的传参都是值传递（传值），都是一个副本，一个拷贝。
      // 拷贝的内容是非引用类型（int、string、struct等这些），在函数中就无法修改原内容数据；
      // 拷贝的内容是引用类型（interface、指针、map、slice、chan等这些），这样就可以修改原内容数据。
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
```




------
