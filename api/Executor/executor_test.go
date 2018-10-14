package main

import (
	"log"
	"testing"
)

func TestGitTree(t *testing.T) {
	job := Executor{}
	job.Env = make(map[string]string)
	// job.Env["ZDOPS_DeployType"] = "appconf"
	job.Env["ZDOPS_AppConfPath"] = "https://code.daledi.cn/cmdb-mybbs/BBSCode-nginx-conf.git"
	job.Env["ZDOPS_GitToken"] = "4X-w5kqayVKGpgtdoMsy"
	job.Env["ZDOPS_AppInstallPath"] = "/Users/dale/www"
	job.Env["ZDOPS_Project"] = "mybbs"
	job.Env["ZDOPS_Cluster"] = "BBSCode"
	job.Env["ZDOPS_App"] = "nginx"
	job.Env["ZDOPS_AppConfVer"] = "cb38126faf80ea5490cb37b16c0c7a8f7f94dbdd"
	if err := job.appConf(); err != nil {
		log.Fatalf("%v\n", err)
	}
}
