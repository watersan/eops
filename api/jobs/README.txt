
JobRun: 作业执行接口
函数参数：
# 支持通过项目，集群，应用过滤主机。允许按比例分次执行（主机列表方式不可用），同时指定每次执行的间隔时间。比如：
# A、10%，20%，30%，40%，分四次执行，间隔2个小时。每次只对指定的项目|集群|应用所对应的百分比数量的机器执行作业。
# B、10%，30%，60%，分3次执行。
# 支持直接提供主机列表，列表允许主机名和IP。当直接提供主机列表时，上面的过滤参数失效。
# 作业名称，参数。

提供给执行器的参数：
# 作业日志的id：用于作业将运行结果通过回调接口发送回来时，对应作业记录。
# 作业名称，作业参数
# 主机名
# 作业内容的md5.如果是作业流程，提供每个作业内容的md5.执行器会根据md5判断是否需要从git下载。
# 环境变量

JobScheduleScaner: 计划作业扫描，单独运行与一个协程
# 计划作业在提交时，就保存到两个地方：mgdb和内部缓存。
# 在程序启动时，要主动从mgdb加载计划作业到内部缓存。
# 可以考虑time.AfterFunc()，来完成计划执行。这样就不用运行单独的协程进行计划任务扫描了。
# 对于已经过时的任务，立即执行。

执行器设计：
# 后台运行作业，作业完成后，回调日志接口提交运行结果，然后退出。
# 限制服务器同时运行的作业数
# 可以设置nice或限制cpu使用。
# 设置超时时间，超过时间强行终止作业运行，并提交结果。
# 可以通过ssh中转。ansible zabbix --ssh-common-args "-o 'ProxyCommand ssh -W %h:%p dale@dev.daledi.cn -p 51006'" -m shell -a "hostname;who"

git仓库命名规范：
  cmdb中的项目（project）对应gitlab中的组(添加"cmdb-"前缀)。应用在gitlab中仓库名称：cmdb中集群名称和应用名称，用中划线分隔。
  应用配置的仓库名称在应用仓库后追加：-conf。
  应用模板单独一个组：apptpl，项目名称为模板名称。
  例如：
  组名                  仓库
  cmdb-gfanbbs          bbs_php-bbs
  cmdb-gfanbbs          bbs_php-bbs-conf
  cmdb-gfanbbs          bbs_php-nginx-conf
  apptpl                nginx
  apptpl                nginx-conf

作业日志：
# 一个作业一条日志（包括作业流），记录每个执行作业主机的执行结果
# 主机的执行结果中，保存每个作业的执行结果。已数组形式，按序保存。

ops设置的变量：
ZDOPS_APP
ZDOPS_CLUSTER
ZDOPS_PROJECT
ZDOPS_APPVER
ZDOPS_APPCONFVER
ZDOPS_DeployType
ZDOPS_SavePath
ZDOPS_JobID
ZDOPS_CallBack
ZDOPS_Token        获取job内容和callback时使用
ZDOPS_JobAPI     获取job内容
ZDOPS_ServiceName  启动的服务名称
ZDOPS_DeployDir    目标服务器部署的目录

WorkDir
DeployDir
ExecutorPath

Job_PreInstall
Job_PreInstallArgs
Job_chkconfig
Job_chkconfigArgs
Job_Install
Job_InstallArgs
Job_PostInstall
Job_PostInstallArgs
Job_Rollback

#ansible配置
ZDOPS_HOSTNAME
ZDOPS_SYSTEM_ENV  主机所在的环境：开发，测试，正式
ZDOPS_LOCATION    主机所在位置

#放弃
ZDOPS_GitScheme
ZDOPS_GitHost
ZDOPS_GitProject
ZDOPS_GitToken

ansible Inventory配置：
  # host变量：带*的变量必须配置
    • ZDOPS_HOSTNAME *
    • ansible_ssh_common_args="-o 'ProxyCommand ssh -W %h:%p dale@dev.daledi.cn -p 51006'"
    • ansible_port=5555
    • ansible_host=192.0.2.50 *
    • ZDOPS_SYSTEM_ENV *
    • ZDOPS_LOCATION *
  #
[all:vars]
ZDOPS_SYSTEM_ENV=test

[appname]
host1 ansible_port=5555 ansible_host=192.0.2.50 ZDOPS_HOSTNAME=host1 ansible_ssh_common_args="-o 'ProxyCommand ssh -W %h:%p dale@dev.daledi.cn -p 51006'"
host2 ZDOPS_HOSTNAME=host2 ZDOPS_LOCATION=aliyun

[appname:vars]
ZDOPS_APP=appname

[clustername:children]
appname1
appname2
[clustername:vars]
ZDOPS_CLUSTER = clustername

[projectname:children]
clustername1
clustername2

[projectname:vars]
ZDOPS_PROJECT = projectname

jobs.yml: ansible-playbook -e '{"ZDOPS_HOSTS":"app","ZDOPS_USER":"dale"}'
- hosts: "{{job_hosts}}"
  remote_user: "{{job_sshuser}}"
  gather_facts: false
  environment:
    # Inventory变量
    ZDOPS_APP: "{{ZDOPS_APP}}"
    ZDOPS_HOSTNAME: "{{ZDOPS_HOSTNAME}}"
    ZDOPS_SYSTEM_ENV: "{{ZDOPS_SYSTEM_ENV}}"
    ZDOPS_LOCATION: "{{ZDOPS_LOCATION}}"
    ZDOPS_HOSTIP: "{{ansible_host}}"
    ZDOPS_GROUPTYPE: "{{grouptype}}"
    ZDOPS_GROUPNAME: "{{groupname}}"
    ZDOPS_CLUSTER: "{{ZDOPS_CLUSTER}}"
    ZDOPS_PROJECT: "{{ZDOPS_PROJECT}}"
    # 作业系统设置的变量
    ZDOPS_SavePath: "{{ZDOPS_SavePath}}"
    ZDOPS_JobID: "{{ZDOPS_JobID}}"
    ZDOPS_CallBack: "{{ZDOPS_CallBack}}"
    ZDOPS_Token: "{{ZDOPS_Token}}"
    ZDOPS_JobAPI: "{{ZDOPS_JobAPI}}"

  tasks:
    - name: run command (jobargs is base64_encode)
      command: "{{ExecutorPath}} {{job_front}} -j {{job_path}}/{{jobs_name}} -a {{job_args}}"



jobflows.yml


部署任务的脚本不能在目标机器上后台运行
deploy.yml
- hosts: "{{ZDOPS_HOSTS}}"
  remote_user: "{{ZDOPS_USER}}"
  gather_facts: false
  #vars_files:
  #- "{{WorkDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/vars.yml"
  environment:
    ZDOPS_APP: "{{ ZDOPS_APP }}"
    ZDOPS_CLUSTER: "{{ ZDOPS_CLUSTER }}"
    ZDOPS_PROJECT: "{{ ZDOPS_PROJECT }}"
    ZDOPS_APPVER: "{{ ZDOPS_APPVER }}"
    ZDOPS_APPCONFVER: "{{ ZDOPS_APPCONFVER }}"
    ZDOPS_DeployType: "{{ ZDOPS_DeployType }}"
    ZDOPS_JobID: "{{ ZDOPS_JobID }}"
    ZDOPS_SavePath: "{{ ZDOPS_SavePath }}"
    ZDOPS_CallBack: "{{ ZDOPS_CallBack }}"
    ZDOPS_JobAPI: "{{ ZDOPS_JobAPI }}"
    ZDOPS_ServiceName: "{{ ZDOPS_ServiceName }}"
    ZDOPS_HOSTNAME: "{{ ZDOPS_HOSTNAME }}"
    ZDOPS_SYSTEM_ENV: "{{ ZDOPS_SYSTEM_ENV }}"
    ZDOPS_LOCATION: "{{ ZDOPS_LOCATION }}"
    ZDOPS_Token: "{{ ZDOPS_Token }}"

  tasks:
  - name: preinstall
    command: "{{ExecutorPath}} -f {{job_newest}} -j {{Job_PreInstall}} -a {{Job_PreInstallArgs}}"
    when: Job_PreInstall is defined
  - name: copy app package to server
    block:
      - unarchive:
        # App_Source = http://source.abc.com/
        src: "{{App_Source}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/packages/{{ZDOPS_APPVER}}.zip"
        #src: "{{WorkDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/packages/{{ZDOPS_APPVER}}.zip"
        dest: "{{ZDOPS_DeployDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/history/app_{{ZDOPS_APPVER}}"
        owner: root
        mode: 755
        remote_src: yes
      register: CopyApp
      #ignore_errors: True
      when: ZDOPS_DeployType != "appconf" or ZDOPS_DeployType != "apptplconf"
    rescue:
      - command: "{{ExecutorPath}} -f {{job_newest}} -b {{CopyApp | to_json}} -j {{Job_Rollback}} -a package"
      - command: /bin/false
  - name: copy config file
    block:
      - file:
          path: '{{ZDOPS_DeployDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/history/conf_{{ZDOPS_APPCONFVER}}/{{ item.path }}'
          state: directory
          mode: 0755
        when: item.state == 'directory'
        with_filetree: "{{WorkDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/conf"
      - template:
          src: '{{ item.src }}'
          dest: "{{ZDOPS_DeployDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/history/conf_{{ZDOPS_APPCONFVER}}/{{ item.path }}"
          mode: '{{ item.mode }}'
        when: item.state == 'file'
        with_filetree: "{{WorkDir}}/{{ZDOPS_PROJECT}}/{{ZDOPS_CLUSTER}}/{{ZDOPS_APP}}/conf"
      #ignore_errors: True
      register: CopyAppConf
      when: ZDOPS_DeployType != "app"
    rescue:
      - command: "{{ExecutorPath}} -f {{job_newest}} -b {{CopyAppConf | to_json}} -j {{Job_Rollback}} -a appconf"
      - command: /bin/false
  - name: Install app or check config
    block:
      - command: "{{ExecutorPath}} -f {{job_newest}} -j {{Job_Install}} -a {{Job_InstallArgs}}"
        when: Job_Install is defined and (ZDOPS_DeployType == "app" or ZDOPS_DeployType == "apptpl")
      - command: "{{ExecutorPath}} -f {{job_newest}} -j {{Job_chkconfig}} -a {{Job_chkconfigArgs}}"
        when: Job_chkconfig is defined
      - service: name="{{ZDOPS_GroupName}}" state="reloaded"
        when: ZDOPS_DeployType == "appconf" or ZDOPS_DeployType == "apptplconf"
      - service: name="{{ZDOPS_GroupName}}" state="restarted"
        when: ZDOPS_DeployType == "app" or ZDOPS_DeployType == "apptpl"
      register: ResultInstall
    rescue:
      - command: "{{ExecutorPath}} -f {{job_newest}} -b {{ResultInstall | to_json}} -j {{Job_Rollback}} -a install"
      - command: /bin/false
  - name: PostInstall
    block:
      - command: "{{ExecutorPath}} -f {{job_newest}} -j {{Job_PostInstall}} -a {{Job_PostInstallArgs}}"
        register: ResultVerify
        when: Job_PostInstall is defined
    rescue:
      - command: "{{ExecutorPath}} -f {{job_newest}} -b {{ResultVerify | to_json}} -j {{Job_Rollback}}  -a install"
      - command: /bin/false

build.yml  #构建作业
- hosts: "{{ZDOPS_HOSTS}}"
  remote_user: "{{ZDOPS_USER}}"
  gather_facts: false
  environment:
    ZDOPS_APP: "{{ ZDOPS_APP }}"
    ZDOPS_CLUSTER: "{{ ZDOPS_CLUSTER }}"
    ZDOPS_PROJECT: "{{ ZDOPS_PROJECT }}"
    ZDOPS_APPVER: "{{ ZDOPS_APPVER }}"
    ZDOPS_APPCONFVER: "{{ ZDOPS_APPCONFVER }}"
    ZDOPS_DeployType: "{{ ZDOPS_DeployType }}"
    ZDOPS_JobID: "{{ ZDOPS_JobID }}"
    ZDOPS_SavePath: "{{ ZDOPS_SavePath }}"
    ZDOPS_CallBack: "{{ ZDOPS_CallBack }}"
    ZDOPS_JobAPI: "{{ ZDOPS_JobAPI }}"
    ZDOPS_ServiceName: "{{ ZDOPS_ServiceName }}"
    ZDOPS_HOSTNAME: "{{ ZDOPS_HOSTNAME }}"
    ZDOPS_SYSTEM_ENV: "{{ ZDOPS_SYSTEM_ENV }}"
    ZDOPS_LOCATION: "{{ ZDOPS_LOCATION }}"
    ZDOPS_Token: "{{ ZDOPS_Token }}"

  tasks:
  - name: build job
    command: "{{ExecutorPath}} -f {{job_newest}} -j {{Job_Build}} -a {{Job_BuildArgs}}"
    register: ResultBuild
    ignore_errors: True
  - name:
