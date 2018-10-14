package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"
)

type RouteList func(*common.APIRoute)

//TestInit 测试的初始化
func Init(rlist RouteList, privilege func(*httpsvr.Context)) (*httpsvr.Router, *common.EopsConf) {
	mgdbSession, err := mgo.DialWithTimeout(
		"mongodb://eops:lejkkgt0uPXw6ArNGJKZ@localhost/eops",
		time.Duration(3)*time.Second,
	)
	if err != nil {
		fmt.Printf("mongodb failed. %v", err)
		return nil, nil
	}
	//rpool := cache.New("map", "127.0.0.1:6379", "r7lebWzDLzOC1E0Cv8OM")
	mgdbSession.SetSafe(&mgo.Safe{WMode: "majority", J: true})
	mgdb := mgdbSession.DB("")
	econf := common.Init()
	econf.SetMdb(mgdb)
	econf.Log = log.New(os.Stdout, "[test]", 7)
	econf.Level = 15
	//econf.Mgo = common.MgoDB{mgdb}
	//econf.Route = common.NewRoute()
	//econf.GOPT = options.GOPT()
	router := httpsvr.NewRouter("EOPS/test")
	router.PanicHandler = func(ctx *httpsvr.Context) {
		econf := ctx.Keys["econf"].(*common.EopsConf)
		if err := recover(); err != nil {
			if econf.Level == 15 {
				httprequest, _ := httputil.DumpRequest(ctx.Req, false)
				econf.LogDebug("[Recovery] Request: %s\n", string(httprequest))
			}
			econf.LogDebug("[Recovery] Error: %s\n", fmt.Sprintf("%+v", errors.Errorf("%v", err)))
			// econf.LogDebug("[Recovery] Error: %v\n", err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
		} else {
			if ctx.Response == nil {
				ctx.Response = 0
			}
			switch ctx.Response.(type) {
			case *common.RespMsg:
				ctx.Writer.Write(ctx.Response.(*common.RespMsg).Bytes())
			case []byte:
				ctx.Writer.Write(ctx.Response.([]byte))
			case string:
				ctx.Writer.Write([]byte(ctx.Response.(string)))
			case int:
				rmsg := common.NewRespMsg()
				code := ctx.Response.(int)
				if code > 0 {
					rmsg.SetCode(ctx.Response.(int), econf.Err)
				}
				ctx.Writer.Write(rmsg.Bytes())
			default:
				if out, err := json.Marshal(ctx.Response); err == nil {
					ctx.Writer.Write(out)
				} else {
					ctx.Writer.WriteHeader(http.StatusResetContent)
				}
			}
		}
	}

	if rlist != nil {
		rlist(econf.Route)
		for k, v := range econf.Route.API {
			t := strings.Split(k, "|")
			if v.Perm > 0 {
				router.AddRoute(string(t[1]), econf.GOPT.Prefix+string(t[0]), econf, privilege, v.Handler)
			} else {
				router.AddRoute(string(t[1]), econf.GOPT.Prefix+string(t[0]), econf, v.Handler)
			}
		}
	}
	return router, econf
}

//HttpReq 构造http request
func HttpReq(method, u string, body io.Reader, params url.Values) (*http.Request, error) {
	if params == nil {
		params = make(url.Values)
	}
	params.Set("keyid", "Lk9UdRA7hG3Ipsls95Rd")
	params.Set("secret", "TddzZe17doKA03RhrYTvvFY5s612zHzNgrHkgAJKKcwxGyM77omJMUYrrlGxJccB")
	req, err := http.NewRequest(method, u+"?"+params.Encode(), body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, err
}
