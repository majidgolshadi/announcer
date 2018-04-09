package client_announcer

import (
	"errors"
	"fmt"
)

type repository struct {
	clusters map[string]*Cluster
}

func RepositoryFactory(zookeeperAddress string, namespace string) *repository {
	r := &repository{
		clusters:   make(map[string]*Cluster),
	}

	r.watchOnZookeeper(zookeeperAddress, namespace)
	return r
}

func (r *repository) setCluster(name string, cluster *Cluster) error {
	if r.clusters[name] != nil {
		return errors.New(fmt.Sprintf("cluster %s exists", name))
	}

	r.clusters[name] = cluster
	return nil
}

func (r *repository) GetCluster(name string) (*Cluster, error) {
	if r.clusters[name] == nil {
		return nil, errors.New(fmt.Sprintf("cluster %s does not exist", name))
	}

	return r.clusters[name], nil
}

func (r *repository) watchOnZookeeper(zookeeperAddr string, namespace string) {

}