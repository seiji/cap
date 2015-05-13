package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"strconv"

	"golang.org/x/crypto/ssh"
)

var (
	// PASS = flag.String("pass", os.Getenv("SOCKSIE_SSH_PASSWORD"), "ssh password")
	sshConfig SshConfigFile
)

type SshConfigHost struct {
	Host string
	HostName string
	User string
	Port int
	IdentityFile string
}

func (c *SshConfigHost) Merge(aConfig SshConfigHost) {
	if aConfig.HostName != "" {
		c.HostName = aConfig.HostName
	}
	if aConfig.User != "" {
		c.User = aConfig.User
	}
	if aConfig.Port != 0 {
		c.Port = aConfig.Port
	}
	if aConfig.IdentityFile != "" {
		c.IdentityFile = aConfig.IdentityFile
	}
}

func (c *SshConfigHost) GetSigner() (ssh.Signer, error) {
	bytes, err := loadIdentity(c.IdentityFile)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(bytes)
}

type SshConfigFile []SshConfigHost

func (f SshConfigFile) Len() int {
	return len(f)
}

func (f SshConfigFile) Less(i, j int) bool {
	return len(f[i].Host) < len(f[j].Host)
}

func (f SshConfigFile) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type SshConnection struct {
	Client *ssh.Client
	Host *SshConfigHost
	Err error
}

func (conn *SshConnection) Exec(name string, args ...string) string {
	if conn.Err != nil {
		return ""
	}
	var sess *ssh.Session

	sess, conn.Err = conn.Client.NewSession()
	if conn.Err != nil {
		return ""
	}
	defer sess.Close()

	var stdout, stderr bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stderr

	cmd := strings.Join(append([]string{name}, args...), " ")
	LogCmd(conn.Host.Host, cmd)
	conn.Err = sess.Run(cmd)
	if conn.Err != nil {
		conn.Err = errors.New(Chomp(stderr.String()))
		return ""
	}

	ret := Chomp(stdout.String())
	if len(ret) > 0 {
		LogOut(conn.Host.Host, ret)
	}
	return ret
}

func (conn *SshConnection) SshIsExist(path string) bool {
	if conn.Err != nil {
		return false
	}
	out := conn.Exec(fmt.Sprintf("if [ -e '%s' ]; then echo -n 'true'; fi", path) )
	return out == "true"
}


func init() {
	path := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	parseConfigFile(path)
}

func parseConfigFile(path string) {
	bytes, _ := ioutil.ReadFile(path)
	sshConfig = make(SshConfigFile, 0)
	for _, section := range strings.Split(string(bytes), "Host ") {
		section = strings.TrimSpace(section);
		if section == "" || strings.HasPrefix(section, "#") == true {
			continue
		}
		configHost := SshConfigHost{}
		for n, line := range strings.Split(section, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") == true {
				continue
			}

			line = strings.ToLower(line)
			if n == 0 {
				configHost.Host = line
			} else if strings.HasPrefix(line, "hostname") {
				configHost.HostName = trimPrefixSpace(line, "hostname")
			} else if strings.HasPrefix(line, "user") {
				configHost.User = trimPrefixSpace(line, "user")
			} else if strings.HasPrefix(line, "port") {
				configHost.Port, _ = strconv.Atoi(trimPrefixSpace(line, "port"))
			} else if strings.HasPrefix(line, "identityfile") {
				configHost.IdentityFile = trimPrefixSpace(line, "identityfile")
			}
		}
		sshConfig = append(sshConfig, configHost)
	}
	sort.Sort(sshConfig)
}

func trimPrefixSpace(s, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(s, prefix))
}

func getSshConfigHost(host string) *SshConfigHost {
	configHost := SshConfigHost{Host: host}
	for _, c := range sshConfig {
		if strings.ContainsAny(c.Host, "*?") {
			m, _ := filepath.Match(c.Host, host)
			if m {
				configHost.Merge(c)
			}
		} else {
			if c.Host == host {
				configHost.Merge(c)
			}
		}
	}
	return &configHost
}

func loadIdentity(path string) ([]byte, error) {
	if strings.HasPrefix(path, "~") {
		path = strings.Replace(path, "~", os.Getenv("HOME"), 1)
	}
	return ioutil.ReadFile(path)
}

func newConnection(s server) (*SshConnection) {
	h := getSshConfigHost(s.Host)
	if s.Port != 0 {
		h.Port = s.Port
	}
	if s.User != "" {
		h.User = s.User
	}

	signer, err := h.GetSigner()
	if err != nil {
		return nil
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	cli, err := ssh.Dial("tcp", net.JoinHostPort(h.HostName, strconv.Itoa(h.Port)), config)
	return &SshConnection{Client: cli, Host: h, Err: err}
}

