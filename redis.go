package main

import (
	"fmt"
	"strings"
)

type RedisCli struct {
	docker *Docker
	id     string
}

func NewRedisCli(docker *Docker, containerName string) (*RedisCli, error) {
	containerId, ok := docker.FindContainerId(containerName)
	if !ok {
		return nil, fmt.Errorf("unable to find container name: %s", containerName)
	}

	r := &RedisCli{
		docker: docker,
		id:     containerId,
	}
	return r, nil
}

func (r *RedisCli) Get(key string) (string, error) {
	res, err := r.exec("get", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(res.StdOut, "\n"), nil
}

func (r *RedisCli) Keys(pattern string) ([]string, error) {
	res, err := r.exec("keys", pattern)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSuffix(res.StdOut, "\n"), "\n"), nil
}

func (r *RedisCli) exec(commands ...string) (*ExecResult, error) {
	return r.docker.Exec(r.id, append([]string{"redis-cli", "--raw"}, commands...)...)
}
