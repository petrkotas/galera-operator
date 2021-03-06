// Copyright 2017 The etcd-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"os"
	"testing"
	"time"

	api "github.com/beekhof/galera-operator/pkg/apis/galera/v1alpha1"
	"github.com/beekhof/galera-operator/pkg/util/k8sutil"
	"github.com/beekhof/galera-operator/test/e2e/e2eutil"
	"github.com/beekhof/galera-operator/test/e2e/framework"

	"github.com/sirupsen/logrus"
)

func TestCreateClusterDummy(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}

	f := framework.Global

	labels := map[string]string{
		"testlabel": "createOnly",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	logrus.Info("Creating cluster")
	origEtcd := e2eutil.NewCluster("beekhof-dummy-", 3, "", labels, annotations)
	logrus.Infof("Cluster pods: %v [%v]", origEtcd.Spec, origEtcd.Spec.DeepCopy())
	// origEtcd.Spec.Pod.AntiAffinity = true
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, origEtcd)
	logrus.Infof("Created pods: %v [%v]", testEtcd.Spec, testEtcd.Spec.Pod)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	logrus.Info(time.Now(), "Waiting")
	time.Sleep(120 * time.Second)
	logrus.Info(time.Now(), "Done Waiting")

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 60, testEtcd); err != nil {
		t.Fatalf("failed to create 3 members etcd cluster: %v", err)
	}

}

func TestCreateClusterGalera(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}

	f := framework.Global

	labels := map[string]string{
		"testlabel": "createOnly",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	origEtcd := e2eutil.NewCluster("galera-", 3, "quay.io/beekhof/galera:latest", labels, annotations)
	// origEtcd.Spec.Pod.AntiAffinity = true
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, origEtcd)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	t.Log(time.Now(), "Waiting")
	time.Sleep(120 * time.Second)
	t.Log(time.Now(), "Done Waiting")

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 60, testEtcd); err != nil {
		t.Fatalf("failed to create 3 members etcd cluster: %v", err)
	}

}

func TestConnectPod(t *testing.T) {
	f := framework.Global

	stdout, stderr, err := k8sutil.ExecCommandInPodWithFullOutput(f.Log, f.KubeClient, f.Namespace,
		"test-galera-qgd49-0001", "ls", "-al")
	// stdout, stderr, err := f.ExecCommandInPodWithFullOutput("test-galera-qgd49-0001", "ls")
	f.Log.Infof("out: %v, err: %v", stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateAClusterOnly(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}

	labels := map[string]string{
		"testlabel": "createOnly",
	}
	annotations := map[string]string{
		"testannotation": "testannotationvalue",
	}

	f := framework.Global
	origEtcd := e2eutil.NewCluster("test-galera-", 3, "", labels, annotations)
	// origEtcd.Spec.Pod.AntiAffinity = true
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, origEtcd)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 60, testEtcd); err != nil {
		t.Fatalf("failed to create 3 members etcd cluster: %v", err)
	}
	time.Sleep(60000)
}

// TestPauseControl tests the user can pause the operator from controlling
// an etcd cluster.
func TestPauseControl(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}

	f := framework.Global
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, e2eutil.NewCluster("test-etcd-", 3, "", nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	names, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 6, testEtcd)
	if err != nil {
		t.Fatalf("failed to create 3 members etcd cluster: %v", err)
	}

	updateFunc := func(cl *api.ReplicatedStatefulSet) {
		cl.Spec.Paused = true
	}
	if testEtcd, err = e2eutil.UpdateCluster(f.CRClient, testEtcd, 10, updateFunc); err != nil {
		t.Fatalf("failed to pause control: %v", err)
	}

	// TODO: this is used to wait for the CR to be updated.
	// TODO: make this wait for reliable
	time.Sleep(5 * time.Second)

	if err := e2eutil.KillMembers(f.KubeClient, f.Namespace, names[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := e2eutil.WaitUntilPodSizeReached(t, f.KubeClient, 2, 1, testEtcd); err != nil {
		t.Fatalf("failed to wait for killed member to die: %v", err)
	}
	if _, err := e2eutil.WaitUntilPodSizeReached(t, f.KubeClient, 3, 1, testEtcd); err == nil {
		t.Fatalf("cluster should not be recovered: control is paused")
	}

	updateFunc = func(cl *api.ReplicatedStatefulSet) {
		cl.Spec.Paused = false
	}
	if testEtcd, err = e2eutil.UpdateCluster(f.CRClient, testEtcd, 10, updateFunc); err != nil {
		t.Fatalf("failed to resume control: %v", err)
	}

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 6, testEtcd); err != nil {
		t.Fatalf("failed to resize to 3 members etcd cluster: %v", err)
	}
}

func TestEtcdUpgrade(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}
	f := framework.Global
	origEtcd := e2eutil.NewCluster("test-etcd-", 3, "", nil, nil)
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, origEtcd)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	err = e2eutil.WaitSizeAndVersionReached(t, f.KubeClient, "3.1.10", 3, 6, testEtcd)
	if err != nil {
		t.Fatalf("failed to create 3 members etcd cluster: %v", err)
	}

	targetVersion := "3.2.10"
	updateFunc := func(cl *api.ReplicatedStatefulSet) {
		cl = e2eutil.ClusterWithVersion(cl, targetVersion)
	}
	_, err = e2eutil.UpdateCluster(f.CRClient, testEtcd, 10, updateFunc)
	if err != nil {
		t.Fatalf("fail to update cluster version: %v", err)
	}

	// We have seen in k8s 1.7.1 env it took 35s for the pod to restart with the new image.
	err = e2eutil.WaitSizeAndVersionReached(t, f.KubeClient, targetVersion, 3, 10, testEtcd)
	if err != nil {
		t.Fatalf("failed to wait new version etcd cluster: %v", err)
	}
}
