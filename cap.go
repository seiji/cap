package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/codegangsta/cli"
)

var (
	config *Config
	localhost, _ = os.Hostname()

	logLevel int

	cachedCopy,
	deployTo,
	releasesPath,
	sharedPath,
	webUser string

	sharedDirs,
	sharedFiles,
	writableDirs []string

	env map[string]string
)

func setup(stage string) {
	config = newConfig(stage)

	fmt.Printf("config.Servers = %+v\n", config.Servers)
	//logLevel = LOG_TRACE
	logLevel = config.LogLevel

	cachedCopy = config.DeployCachedCopy
	deployTo = config.DeployTo
	releasesPath = filepath.Join(deployTo, "releases")
	sharedPath = filepath.Join(deployTo, "shared")
	webUser = config.WebUser

	sharedDirs = config.SharedDirs
	sharedFiles = config.SharedFiles
	writableDirs = config.WritableDirs
}

func capInit(c *cli.Context) {
	setup(c.GlobalString("stage"))
	Log("not implemented")
}

func capDeploy(c *cli.Context) {
	setup(c.GlobalString("stage"))
	if len(config.Servers) == 0 {
		return
	}

	ts := time.Now().Local().Format("20060102150405")
	var rev string
	var err error

	rev, err = updateSrc(branch)
	src, err := syncSrc()
	defer os.RemoveAll(src)
	if err != nil {
		LogError(localhost, err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(config.Servers))

	for _, v := range config.Servers {
		LogInfo(v.Host, "##### Start deploy ##### \n")
		go func(s server) {
			defer wg.Done()
			err := syncDest(s, src)
			if err != nil {
				return
			}

			conn := newConnection(s)
			release := finalizeUpdate(conn, ts, rev)
			Symfony(conn, release)
			createSymlink(conn, release)

			if conn.Err != nil {
				conn.Err = nil
				LogInfo(s.Host, "Roleback\n")
				conn.Exec("rm", "-Rf", release)
			}
		}(v)
	}
	wg.Wait()
}

func capSetup(c *cli.Context) {
	setup(c.GlobalString("stage"))
	if len(config.Servers) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(config.Servers))

	for _, v := range config.Servers {
		LogInfo(v.Host, "-----> Start deploy setup \n")
		go func(s server) {
			defer wg.Done()
			conn := newConnection(s)

			conn.Exec("mkdir", "-p", deployTo, releasesPath, sharedPath)
			conn.Exec("chmod", "g+w", deployTo, releasesPath, sharedPath)
			lenText := LogInfo(s.Host, "Creates dirs of deploy, releases, shared")
			if conn.Err != nil {
				LogNG(lenText)
				LogError(s.Host, conn.Err)
				return
			}
			LogOK(lenText)
			SymfonySetup(conn)
		}(v)
	}
	wg.Wait()
}


func updateSrc(branch string) (string, error) {
	if IsExist(cachedCopy) {
		return GitSync(branch)
	} else {
		return GitCheckout(branch)
	}
}

func syncSrc() (dst string, err error) {
	cli := &Cmd{}
	defer func() {
		lenText := LogInfo(localhost, "Update local src")
		if cli.err != nil {
			LogNG(lenText)
			LogError(localhost, cli.err)
			dst = ""
			err = cli.err
			return
		}
		LogOK(lenText)
	}()

	src := filepath.Join(cachedCopy, config.DeploySubdir)
	dst, cli.err = TempDir()

	args := []string{"-lrpta"}
	copyExclude := config.DeoloyExclude

	if len(copyExclude) > 0 {
		for _, v := range copyExclude {

			args = append(args, "--exclude", v)
		}
		args = append(args, src + "/", dst)
		cli.Exec("rsync", args...)
	} else {
		cli.Exec("cp", "-RPp", src, dst)
	}

	return dst, nil
}

func syncDest(s server, src string) (err error) {
	cli := &Cmd{}
	defer func() {
		lenText := LogInfo(s.Host, fmt.Sprintf("Rsync from [%s] to [%s]", localhost, s.Host))
		if cli.err != nil {
			LogNG(lenText)
			LogError(s.Host, cli.err)
			err = cli.err
			return
		}
		LogOK(lenText)
	}()

	remoteSrc := fmt.Sprintf("%s@%s:%s", s.User, s.Host, filepath.Join(sharedPath, "cached-copy"))
	cli.Exec("rsync", "-lrptauz", "--delete", "-e", "ssh", src + "/", remoteSrc)

	return nil
}

func pathCached() string {
	return filepath.Join(sharedPath,"cached-copy")
}

func pathRelease(ts string) string {
	return filepath.Join(releasesPath, ts)
}

func pathCurrent() string {
	return filepath.Join(deployTo, "current")
}

func finalizeUpdate(conn *SshConnection, ts, rev string) string {
	if conn.Err != nil {
		return ""
	}

	src := pathCached()
	release := pathRelease(ts)

	conn.Exec("cp", "-RPp", src, release)
	conn.Exec("echo", rev, ">", filepath.Join(release, "REVISION"))
	conn.Exec("chmod", "-R", "g+w", release)

	lenText := LogInfo(conn.Host.Host, "Create latest release")
	if conn.Err != nil {
		LogNG(lenText)
		LogError(conn.Host.Host, conn.Err)
		return ""
	}
	LogOK(lenText)

	for _, v := range sharedDirs {
		conn.Exec("mkdir", "-p", filepath.Join(sharedPath, v))
		conn.Exec(fmt.Sprintf(
			"sh -c 'if [ -d %s/%s ] ; then rm -rf %s/%s; fi'",
			release,
			v,
			release,
			v,
		))
		conn.Exec("ln", "-nfs", filepath.Join(sharedPath, v), filepath.Join(release,v))
	}

	for _, v := range sharedFiles {
		conn.Exec("mkdir", "-p", filepath.Join(sharedPath, path.Dir(v)))
		conn.Exec("touch", filepath.Join(sharedPath, v))
		conn.Exec("ln", "-nfs", filepath.Join(sharedPath, v), filepath.Join(release,v))
	}

	lenText = LogInfo(conn.Host.Host, "Symlinks static directories and static files")
	if conn.Err != nil {
		LogNG(lenText)
		LogError(conn.Host.Host, conn.Err)
		return ""
	}
	LogOK(lenText)

	return release
}

func createSymlink(conn *SshConnection, dst string) {
	if conn.Err != nil {
		return
	}

	defer func() {
		lenText := LogInfo(conn.Host.Host, "Updates latest release")
		if conn.Err != nil {
			LogNG(lenText)
			LogError(conn.Host.Host, conn.Err)
			return
		}
		LogOK(lenText)
	}()

	cur := pathCurrent()

	conn.Exec("rm", "-f", cur, "&&", "ln", "-vs", dst, cur)
	conn.Exec(
		"ls", "-1dt", filepath.Join(releasesPath, "*"),
		" | ", "tail", "-n", fmt.Sprintf("+%d", config.DeployKeepReleases + 1),
		" | ", "xargs", "rm", "-rf",
	)
}
