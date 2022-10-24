/*
 *  Copyright (c) 2021, WSO2 Inc. (http://www.wso2.org).
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package xds

import (
	"APKManagementServer/internal/logger"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	apkmgt_api "github.com/wso2/product-microgateway/adapter/pkg/discovery/api/wso2/discovery/apkmgt"
	wso2_cache "github.com/wso2/product-microgateway/adapter/pkg/discovery/protocol/cache/v3"
	wso2_resource "github.com/wso2/product-microgateway/adapter/pkg/discovery/protocol/resource/v3"
)

var (
	apiCache wso2_cache.SnapshotCache
	// The labels with partition IDs are stored here. <LabelHirerarchy>-P:<partition_ID>
	// TODO: (VirajSalaka) change the implementation of the snapshot library to provide the same information.
	introducedLabels map[string]bool
	apiCacheMutex    sync.Mutex
	Sent             bool = true
)

const (
	maxRandomInt int    = 999999999
	typeURL      string = "type.googleapis.com/wso2.discovery.ga.Api"
)

// IDHash uses ID field as the node hash.
type IDHash struct{}

// ID uses the node ID field
func (IDHash) ID(node *corev3.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

var _ wso2_cache.NodeHash = IDHash{}

func init() {
	apiCache = wso2_cache.NewSnapshotCache(false, IDHash{}, nil)
	rand.Seed(time.Now().UnixNano())
	introducedLabels = make(map[string]bool, 1)
}

// GetAPICache returns API Cache
func GetAPICache() wso2_cache.SnapshotCache {
	return apiCache
}

//FeedData mock data
func FeedData() {
	logger.LoggerXdsServer.Debug("adding mock data")
	version := rand.Intn(maxRandomInt)
	api := &apkmgt_api.Api{
		ApiUUID:      "apiUUID1",
		RevisionUUID: "revisionUUID1",
	}

	newSnapshot, _ := wso2_cache.NewSnapshot(fmt.Sprint(version), map[wso2_resource.Type][]types.Resource{
		wso2_resource.APKMgtType: {api},
	})
	apiCacheMutex.Lock()
	apiCache.SetSnapshot(context.Background(), "mine", newSnapshot)
	apiCacheMutex.Unlock()
	logger.LoggerXdsServer.Debug("added mock data")
}

// SetEmptySnapshot sets an empty snapshot into the apiCache for the given label
// this is used to set empty snapshot when there are no APIs available for a label
func SetEmptySnapshot(label string) error {
	version := rand.Intn(maxRandomInt)
	newSnapshot, err := wso2_cache.NewSnapshot(fmt.Sprint(version), map[wso2_resource.Type][]types.Resource{
		wso2_resource.APKMgtType: {},
	})
	if err != nil {
		logger.LoggerXdsServer.Error("Error creating empty snapshot. error: ", err.Error())
		return err
	}
	apiCacheMutex.Lock()
	defer apiCacheMutex.Unlock()
	//performing null check again to avoid race conditions
	_, errSnap := apiCache.GetSnapshot(label)
	if errSnap != nil && strings.Contains(errSnap.Error(), "no snapshot found for node") {
		errSetSnap := apiCache.SetSnapshot(context.Background(), label, newSnapshot)
		if errSetSnap != nil {
			logger.LoggerXdsServer.Error("Error setting empty snapshot to apiCache. error : ", errSetSnap.Error())
			return errSetSnap
		}
	}
	return nil
}
