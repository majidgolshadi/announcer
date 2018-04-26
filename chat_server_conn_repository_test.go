package client_announcer

import "testing"

func TestRepoClusterFactory(t *testing.T) {
	defaultCluster := ""
	repo, err := ChatServerClusterRepositoryFactory(defaultCluster)
	if err == nil {
		t.Fail()
	}

	defaultCluster = "defaultClusterName"
	repo, err = ChatServerClusterRepositoryFactory(defaultCluster)
	if err != nil {
		t.Fail()
	}

	if repo.defaultCluster != defaultCluster {
		t.Fail()
	}
}

func TestSaveNewCluster(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)
	cluster := &Cluster{}
	if err := repo.Save(defaultCluster, cluster); err != nil {
		t.Fail()
	}
}

func TestSaveClusterOnSameName(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)
	cluster := &Cluster{}
	if err := repo.Save(defaultCluster, cluster); err != nil {
		t.Fail()
	}

	if err := repo.Save(defaultCluster, cluster); err == nil {
		t.Fail()
	}
}

func TestGetWithEmptyName(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)
	if _, err := repo.Get(""); err == nil {
		t.Fail()
	}

}

func TestGetNotExistCluster(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)
	if _, err := repo.Get(defaultCluster); err == nil {
		t.Fail()
	}
}

func TestGetWithName(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)

	repo.Save(defaultCluster, &Cluster{})
	if _, err := repo.Get(defaultCluster); err != nil {
		t.Fail()
	}
}
