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
	"fmt"
	"os"
	"testing"

	api "github.com/beekhof/galera-operator/pkg/apis/galera/v1alpha1"
	"github.com/beekhof/galera-operator/test/e2e/e2eutil"
	"github.com/beekhof/galera-operator/test/e2e/framework"
)

func TestResizeCluster3to5(t *testing.T) {
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

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 6, testEtcd); err != nil {
		t.Fatalf("failed to create 3 members etcd cluster: %v", err)
	}
	fmt.Println("reached to 3 members cluster")

	updateFunc := func(cl *api.ReplicatedStatefulSet) {
		cl.Spec.Replicas = 5
	}
	if _, err := e2eutil.UpdateCluster(f.CRClient, testEtcd, 10, updateFunc); err != nil {
		t.Fatal(err)
	}

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 5, 6, testEtcd); err != nil {
		t.Fatalf("failed to resize to 5 members etcd cluster: %v", err)
	}
}

func TestResizeCluster5to3(t *testing.T) {
	if os.Getenv(envParallelTest) == envParallelTestTrue {
		t.Parallel()
	}
	f := framework.Global
	testEtcd, err := e2eutil.CreateCluster(t, f.CRClient, f.Namespace, e2eutil.NewCluster("test-etcd-", 5, "", nil, nil))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := e2eutil.DeleteCluster(t, f.CRClient, f.KubeClient, testEtcd); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 5, 9, testEtcd); err != nil {
		t.Fatalf("failed to create 5 members etcd cluster: %v", err)
	}
	fmt.Println("reached to 5 members cluster")

	updateFunc := func(cl *api.ReplicatedStatefulSet) {
		cl.Spec.Replicas = 3
	}
	if _, err := e2eutil.UpdateCluster(f.CRClient, testEtcd, 10, updateFunc); err != nil {
		t.Fatal(err)
	}

	if _, err := e2eutil.WaitUntilSizeReached(t, f.CRClient, 3, 6, testEtcd); err != nil {
		t.Fatalf("failed to resize to 3 members etcd cluster: %v", err)
	}
}
