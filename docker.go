package main

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io/ioutil"
	"strings"
)

type Docker struct {
	cli *client.Client
}

func (r *Docker) Exec(containerId string, commands ...string) (*ExecResult, error) {
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          commands,
	}

	resp, err := r.cli.ContainerExecCreate(context.Background(), containerId, config)
	if err != nil {
		return nil, err
	}

	return r.inspectExecResp(context.Background(), resp.ID)
}

func (r *Docker) FindContainerId(containerName string) (string, bool) {
	if containers, err := r.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true}); err == nil {
		for _, container := range containers {
			if contains(container.Names, containerName) {
				return container.ID, true
			}
		}
	}

	return "", false
}

func (r *Docker) inspectExecResp(ctx context.Context, execId string) (*ExecResult, error) {
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

type ExecResult struct {
	ExitCode int
	StdOut   string
	StdErr   string
}

func contains(array []string, needle string) bool {
	for _, i := range array {
		if strings.EqualFold(i, needle) {
			return true
		}
	}
	return false
}
