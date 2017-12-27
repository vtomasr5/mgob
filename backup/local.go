package backup

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	sh "github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	"github.com/vtomasr5/mgob/config"
)

func getURIHost(plan config.Plan) string {
	var host string
	if plan.Target.Type == "sharding" {
		host = strings.Join(plan.Target.Host.Mongos, ",")
	}
	if plan.Target.Type == "replicaset" {
		host = strings.Join(plan.Target.Host.Mongod, ",")
	}
	if plan.Target.Type == "standalone" {
		host = plan.Target.Host.Mongod[0]
	}
	return host
}

func dump(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
	archive := fmt.Sprintf("%v/%v-%v.gz", tmpPath, plan.Name, ts.Format("2006-01-02T15:04:05"))
	log := fmt.Sprintf("%v/%v-%v.log", tmpPath, plan.Name, ts.Format("2006-01-02T15:04:05"))

	host := getURIHost(plan)

	dump := fmt.Sprintf("mongodump --archive=%v --gzip --host %v ", archive, host)
	if plan.Target.Database != "" {
		dump += fmt.Sprintf("--db %v ", plan.Target.Database)
	}
	if plan.Target.Username != "" && plan.Target.Password != "" {
		dump += fmt.Sprintf("-u %v -p %v", plan.Target.Username, plan.Target.Password)
	}
	fmt.Println("COMMAND: ", dump)
	output, err := sh.Command("/bin/sh", "-c", dump).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", "", errors.Wrapf(err, "mongodump log %v", ex)
	}
	logToFile(log, output)

	return archive, log, nil
}

func logToFile(file string, data []byte) error {
	if len(data) > 0 {
		err := ioutil.WriteFile(file, data, 0644)
		if err != nil {
			return errors.Wrapf(err, "writing log %v failed", file)
		}
	}

	return nil
}

func applyRetention(path string, retention int) error {
	gz := fmt.Sprintf("cd %v && rm -f $(ls -1t *.gz | tail -n +%v)", path, retention+1)
	err := sh.Command("/bin/sh", "-c", gz).Run()
	if err != nil {
		return errors.Wrapf(err, "removing old gz files from %v failed", path)
	}

	log := fmt.Sprintf("cd %v && rm -f $(ls -1t *.log | tail -n +%v)", path, retention+1)
	err = sh.Command("/bin/sh", "-c", log).Run()
	if err != nil {
		return errors.Wrapf(err, "removing old log files from %v failed", path)
	}

	return nil
}

// TmpCleanup remove files older than one day
func TmpCleanup(path string) error {
	rm := fmt.Sprintf("find %v -mtime +%v -type f -delete", path, 1)
	err := sh.Command("/bin/sh", "-c", rm).Run()
	if err != nil {
		return errors.Wrapf(err, "%v cleanup failed", path)
	}

	return nil
}
