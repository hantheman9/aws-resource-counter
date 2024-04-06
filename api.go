/******************************************************************************
Cloud Resource Counter
File: api.go

Summary: ServiceFactory, abstract services and the AWS Service Factory implementation.
******************************************************************************/

package main

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/lightsail/lightsailiface"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/networkfirewall/networkfirewalliface"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/aws/aws-sdk-go/service/opensearchservice/opensearchserviceiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/redshift/redshiftiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// DefaultRegion is used if the caller does not supply a region
// on the command line or the profile does not have a default
// region associated with it.
const DefaultRegion = "us-east-1"

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Abstract Services (hides details of Cloud Provider API)
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// AccountIDService is a struct that knows how get the AWS
// Account ID using an object that implements the Security
// Token Service API interface.
type AccountIDService struct {
	Client stsiface.STSAPI
}

// Account uses the supplied AccountIDService to invoke
// the associated GetCallerIdentity method on the struct's
// Client object.
func (aids *AccountIDService) Account() (string, error) {
	// Construct the input parameter
	input := &sts.GetCallerIdentityInput{}

	// Get the caller's identity
	result, err := aids.Client.GetCallerIdentity(input)
	if err != nil {
		return "", err
	}

	return *result.Account, nil
}

// EC2InstanceService is a struct that knows how to get the
// descriptions of all EC2 instances as well as accessbile
// regions using an object that implements the Elastic
// Compute Cloud API interface.
type EC2InstanceService struct {
	Client ec2iface.EC2API
}

// InspectInstances takes an input filter specification (for the types of instances)
// and a function to evaluate a DescribeInstanceOutput struct. The supplied function
// can determine when to stop iterating through EC2 instances.
func (ec2i *EC2InstanceService) InspectInstances(input *ec2.DescribeInstancesInput,
	fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	return ec2i.Client.DescribeInstancesPages(input, fn)
}

// GetRegions returns the list of available regions for EC2 instances based on the
// set of input parameters.
func (ec2i *EC2InstanceService) GetRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	return ec2i.Client.DescribeRegions(input)
}

// InspectVolumes takes an input filter specification (for the types of volumes)
// and a function to evalatuate a DescribeVolumesOutput struct. The supplied function
// can determine when to stop iterating through EBS volumes.
func (ec2i *EC2InstanceService) InspectVolumes(input *ec2.DescribeVolumesInput,
	fn func(*ec2.DescribeVolumesOutput, bool) bool) error {
	return ec2i.Client.DescribeVolumesPages(input, fn)
}

// ElastiCacheService is a struct that knows how to interact with AWS ElastiCache CacheClusters.
type ElastiCacheService struct {
	Client elasticacheiface.ElastiCacheAPI
}

// GetElastiCacheService returns an instance of an ElastiCacheService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetElastiCacheService(regionName string) *ElastiCacheService {
	var client elasticacheiface.ElastiCacheAPI
	if regionName == "" {
		client = elasticache.New(awssf.Session)
	} else {
		client = elasticache.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &ElastiCacheService{
		Client: client,
	}
}

// DocDBService is a struct that knows how to interact with AWS DocumentDB instances.
type DocDBService struct {
	Client docdbiface.DocDBAPI
}

// GetDocDBService returns an instance of a DocDBService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetDocDBService(regionName string) *DocDBService {
	var client docdbiface.DocDBAPI
	if regionName == "" {
		client = docdb.New(awssf.Session)
	} else {
		client = docdb.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &DocDBService{
		Client: client,
	}
}

type CloudFrontService struct {
	Client cloudfrontiface.CloudFrontAPI
}

func (awssf *AWSServiceFactory) GetCloudFrontService() *CloudFrontService {
	client := cloudfront.New(awssf.Session)
	return &CloudFrontService{
		Client: client,
	}
}

// APIGatewayService struct definition
type APIGatewayService struct {
	Client apigatewayiface.APIGatewayAPI
}

// GetAPIGatewayService method to get an instance of APIGatewayService
func (awssf *AWSServiceFactory) GetAPIGatewayService(regionName string) *APIGatewayService {
	var client apigatewayiface.APIGatewayAPI
	if regionName == "" {
		client = apigateway.New(awssf.Session)
	} else {
		client = apigateway.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &APIGatewayService{
		Client: client,
	}
}

type APIGatewayV2Service struct {
	Client apigatewayv2iface.ApiGatewayV2API
}

func (awssf *AWSServiceFactory) GetAPIGatewayV2Service(regionName string) *APIGatewayV2Service {
	var client apigatewayv2iface.ApiGatewayV2API
	if regionName == "" {
		client = apigatewayv2.New(awssf.Session)
	} else {
		client = apigatewayv2.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}
	return &APIGatewayV2Service{Client: client}
}

type ELBService struct {
	Client elbiface.ELBAPI
}

// GetELBService returns an instance of an ELBService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetELBService(regionName string) *ELBService {
	var client elbiface.ELBAPI
	if regionName == "" {
		client = elb.New(awssf.Session)
	} else {
		client = elb.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &ELBService{
		Client: client,
	}
}

// ELBv2Service is a struct that knows how to interact with AWS Elastic Load Balancers V2.
type ELBv2Service struct {
	Client elbv2iface.ELBV2API
}

// GetELBv2Service returns an instance of an ELBv2Service associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetELBv2Service(regionName string) *ELBv2Service {
	var client elbv2iface.ELBV2API
	if regionName == "" {
		client = elbv2.New(awssf.Session)
	} else {
		client = elbv2.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &ELBv2Service{
		Client: client,
	}
}

type DynamoDBService struct {
	Client dynamodbiface.DynamoDBAPI
}

// GetDynamoDBService returns an instance of a DynamoDBService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetDynamoDBService(regionName string) *DynamoDBService {
	var client dynamodbiface.DynamoDBAPI
	if regionName == "" {
		client = dynamodb.New(awssf.Session)
	} else {
		client = dynamodb.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &DynamoDBService{
		Client: client,
	}
}

// NetworkFirewallService is a struct that knows how to interact with AWS Network Firewall.
type NetworkFirewallService struct {
	Client networkfirewalliface.NetworkFirewallAPI
}

// GetNetworkFirewallService returns an instance of a NetworkFirewallService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetNetworkFirewallService(regionName string) *NetworkFirewallService {
	var client networkfirewalliface.NetworkFirewallAPI
	if regionName == "" {
		client = networkfirewall.New(awssf.Session)
	} else {
		client = networkfirewall.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &NetworkFirewallService{
		Client: client,
	}
}

// IAMService is a struct that knows how to get the
// list of IAM users using an object that implements the IAM API interface.
type IAMService struct {
	Client iamiface.IAMAPI
}

// ListUsers wraps the ListUsers method of the IAM API.
func (s *IAMService) ListUsers(input *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	return s.Client.ListUsers(input)
}

// Add to AWSServiceFactory to include a method for getting an IAMService instance
func (awssf *AWSServiceFactory) GetIAMService() *IAMService {
	return &IAMService{
		Client: iam.New(awssf.Session),
	}
}

// RedshiftService is a struct that knows how to interact with AWS Redshift clusters.
type RedshiftService struct {
	Client redshiftiface.RedshiftAPI
}

// GetRedshiftService returns an instance of a RedshiftService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetRedshiftService(regionName string) *RedshiftService {
	var client redshiftiface.RedshiftAPI
	if regionName == "" {
		client = redshift.New(awssf.Session)
	} else {
		client = redshift.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &RedshiftService{
		Client: client,
	}
}

// OpenSearchService is a struct that knows how to interact with AWS OpenSearch domains.
type OpenSearchService struct {
	Client opensearchserviceiface.OpenSearchServiceAPI
}

// GetOpenSearchService returns an instance of an OpenSearchService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetOpenSearchService(regionName string) *OpenSearchService {
	var client opensearchserviceiface.OpenSearchServiceAPI
	if regionName == "" {
		client = opensearchservice.New(awssf.Session)
	} else {
		client = opensearchservice.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &OpenSearchService{
		Client: client,
	}
}

// RDSInstanceService is a struct that knows how to get the
// descriptions of all RDS instances using an object that
// implements the Relational Database Service API interface.
type RDSInstanceService struct {
	Client rdsiface.RDSAPI
}

// InspectInstances takes an input filter specification (for the types of instances)
// and a function to evaluate a DescribeDBInstancesOutput struct. The supplied function
// can determine when to stop iterating through RDS instances.
func (rdsis *RDSInstanceService) InspectInstances(input *rds.DescribeDBInstancesInput,
	fn func(*rds.DescribeDBInstancesOutput, bool) bool) error {
	return rdsis.Client.DescribeDBInstancesPages(input, fn)
}

// S3Service is a struct that knows how to get all of the S3 buckets using an object
// that implements the Simple Storage Service API interface.
type S3Service struct {
	Client s3iface.S3API
}

// ListBuckets takes an input filter specification (for the types of S3 buckets) and
// returns a ListBucketsOutput struct.
func (s3s *S3Service) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return s3s.Client.ListBuckets(input)
}

// LambdaService is a struct that knows how to get all of the Lambda functions using
// an object that implements the Lambda API interface
type LambdaService struct {
	Client lambdaiface.LambdaAPI
}

// ListFunctions takes an input structure to identify specific lambda functions along
// with a function which is supplied a "page" of lambda functions.
func (ls *LambdaService) ListFunctions(input *lambda.ListFunctionsInput,
	fn func(*lambda.ListFunctionsOutput, bool) bool) error {
	return ls.Client.ListFunctionsPages(input, fn)
}

// ContainerService is a struct that knows how to get a list of all task definition
// and get a description of each one.
type ContainerService struct {
	Client ecsiface.ECSAPI
}

// ListTaskDefinitions takes an input specification (ListTaskDefinitionsInput) and
// a function that is invoked for each page of results (ListTaskDefinitionsOutput).
// This allows a caller to obtain a list of all task definitions.
func (cs *ContainerService) ListTaskDefinitions(input *ecs.ListTaskDefinitionsInput,
	fn func(output *ecs.ListTaskDefinitionsOutput, lastPage bool) bool) error {
	return cs.Client.ListTaskDefinitionsPages(input, fn)
}

// InspectTaskDefinition takes an input specification (DescribeTaskDefinitionInput)
// that describes a single task definition and returns information about it.
func (cs *ContainerService) InspectTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	return cs.Client.DescribeTaskDefinition(input)
}

// LightsailService is a struct that knows how to get a list of all Lightsail
// instances and availble regions.
type LightsailService struct {
	Client lightsailiface.LightsailAPI
}

// GetRegions returns a list of available regions for Lightsail instances
func (lss *LightsailService) GetRegions(input *lightsail.GetRegionsInput) (*lightsail.GetRegionsOutput, error) {
	return lss.Client.GetRegions(input)
}

// InspectInstances returns a full description of all Lightsail instances.
func (lss *LightsailService) InspectInstances(input *lightsail.GetInstancesInput) (*lightsail.GetInstancesOutput, error) {
	return lss.Client.GetInstances(input)
}

// EKSService is a struct that knows how to get a list of all EKS clusters and
// describes the clusters
type EKSService struct {
	Client eksiface.EKSAPI
}

// ListClusters takes an input filter specification and a function
// to evaluate a ListClustersOutput struct. The supplied function
// can determine when to stop iterating through EKS clusters.
func (eksi *EKSService) ListClusters(input *eks.ListClustersInput,
	fn func(*eks.ListClustersOutput, bool) bool) error {
	return eksi.Client.ListClustersPages(input, fn)
}

// ListNodeGroups takes an input filter specification and a function
// to evaluate a ListNodeGroupsOutput struct. The supplied function
// can determine when to stop iterating through Nodegroups.
func (eksi *EKSService) ListNodeGroups(input *eks.ListNodegroupsInput,
	fn func(*eks.ListNodegroupsOutput, bool) bool) error {
	return eksi.Client.ListNodegroupsPages(input, fn)
}

// DescribeNodegroups returns a full description of a Nodegroup
func (eksi *EKSService) DescribeNodegroups(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error) {
	return eksi.Client.DescribeNodegroup(input)
}

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Abstract Service Factory (provides access to all Abstract Services)
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// ServiceFactory is the generic interface for our Cloud Service provider.
type ServiceFactory interface {
	Init()
	GetCurrentRegion() string
	GetAccountIDService() *AccountIDService
	GetEC2InstanceService(string) *EC2InstanceService
	GetEKSService(string) *EKSService
	GetRDSInstanceService(string) *RDSInstanceService
	GetS3Service() *S3Service
	GetLambdaService(string) *LambdaService
	GetContainerService(string) *ContainerService
	GetLightsailService(string) *LightsailService
	GetIAMService() *IAMService
	GetOpenSearchService(string) *OpenSearchService
	GetRedshiftService(string) *RedshiftService
	GetElastiCacheService(string) *ElastiCacheService
	GetDocDBService(string) *DocDBService
	GetCloudFrontService() *CloudFrontService
	GetAPIGatewayService(string) *APIGatewayService
	GetAPIGatewayV2Service(string) *APIGatewayV2Service
	GetELBService(string) *ELBService
	GetELBv2Service(string) *ELBv2Service
	GetDynamoDBService(string) *DynamoDBService
	GetNetworkFirewallService(string) *NetworkFirewallService
}

// AWSServiceFactory is a struct that holds a reference to
// an actual AWS Session object (pointer) and uses it to return
// other specialized services, such as the AccountIDService.
// It also accepts a profile name, overriding region and file
// to use to send trace information.
type AWSServiceFactory struct {
	Session     *session.Session
	ProfileName string
	RegionName  string
	TraceWriter io.Writer
	UseSSO      bool
}

// Init initializes the AWS service factory by creating an
// initial AWS Session object (pointer). It inspects the profiles
// in the current user's directories and prepares the session for
// tracing (if requested).
func (awssf *AWSServiceFactory) Init() {
	config := &aws.Config{}

	// Was a region specified by the user?
	if awssf.RegionName != "" {
		// Add it to the configuration
		config = config.WithRegion(awssf.RegionName)
	}

	// Was tracing specified by the user?
	if awssf.TraceWriter != nil {
		// Enable logging of AWS Calls with Body
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)

		// Enable a logger function which writes to the Trace file
		config = config.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
			fmt.Fprintln(awssf.TraceWriter, args...)
		}))
	}

	// Construct our session Options object
	options := session.Options{
		Config: *config,
	}

	// options to set if using SSO
	if awssf.UseSSO {
		options.SharedConfigState = session.SharedConfigEnable
		options.Profile = awssf.ProfileName
	} else {
		// Create an initial configuration object which defines our chain
		// of credentials providers: first, honor a supplied profile name,
		// if that fails, look for the environment variables.
		options.Config.Credentials = credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.SharedCredentialsProvider{
					Profile: awssf.ProfileName,
				},
				&credentials.EnvProvider{},
			},
		)
	}

	// Ensure that we have a session
	sess := session.Must(session.NewSessionWithOptions(options))

	// Does this session have a region? If not, use the default region
	if *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String(DefaultRegion)})
	}

	// Store the session in our struct
	awssf.Session = sess
}

// GetCurrentRegion returns the name of the current region.
func (awssf *AWSServiceFactory) GetCurrentRegion() string {
	return *awssf.Session.Config.Region
}

// GetAccountIDService returns an instance of an AccountIDService associated
// with our session.
func (awssf *AWSServiceFactory) GetAccountIDService() *AccountIDService {
	return &AccountIDService{
		Client: sts.New(awssf.Session),
	}
}

// GetEC2InstanceService returns an instance of an EC2InstanceService associated
// with our session. The caller can supply an optional region name to contruct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetEC2InstanceService(regionName string) *EC2InstanceService {
	// Construct our service client
	var client ec2iface.EC2API
	if regionName == "" {
		client = ec2.New(awssf.Session)
	} else {
		client = ec2.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &EC2InstanceService{
		Client: client,
	}
}

// GetRDSInstanceService returns an instance of an RDSInstanceService associated
// with our session. The caller can supply an optional region name to construct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetRDSInstanceService(regionName string) *RDSInstanceService {
	// Construct our service client
	var client rdsiface.RDSAPI
	if regionName == "" {
		client = rds.New(awssf.Session)
	} else {
		client = rds.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &RDSInstanceService{
		Client: client,
	}
}

// GetS3Service returns an instance of an S3Service associated with the current session.
// There is currently no way to accept a different region name.
func (awssf *AWSServiceFactory) GetS3Service() *S3Service {
	return &S3Service{
		Client: s3.New(awssf.Session),
	}
}

// GetLambdaService returns an instance of a LambdaService associated with the our session.
// The caller can supply an optional region name to construct an instance associated with
// that region.
func (awssf *AWSServiceFactory) GetLambdaService(regionName string) *LambdaService {
	// Construct our service client
	var client lambdaiface.LambdaAPI
	if regionName == "" {
		client = lambda.New(awssf.Session)
	} else {
		client = lambda.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &LambdaService{
		Client: client,
	}
}

// GetContainerService returns an instance of a ContainerService associated with our session.
// The caller can supply an optional region name to construct an instance associated with
// that region.
func (awssf *AWSServiceFactory) GetContainerService(regionName string) *ContainerService {
	// Construct our service client
	var client ecsiface.ECSAPI
	if regionName == "" {
		client = ecs.New(awssf.Session)
	} else {
		client = ecs.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &ContainerService{
		Client: client,
	}
}

// GetLightsailService returns an instance of a LightsailService associated with our session.
// The caller can supply an optional region name to construct an instance associated with
// that region.
func (awssf *AWSServiceFactory) GetLightsailService(regionName string) *LightsailService {
	// Construct our service client
	var client lightsailiface.LightsailAPI
	if regionName == "" {
		client = lightsail.New(awssf.Session)
	} else {
		client = lightsail.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &LightsailService{
		Client: client,
	}
}

// GetEKSService returns an instance of an EKSService associated
// with our session. The caller can supply an optional region name to contruct
// an instance associated with that region.
func (awssf *AWSServiceFactory) GetEKSService(regionName string) *EKSService {
	// Construct our service client
	var client eksiface.EKSAPI
	if regionName == "" {
		client = eks.New(awssf.Session)
	} else {
		client = eks.New(awssf.Session, aws.NewConfig().WithRegion(regionName))
	}

	return &EKSService{
		Client: client,
	}
}
