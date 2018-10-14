package cmdb

import (
	"fmt"
	"os"
	"strconv"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/dale-di/ucloud-sdk-go/service/uhost"
	"github.com/dale-di/ucloud-sdk-go/ucloud"
	"github.com/dale-di/ucloud-sdk-go/ucloud/auth"
)

func ImportFromCSV(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	id := ctx.Req.Form.Get("id")
	cdenv := ctx.Req.Form.Get("env")
	rIDC := &MgoIDCProvider{}
	var err error
	if _, err = econf.Mgo.Query(mgoIDCProvider, bson.ObjectIdHex(id), rIDC, 0, 0, nil); err != nil {
		econf.LogWarning("[cmdb] Not found job recoder %s with: %v\n", id, err)
		ctx.Response = 1301
		return
	}
	var csvname string
	_, csvname, err = ctx.Upload("/tmp", `\.csv$`)
	if err != nil {
		econf.LogWarning("[CMDB] upload csv %s failed: %v\n", csvname, err)
		ctx.Response = 1309
		return
	}
	var records []common.V
	fieldList := []string{"ip", "hostid", "hostname", "eip", "cpu", "memory", "disk", "os", "osver"}
	records, err = common.ReadCSV("/tmp/"+csvname, fieldList)
	if err != nil {
		econf.LogDebug("[CMDB] read csv failed: %v\n", err)
		ctx.Response = 1309
		return
	}
	var total, insCount int
	total = len(records)
	c := econf.Mgo.Coll(mgoResource)
	for _, item := range records {
		cpu, _ := strconv.Atoi(item["cpu"].(string))
		memory, _ := strconv.Atoi(item["memory"].(string))
		data := MgoResource{
			HostID:     item["hostid"].(string),
			HostName:   item["hostname"].(string),
			CPU:        cpu,
			Memory:     memory,
			IP:         item["ip"].(string),
			OS:         item["os"].(string),
			EIP:        item["eip"].(string),
			Disk:       item["disk"].(string),
			Location:   rIDC.Name,
			Modifytime: time.Now(),
			Createtime: time.Now(),
		}
		if cdenv != "" {
			data.Environment = cdenv
		}
		if err = common.InputFilter(&data, 1, nil, nil); err == nil {
			if err = c.Insert(data); err == nil {
				insCount++
			}
		}
		if err != nil {
			econf.LogDebug("[CMDB] csv insert record failed: %v\n", err)
		}
	}
	rmsg := common.NewRespMsg()
	rmsg.TotalRecord = total
	rmsg.DataList = common.V{"insert": insCount}
	ctx.OpLog.WriteString(fmt.Sprintf("向%s导入%d个资源", rIDC.Name, total))
	ctx.Response = rmsg
}

func ImportFromYun(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	id := ctx.Req.Form.Get("id")
	cdenv := ctx.Req.Form.Get("env")
	rYun := &MgoYunProvider{}
	var err error
	if _, err = econf.Mgo.Query(mgoYunProvider, bson.ObjectIdHex(id), rYun, 0, 0, nil); err != nil {
		econf.LogWarning("[cmdb] Not found job recoder %s with: %v\n", id, err)
		ctx.Response = 1301
		return
	}
	var total, upCount, insCount int
	switch rYun.Name {
	case "aliyun":
		total, upCount, insCount, err = fromAliyun(rYun, econf, cdenv)
		if err != nil {
			econf.LogDebug("[CMDB] aliyun import failed: %v; %d instance failed to import.\n", err, total-upCount-insCount)
		}
	case "aws":
		total, upCount, insCount, err = fromAws(rYun, econf, cdenv)
		if err != nil {
			econf.LogDebug("[CMDB] aws import failed: %v; %d instance failed to import.\n", err, total-upCount-insCount)
		}
	case "ucloud":
		total, upCount, insCount, err = fromUcloud(rYun, econf, cdenv)
		if err != nil {
			econf.LogDebug("[CMDB] Ucloud import failed: %v; %d instance failed to import.\n", err, total-upCount-insCount)
		}
	}
	rmsg := common.NewRespMsg()
	rmsg.TotalRecord = total
	rmsg.DataList = common.V{"update": upCount, "insert": insCount}
	ctx.OpLog.WriteString(fmt.Sprintf("从%s导入%d个资源", rYun.Name, total))
	ctx.Response = rmsg
}

func fromAliyun(rYun *MgoYunProvider, econf *common.EopsConf, cdenv string) (int, int, int, error) {
	var ecsClient *ecs.Client
	var total, n1, n2, upCount, insCount int
	var err error
	syncTime := time.Now()
	for _, region := range rYun.Regions {
		ecsClient, err = ecs.NewClientWithAccessKey(
			region,
			rYun.Key,
			rYun.Secret,
		)
		if err != nil {
			return 0, 0, 0, err
		}
		request := ecs.CreateDescribeInstancesRequest()
		request.PageSize = "100"
		var response *ecs.DescribeInstancesResponse
		response, err = ecsClient.DescribeInstances(request)
		if err != nil {
			// econf.LogDebug("[CMDB] aliyun DescribeInstances failed: %v\n", err)
			return 0, 0, 0, err
		}
		totalpage := int(response.TotalCount/100) + 1
		total += response.TotalCount
		n1, n2, err = aliInstances(response.Instances, ecsClient, econf, cdenv)
		upCount += n1
		insCount += n2
		if err != nil {
			return total, upCount, insCount, err
		}
		for i := 2; i <= totalpage; i++ {
			request.PageNumber = requests.NewInteger(i)
			response, err = ecsClient.DescribeInstances(request)
			if err != nil {
				return total, upCount, insCount, err
			}
			n1, n2, err = aliInstances(response.Instances, ecsClient, econf, cdenv)
			upCount += n1
			insCount += n2
			if err != nil {
				return total, upCount, insCount, err
			}
		}
	}
	c := econf.Mgo.Coll(mgoResource)
	if _, err := c.RemoveAll(common.V{"modifytime": common.V{"$lt": syncTime}, "location": "aliyun"}); err != nil {
		econf.LogInfo("[CMDB] clean resource from aliyun failed: %v\n", err)
	}
	return total, upCount, insCount, nil
}
func aliInstances(instances ecs.Instances, ecsClient *ecs.Client, econf *common.EopsConf, cdenv string) (int, int, error) {
	var upCount, insCount int
	var err error
	var response *ecs.DescribeDisksResponse
	c := econf.Mgo.Coll(mgoResource)
	for _, instance := range instances.Instance {
		if instance.Status != "Running" {
			continue
		}
		req := ecs.CreateDescribeDisksRequest()
		req.InstanceId = instance.InstanceId
		response, err = ecsClient.DescribeDisks(req)
		if err != nil {
			return upCount, insCount, err
		}
		var diskinfo string
		for _, disk := range response.Disks.Disk {
			path := disk.Device
			if disk.Type == "system" {
				path = "/"
			} else if disk.DiskName != "" {
				path = disk.DiskName
			}
			diskinfo += fmt.Sprintf("%s:%dG;", path, disk.Size)
		}
		eip := instance.EipAddress.IpAddress
		if len(instance.PublicIpAddress.IpAddress) > 0 {
			eip = instance.PublicIpAddress.IpAddress[0]
		}
		data := MgoResource{
			HostID:     instance.InstanceId,
			HostName:   instance.HostName,
			CPU:        instance.Cpu,
			Memory:     instance.Memory,
			Region:     instance.RegionId,
			IP:         instance.NetworkInterfaces.NetworkInterface[0].PrimaryIpAddress,
			OS:         instance.OSName,
			EIP:        eip,
			Disk:       diskinfo,
			Location:   "aliyun",
			Modifytime: time.Now(),
		}
		if cdenv != "" {
			data.Environment = cdenv
		}
		var cinfo *mgo.ChangeInfo
		if cinfo, err = c.Upsert(common.V{"hostid": instance.InstanceId}, data); err == nil {
			if cinfo.UpsertedId != nil {
				insCount++
			} else {
				upCount++
			}
		} else {
			econf.LogDebug("[CMDB] host %s import from aliyun failed: %v\n", instance.InstanceId, err)
		}
	}
	return upCount, insCount, nil
}

//AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY
func fromAws(rYun *MgoYunProvider, econf *common.EopsConf, cdenv string) (int, int, int, error) {
	//rmsg := common.NewRespMsg()
	os.Setenv("AWS_ACCESS_KEY_ID", rYun.Key)
	os.Setenv("AWS_SECRET_ACCESS_KEY", rYun.Secret)
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	}()
	var err error
	var total, upCount, insCount int
	syncTime := time.Now()
	c := econf.Mgo.Coll(mgoResource)
	for _, region := range rYun.Regions {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(region),
		}))
		ec2Svc := ec2.New(sess)
		var result *ec2.DescribeInstancesOutput
		result, err = ec2Svc.DescribeInstances(nil)
		if err != nil {
			return total, upCount, insCount, err
		}
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				total++
				var name string
				if *instance.State.Code > 16 {
					continue
				}
				for _, tag := range instance.Tags {
					if *tag.Key == "name" || *tag.Key == "Name" {
						name = *tag.Value
					}
				}
				data := MgoResource{
					HostID:     *instance.InstanceId,
					HostName:   name,
					Type:       *instance.InstanceType,
					Region:     *instance.Placement.AvailabilityZone,
					IP:         *instance.PrivateIpAddress,
					Location:   "aws",
					Modifytime: time.Now(),
				}
				if instance.PublicIpAddress != nil {
					data.EIP = *instance.PublicIpAddress
				}
				if cdenv != "" {
					data.Environment = cdenv
				}
				var cinfo *mgo.ChangeInfo
				if cinfo, err = c.Upsert(common.V{"hostid": *instance.InstanceId}, data); err == nil {
					if cinfo.UpsertedId != nil {
						insCount++
					} else {
						upCount++
					}
				} else {
					econf.LogDebug("[CMDB] host %s import from aws failed: %v\n", *instance.InstanceId, err)
				}
			}
		}
	}
	if _, err := c.RemoveAll(common.V{"modifytime": common.V{"$lt": syncTime}, "location": "aws"}); err != nil {
		econf.LogInfo("[CMDB] clean resource from aws failed: %v\n", err)
	}
	return total, upCount, insCount, nil
}

func fromUcloud(rYun *MgoYunProvider, econf *common.EopsConf, cdenv string) (int, int, int, error) {
	// var err error
	var total, upCount, insCount, n1, n2 int
	// c := econf.Mgo.Coll(mgoResource)
	syncTime := time.Now()
	for _, region := range rYun.Regions {
		hostsvc := uhost.New(&ucloud.Config{
			Credentials: &auth.KeyPair{
				PublicKey:  rYun.Key,
				PrivateKey: rYun.Secret,
			},
			Region:    region,
			ProjectID: rYun.ProjectID,
		})
		describeParams := uhost.DescribeUHostInstanceParams{
			Region: region,
			Limit:  100,
			Offset: 0,
		}
		response, err := hostsvc.DescribeUHostInstance(&describeParams)
		if err != nil {
			econf.LogDebug("[CMDB] DescribeUHostInstance failed: %v", err)
			continue
		}
		n1, n2, err = ucloudInstances(response.UHostSet, region, econf, cdenv)
		upCount += n1
		insCount += n2
		total = response.TotalCount
		totalpage := int(total/10) + 1
		for i := 1; i < totalpage; i++ {
			describeParams.Offset += describeParams.Limit
			response, err = hostsvc.DescribeUHostInstance(&describeParams)
			if err != nil {
				return total, upCount, insCount, err
			}
			n1, n2, err = ucloudInstances(response.UHostSet, region, econf, cdenv)
			upCount += n1
			insCount += n2
			if err != nil {
				return total, upCount, insCount, err
			}
		}
	}
	c := econf.Mgo.Coll(mgoResource)
	if _, err := c.RemoveAll(common.V{"modifytime": common.V{"$lt": syncTime}, "location": "ucloud"}); err != nil {
		econf.LogInfo("[CMDB] clean resource from aws failed: %v\n", err)
	}
	return total, upCount, insCount, nil
}

func ucloudInstances(instances uhost.UHostSetArray, region string, econf *common.EopsConf, cdenv string) (int, int, error) {
	var upCount, insCount int
	var err error
	c := econf.Mgo.Coll(mgoResource)
	for _, instance := range instances {
		var ip, diskinfo, eip string
		for _, ipset := range instance.IPSet {
			if ipset.IPId == "" {
				ip = ipset.IP
			} else {
				eip = ipset.IP
			}
		}
		for _, disk := range instance.DiskSet {
			diskinfo += fmt.Sprintf("%s:%dG;", disk.Drive, disk.Size)
		}
		data := MgoResource{
			HostID:     instance.UHostId,
			HostName:   instance.Name,
			CPU:        instance.CPU,
			Memory:     instance.Memory,
			Region:     region,
			IP:         ip,
			EIP:        eip,
			Disk:       diskinfo,
			Location:   "ucloud",
			Modifytime: time.Now(),
		}
		if cdenv != "" {
			data.Environment = cdenv
		}
		var cinfo *mgo.ChangeInfo
		if cinfo, err = c.Upsert(common.V{"hostid": instance.UHostId}, data); err == nil {
			if cinfo.UpsertedId != nil {
				insCount++
			} else {
				upCount++
			}
		} else {
			econf.LogInfo("[CMDB] host %s import from ucloud failed: %v\n", instance.UHostId, err)
		}
	}
	return upCount, insCount, nil
}
