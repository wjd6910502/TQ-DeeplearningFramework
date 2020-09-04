golang
基于go-kit组件构建服务，用truss自动生成代码框架，组件定义协议，支持rpc，http，debug模式等

master:
  负责调用方的参数解析和启动任务消息处理
  
slave:
  负责任务的分发和状态机管理
  
流程图如下：
   ![image.png](/uploads/E3FBC3EF41134A648E60B9296110A9D6/image.png)



