# golang-demo
常用的golang学习资料记录

## 基础知识链接
1. go开发版本V1.18.3 Or Higher -- IDE vscode or goland
2. [基础语法](https://www.runoob.com/go/go-tutorial.html)
3. [基础框架 iris (只使用iris的路由功能)](https://www.topgoer.com/Iris/%E8%B7%AF%E7%94%B1)
4. MySQL框架 [goqu -- mysql构建](http://doug-martin.github.io/goqu/docs/expressions.html) + [sqlx --sql执行](https://jmoiron.github.io/sqlx/) [中文介绍文档](https://www.tizi365.com/archives/100.html) 
5. redis框架 [go-redis](https://github.com/go-redis/redis)
6. [es](https://github.com/elastic/go-elasticsearch)
7. [beanstalkd](https://github.com/beanstalkd/go-beanstalk)
8. [etcd 微服务管理 类似于zookeeper](https://etcd.io/docs/v3.6/dev-internal/modules/)

## 项目结构
### internal
1. consts -- 常有的配置 和 全局静态值
2. handler -- 所有的控制，处理请求的入口
3. logic -- 所有接口的业务逻辑
4. middleware -- iris 中间件的文件
5. models -- 所有的数据库表的对应模型
6. router -- 路由管理


### doraemon
1. helper --- 常有的工具类文件
2. .yaml 项目的环境配置

### main -- 项目入口