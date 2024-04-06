package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
)

// Define a mock struct to implement the ElastiCacheAPI interface
type mockElastiCacheClient struct {
	elasticacheiface.ElastiCacheAPI
	Resp elasticache.DescribeCacheClustersOutput
}

// Mock DescribeCacheClusters function
func (m *mockElastiCacheClient) DescribeCacheClusters(input *elasticache.DescribeCacheClustersInput) (*elasticache.DescribeCacheClustersOutput, error) {
	return &m.Resp, nil
}

// Mock data for ElastiCache CacheClusters
var mockElastiCacheData = map[string]*elasticache.DescribeCacheClustersOutput{
	"us-east-1": {
		CacheClusters: []*elasticache.CacheCluster{
			{CacheClusterId: aws.String("cluster-1")},
			{CacheClusterId: aws.String("cluster-2")},
		},
	},
	"us-west-2": {
		CacheClusters: []*elasticache.CacheCluster{
			{CacheClusterId: aws.String("cluster-3")},
		},
	},
}

func TestElastiCacheClusterCounts(t *testing.T) {
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{RegionName: "us-east-1", ExpectedCount: 2},
		{RegionName: "us-west-2", ExpectedCount: 1},
		{AllRegions: true, ExpectedCount: 3},
	}

	for _, c := range cases {
		sf := mockServiceFactory{
			MockElastiCacheClient: func(regionName string) elasticacheiface.ElastiCacheAPI {
				return &mockElastiCacheClient{Resp: *mockElastiCacheData[regionName]}
			},
		}
		am := &mockActivityMonitor{}

		actualCount := ElastiCacheClusterCounts(&sf, am, c.AllRegions)

		if c.ExpectError {
			if !am.ErrorOccurred {
				t.Error("Expected an error to occur, but it did not")
			}
		} else if am.ErrorOccurred {
			t.Errorf("Unexpected error occurred: %s", am.ErrorMessage)
		} else if actualCount != c.ExpectedCount {
			t.Errorf("Error: ElastiCacheClusterCounts returned %d; expected %d", actualCount, c.ExpectedCount)
		}
	}
}