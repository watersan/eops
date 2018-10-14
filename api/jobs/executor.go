package jobs

import (
	"os/exec"
	"syscall"

	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/common"
)

//JobsQueue 作业执行队列
var JobsQueue chan *JobExecutor

//StartJobExecutor 启动作业执行器
func StartJobExecutor(econf *common.EopsConf) {
	//Bug[parallelJob] 当前的方式无法真正的控制同时执行job的主机数量
	JobsQueue = make(chan *JobExecutor, econf.GOPT.ParallelJobExe)
	go func() {
		for {
			select {
			case je := <-JobsQueue:
				//econf.LogDebug("new job to run")
				//getone -> schan
				go je.CallExecutor(econf)
			}
		}
	}()
}

//CallExecutor 调用执行器，向主机发送任务，并执行。
func (je *JobExecutor) CallExecutor(econf *common.EopsConf) {
	vars := common.V{"opsenv": je.ExtraVars}
	// envBuf, _ := json.Marshal(je.ExtraVars)
	// extraVars := string(envBuf)

	executor := exec.Command(
		econf.GOPT.Ansible,
		"-i",
		econf.GOPT.WorkDir+"/hosts.py",
		"-e",
		string(vars.Json("")),
		"-e job_hosts="+je.Host,
		"-f",
		econf.GOPT.AnsibleFork,
		je.PlayBook,
	)
	econf.LogDebug("[Job] %v\n", executor.Args)
	// for _, key := range []string{"USER", "SSH_AUTH_SOCK", "LOGNAME", "PATH", "HOME"} {
	// 	if v := os.Getenv(key); v != "" {
	// 		executor.Env = append(executor.Env, key+"="+v)
	// 	}
	// }
	out, err := executor.Output()
	econf.LogDebug("[Job] errout: %v; output: %s\n", err, string(out))
	if err != nil {
		sys := executor.ProcessState.Sys().(syscall.WaitStatus)
		c := econf.Mgo.Coll(mgoRunlogs)
		skey := bson.M{"_id": bson.ObjectIdHex(je.ExtraVars["ZDOPS_TaskLogID"])}
		filArgs := bson.M{
			"status": sys.ExitStatus(),
			"msg":    err.Error(),
		}
		if err := c.Update(skey, bson.M{"$set": filArgs}); err != nil {
			econf.LogWarning("[Job] update %s: %v\n", mgoRunlogs, err)
		}
	}
}
