package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/vtomasr5/mgob/config"
)

// Restore copies the archived backup to the destination directory
func Restore(plan config.Plan, tmpPath string, storagePath string) (Result, error) {
	t1 := time.Now()
	planDir := fmt.Sprintf("%v/%v", storagePath, plan.Name)

	var archive string
	err := filepath.Walk(plan.SFTP.BackupDir,
		func(path string, f os.FileInfo, err error) error {
			if strings.Contains(path, "gz") {
				archive, err = newestFile(path)
				if err != nil {
					logrus.Error(err)
				}
			}
			return nil
		},
	)
	res := Result{
		Plan:      plan.Name,
		Timestamp: t1.UTC(),
		Status:    500,
	}
	_, res.Name = filepath.Split(archive)

	if err != nil {
		return res, err
	}

	if plan.Scheduler.Retention > 0 {
		err = applyRetention(planDir, plan.Scheduler.Retention)
		if err != nil {
			return res, errors.Wrap(err, "retention job failed")
		}
	}

	file := filepath.Join(planDir, res.Name)

	if plan.SFTP != nil {
		sftpOutput, err := sftpDownload(file, plan)
		if err != nil {
			return res, err
		}
		logrus.WithField("plan", plan.Name).Info(sftpOutput)
	}

	// TODO: support restore from S3
	// if plan.S3 != nil {
	// 	s3Output, err := s3Download(file, plan)
	// 	if err != nil {
	// 		return res, err
	// 	}
	// 	logrus.WithField("plan", plan.Name).Infof("S3 upload finished %v", s3Output)
	// }

	t2 := time.Now()
	res.Status = 200
	res.Duration = t2.Sub(t1)
	return res, nil
}
