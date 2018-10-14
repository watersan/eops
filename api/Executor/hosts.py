#!/usr/bin/env python
# coding=utf8
import os.path,os
import json
import requests
import optparse
import ConfigParser
import time


class Opsv2Inventory(object):
    def __init__(self):
        self.cache_file = "/tmp/opsv2_hosts.json"
        self.opsv2url = "http://127.0.0.1:5000/eops"
        self.keyid = "gRfTjFOgxPeNwOtx"
        self.secret = "oI66yeyHfZITi13M5Fuyvd5eFNfYkgTDaTHIyZferhtiULtfkpKEwRZoLDXF2FwI"
        self.cache_max_age = 60
        self.lock = "/tmp/opsv2_hosts.lock"
        self.inventory = dict(
            hosts = [],
            _meta = dict(hostvars = {}),
        )
        self.cliArgs()
        self.readConf()
        if self.is_cache_valid() == False or self.args.refresh:
            # if os.path.isfile(self.lock):
            #     while os.path.isfile(self.lock) == False:
            #         time.sleep(1)
            #     self.load_from_cache()
            # else:
            self.updateCache()
        else:
            self.load_from_cache()
        print json.dumps(self.inventory, sort_keys=True, indent=2)

    def cliArgs(self):
        parser = optparse.OptionParser()
        parser.add_option("--refresh", action="store_true", dest="refresh", default=False, help='refresh cache')
        parser.add_option("--list", action="store_true", dest="list", default=True, help='list all hosts')
        self.args,tmp = parser.parse_args()
    def is_cache_valid(self):
        if os.path.isfile(self.cache_file):
            mod_time = os.path.getmtime(self.cache_file)
            current_time = time.time()
            if (mod_time + self.cache_max_age) > current_time:
                return True
        return False
    def load_from_cache(self):
        cache = open(self.cache_file, 'r')
        json_inventory = cache.read()
        self.inventory = json.loads(json_inventory)

    def readConf(self):
        config = ConfigParser.SafeConfigParser()
        if len(config.read(os.path.dirname(os.path.realpath(__file__)) + '/opsv2.ini')) == 0:
            return
        for item in ["cache_file","opsv2url","keyid","secret","cache_max_age"]:
            if config.has_option("opsv2",item):
                setattr(self,item,config.get('opsv2', item))
    def updateCache(self):
        lockf = file(self.lock,'w')
        lockf.close()
        payload = dict(keyid = self.keyid,secret=self.secret)
        r = requests.get('%s/cmdb/svcprovider' % self.opsv2url, params=payload)
        svcinfo = json.loads(r.content)
        proxy = {}
        for item in svcinfo["dataList"]:
            if "proxy" in item:
                proxy[item["name"]] = item["proxy"]
        r = requests.get('%s/cmdb/resourceall' % self.opsv2url, params=payload)
        hostinfo = json.loads(r.content)
        r = requests.get('%s/cmdb/project' % self.opsv2url, params=payload)
        projectinfo = json.loads(r.content)
        r = requests.get('%s/cmdb/cluster' % self.opsv2url, params=payload)
        clusterinfo = json.loads(r.content)
        r = requests.get('%s/cmdb/apps' % self.opsv2url, params=payload)
        appsinfo = json.loads(r.content)
        cdenv = {}
        for host in hostinfo["dataList"]:
            if host["hostname"] == "":
                continue
            hname = host["hostname"]
            port = "22"
            ip = host["ip"]
            if "eip" in host and host["eip"] != "":
                ip = host["eip"]
            try:
                i = ip.index(':')
                port = ip[i+1:]
                ip1 = ip[:i]
                ip = ip1
            except:
                pass
            self.inventory["_meta"]["hostvars"][hname] = dict(
                ZDOPS_HostName = host["hostname"],
                ansible_host = ip,
                ansible_port = port,
                ZDOPS_CDEnv = host["environment"],
                ZDOPS_Location = host["location"],
                ZDOPS_CPU = host["cpu"],
                ZDOPS_Memory = host["memory"],
            )
            for diskinfo in host["disk"].split(';'):
                diskitems = diskinfo.split(':')
                if len(diskitems) == 2:
                    dname,dsize = diskitems[0],diskitems[1]
                    if dsize[-1] == 'G':
                        dsize = dsize[:-1]
                    if dname[1:] == "":
                        dname = "root"
                    else:
                        dname = dname[1:]
                    self.inventory["_meta"]["hostvars"][hname]["ZDOPS_Disk_"+dname] = dsize
            cdenv[host["environment"]] = True
            if host["location"] in proxy and len(proxy[host["location"]]) > 1:
                proxyinfo = proxy[host["location"]].split(":")
                port = "22"
                if len(proxyinfo) == 2:
                    port = proxyinfo[1]
                self.inventory["_meta"]["hostvars"][hname]["ansible_ssh_common_args"] = "-o 'ProxyCommand ssh -W %%h:%%p %s -p %s'" % (proxyinfo[0],port)
            self.inventory["hosts"].append(host["hostname"])

        buf = dict(project = {},cluster={})
        for p in projectinfo["dataList"]:
            buf["project"][p["_id"]] = p["name"]
            for env,v in cdenv.iteritems():
                self.inventory[env+"-"+p["_id"]] = dict(
                    children = [],
                    vars = dict(
                        ZDOPS_GroupType = "project",
                        ZDOPS_GroupName=p["name"],
                    ),
                )
        for c in clusterinfo["dataList"]:
            buf["cluster"][c["_id"]] = dict(name = c["name"],pid=c["project"])
            for env,v in cdenv.iteritems():
                self.inventory[env+"-"+c["project"]]["children"].append(c["_id"])
                self.inventory[env+"-"+c["_id"]] = dict(
                    children = [],
                    vars = dict(
                        ZDOPS_GroupType = "cluster",
                        ZDOPS_GroupName=c["name"],
                        ZDOPS_Project=buf["project"][c["project"]],
                    )
                )
        depend = dict()
        for app in appsinfo["dataList"]:
            aid = app["_id"]
            pid = buf["cluster"][app["cluster"]]["pid"]
            for env,v in cdenv.iteritems():
                #hosts = app["environment"][env]["hosts"]
                try:
                    hosts = app["environment"][env]["hosts"]
                except:
                    continue
                self.inventory[env+"-"+aid] = dict(
                    hosts = app["environment"][env]["hosts"],
                    vars = dict(
                        ZDOPS_GroupType = "apps",
                        ZDOPS_GroupName=app["name"],
                        ZDOPS_Cluster = buf["cluster"][app["cluster"]]["name"],
                        ZDOPS_Project=buf["project"][pid],
                    ),
                )
                if len(app["depend"]) > 0:
                    for depapp in app["depend"]:
                        if depapp in self.inventory:
                            self.inventory[env+"-"+app["cluster"]]["children"].remove(depapp)
                            self.inventory[env+"-"+depapp]["hosts"] = app["hosts"]
                        else:
                            depend[env+"-"+depapp] = env+"-"+aid
                if aid in depend:
                    self.inventory[aid]["hosts"] = self.inventory[depend[aid]]["hosts"]
                else:
                    self.inventory[env+"-"+app["cluster"]]["children"].append(app["_id"])

                for h in app["environment"][env]["hosts"]:
                    try:
                        self.inventory["hosts"].remove(h)
                    except:
                        pass
        cache = file(self.cache_file,'w')
        cache.write(json.dumps(self.inventory,sort_keys=True, indent=2))
        cache.close()
        os.unlink(self.lock)
        #print json.dumps(inventory, sort_keys=True, indent=2)

if __name__ == "__main__":
    Opsv2Inventory()
