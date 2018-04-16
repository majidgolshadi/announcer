package client_announcer

import (
	"encoding/json"
	"errors"
	"fmt"
)

type chatServerClusterRepository struct {
	clusters       map[string]*Cluster
	zk             *zookeeper
	zkNode         string
	defaultCluster string
}

func ChatServerClusterRepositoryFactory(defaultCluster string) (repo *chatServerClusterRepository, err error) {
	repo = &chatServerClusterRepository{
		clusters:       make(map[string]*Cluster),
		defaultCluster: defaultCluster,
	}

	return
}

func (r *chatServerClusterRepository) Save(name string, cluster *Cluster) error {
	if r.clusters[name] != nil {
		return errors.New(fmt.Sprintf("cluster %s exists", name))
	}

	r.clusters[name] = cluster

	if r.zk != nil {
		path := fmt.Sprintf("%s/%s", r.zkNode, name)
		return r.zk.save(path, cluster.toJson())
	}

	return nil
}

func (r *chatServerClusterRepository) Get(name string) (*Cluster, error) {
	if name == "" {
		name = r.defaultCluster
	}

	if r.clusters[name] == nil {
		return nil, errors.New(fmt.Sprintf("cluster %s does not exist", name))
	}

	return r.clusters[name], nil
}

func (r *chatServerClusterRepository) Delete(name string) {
	r.clusters[name].Close()
	r.clusters[name] = nil
	delete(r.clusters, name)
}

func (r *chatServerClusterRepository) ListenToZookeeper(zookeeperAddress string, zkNode string) (err error) {
	if r.zk, err = zookeeperFactory(zookeeperAddress); err != nil {
		return err
	}

	go func() {
		chEvent, err := r.zk.childrenW(zkNode)
		if err != nil {
			println("children watcher error: ", err.Error())
			return
		}

		for zkevent := range chEvent {
			if zkevent.Event == DELETE_EVENT {
				r.Delete(zkevent.ZNode)
			} else {
				r.append(zkevent.ZNode)
			}
		}
	}()

	return
}

func (r *chatServerClusterRepository) append(zNode string) error {
	path := fmt.Sprintf("%s/%s", r.zkNode, zNode)
	clusterJson, err := r.zk.get(path)
	if err != nil {
		return err
	}

	cluster, err := r.restoreCluster(clusterJson)
	if err != nil {
		return err
	}

	r.clusters[zNode] = cluster
	return nil
}

func (r *chatServerClusterRepository) restoreCluster(clusterData []byte) (*Cluster, error) {
	clusterTmp := &Cluster{}
	if err := json.Unmarshal(clusterData, clusterTmp); err != nil {
		return nil, err
	}

	if clusterTmp.Client != nil {
		return ClusterClientFactory(
			clusterTmp.Client.Username, clusterTmp.Client.Password, clusterTmp.Client.Domain,
			clusterTmp.Client.Duration, clusterTmp.Addresses)
	}

	return ClusterComponentFactory(clusterTmp.Component.Name, clusterTmp.Component.Secret, clusterTmp.Addresses)
}

func (r *chatServerClusterRepository) Close() {
	if r.zk != nil {
		r.zk.close()
	}

	for _, cluster := range r.clusters {
		cluster.Close()
	}
}
