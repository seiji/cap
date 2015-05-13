package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Cmd struct {
	err error
}

func (c *Cmd) Err() error {
	return c.err
}

func (c *Cmd) SetErr(e error) {
	c.err = e
}

func (c *Cmd) Exec(name string, args ... string) {
	if c.err != nil {
		return
	}
	c.err = CmdExec(name, args...)
}

func (c *Cmd) Output(name string, args ... string) string {
	var out string
	if c.err != nil {
		return ""
	}
	out, c.err = CmdOutput(name, args...)

	return out
}

func CmdExec(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	LogCmd(localhost, strings.Join(cmd.Args, " "))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		err = errors.New(Chomp(stderr.String()))
		return err
	}

	out := Chomp(stdout.String())
	if len(out) > 0 {
		LogOut(localhost, out)
	}

	return err
}

func CmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	LogCmd(localhost, strings.Join(cmd.Args, " "))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		err = errors.New(Chomp(stderr.String()))
		return "", err
	}

	ret := Chomp(string(b[:]))
	if len(ret) > 0 {
		LogOut(localhost, ret)
	}
	return ret, err
}

func newCmd(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func CmdPipe(c1, c2 *exec.Cmd) error {
	c2.Stdin, _ = c1.StdoutPipe()

	var out bytes.Buffer
	c2.Stdout = &out

	fmt.Printf("$ %s | %s \n", strings.Join(c1.Args, " "), strings.Join(c2.Args, " "))
	err := c2.Start()
	if err != nil {
		return err
	}

	err = c1.Run()
	if err != nil {
		return err
	}

	err = c2.Wait()
	if err != nil {
		return err
	}

	if out.Len() > 0 {
		fmt.Printf("%s\n", out.String())
	}
	return nil
}
