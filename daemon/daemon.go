package daemon

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	//DAEMONKEY 环境变量
	DAEMONKEY = "DAEMONKEY"
	// DAEMONVALUE 环境变量-值
	DAEMONVALUE = "1"
)

// Daemon 启动后台进程
func Daemon(pidfile string, ifiles []string, username string) (err error) {
	if os.Getenv(DAEMONKEY) == DAEMONVALUE {
		err = pidFile(pidfile)
		if err != nil {
			os.Exit(1)
		}
		return
	}
	fileds := []*os.File{os.Stdin, os.Stdout, os.Stderr}
	var f *os.File
	for k, v := range ifiles {
		if f, err = os.OpenFile(v, os.O_APPEND|os.O_CREATE, 0644); err == nil {
			fileds[k] = f
		}
	}
	var credential *syscall.Credential
	var userinfo *user.User
	if os.Getuid() == 0 && username != "" {
		if userinfo, err = user.Lookup(username); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		credential = new(syscall.Credential)
		var i int
		i, _ = strconv.Atoi(userinfo.Uid)
		credential.Uid = uint32(i)
		i, _ = strconv.Atoi(userinfo.Gid)
		credential.Gid = uint32(i)
	}
	var originalWD, _ = os.Getwd()
	attr := &os.ProcAttr{
		Dir:   originalWD,
		Env:   append(os.Environ(), DAEMONKEY+"="+DAEMONVALUE),
		Files: fileds,
		Sys: &syscall.SysProcAttr{
			//Chroot:     d.Chroot,
			Credential: credential,
			//Setsid:     true,
		},
	}
	var abspath string
	var child *os.Process
	if abspath, err = filepath.Abs(os.Args[0]); err != nil {
		os.Exit(3)
		return
	}
	if child, err = os.StartProcess(abspath, os.Args, attr); err != nil {
		os.Exit(2)
	}
	if child != nil {
		os.Exit(0)
	}
	return
}

func pidFile(pidfile string) (err error) {
	if pidfile == "" {
		return
	}
	if err = os.MkdirAll(filepath.Dir(pidfile), os.FileMode(0755)); err != nil {
		return
	}
	os.Remove(pidfile)
	var f *os.File
	if f, err = os.OpenFile(pidfile, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		return
	}
	_, err = fmt.Fprintf(f, "%d", os.Getpid()) //os.Getpid()
	if err != nil {
		return
	}
	err = f.Close()
	return
}
