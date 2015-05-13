package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	head,
	origin string
	enableSubmodule bool
)

func init() {
	enableSubmodule = false
}

func GitRevision(branch string) (string, error) {
	prev, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	defer os.Chdir(prev)
	os.Chdir(cachedCopy)

	out, err := CmdOutput(
		"git",
		"rev-list",
		"--max-count=1",
		// "--abbrev-commit",
		strings.Join([]string{"origin", branch}, "/"),
	)
	if err != nil {
		return "", err
	}

	return Chomp(out), nil
}

func archive(branch string) error {
	tempDir, _:= TempDir()
	defer os.RemoveAll(tempDir)

	cmd1 := newCmd(
		"git",
		"archive",
		branch,
	)
	cmd2 := newCmd(
		"tar",
		"-x",
		"--strip-components",
		"0",
		"-f",
		"-",
		"-C",
		tempDir,
	)
	err := CmdPipe(cmd1, cmd2)
	if err != nil {
		fmt.Printf("err = %+v\n", err)
		return err
	}

	return nil
}

func GitCheckout(branch string) (string, error) {
	lenText := LogInfo(localhost, "Git checkout repository ")

	var rev, prev string
	cli := &Cmd{}

	if !IsExist(cachedCopy) {
		cli.Exec("git","clone","-b", config.GitBranch, config.GitRepoURL, config.DeployCachedCopy)
	}

	prev, cli.err = filepath.Abs(".")

	defer os.Chdir(prev)
	os.Chdir(cachedCopy)

	rev, cli.err = GitRevision(branch)

	if cli.Output("git", "branch", "--list", "deploy")  == "" {
		cli.Exec("git", "checkout", "-b", "deploy", rev)
	}

	if enableSubmodule {
		cli.Exec("git","submodule","init",)
		cli.Exec("git","submodule","sync",)
		cli.Exec("git","submodule","update","--init",)
	}

	if cli.err != nil {
		LogNG(lenText)
		LogError(localhost, cli.err)
		return "", cli.err
	}
	LogOK(lenText)

	return rev, nil
}

func GitSync(branch string) (string, error) {
	lenText := LogInfo(localhost, "Git fetch repository ")

	var rev, prev string
	cli := &Cmd{}
	prev, cli.err = filepath.Abs(".")

	defer os.Chdir(prev)
	os.Chdir(cachedCopy)

	cli.Exec("git","fetch","origin",)
	cli.Exec("git","fetch","--tags", "origin",)

	rev, cli.err = GitRevision(branch)
	cli.Exec("git","reset","--hard",rev)

	if enableSubmodule {
		cli.Exec("git","submodule","init",)
		cli.Exec("git","submodule","sync",)
		cli.Exec("git","submodule","update","--init",)
	}

	cli.Exec("git","clean","-d","-x","-f",)
	if cli.err != nil {
		LogNG(lenText)
		LogError(localhost, cli.err)
		return "", cli.err
	}
	LogOK(lenText)

	return rev, nil
}
