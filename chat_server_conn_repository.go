package client_announcer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

type chatServerConnRepository struct {
	clusters    map[string]*Cluster
	zkConn      *zk.Conn
	zkNamespace string
	defaultCluster string
}

func ChatServerConnRepositoryFactory(zookeeperAddress string, namespace string) (repo *chatServerConnRepository, err error) {
	repo = &chatServerConnRepository{
		clusters:    make(map[string]*Cluster),
		zkNamespace: namespace,
	}

	zkServers := strings.Split(zookeeperAddress, ",")
	if repo.zkConn, _, err = zk.Connect(zkServers, time.Minute); err != nil {
		return nil, err
	}

	err = repo.initRepository()
	err = repo.watchOnZookeeper()

	return
}

func (r *chatServerConnRepository) SetDefaultCluster(defaultCluster string) {
	r.defaultCluster = defaultCluster
}

func (r *chatServerConnRepository) SetCluster(name string, cluster *Cluster) error {
	if r.clusters[name] != nil {
		return errors.New(fmt.Sprintf("cluster %s exists", name))
	}

	r.clusters[name] = cluster
	r.storeCluster(name, cluster)
	return nil
}

func (r *chatServerConnRepository) GetCluster(name string) (*Cluster, error) {
	if name == "" {
		name = r.defaultCluster
	}

	if r.clusters[name] == nil {
		return nil, errors.New(fmt.Sprintf("cluster %s does not exist", name))
	}

	return r.clusters[name], nil
}

func (r *chatServerConnRepository) deleteCluster(name string) {
	r.clusters[name].Close()
	r.clusters[name] = nil
	delete(r.clusters, name)
}

// TODO: implement zookeeper watch
func (r *chatServerConnRepository) watchOnZookeeper() (err error) {
	//_, _, eventCh, err := r.zkConn.ExistsW(r.zkNamespace)
	//if err != nil {
	//	return err
	//}
	//
	//go func() {
	//	for ev := range eventCh{
	//		var (
	//			clusterList []string
	//			err error
	//		)
	//		clusterList,_, eventCh, err = r.zkConn.ChildrenW(r.zkNamespace)
	//		if err != nil {
	//			println(err.Error())
	//			continue
	//		}
	//	}
	//}()

	return nil
}

func (r *chatServerConnRepository) initRepository() error {
	data, _, err := r.zkConn.Children(r.zkNamespace)
	if err != nil {
		return err
	}

	for _, name := range data {
		path := fmt.Sprintf("%s/%s", r.zkNamespace, name)
		clusterData, _, _ := r.zkConn.Get(path)
		cluster, err := r.restoreCluster(clusterData)

		if err != nil {
			return err
		}

		r.clusters[name] = cluster
	}

	return nil
}

func (r *chatServerConnRepository) restoreCluster(clusterData []byte) (*Cluster, error) {
	clusterTmp := &Cluster{}
	if err := json.Unmarshal(clusterData, clusterTmp); err != nil {
		return nil, err
	}

	if clusterTmp.Client != nil {
		return ClusterClientFactory(clusterTmp.Client.Username, clusterTmp.Client.Password, clusterTmp.Client.Domain,
			clusterTmp.Client.Duration, clusterTmp.Addresses)
	}

	return ClusterComponentFactory(clusterTmp.Component.Name, clusterTmp.Component.Secret, clusterTmp.Addresses)
}

func (r *chatServerConnRepository) storeCluster(name string, cluster *Cluster) error {
	json, err := json.Marshal(cluster)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", r.zkNamespace, name)
	_, err = r.zkConn.Create(path, json, int32(0), zk.WorldACL(zk.PermAll))

	return err
}

func (r *chatServerConnRepository) Close() {
	r.zkConn.Close()

	for _, cluster := range r.clusters {
		cluster.Close()
	}
}
