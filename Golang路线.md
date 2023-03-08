# Golang路线

# golang基础+MySQL基础

## 具体要求

* #### 配置golang环境，使用golang书写代码

* #### 掌握golang基本数据类型的声明和使用

* #### 掌握golang引用类型的声明和使用（map，切片，channel）

* #### 掌握函数，结构体，接口，网络编程使用，以及一些常用包

* #### 了解并发，指针，反射

## 考核要求（开发一个使用MySQL存储数据的golang聊天室）

* #### 实现登录注册注销等功能

* #### 实现私发，群发，查看在线好友，添加好友，删除好友等功能





# 一：Gorm——GO语言操作数据库

​	**gorm官方中文文档，看这一个就够了**：[GORM 指南 | GORM - The fantastic ORM library for Golang, aims to be developer friendly.](https://gorm.io/zh_CN/docs/)

​	**视频指南**：[【GORM教学】手把手带你入门GORM（手把手系列复活了）_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1E64y1472a?spm_id_from=333.337.search-card.all.click)，视频作者是``gin-vue-admin``的作者

​	**mysql可视化工具推荐**：navicat ，破解教程：



## 具体要求

* #### 掌握连接数据库及调整数据库配置

* #### 数据库的增删改查

* #### 掌握Has One，Has Many，Belong To，Many To Many 等关系

* #### 掌握Preload，Joins预加载，事务

* #### 掌握关联模式，掌握连表查询





# 二：Redis

**redis教程**：[Redis 教程 | 菜鸟教程 (runoob.com)](https://www.runoob.com/redis/redis-tutorial.html)

**redis可视化工具推荐**：another-reids

## 具体要求

* #### 掌握电脑本地安装redis，学习docker后掌握服务器通过docker启动redis

* #### 熟练redis的配置文件，能够通过配置文件启动redis服务

* #### 掌握redis五大数据类型

* #### 能够通过命令行现实redis的增删改查等

* #### 拓展掌握主从复制，集群，哨兵模式





# 三：Go-redis——Go语言操作redis

**go-redis官方文档（外网**）：[Go Redis [getting started guide\] (uptrace.dev)](https://redis.uptrace.dev/guide/go-redis.html)  

## 具体要求

* #### 掌握go-redis进行五大数据类型的增删查改





# 四：Docker

**菜鸟教程**：[Docker 教程 | 菜鸟教程 (runoob.com)](https://www.runoob.com/docker/docker-tutorial.html)

**狂神视频**：[【狂神说Java】Docker最新超详细版教程通俗易懂_哔哩哔哩_bilibili](https://www.bilibili.com/video/BV1og4y1q7M4?spm_id_from=333.999.0.0)

**Docker官方文档**：[Get Docker | Docker Documentation](https://docs.docker.com/get-docker/)

## 具体要求

* #### 了解Docker的起源历史

* #### 掌握阿里云服务器安装Docker，掌握阿里云服务器后台的基本使用（配置服务器操作系统，开放安全组等等）

* #### 掌握Docker镜像命令，容器命令（进入，查看，删除，停止等等）

* #### 掌握容器数据卷挂载

* #### 熟练编写基础的Dockerfile用于生成镜像(项目中编写可把项目打包为镜像)

* #### 掌握Docker部署项目，了解shell命令（用于编写脚本文件，现实一键部署）

* #### 提升掌握Docker网络



## 实战练习

* #### Docker安装MySQL

* #### Docker安装Redis（通过redis配置文件启动，并实现配置文件和数据挂载）

* #### 实现MySQL的主从复制

* #### 实现Redis的主从复制


## 实战提升（作品完成后部署时再来学习）

* #### Docker启动Nginx

* #### 编写Dockerfile文件为go项目打包镜像

* #### 使用端口映射和nginx两种方式部署项目







# 五：Gin框架基础

**参考博客**：[Gin框架介绍及使用 | 李文周的博客 (liwenzhou.com)](https://www.liwenzhou.com/posts/Go/gin/#autoid-0-4-1)

**官方文档**：[文档 | Gin Web Framework (gin-gonic.com)](https://gin-gonic.com/zh-cn/docs/)

## 具体要求

* #### 了解gin渲染（Json渲染）

* #### 掌握gin框架获取参数（Query方法，PostForm方法，Param方法）

* #### 掌握gin框架参数绑定到结构体（ShouldBind方法，ShouldBindJson方法等）

* #### 掌握gin路由，路由组，GET，POST等请求方式 

* #### 掌握gin中间件，如全局中间件，局部中间件，路由组中间件

* #### 掌握gin.Context 上下文





# 六：Gin框架进阶知识

## 具体要求

* #### 掌握viper，使用viper加载读取配置文件

* #### 掌握Zap，使用Zap来记录日志

* #### 掌握Validator，使用Validator是进行参数校验

* #### 掌握JWT认账，了解Cookie和Session

* #### 了解Air热加载，swagger等工具

* #### 完成作品bullbell

* #### 了解Gin-Vue-Admin框架

**参考视频**：bullbell视频，超详细，不仅有上面的知识讲解，而且一步一步搭建了一个Go Web脚手架，手把手教程。

**视频链接**：





# 七：Gin框架作品实战要求

* #### 独立原创，实现接口不低于10个，开发尽量用来解决实际需求

* #### 进阶知识需要全部应用

* #### 在数据库存储中体现一对一和一对多关系存储

* #### 实现MySQL和Redis多数据库存储，拓展实现主从复制

* #### 实现通过dockerfile文件来生成镜像，并通过端口映射和nginx两种方式部署服务器

* #### 拓展实现编写shell命令自动化部署





# 八：git&Coding

## 具体要求

* #### 掌握本地安装git，掌握docker容器内安装git，了解git历史

* #### 掌握git基本命令并推送代码到Coding，克隆代码

* #### 熟练使用Goland解决代码冲突（Goland上集成的有git工具）

* #### 能够在Coding上面建立自己的分支并申请合并代码





