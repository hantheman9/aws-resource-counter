package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/aws/aws-sdk-go/service/opensearchservice/opensearchserviceiface"

	"github.com/expel-io/aws-resource-counter/mock"
)

// Fake OpenSearch Domain Data
var openSearchDomainsPerRegion = map[string]*opensearchservice.ListDomainNamesOutput{
	"us-east-1": {
		DomainNames: []*opensearchservice.DomainInfo{
			{DomainName: aws.String("domain1")},
			{DomainName: aws.String("domain2")},
		},
	},
	"us-east-2": {
		DomainNames: []*opensearchservice.DomainInfo{
			{DomainName: aws.String("domain3")},
		},
	},
	"af-south-1": {
		DomainNames: []*opensearchservice.DomainInfo{},
	},
}