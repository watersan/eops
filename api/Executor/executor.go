package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/jobs"
	"opsv2.cn/opsv2/api/opsv2tmpl"
	"opsv2.cn/opsv2/daemon"
)

/*
1、所有参数通过环境变量获取，flag部分也会被环境变量覆盖。只处理ZDOPS_开始的环境变量。
2、task也通过环境变量获取。将TaskAttr结构json后，再base64_encode。
3、
*/

//Executor 提供给执行器的参数
type Executor struct {
	//ExeAttr *jobs.ExeAttr `json:"exeattr"`
	Env    map[string]string
	OutBuf *bytes.Buffer
	ErrBuf *bytes.Buffer
	/*Status 定义
	0：执行成功
	3：执行失败
	900：回滚执行成功
	901：回滚执行失败
	911：程序异常
	912：没有任务
	*/
	Status int
	Result *jobs.HostLogs
}

var (
	//Debug 调试
	Debug         = flag.Bool("d", false, "Debug")
	front         = flag.Bool("f", false, "Front runing")
	newestJob     = flag.Bool("n", false, "newest job's version in git")
	jobName       = flag.String("j", "no", "job name")
	jobArgs       = flag.String("a", "", "job args")
	prevJobResult = flag.String("b", "no", "previous job exec result")
	lkey          = []byte("o[$WqlN{iF~\\w`Im=oTY>}\"+iEG.d>QC")
	sercet        = []byte("x1nJDZzi6rS1B4d7LPx0nEXYN0F3JSP0pPTY65eN7zOl9e9VA0rZ3fYN1VUd7cnsreBf3qcSghDdn6ufpx7f37r6Xf2TDaeHFYMS")
	savePath      = "/data/scripts"
	outFile       = "/tmp/executor.out"
	errFile       = "/tmp/executor.err"
	logFile       = "/tmp/executor.log"
	logger        = &common.Logger{}
)

func readFile(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data := make([]byte, 1000)
	count, err := file.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:count], nil
}

func saveFile(name string, content []byte, aes bool) error {
	var perm os.FileMode
	perm = 0755
	//var decodeContent []byte
	if aes {
		perm = 0600
		encryptData, err := common.AesEncrypt(content, lkey)
		if err != nil {
			return err
		}
		encodeBase64 := base64.StdEncoding
		content = make([]byte, encodeBase64.EncodedLen(len(encryptData)))
		encodeBase64.Encode(content, encryptData)
	}
	if _, err := os.Stat(name); err == nil {
		os.Remove(name)
	}
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	return err
}

func (jobExecutor *Executor) osInfo() {
	executor := exec.Command("/bin/uname")
	out, err := executor.Output()
	if err != nil {
		jobExecutor.Env["ZDOPS_OS_Type"] = "unknown"
		return
	}
	jobExecutor.Env["ZDOPS_OS_Type"] = string(out)
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer file.Close()
	data := make([]byte, 1024)
	if _, err = file.Read(data); err != nil {
		return
	}
	oskeys := []string{"NAME", "VERSION", "ID", "PRETTY_NAME", "VERSION_ID"}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		l := len(line)
		for _, oskey := range oskeys {
			//oskey加上后面的等号
			oskeylen := len(oskey) + 1
			if l <= oskeylen {
				continue
			}
			if bytes.Compare(line[:oskeylen], []byte(oskey+"=")) == 0 {
				var v []byte
				if line[oskeylen] == '"' {
					v = line[oskeylen+1 : l-1]
				} else {
					v = line[oskeylen:]
				}
				jobExecutor.Env["ZDOPS_OS_"+oskey] = string(v)
				continue
			}
		}
	}

	return
}
func (jobExecutor *Executor) checkSum(job *jobs.TaskAttr) bool {
	// if _, ok := jobExecutor.Env["ZDOPS_JobVer"]; !ok {
	// 	return false
	// }
	name := jobExecutor.Env["ZDOPS_SavePath"] + "/" + job.Path + "/" + job.Name
	data, err := readFile(name + ".sum")
	if err != nil {
		logger.LogDebug("Read %s.sum failed: %v\n", name, err)
		return false
	}
	var fmd5 []byte
	fmd5, err = common.FileMd5(name, sercet)
	if err != nil {
		logger.LogWarning("%s md5 failed: %v\n", name, err)
		return false
	}
	secretdata := make([]byte, 1000)
	var count int
	count, err = base64.StdEncoding.Decode(secretdata, data)
	if err != nil {
		logger.LogWarning("%s.sum base64_decode failed: %v\n", name, err)
		return false
	}
	var sumdata []byte
	sumdata, err = common.AesDecrypt(secretdata[:count], lkey)
	if err != nil {
		logger.LogWarning("%s.sum AesDecrypt failed: %v\n", name, err)
		return false
	}
	if bytes.Compare(sumdata[:39], []byte(job.CommitID)) != 0 ||
		bytes.Compare(sumdata[40:], fmd5) != 0 {
		return false
	}
	return true
}

type gittree struct {
	Path string
	Type string `json:"type"`
	Mode string
	Name string
	id   string
}

//获取配置文件，并通过模板模块进行转换。
func (jobExecutor *Executor) appConf() error {
	var giturl, ref, savepath string
	var err error
	//判断部署类型是否是配置文件或模板配置文件
	if strings.Index(jobExecutor.Env["ZDOPS_DeployType"], "apptplconf") == -1 &&
		strings.Index(jobExecutor.Env["ZDOPS_DeployType"], "appconf") == -1 {
		return nil
	}
	// if strings.Index(jobExecutor.Env["ZDOPS_DeployType"], "apptplconf") >= 0 {
	// 	giturl = jobExecutor.Env["ZDOPS_AppTplConfPath"]
	// 	ref = jobExecutor.Env["ZDOPS_AppTplConfVer"]
	// 	savepath = fmt.Sprintf("%s/conf/apptpl/%s/%s",
	// 		jobExecutor.Env["ZDOPS_AppInstallPath"],
	// 		jobExecutor.Env["ZDOPS_AppTpl"],
	// 		ref,
	// 	)
	// } else if strings.Index(jobExecutor.Env["ZDOPS_DeployType"], "appconf") >= 0 {
	giturl = jobExecutor.Env["ZDOPS_AppConfPath"]
	ref = jobExecutor.Env["ZDOPS_AppConfVer"]
	savepath = fmt.Sprintf("%s/conf/%s/%s/%s/%s",
		jobExecutor.Env["ZDOPS_AppInstallPath"],
		jobExecutor.Env["ZDOPS_Project"],
		jobExecutor.Env["ZDOPS_Cluster"],
		jobExecutor.Env["ZDOPS_App"],
		ref,
	)
	// } else {
	// 	return nil
	// }
	if giturl[:7] != "http://" && giturl[:8] != "https://" {
		err = fmt.Errorf("git url is invalid")
		return err
	}
	urllen := len(giturl)
	if giturl[urllen-4:] == ".git" {
		urllen -= 4
	}
	var i int
	if i = strings.IndexByte(giturl[9:], '/'); i < 0 {
		err = fmt.Errorf("git url do not contain a project")
		return err
	}
	i += 9
	git := common.GitAction{
		Action:   "tree",
		Method:   "GET",
		Projects: giturl[i+1 : urllen],
		Url:      giturl[:i],
		GitToken: jobExecutor.Env["ZDOPS_GitToken"],
	}
	//获取project的目录数，遍历所有文件进行下载并进行模板匹配
	if err = git.GitAPI(); err != nil {
		return err
	}
	gittrees := []gittree{}
	if err = json.Unmarshal(git.Result, &gittrees); err != nil {
		return err
	}
	git.Ref = ref
	var dir os.FileInfo
	for _, gitfile := range gittrees {
		if gitfile.Type == "blob" {
			fmt.Printf("gitfile: %s\n", gitfile.Name)
			git.Path = gitfile.Path
			git.Result = []byte{}
			git.Action = "files"
			if err = git.GitAPI(); err != nil {
				return err
			}
			gfile := common.GitFile{}
			if err = json.Unmarshal(git.Result, &gfile); err != nil {
				return err
			}
			//截取目录部分，并判断目录是否存在，否则进行创建
			i := strings.LastIndexByte(gitfile.Path, '/')
			path := savepath
			if i > 0 {
				path += "/" + gitfile.Path[:i]
			}
			dir, err = os.Stat(path)
			if err == nil {
				if !dir.IsDir() {
					err = fmt.Errorf("%s is exist，but is not dir", path)
					return err
				}
			}
			os.MkdirAll(path, 0755)
			var f *os.File
			if f, err = os.OpenFile(path+"/"+gitfile.Name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644); err != nil {
				logger.LogDebug("[conf] can't open file to write: %v\n", err)
				return err
			}
			myBase64 := base64.StdEncoding
			confContent := make([]byte, myBase64.EncodedLen(len(gfile.Content)))
			if _, err = myBase64.Decode(confContent, []byte(gfile.Content)); err != nil {
				return err
			}
			var tmpl *template.Template
			tmpl, err = template.New("conf").Funcs(opsv2tmpl.FuncMap).Parse(string(confContent))
			if err != nil {
				return err
			}
			err = tmpl.Execute(f, jobExecutor.Env)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				os.Remove(f.Name())
				return err
			}
		}
	}
	return nil
}

func (jobExecutor *Executor) upload() error {
	tmp, err := readFile("/tmp/ZDOPS_Build-" + jobExecutor.Env["ZDOPS_TaskLogID"])
	if err != nil {
		return err
	}
	fpath := string(tmp)
	var file *os.File
	file, err = os.Open(fpath)
	if err != nil {
		return err
	}
	defer file.Close()
	var fstat os.FileInfo
	if fstat, err = os.Stat(fpath); err != nil {
		return err
	}
	var fname string
	if n := strings.LastIndexByte(fpath, '/'); n >= 0 {
		fname = fpath[n+1:]
	} else {
		fname = fpath
	}
	query := url.Values{}
	query.Set("keyid", jobExecutor.Env["ZDOPS_KeyID"])
	query.Set("secret", jobExecutor.Env["ZDOPS_Secret"])
	var fmd5 []byte
	fmd5, err = common.FileMd5(fpath, []byte{})
	if err != nil {
		logger.LogWarning("%s md5 failed: %v\n", fpath, err)
		return err
	}
	query.Set("md5", hex.EncodeToString(fmd5))
	// params.Set("build", common.TrueStr)
	if jobExecutor.Env["ZDOPS_DeployType"] == common.ApptplName {
		query.Set(common.ApptplName, jobExecutor.Env["ZDOPS_AppTplName"])
	} else {
		query.Set("p", jobExecutor.Env["ZDOPS_Project"])
		query.Set("c", jobExecutor.Env["ZDOPS_Cluster"])
		query.Set("a", jobExecutor.Env["ZDOPS_App"])
	}
	var apiurl *url.URL
	apiurl, err = url.ParseRequestURI(jobExecutor.Env["ZDOPS_CallBack"])
	if err != nil {
		return err
	}
	apiurl.Path = "/eops/deploy/upload"
	apiurl.RawPath = "/eops/deploy/upload"
	apiurl.RawQuery = query.Encode()

	totalsize := fstat.Size()
	chunksize := 1048576
	var n int
	var i int64
	iseof := false
	for {
		p := make([]byte, chunksize)
		if n, err = file.ReadAt(p, i); err == io.EOF {
			iseof = true
		} else if err != nil {
			return err
		}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("files", fname)
		if err != nil {
			return err
		}
		part.Write(p)
		err = writer.Close()
		if err != nil {
			return err
		}
		rc := ioutil.NopCloser(body)
		header := make(http.Header)
		header.Set("User-Agent", "Job Executor")
		header.Set("Content-Type", writer.FormDataContentType())
		header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", i, i+int64(n), totalsize))
		header.Set("X-APP-Ver", "1.0")
		logger.LogDebug("[upload] bytes %d-%d/%d", i, i+int64(n), totalsize)
		// header.Add("Content-Type", "application/json")
		req := &http.Request{
			Method:     "POST",
			URL:        apiurl,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     header,
			Body:       rc,
			Host:       apiurl.Host,
		}
		client := &http.Client{Timeout: time.Second * 10}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("[upload] http status not ok: %v", resp)
		}
		var content []byte
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			// logger.LogDebug("[git] read response failed\n")
			return fmt.Errorf("no upload api response body: %v\n", err)
		}
		result := common.RespMsg{}
		err = json.Unmarshal(content, &result)
		if err != nil {
			return err
		}
		if result.Code != 0 {
			err = errors.New(result.Message)
			return err
		}

		i += int64(n)

		if n < chunksize || iseof {
			break
		}
	}
	return nil
}

/*
	下载软件包，并下载软件包的md5文件，进行md5对比
*/
func (jobExecutor *Executor) download() error {
	if jobExecutor.Env["ZDOPS_Build"] == common.TrueStr {
		return nil
	}
	dtype := jobExecutor.Env["ZDOPS_DeployType"]
	if dtype == "" || (dtype[:6] != "apptpl" && dtype[:3] != "app") {
		return nil
	}
	apiurl, err := url.ParseRequestURI(jobExecutor.Env["ZDOPS_AppSource"])
	if err != nil {
		return err
	}
	var fname string
	if jobExecutor.Env["ZDOPS_DeployType"][:3] == "app" {
		apiurl.Path += fmt.Sprintf("%s/%s/%s-%s.zip",
			jobExecutor.Env["ZDOPS_Project"],
			jobExecutor.Env["ZDOPS_Cluster"],
			jobExecutor.Env["ZDOPS_App"],
			jobExecutor.Env["ZDOPS_AppVer"],
		)
	} else {
		apiurl.Path += fmt.Sprintf("apptpl/%s-%s.zip",
			jobExecutor.Env["ZDOPS_AppTpl"],
			jobExecutor.Env["ZDOPS_AppTplVer"],
		)
	}
	fname = fmt.Sprintf("%s/%s/%s",
		jobExecutor.Env["ZDOPS_AppInstallPath"],
		jobExecutor.Env["ZDOPS_Project"],
		jobExecutor.Env["ZDOPS_Cluster"],
	)
	if err = os.MkdirAll(fname, 0755); err != nil {
		return err
	}
	fname += jobExecutor.Env["ZDOPS_App"] + "-" + jobExecutor.Env["ZDOPS_AppVer"] + ".zip"
	query := url.Values{}
	query.Set("keyid", jobExecutor.Env["ZDOPS_KeyID"])
	query.Set("secret", jobExecutor.Env["ZDOPS_Secret"])
	apiurl.RawQuery = query.Encode()
	header := make(http.Header)
	header.Add("User-Agent", "Job Executor")
	header.Add("Content-Type", "application/zip")
	req := &http.Request{
		Method:     "GET",
		URL:        apiurl,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       apiurl.Host,
	}
	client := &http.Client{Timeout: time.Second * 1800}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogDebug("[down] http failed: %v\n", resp)
		return err
	}
	if resp.StatusCode != 200 {
		logger.LogDebug("[down] http status not ok: %v\n", resp)
		err = errors.New(resp.Status)
		return err
	}
	var f *os.File
	if f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644); err != nil {
		logger.LogDebug("[down] can't open file to write: %v\n", err)
		return err
	}
	var wn int64
	wn, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		logger.LogWarning("[down] Write file %s failed: %v\n", fname, err)
		os.Remove(f.Name())
		return err
	}
	resp.Body.Close()
	var downlen int
	if downlen, err = strconv.Atoi(resp.Header.Get("Content-Length")); err != nil || int64(downlen) != wn {
		os.Remove(f.Name())
		return fmt.Errorf("[down] Content-Length: %d not equal to recv length: %d", downlen, wn)
	}
	req.URL.Path += ".md5"
	client = &http.Client{Timeout: time.Second * 1800}
	resp, err = client.Do(req)
	if err != nil {
		logger.LogDebug("[down] http failed: %v\n", resp)
		return nil
	}
	if resp.StatusCode == 200 {
		var content []byte
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.LogDebug("[down] read response failed\n")
			os.Remove(f.Name())
			return err
		}
		var fmd5 []byte
		if fmd5, err = common.FileMd5(fname, []byte{}); err != nil || bytes.Compare(content, fmd5) != 0 {
			os.Remove(f.Name())
			return err
		}
	}
	return nil
}

//GetJobContent 获取文件，并保存到本地
func (jobExecutor *Executor) GetJobContent(job *jobs.TaskAttr) error {
	apiurl, err := url.ParseRequestURI(jobExecutor.Env["ZDOPS_JobAPI"])
	if err != nil {
		return err
	}
	// n := strings.LastIndexByte(*jobName, '/')
	// jname := *jobName
	query := url.Values{}
	query.Set("path", job.Path)
	query.Set("name", job.Name)
	query.Set("keyid", jobExecutor.Env["ZDOPS_KeyID"])
	query.Set("secret", jobExecutor.Env["ZDOPS_Secret"])
	if *newestJob == false {
		query.Set("releasejob", common.TrueStr)
	}
	apiurl.RawQuery = query.Encode()
	header := make(http.Header)
	header.Add("User-Agent", "Job Executor")
	header.Add("Content-Type", "application/json")
	req := &http.Request{
		Method:     "GET",
		URL:        apiurl,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       apiurl.Host,
	}
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogDebug("[git] http failed: %v\n", resp)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.LogDebug("[git] http status not ok: %v\n", resp)
		err = errors.New(resp.Status)
		return err
	}
	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogDebug("[git] read response failed\n")
		return err
	}
	result := common.RespMsg{}
	err = json.Unmarshal(content, &result)
	if err != nil {
		return err
	}
	if result.Code != 0 {
		err = errors.New(result.Message)
		return err
	}
	jobinfo := result.DataList.([]interface{})[0].(map[string]interface{})
	jobcontent := jobinfo["content"].(string)
	_, err = os.Stat(jobExecutor.Env["ZDOPS_SavePath"] + "/" + job.Path)
	if err != nil {
		os.MkdirAll(jobExecutor.Env["ZDOPS_SavePath"]+"/"+job.Path, 0755)
	}
	myBase64 := base64.StdEncoding
	scriptContent := make([]byte, myBase64.EncodedLen(len(jobcontent)))
	if _, err = myBase64.Decode(scriptContent, []byte(jobcontent)); err != nil {
		return err
	}
	if err = saveFile(jobExecutor.Env["ZDOPS_SavePath"]+"/"+job.Path+"/"+job.Name, scriptContent, false); err != nil {
		return err
	}
	var fmd5 []byte
	fmd5, err = common.FileMd5(jobExecutor.Env["ZDOPS_SavePath"]+"/"+job.Path+"/"+job.Name, sercet)
	if err != nil {
		return err
	}
	vinfo := make([]byte, 56)
	copy(vinfo[:39], []byte(jobinfo["commitid"].(string)))
	copy(vinfo[40:], fmd5)
	err = saveFile(jobExecutor.Env["ZDOPS_SavePath"]+"/"+job.Path+"/"+job.Name+".sum", vinfo, true)
	return err
}

//CallBack 回调接口
func (jobExecutor *Executor) CallBack() error {
	if jobExecutor.Env["ZDOPS_Build"] != "" {
		if err := jobExecutor.upload(); err != nil {
			return fmt.Errorf("Upload failed: %v\n", err)
		}
	}
	apiurl, err := url.ParseRequestURI(jobExecutor.Env["ZDOPS_CallBack"])
	if err != nil {
		return err
	}
	query := url.Values{}
	query.Set("keyid", jobExecutor.Env["ZDOPS_KeyID"])
	query.Set("secret", jobExecutor.Env["ZDOPS_Secret"])
	apiurl.RawQuery = query.Encode()
	header := make(http.Header)
	header.Add("User-Agent", "Job Executor")
	header.Add("Content-Type", "application/json")
	req := http.Request{
		Method:     "POST",
		URL:        apiurl,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       apiurl.Host,
	}
	// if jobExecutor.Env["ZDOPS_LastJob"] == "yes" && jobExecutor.Result.Jobs[0].Name != "lastjob" {
	// 	jobExecutor.Result.Jobs = append(jobExecutor.Result.Jobs, &jobs.JobResult{Name: "lastjob"})
	// }
	cbr := jobs.CBResult{
		ID:     jobExecutor.Env["ZDOPS_TaskLogID"],
		Env:    jobExecutor.Env,
		Result: jobExecutor.Result,
		Status: jobExecutor.Status,
	}
	jsonData, _ := json.Marshal(cbr)
	logger.LogDebug("json: %s\n", string(jsonData))
	req.Body = ioutil.NopCloser(bytes.NewReader(jsonData))
	client := http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(&req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		return err
	}
	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	result := common.RespMsg{}
	err = json.Unmarshal(content, &result)
	if err != nil {
		return err
	}
	if result.Code != 0 {
		err = errors.New(result.Message)
		return err
	}
	return nil
}

//4X-w5kqayVKGpgtdoMsy
func main() {
	flag.Parse()
	jobExecutor := Executor{
		Env:    make(map[string]string),
		OutBuf: bytes.NewBuffer([]byte{}),
		ErrBuf: bytes.NewBuffer([]byte{}),
		Result: &jobs.HostLogs{},
	}
	logf, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("log open failed. %v", err)
		os.Exit(1)
	}
	log := log.New(logf, "", 0)
	log.SetFlags(7)
	logger.Log = log
	logger.Level = 7
	if *Debug {
		logger.Level = 15
	}
	cmdEnv := []string{}
	logger.LogDebug("Env: %s\n", os.Environ())
	for _, envstr := range os.Environ() {
		if envstr[:6] != "ZDOPS_" && envstr[:4] != "job_" {
			continue
		}
		cmdEnv = append(cmdEnv, envstr)
		if n := strings.IndexByte(envstr, '='); n > 0 {
			value := envstr[n+1:]
			switch envstr[:n] {
			case "ZDOPS_ExecFront":
				if value == common.TrueStr {
					*front = true
				}
			case "ZDOPS_NewestJob":
				if value == common.TrueStr {
					*newestJob = true
				}
			case "ZDOPS_ExecDebug":
				if value == common.TrueStr {
					logger.Level = 15
				}
			default:
				jobExecutor.Env[envstr[:n]] = value
			}
			//logger.LogInfo("ENV: %s= %s\n", envstr[:n], envstr[n+1:])
		}
	}

	dmfiles := []string{"/dev/stdin", outFile, errFile}
	if *front == false {
		daemon.Daemon("", dmfiles, "")
	}
	// sigs := make(chan os.Signal, 10)
	// signal.Notify(sigs, syscall.SIGHUP)
	// go func(sigs chan os.Signal) {
	// 	for {
	// 		select {
	// 		case <-sigs:
	// 			continue
	// 		}
	// 	}
	// }(sigs)
	/*
		通过Notify接收到信号后，主进程不会退出，但是exec.Command的命令还是会收到HUP信号并退出。
		用Ignore忽视信号，exec.Command也会忽视。
	*/
	signal.Ignore(syscall.SIGHUP)

	if *jobName != "no" {
		if jobExecutor.Env["ZDOPS_Task"] == "" {
			jobExecutor.Env["ZDOPS_Task"] = *jobName
			//jobExecutor.Env["ZDOPS_Task1_Args"] = *jobArgs
		}
		// logger.LogCritical("no job")
		// os.Exit(1)
	}
	taskIndex := -1

	defer func() {
		if err := recover(); err != nil {
			logger.LogCritical("[Recovery] Error: %s\n", fmt.Sprintf("%+v", errors.Errorf("%v", err)))
			jobExecutor.Status = 911
			if cberr := jobExecutor.CallBack(); cberr != nil {
				logger.LogCritical("CallBack failed: %s", cberr.Error())
			}
		} else {
			if taskIndex >= 0 {
				jobExecutor.Result.Jobs[taskIndex].ErrOut = jobExecutor.ErrBuf.String()
			} else {
				jobExecutor.Status = 912
			}
			if cberr := jobExecutor.CallBack(); cberr != nil {
				logger.LogCritical("CallBack failed: %s", cberr.Error())
			}
		}
	}()
	// if taskNum > 0 {
	// 	jobExecutor.Result.Jobs = make([]*jobs.JobResult, taskNum)
	// } else {
	// 	return
	// }
	var ok bool
	if jobExecutor.Result.HostName, ok = jobExecutor.Env["ZDOPS_HostName"]; !ok {
		jobExecutor.Result.HostName, _ = os.Hostname()
	}
	if _, ok = jobExecutor.Env["ZDOPS_SavePath"]; !ok {
		jobExecutor.Env["ZDOPS_SavePath"] = savePath
	}
	var tmp []byte
	var taskstr string
	if taskstr, ok = jobExecutor.Env["ZDOPS_Task"]; !ok {
		logger.LogWarning("No Task")
		return
	}
	task := &jobs.TaskAttr{}
	if tmp, err = base64.StdEncoding.DecodeString(taskstr); err == nil {
		logger.LogDebug("Task json: %s\n", string(tmp))
		if err = json.Unmarshal(tmp, task); err != nil {
			logger.LogWarning("The task %d is not TaskAttr struct: %s\n", taskIndex, taskstr)
			//jobExecutor.ErrBuf.WriteString(fmt.Sprintf("The task %d is not TaskAttr struct", taskIndex))
			return
		}
	} else {
		logger.LogWarning("The task %d cant decode_base64: %s\n", taskIndex, taskstr)
		//jobExecutor.ErrBuf.WriteString(fmt.Sprintf("The task %d cant decode_base64", taskIndex))
		return
	}
	delete(jobExecutor.Env, "ZDOPS_Task")
	taskIndex = 0
	if err = jobExecutor.download(); err != nil {
		jobExecutor.Result.Jobs[0].ErrOut = err.Error()
		return
	}
	if err = jobExecutor.appConf(); err != nil {
		jobExecutor.Result.Jobs[0].ErrOut = err.Error()
		return
	}
	//begintask := task
	for {
		jobResult := &jobs.JobResult{}
		jobResult.Name = task.Path + "/" + task.Name
		jobResult.Argv = task.Argv
		jobResult.Alias = task.Alias
		jobExecutor.Result.Jobs = append(jobExecutor.Result.Jobs, jobResult)
		logger.LogDebug("Task %s: %s\n", task.Alias, jobResult.Name)
		jobExecutor.OutBuf.Reset()
		jobExecutor.ErrBuf.Reset()

		/*
			如果没有版本信息（commitid），或校验本地文件失败（对比commitid和文件md5）时，
			会从git重新拉取最新版本内容。
		*/
		if !jobExecutor.checkSum(task) {
			err = jobExecutor.GetJobContent(task)
			if err != nil {
				logger.LogCritical("Get git %s failed: %v\n", *jobName, err)
				jobExecutor.ErrBuf.WriteString(err.Error())
				jobExecutor.Result.Jobs[taskIndex].Code = -9
				return
			}
		}
		logger.LogDebug("Create exec: %s\n", jobResult.Name)
		cmd := exec.Command(jobExecutor.Env["ZDOPS_SavePath"]+"/"+task.Path+"/"+task.Name, task.Argv)
		cmd.Env = append(os.Environ(), cmdEnv...)
		if taskIndex >= 1 {
			prevjob := make([]string, 4)
			prevjob[0] = "ZDOPS_PrevJobName=" + jobExecutor.Result.Jobs[taskIndex-1].Name
			prevjob[1] = "ZDOPS_PrevJobArgv=" + jobExecutor.Result.Jobs[taskIndex-1].Argv
			prevjob[2] = fmt.Sprintf("ZDOPS_PrevJobCode=%d", jobExecutor.Result.Jobs[taskIndex-1].Code)
			prevjob[3] = "ZDOPS_PrevJobAlias=" + jobExecutor.Result.Jobs[taskIndex-1].Alias
			cmd.Env = append(cmd.Env, prevjob...)
		}
		var userinfo *user.User
		if userinfo, err = user.Lookup(task.User); err != nil {
			logger.LogCritical("The user %s is invalid", task.User)
			jobExecutor.ErrBuf.WriteString("Invalid User")
			jobResult.Code = -9
			return
		}
		credential := new(syscall.Credential)
		var i int
		i, _ = strconv.Atoi(userinfo.Uid)
		credential.Uid = uint32(i)
		i, _ = strconv.Atoi(userinfo.Gid)
		credential.Gid = uint32(i)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: credential,
		}
		cmd.Stdout = jobExecutor.OutBuf
		cmd.Stderr = jobExecutor.ErrBuf
		logger.LogDebug("Start task: %s\n", jobResult.Name)
		err = cmd.Start()
		tbegin := time.Now().UnixNano()
		if err != nil {
			logger.LogCritical("%s/%s start failed: %v\n", task.Path, task.Name, err)
			jobExecutor.ErrBuf.WriteString(err.Error())
			jobResult.Code = -10
			return
		}
		syscall.Setpriority(syscall.PRIO_PROCESS, cmd.Process.Pid, task.Priority)
		time.AfterFunc(time.Duration(task.Timeout)*time.Second, func() { cmd.Process.Kill() })
		err = cmd.Wait()
		if err != nil {
			logger.LogDebug("Task %s exec failed: %v\n", jobResult.Name, err)
			jobExecutor.ErrBuf.Write([]byte(err.Error()))
		}
		// if task.Priority > -10 {
		// 	syscall.Setpriority(syscall.PRIO_PROCESS, os.Getpid(), -10)
		// }
		sys := cmd.ProcessState.Sys().(syscall.WaitStatus)
		logger.LogDebug("Task %s(%s) exec result: %d\n", jobResult.Name, jobResult.Alias, sys.ExitStatus())
		runtime := time.Now().UnixNano() - tbegin
		sysusage := cmd.ProcessState.SysUsage().(*syscall.Rusage)
		jobResult.RunTime = float32(runtime / 1e6)
		jobResult.UsedCPU = float32(cmd.ProcessState.UserTime()+cmd.ProcessState.SystemTime()) * 10000 / float32(runtime)
		jobResult.UsedMEM = int(sysusage.Maxrss / 1024)
		jobResult.Code = sys.ExitStatus()
		if jobResult.Code != 0 {
			jobExecutor.Status = 3
			if jobExecutor.Status == 900 {
				jobExecutor.Status = 901
			}
		}
		jobResult.StdOut = jobExecutor.OutBuf.String()
		jobResult.ErrOut = jobExecutor.ErrBuf.String()
		//jobExecutor.Result.Jobs[0].Env = jobExecutor.Env
		if jobResult.Code == 0 {
			if task.Next != nil {
				task = task.Next
			} else {
				return
			}
		} else {
			if task.Rescue != nil {
				task = task.Rescue
				jobExecutor.Status = 900
			} else {
				return
			}
		}
		taskIndex++
	}
	// if *prevJobResult != "no" {
	// 	prevjob := make(map[string]interface{})
	// 	if err = json.Unmarshal([]byte(*prevJobResult), &prevjob); err == nil {
	// 		cmdEnv = append(cmdEnv, "ZDOPS_PrevJob-stdout="+prevjob["stdout"].(string))
	// 		cmdEnv = append(cmdEnv, "ZDOPS_PrevJob-stderr="+prevjob["stderr"].(string))
	// 		cmdEnv = append(cmdEnv, "ZDOPS_PrevJob-rc="+strconv.FormatInt(int64(prevjob["rc"].(float64)), 10))
	// 	}
	// }

	//fmt.Printf("total: %d; user: %d; sys: %d\n", time.Now().UnixNano()-tbegin, cmd.ProcessState.UserTime(), cmd.ProcessState.SystemTime())
	//fmt.Printf("Maxrss: %d; Ixrss: %d; Idrss: %d; Isrss: %d\n", sysusage.Maxrss, sysusage.Ixrss, sysusage.Idrss, sysusage.Isrss)
	// fmt.Printf("err: %v\n", err)
	// fmt.Printf("exit: %v\n", sys.ExitStatus())
	// fmt.Printf("out: %s\n", exe.OutBuf.String())
	// fmt.Printf("err: %s", exe.ErrBuf.String())
}
