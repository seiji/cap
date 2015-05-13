package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	PATH_CONFIG = "config/deploy.toml"
)

var (
	maxHostLength int
)

type ConfigApp struct {
	Name string
	GitRepoURL string `toml:"git_repo_url"`
	GitBranch string `toml:"git_branch"`
	GItSubmodules bool `toml:"git_submodules"`
	WritableDirs []string `toml:"writable_dirs"`
	SharedDirs []string `toml:"shared_dirs"`
	SharedFiles []string `toml:"shared_files"`
	StageDir string `toml:"stage_dir"`
	StageDefault string `toml:"stage_default"`
}

type ConfigServer struct {
	SSHUser string `toml:"ssh_user"`
	SSHPort int `toml:"ssh_port"`
	SSHPass string `toml:"ssh_pass"`
	SSHForwardAgent bool `toml:"ssh_forward_agent"`
	WebUser string `toml:"web_user"`
	PHPBin string `toml:"php_bin"`
}

type ConfigDeploy struct {
	DeployVia string `toml:"deploy_via"`
	DeployTo string `toml:"deploy_to"`
	DeploySubdir string `toml:"deploy_subdir"`
	DeployKeepReleases int `toml:"deploy_keep_releases"`
	DeployCachedCopy string `toml:"deploy_cached_copy"`
	DeoloyExclude []string `toml:"deploy_exclude"`
}

type server struct {
	Host string
	Port int
	User string
	Pass string
	Roles []string
	WebUser string `toml:"web_user"`
	PHPBin string `toml:"php_bin"`
}

type Config struct {
	LogLevel int `toml:"log_level"`
	Env []string
	*ConfigApp
	*ConfigServer
	*ConfigDeploy
	Servers []server
}

func newConfig(stage string) *Config {
	var baseConf Config
	bytes, err := ioutil.ReadFile(PATH_CONFIG)
	if err != nil {
		LogError(localhost,err)
		return nil
	}

	if _, err := toml.Decode(string(bytes), &baseConf); err != nil {
		LogError(localhost,err)
		return nil
	}

	stagePath := filepath.Join(baseConf.StageDir, stage + ".toml")

	var conf Config
	bytesStage, err := ioutil.ReadFile(stagePath)
	if err != nil {
		LogError(localhost,err)
		return nil
	}

	mergedBytes := append(bytes, bytesStage...)
	if _, err := toml.Decode(string(mergedBytes), &conf); err != nil {
		LogError(localhost,err)
		return nil
	}


	servers := make([]server, 0)
	for _, v := range conf.Servers {
		maxHostLength = MaxInt(maxHostLength, len(v.Host))
		if v.User == "" {
			v.User = conf.SSHUser
		}
		if v.Port == 0 {
			v.Port = conf.SSHPort
		}
		if v.WebUser == "" {
			v.WebUser = conf.WebUser
		}
		if v.PHPBin == "" {
			v.PHPBin = conf.PHPBin
		}
		servers = append(servers, v)
	}
	conf.Servers = servers

	return &conf
}

