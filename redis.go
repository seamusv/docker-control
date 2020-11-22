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
	res, err := r.docker.Exec(r.id, "redis-cli", "--raw", "get", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(res.StdOut, "\n"), nil
}

func (r *RedisCli) Keys(pattern string) ([]string, error) {
	res, err := r.docker.Exec(r.id, "redis-cli", "--raw", "keys", pattern)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSuffix(res.StdOut, "\n"), "\n"), nil
}
