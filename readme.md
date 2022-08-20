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
   
   ```
5. sqlx -- 待完善
6. hprose -- 待完善
7. nats
8. websocket 使用 -- 带完善
9. context 包
10. net/http 包
11. 单元测试包 testing
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
4. slice 切片
5. goroutine 介绍
6. 读取本地文件 --- consts/config.go
7. 反射





------

## 6.git使用的规范
1. 分支管理 -- 同一分支 强制使用 git pull 使用 rebase
    * master 分支 -- 主分支，对应当前线上版本
    * develop 分支 -- 开发分支
    * feature 分支 -- 功能分支，开发新功能的分支，并且由develop分支切出来, feature/new_task,开发完成后需要删除。
    * release 分支 -- 发布分支，新功能合并到 develop 分支，准备发布新版本时使用的分支
    * hotfix 分支 -- 紧急修复线上 bug 分支
2. 提交信息规范 --  不要随意添加commit，必须说明此commit的功能 example: (fix: 用户登录参数校验)
    * feat/add: 新功能
    * fix: 修复 bug
    * update: 更新内容
    * docs: 文档变动
    * style: 格式调整，对代码实际运行没有改动，例如添加空行、格式化等
    * refactor: bug 修复和添加新功能之外的代码改动
    * perf: 提升性能的改动
    * test: 添加或修正测试代码
    * chore: 构建过程或辅助工具和库（如文档生成）的更改