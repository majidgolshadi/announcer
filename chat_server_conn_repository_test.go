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

func TestGetExistCluster(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)

	repo.Save(defaultCluster, &Cluster{})
	if _, err := repo.Get(defaultCluster); err != nil {
		t.Fail()
	}
}

func TestDeleteCluster(t *testing.T) {
	defaultCluster := "default"
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)

	repo.Save(defaultCluster, &Cluster{})
	repo.Delete(defaultCluster)
	if repo, err := repo.Get(defaultCluster); err == nil || repo != nil {
		t.Fail()
	}
}

func TestRestoreClientCluster(t *testing.T) {
	defaultCluster := "default"
	jsonCluster := []byte(`{"rateLimit":2, "sendRetry": 2, "client":{"username": "username", "password": "password", "pingInterval": 2, "domain": "soroush.ir", "resource": "announcer"}}`)
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)
	cluster, err := repo.restoreCluster(jsonCluster)
	if err != nil {
		t.Fail()
	}

	if cluster.RateLimit != 2 || cluster.SendRetry != 2 {
		t.Fail()
	}

	if cluster.Client.Username != "username" || cluster.Client.Password != "password" ||
		cluster.Client.PingInterval != 2 || cluster.Client.Domain != "soroush.ir" ||
		cluster.Client.Resource != "announcer" {
		t.Fail()
	}

	if cluster.Component.Name != "" || cluster.Component.Secret != "" ||
		cluster.Component.PingInterval != 0 || cluster.Component.Domain != "" {
		t.Fail()
	}
}

func TestRestoreComponentCluster(t *testing.T) {
	defaultCluster := "default"
	jsonCluster := []byte(`{"rateLimit":2, "sendRetry": 2, "component":{"name": "name", "secret": "secret", "pingInterval": 2, "domain": "soroush.ir", "resource": "announcer"}}`)
	repo, _ := ChatServerClusterRepositoryFactory(defaultCluster)
	cluster, err := repo.restoreCluster(jsonCluster)
	if err != nil {
		t.Fail()
	}

	if cluster.RateLimit != 2 || cluster.SendRetry != 2 {
		t.Fail()
	}

	if cluster.Component.Name != "name" || cluster.Component.Secret != "secret" ||
		cluster.Component.PingInterval != 2 || cluster.Component.Domain != "soroush.ir" {
		t.Fail()
	}

	if cluster.Client.Username != "" || cluster.Client.Password != "" ||
		cluster.Client.PingInterval != 0 || cluster.Client.Domain != "" ||
		cluster.Client.Resource != "" {
		t.Fail()
	}
}
