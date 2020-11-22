package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io/ioutil"
	"strings"
)

type RedisCli struct {
	cli *client.Client
	id  string
}

func NewRedisCli(cli *client.Client, containerName string) (*RedisCli, error) {
	containerId, ok := findContainerId(cli, containerName)
	if !ok {
		return nil, fmt.Errorf("unable to find container name: %s", containerName)
	}

	r := &RedisCli{
		cli: cli,
		id:  containerId,
	}
	return r, nil
}

func (r *RedisCli) Get(key string) (string, error) {
	res, err := r.exec("redis-cli", "--raw", "get", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(res.StdOut, "\n"), nil
}

func (r *RedisCli) Keys(pattern string) ([]string, error) {
	res, err := r.exec("redis-cli", "--raw", "keys", pattern)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSuffix(res.StdOut, "\n"), "\n"), nil
}

func (r *RedisCli) exec(commands ...string) (*ExecResult, error) {
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          commands,
	}

	resp, err := r.cli.ContainerExecCreate(context.Background(), r.id, config)
	if err != nil {
		return nil, err
	}

	return r.inspectExecResp(context.Background(), resp.ID)
}

func (r *RedisCli) inspectExecResp(ctx context.Context, execId string) (*ExecResult, error) {
	resp, err := r.cli.ContainerExecAttach(ctx, execId, types.ExecConfig{})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return nil, err
		}
		break

	case <-ctx.Done():
		return nil, ctx.Err()
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return nil, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return nil, err
	}

	res, err := r.cli.ContainerExecInspect(ctx, execId)
	if err != nil {
		return nil, err
	}

	execResult := &ExecResult{
		ExitCode: res.ExitCode,
		StdOut:   string(stdout),
		StdErr:   string(stderr),
	}
	return execResult, nil
}

func findContainerId(cli *client.Client, containerName string) (string, bool) {
	if containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true}); err == nil {
		for _, container := range containers {
			if contains(container.Names, containerName) {
				return container.ID, true
			}
		}
	}

	return "", false
}

func contains(array []string, needle string) bool {
	for _, i := range array {
		if strings.EqualFold(i, needle) {
			return true
		}
	}
	return false
}

type ExecResult struct {
	ExitCode int
	StdOut   string
	StdErr   string
}
