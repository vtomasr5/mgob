package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/vtomasr5/mgob/config"
	"golang.org/x/crypto/ssh"
)

// SSHClient contains the client and the session for an SSH Connection
type SSHClient struct {
	client  *ssh.Client
	session *ssh.Session
}

// NewSSHClient returns a SSH Client struct that contains the Client and the Session
func NewSSHClient(plan config.Plan) (*SSHClient, error) {
	sshConf := &ssh.ClientConfig{
		User: plan.SFTP.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(plan.SFTP.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cli, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", plan.SFTP.Host, plan.SFTP.Port), sshConf)
	if err != nil {
		return nil, errors.Wrapf(err, "SSH dial to %v:%v failed", plan.SFTP.Host, plan.SFTP.Port)
	}
	sess, err := cli.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "SFTP client init %v:%v failed", plan.SFTP.Host, plan.SFTP.Port)
	}
	sshCli := &SSHClient{
		client:  cli,
		session: sess,
	}
	return sshCli, err
}

// Command executes a command through an SSH connection
func (s SSHClient) Command(cmd string) error {
	err := s.session.Run(cmd)
	if err != nil {
		return errors.Wrapf(err, "Failed to run ssh command %v", err)
	}
	return nil
}

func sftpUpload(file string, plan config.Plan) (string, error) {
	t1 := time.Now()
	sshCon, err := NewSSHClient(plan)
	if err != nil {
		return "", errors.Wrapf(err, "SSH dial to %v:%v failed", plan.SFTP.Host, plan.SFTP.Port)
	}
	defer sshCon.session.Close()

	sftpClient, err := sftp.NewClient(sshCon.client)
	if err != nil {
		return "", errors.Wrapf(err, "SFTP client init %v:%v failed", plan.SFTP.Host, plan.SFTP.Port)
	}
	defer sftpClient.Close()

	f, err := os.Open(file)
	if err != nil {
		return "", errors.Wrapf(err, "Opening file %v failed", file)
	}
	defer f.Close()

	_, fname := filepath.Split(file)
	dstPath := filepath.Join(plan.SFTP.Dir, fname)
	sf, err := sftpClient.Create(dstPath)
	if err != nil {
		return "", errors.Wrapf(err, "SFTP %v:%v creating file %v failed", plan.SFTP.Host, plan.SFTP.Port, dstPath)
	}

	_, err = io.Copy(sf, f)
	if err != nil {
		return "", errors.Wrapf(err, "SFTP %v:%v upload file %v failed", plan.SFTP.Host, plan.SFTP.Port, dstPath)
	}
	sf.Close()

	//listSftpBackups(sftpClient, plan.SFTP.Dir)

	t2 := time.Now()
	msg := fmt.Sprintf("SFTP upload finished `%v` -> `%v` Duration: %v",
		file, dstPath, t2.Sub(t1))
	return msg, nil
}

func listSftpBackups(client *sftp.Client, dir string) error {
	list, err := client.ReadDir(fmt.Sprintf("/%v", dir))
	if err != nil {
		return errors.Wrapf(err, "SFTP reading %v dir failed", dir)
	}

	for _, item := range list {
		fmt.Println(item.Name())
	}

	return nil
}
