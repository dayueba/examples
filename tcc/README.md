## tcc example

[docs](https://xjip3se76o.feishu.cn/wiki/wikcnOZnZMMuGWB42PPB0dyIppd)

## 假如面试官让你介绍一下 tcc

tcc将一个分布式事务分成2个阶段，第一个阶段是try，尝试锁定资源，第二个阶段根据try阶段是否成功执行来确定后续操作，如果成功则走 confirm
因为try阶段已经锁定了资源，所以 confirm 不用做任何检查，直接使用已经锁定的资源。如果失败则走 cancel 释放 try 阶段锁定的资源。
避免了长事务，性能更高。

## tcc的实现

一般不会从0开始实现，会借助于框架。比如seata，dtm。在这个例子中使用dtm，需要实现每个分支的 try，cancel，comfirm 操作，注册到dtm。
由dtm执行

## 优缺点
- 避免了长事务，性能好
- 代码量增加
- 不适合长事务