package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"

	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/cache"
	"opsv2.cn/opsv2/api/cmdb"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/deploy"
	"opsv2.cn/opsv2/api/httpsvr"
	"opsv2.cn/opsv2/api/jobs"
	"opsv2.cn/opsv2/api/options"
	"opsv2.cn/opsv2/daemon"
)

const (
	//Version 版本号
	Version = "1"
	//SVer http版本号
	SVer = "EOPS/" + Version
)

var (
	conf    = flag.String("c", "no", "configure file")
	pidfile = flag.String("p", "/data/log/opsv2/opsv2.pid", "the pidfile's path")
	front   = flag.Bool("f", false, "front")
	lport   = flag.String("P", ":5000", "Listen addr")
	logdir  = flag.String("l", "/data/log/opsv2", "log dir")
)

func main() {
	flag.Parse()
	dmfiles := []string{}
	econf := common.Init()
	if *conf != "no" {
		if err := options.ReadConfig(*conf, &econf.GOPT); err != nil {
			log.Fatalf("Cant read %s: %v\n", *conf, err)
		}
	}
	// dmfiles := []string{"", econf.GOPT.LogDir + "/opsv2.out", econf.GOPT.LogDir + "/opsv2.err"}
	if *front == false {
		daemon.Daemon(econf.GOPT.PidFile, dmfiles, "")
	}
	mdbconn := fmt.Sprintf("mongodb://%s:%s@%s/%s",
		econf.GOPT.MdbUser,
		econf.GOPT.MdbPW,
		econf.GOPT.MdbHost,
		econf.GOPT.Mdb,
	)
	mgdbSession, err := mgo.DialWithTimeout(
		// "mongodb://eops:lejkkgt0uPXw6ArNGJKZ@localhost/eops",
		mdbconn,
		time.Duration(econf.GOPT.MdbTimeout)*time.Second,
	)
	if err != nil {
		fmt.Printf("mongodb failed. %v", err)
		os.Exit(1)
	}
	mgdbSession.SetSafe(&mgo.Safe{WMode: "majority", J: true})
	mgdb := mgdbSession.DB("")
	econf.SetMdb(mgdb)
	var accesslog, errlog *os.File
	accesslog, err = os.OpenFile(econf.GOPT.LogDir+"/access.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("access log open failed. %v", err)
		os.Exit(1)
	}
	errlog, err = os.OpenFile(econf.GOPT.LogDir+"/error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("error log open failed. %v", err)
		os.Exit(1)
	}
	econf.AccessLog = log.New(accesslog, "", 7)
	econf.Log = log.New(errlog, "", 7)
	econf.Level = econf.GOPT.LogLevel
	//logger.SetFlags(7)
	// econf := common.Init(mgdb, log.New(errlog, "", 7), log.New(accesslog, "", 7))
	//econf.GOPT.RunEnv = "release"
	auth.RouteList(econf.Route)
	cmdb.RouteList(econf.Route)
	jobs.RouteList(econf.Route)
	deploy.RouteList(econf.Route)
	jobs.StartJobExecutor(econf)
	router := httpsvr.NewRouter(SVer)
	router.NotFound(econf)
	router.ResponseHandler = common.ResponseHandler
	router.PanicHandler = func(ctx *httpsvr.Context) {
		if err := recover(); err != nil {
			econf := ctx.Keys["econf"].(*common.EopsConf)
			if econf.Level == 15 {
				httprequest, _ := httputil.DumpRequest(ctx.Req, false)
				econf.LogDebug("[Recovery] Request: %s\n", string(httprequest))
			}
			econf.LogDebug("[Recovery] Error: %s\n", fmt.Sprintf("%+v", errors.Errorf("%v", err)))
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
		}
	}
	router.AddRoute("GET", "/mon", econf, func(ctx *httpsvr.Context) {
		ctx.NoLog = true
		ctx.Response = []byte("OK")
	})
	router.AddRoute("GET", "/memcache", econf, func(ctx *httpsvr.Context) {
		econfl := ctx.Keys["econf"].(*common.EopsConf)
		//outbuf := bytes.NewBuffer(make([]byte, 1024*1024))
		ctx.Req.ParseForm()
		ctx.Response = []byte{}
		if key, ok := ctx.Req.Form["k"]; ok {
			tmp, _ := econfl.MapCache.Get(key[0])
			if tmp != nil {
				outstr, _ := json.MarshalIndent(tmp, "", "  ")
				ctx.Response = outstr
			}
		} else {
			tmp, _ := json.MarshalIndent(econfl.MapCache.GetAll().(map[string]cache.Item), "", "  ")
			ctx.Response = tmp
		}
	})
	for k, v := range econf.Route.API {
		t := strings.Split(k, "|")
		if v.Perm > 0 {
			router.AddRoute(string(t[1]), econf.GOPT.Prefix+string(t[0]), econf, auth.Privilege, v.Handler)
			router.AddRoute(string(t[1]), fmt.Sprintf("%s/:ver%s", econf.GOPT.Prefix, string(t[0])), econf, auth.Privilege, v.Handler)
		} else {
			router.AddRoute(string(t[1]), econf.GOPT.Prefix+string(t[0]), econf, v.Handler)
			router.AddRoute(string(t[1]), fmt.Sprintf("%s/:ver%s", econf.GOPT.Prefix, string(t[0])), econf, v.Handler)
		}
	}
	//router.AddRoute("GET", "/", "econf", Index, Index2)
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	server := &http.Server{Addr: econf.GOPT.Bind + ":" + econf.GOPT.Port, Handler: router, ErrorLog: econf.Log}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		switch <-ch {
		case syscall.SIGINT, syscall.SIGTERM:
			signal.Stop(ch)
			server.Shutdown(ctx)
		case syscall.SIGUSR2:
			os.Unsetenv(daemon.DAEMONKEY)
			daemon.Daemon(econf.GOPT.PidFile, dmfiles, "")
			server.Shutdown(ctx)
		}
	}()
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Start failed: %v\n", err)
	}
}
