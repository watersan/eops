# 处理步骤
* 创建日志文件，并判断是否是debug级别
* 处理环境变量
* 判断是否是后台运行
* 处理环境变量中的任务信息
* 循环执行任务
  * 判断本地是否有作业文件及作业的版本，否则就从git中下载指定版本
  * 设置作业运行的环境变量，讲上一个作业的运行信息也附加到当前作业的环境变量中
  * 按照要求：用户，优先级等运行作业
  * 将作业的运行结果保存到内存
  * 如果作业执行失败，则执行回滚作业
* 将所有作业的运行结果发送回平台
