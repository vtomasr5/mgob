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

func getURIHost(plan config.Plan, clusterType string) string {
	var host string
	if plan.Target.Type == clusterType {
		host = strings.Join(plan.Target.Backup.Host.Mongos, ",")
	}
	if plan.Target.Type == clusterType {
		host = strings.Join(plan.Target.Backup.Host.Mongod, ",")
	}
	if plan.Target.Type == clusterType {
		host = plan.Target.Backup.Host.Mongod[0]
	}
	return host
}

func _dump(plan config.Plan, archive, host string) ([]byte, error) {
	dump := fmt.Sprintf("mongodump --archive=%v --gzip --host %v ", archive, host)
	if plan.Target.Backup.Database != "" {
		dump += fmt.Sprintf("--db %v ", plan.Target.Backup.Database)
	}
	if plan.Target.Backup.Username != "" && plan.Target.Backup.Password != "" {
		dump += fmt.Sprintf("-u %v -p %v", plan.Target.Backup.Username, plan.Target.Backup.Password)
	}
	fmt.Println("COMMAND: ", dump)
	output, err := sh.Command("/bin/sh", "-c", dump).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return nil, errors.Wrapf(err, "mongodump log %v", ex)
	}
	return output, nil
}

func dump(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
	archive := fmt.Sprintf("%v/%v-%v.gz", tmpPath, plan.Name, ts.Format("2006-01-02T15:04:05"))
	log := fmt.Sprintf("%v/%v-%v.log", tmpPath, plan.Name, ts.Format("2006-01-02T15:04:05"))

	if plan.Target.Type == "sharding" {
		mc := NewMongoClient(plan.Target.Backup.Host.Mongos[0])
		// stop balancer
		err := mc.BalancerStop()
		if err != nil {
			return "", "", errors.Wrapf(err, "failed stoping the mongos balancer")
		}

		// backup mongoc
		for _, host := range plan.Target.Backup.Host.Mongoc {
			output, err := _dump(plan, archive, host)
			if err != nil {
				return "", "", errors.Wrapf(err, "mongodump failed")
			}
			logToFile(log, output)
		}

		// backup shards (mongod)
		for _, host := range plan.Target.Backup.Host.Mongod {
			output, err := _dump(plan, archive, host)
			if err != nil {
				return "", "", errors.Wrapf(err, "mongodump failed")
			}
			logToFile(log, output)
		}

		// start balancer
		err = mc.BalancerStart()
		if err != nil {
			return "", "", errors.Wrapf(err, "failed starting the mongos balancer")
		}

		// check if started
		err = mc.BalancerStatus()
		if err != nil {
			return "", "", errors.Wrapf(err, "failed checking the mongos balancer")
		}
	} else if plan.Target.Type == "replicaset" {
		for _, host := range plan.Target.Backup.Host.Mongod {
			output, err := _dump(plan, archive, host)
			if err != nil {
				return "", "", errors.Wrapf(err, "mongodump failed")
			}
			logToFile(log, output)
		}
	} else if plan.Target.Type == "standalone" {
		host := plan.Target.Type
		output, err := _dump(plan, archive, host)
		if err != nil {
			return "", "", errors.Wrapf(err, "mongodump failed")
		}
		logToFile(log, output)
	} else {
		return "", "", errors.New("target type not compatible")
	}

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
