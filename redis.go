package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"strings"
)

type RedisCli struct {
	Docker
	id string
}

func NewRedisCli(cli *client.Client, containerName string) (*RedisCli, error) {
	docker := Docker{cli: cli}
	containerId, ok := docker.FindContainerId(containerName)
	if !ok {
		return nil, fmt.Errorf("unable to find container name: %s", containerName)
	}

	r := &RedisCli{
		Docker: docker,
		id:     containerId,
	}
	return r, nil
}

func (r *RedisCli) Get(key string) (string, error) {
	res, err := r.exec("get", key)
	if err != nil {
		return "", err
	}
	return res.StdOut, nil
}

func (r *RedisCli) Keys(pattern string) ([]string, error) {
	res, err := r.exec("keys", pattern)
	if err != nil {
		return nil, err
	}
	return strings.Split(res.StdOut, "\n"), nil
}

func (r *RedisCli) exec(commands ...string) (*ExecResult, error) {
	res, err := r.Exec(r.id, append([]string{"redis-cli", "--raw"}, commands...)...)
	if err == nil {
		res.StdOut = strings.TrimSuffix(res.StdOut, "\n")
	}
	return res, err
}
