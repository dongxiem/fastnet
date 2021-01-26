## fastnet 简介

`fastnet` 是一个高性能、轻量级、非阻塞的 TCP 网络库，用户可以使用 `fastnet` 来实现自己的应用层网络协议，从而构建出自己的应用层网络应用；

在某些极端的网络业务场景，比如海量连接、高频短连接、网络小包等场景，`fastnet` 在性能和资源占用上都超过 `Go` 原生 `net` 包的`goroutine-per-connection` 模式。

## fastnet 特性

- 使用 `Epoll` 边沿触发的IO多路复用技术，非阻塞IO，使用 `Reactor` 模式；
- 使用多线程充分利用多核CPU，使用动态扩容 `Ring Buffer` 实现读写缓冲区；
- 支持异步读写操作、支持 `SO_REUSEPORT` 端口重用；
- 灵活的事件定时器，可以定时任务，延时任务；
- 支持 `WebSocket`，同时支持自定义协议，处理 `TCP` 粘包；

**TODO：**

- 更多的负载方式（轮询（Round-Robin）、一致性哈希（`consistent hashing`）等）；
- 