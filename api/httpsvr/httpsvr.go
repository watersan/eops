package httpsvr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2/bson"
)

//Router 自定义http router
type Router struct {
	httprouter.Router
	serverVer       string
	PanicHandler    func(*Context)
	ResponseHandler func(*Context)
}

// type ResponseWriter interface {
//   http.ResponseWriter
//
//   Status() int
//   // Returns the number of bytes already written into the response http body.
//   // See Written()
//   Size() int
//   // Writes the string into the response body.
//   WriteString(string) (int, error)
//
//   // Returns true if the response body was already written.
//   Written() bool
//
//   // Forces to write the http header (status code + headers).
//   WriteHeader()
// }

//ResponseWriter 增加status字段
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

//Context http会话期间的session
type Context struct {
	Req      *http.Request
	Writer   *ResponseWriter
	Params   httprouter.Params
	Keys     map[string]interface{}
	Response interface{}
	Begin    time.Time
	NoLog    bool
	OpLog    *bytes.Buffer
	Body     []byte
	BodyMap  map[string]interface{}
	//	Econf   *common.EopsConf
}

const (
	noWritten   = -1
	maxBodySize = int64(10 << 20)
)

//Handle 增加全局配置信息
type Handle func(ctx *Context)

//type Handle func(http.ResponseWriter, *http.Request, httprouter.Params, interface{})

func NewRouter(sver string) *Router {
	router := &Router{
		Router: *httprouter.New(),
		// accessLog: alog,
		// errLog:    elog,
		serverVer: sver,
	}
	return router
}

//NewResponseWriter 创建新实例
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK, noWritten}
}

//WriteHeader 记录status
func (lrw *ResponseWriter) WriteHeader(code int) {
	if !lrw.Written() {
		lrw.size = 0
		lrw.statusCode = code
		lrw.ResponseWriter.WriteHeader(code)
	}
}

//AbortWithStatus 记录status
func (lrw *ResponseWriter) AbortWithStatus(code int) {
	lrw.size = 0
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

//Write 记录status
func (lrw *ResponseWriter) Write(data []byte) (n int, err error) {
	if !lrw.Written() {
		n, err = lrw.ResponseWriter.Write(data)
		lrw.size += n
	}
	return
}

//WriteString 获取Written
func (lrw *ResponseWriter) WriteString(s string) (n int, err error) {
	n, err = io.WriteString(lrw.ResponseWriter, s)
	lrw.size += n
	return
}

//Written 获取Written
func (lrw *ResponseWriter) Written() bool {
	return lrw.size != noWritten
}

//Status 获取status
func (lrw *ResponseWriter) Status() int {
	return lrw.statusCode
}

//Size 获取size
func (lrw *ResponseWriter) Size() int {
	return lrw.size
}

//ServeHTTP 增加日志输出
func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	lw := NewResponseWriter(w)
	lw.Header().Set("Server", router.serverVer)
	//lw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	router.Router.ServeHTTP(lw, req)
}

//AddRoute 增加route信息
func (router *Router) AddRoute(method, path string, conf interface{}, handlers ...Handle) {
	router.Router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := &Context{
			Req:    r,
			Writer: w.(*ResponseWriter),
			Params: ps,
			Keys:   make(map[string]interface{}),
			Begin:  time.Now(),
			OpLog:  bytes.NewBuffer([]byte{}),
		}
		ctx.Keys["econf"] = conf
		if router.PanicHandler != nil {
			defer router.PanicHandler(ctx)
		}
		for _, handle := range handlers {
			handle(ctx)
			if w.(*ResponseWriter).Written() || ctx.Response != nil {
				break
			}
		}
		if router.ResponseHandler != nil {
			router.ResponseHandler(ctx)
		}
	})
	return
}

//NotFound 增加route信息
func (router *Router) NotFound(conf interface{}) {
	router.Router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Req:    r,
			Writer: w.(*ResponseWriter),
			Keys:   make(map[string]interface{}),
			Begin:  time.Now(),
		}
		ctx.Keys["econf"] = conf
		ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctx.Writer.WriteHeader(http.StatusNotFound)
		if router.PanicHandler != nil {
			defer router.PanicHandler(ctx)
		}
	})
}

//RemoteAddr 获取客户端IP
func (ctx *Context) RemoteAddr(xff bool) string {
	addr := ctx.Req.Header.Get("X-Real-IP")
	if addr == "" {
		addr = ctx.Req.RemoteAddr
		if i := strings.IndexByte(addr, ':'); i >= 0 {
			addr = ctx.Req.RemoteAddr[:i]
		}
		if xff {
			addr = ctx.Req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = ctx.Req.RemoteAddr
			}
		}
	}
	return addr
}

//ReqBody 获取请求中的Body
func (ctx *Context) ReqBody(out interface{}) (b []byte, ecode int, err error) {
	ecode = 0
	if ctx.Req.Body == nil {
		ecode = 1202
		return
	}
	ct := ctx.Req.Header.Get("Content-Type")
	// RFC 2616, section 7.2.1 - empty type
	//   SHOULD be treated as application/octet-stream
	if ct == "" {
		ct = "application/octet-stream"
	}
	ct, _, _ = mime.ParseMediaType(ct)
	switch ct {
	case "application/x-www-form-urlencoded", "application/json":
		reader := io.LimitReader(ctx.Req.Body, maxBodySize+1)
		b, err = ioutil.ReadAll(reader)
		if err != nil {
			ecode = 1205
			return
		}
		if int64(len(b)) > maxBodySize {
			ecode = 1203
			return
		}
	default:
		ecode = 1204
	}
	if out != nil {
		if err = json.Unmarshal(b, out); err != nil {
			ecode = 1206
		}
	}
	return
}

var Chunking = errors.New("chunking")

//Upload 文件上传处理。兼容chunked方式。
/*
	不是chunk方式上传时，保存的文件名为随机生成。
	chunk方式上传时，保存的文件名为提交的文件名。这种方式上传的文件名不能重名。
	返回的参数：提交的文件名，保存的文件名，错误信息。当chunk方式上传时，在最后一个chunk请求处理
	之前，返回的错误信息是：Chunking，表示chunk未完成。
*/
func (ctx *Context) Upload(savedir, allowExt string) (string, string, error) {
	mr, err := ctx.Req.MultipartReader()
	if err != nil {
		return "", "", err
	}
	var part *multipart.Part
	part, err = mr.NextPart()
	if err != nil {
		return "", "", err
	}
	//savedir := econf.GOPT.WorkDir + "/" + project + "/" + cluster + "/" + app
	// var dirst os.FileInfo
	if _, err = os.Stat(savedir); err != nil {
		if err := os.MkdirAll(savedir, 0755); err != nil {
			err = fmt.Errorf("mkdir %s failed: %v", savedir, err)
			return "", "", err
		}
	}
	for {
		if part.FormName() == "files" {
			break
		}
		part, err = mr.NextPart()
		if err != nil {
			return "", "", err
		}
	}
	postname := part.FileName()
	reg := regexp.MustCompile(allowExt)
	tmpname := "multipart-" + bson.NewObjectId().Hex()
	if postname != "" && reg.MatchString(postname) {
		// var b bytes.Buffer
		// io.CopyN(&b, part, 1048576+1)
		// econf.LogDebug("[%s] upload file %s with content: %s\n", logKind, fname, b.String())
		/*
						Content-Range:bytes 0-10485759/202286246
						Content-Type:multipart/form-data; boundary=----WebKitFormBoundaryduq7g7PEGTVOEwWh
						------WebKitFormBoundarydCgVRryb2qJPD2kR
			Content-Disposition: form-data; name="checkbox1"

			csv
			------WebKitFormBoundarydCgVRryb2qJPD2kR
			Content-Disposition: form-data; name="files"; filename="test.csv"
			Content-Type: text/csv


			------WebKitFormBoundarydCgVRryb2qJPD2kR--
		*/
		// wf := bufio.NewWriter(os.Stdout)
		// defer wf.Flush()
		var chunkStart, chunkEnd, chunkTotal int64
		chunk := ctx.Req.Header.Get("Content-Range")
		if chunk != "" {
			i := strings.IndexByte(chunk[6:], '-')
			j := strings.IndexByte(chunk[6:], '/')
			chunkStart, _ = strconv.ParseInt(chunk[6:6+i], 10, 0)
			chunkEnd, _ = strconv.ParseInt(chunk[6+i+1:6+j], 10, 0)
			chunkTotal, _ = strconv.ParseInt(chunk[6+j+1:], 10, 0)
			if i == -1 || j == -1 {
				err = errors.New("InvalidRequest: " + chunk)
				return "", "", err
			}
			tmpname = postname
		}
		//econf.LogDebug("[%s] chunk info: %s\n", logKind, chunk)
		fname := savedir + "/" + tmpname
		var f *os.File
		var fstat os.FileInfo
		if chunkStart == 0 {
			f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
		} else {
			fstat, err = os.Stat(fname)
			if err != nil {
				err = errors.New("no first chunk request")
				return "", "", err
			}
			if fstat.Size() != chunkStart {
				err = errors.New("file size not eq total chunk size")
				return "", "", err
			}
			// fmt.Fprintf(wf, "%s size: %d\n", fstat.Name(), fstat.Size())
			f, err = os.OpenFile(fname, os.O_APPEND|os.O_RDWR, 0644)
		}
		if err != nil {
			err = fmt.Errorf("open file %s failed: %v\n", fname, err)
			return "", "", err
		}
		// var wn int64
		_, err = io.CopyN(f, part, chunkEnd-chunkStart)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		// fmt.Fprintf(wf, "write size: %d\n", wn)
		if err != nil {
			err = fmt.Errorf("Write file %s failed: %v\n", fname, err)
			os.Remove(f.Name())
			return "", "", err
		}
		//文件全部上传完成
		if (chunkTotal != 0 && chunkEnd == chunkTotal) || chunkTotal == 0 {
			return postname, tmpname, nil
		} else {
			return postname, tmpname, Chunking
		}
	}
	err = fmt.Errorf("InvalidRequest")
	return "", "", err
}
