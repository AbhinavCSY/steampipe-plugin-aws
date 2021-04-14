/*
Package aws implements a steampipe plugin for aws.

This plugin provides data that Steampipe uses to present foreign
tables that represent Amazon AWS resources.
*/
package aws

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

const pluginName = "steampipe-plugin-aws"

// Plugin creates this (aws) plugin
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             pluginName,
		DefaultTransform: transform.FromCamel(),
		DefaultGetConfig: &plugin.GetConfig{
			ShouldIgnoreError: isNotFoundError([]string{"ResourceNotFoundException", "NoSuchEntity"}),
		},
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		TableMap: map[string]*plugin.Table{
			"aws_account":                            tableAwsAccount(ctx),
			"aws_acm_certificate":                    tableAwsAcmCertificate(ctx),
			"aws_api_gateway_api_key":                tableAwsAPIGatewayAPIKey(ctx),
			"aws_api_gateway_authorizer":             tableAwsAPIGatewayAuthorizer(ctx),
			"aws_api_gateway_rest_api":               tableAwsAPIGatewayRestAPI(ctx),
			"aws_api_gateway_stage":                  tableAwsAPIGatewayStage(ctx),
			"aws_api_gateway_usage_plan":             tableAwsAPIGatewayUsagePlan(ctx),
			"aws_api_gatewayv2_api":                  tableAwsAPIGatewayV2Api(ctx),
			"aws_api_gatewayv2_domain_name":          tableAwsAPIGatewayV2DomainName(ctx),
			"aws_api_gatewayv2_stage":                tableAwsAPIGatewayV2Stage(ctx),
			"aws_availability_zone":                  tableAwsAvailabilityZone(ctx),
			"aws_backup_plan":                        tableAwsBackupPlan(ctx),
			"aws_backup_vault":                       tableAwsBackupVault(ctx),
			"aws_cloudformation_stack":               tableAwsCloudFormationStack(ctx),
			"aws_cloudtrail_trail":                   tableAwsCloudtrailTrail(ctx),
			"aws_cloudwatch_alarm":                   tableAwsCloudWatchAlarm(ctx),
			"aws_cloudwatch_log_group":               tableAwsCloudwatchLogGroup(ctx),
			"aws_cloudwatch_log_metric_filter":       tableAwsCloudwatchLogMetricFilter(ctx),
			"aws_cloudwatch_log_stream":              tableAwsCloudwatchLogStream(ctx),
			"aws_config_configuration_recorder":      tableAwsConfigConfigurationRecorder(ctx),
			"aws_config_conformance_pack":            tableAwsConfigConformancePack(ctx),
			"aws_dynamodb_backup":                    tableAwsDynamoDBBackup(ctx),
			"aws_dynamodb_global_table":              tableAwsDynamoDBGlobalTable(ctx),
			"aws_dynamodb_table":                     tableAwsDynamoDBTable(ctx),
			"aws_ebs_snapshot":                       tableAwsEBSSnapshot(ctx),
			"aws_ebs_volume":                         tableAwsEBSVolume(ctx),
			"aws_ec2_ami":                            tableAwsEc2Ami(ctx),
			"aws_ec2_application_load_balancer":      tableAwsEc2ApplicationLoadBalancer(ctx),
			"aws_ec2_autoscaling_group":              tableAwsEc2ASG(ctx),
			"aws_ec2_classic_load_balancer":          tableAwsEc2ClassicLoadBalancer(ctx),
			"aws_ec2_gateway_load_balancer":          tableAwsEc2GatewayLoadBalancer(ctx),
			"aws_ec2_instance":                       tableAwsEc2Instance(ctx),
			"aws_ec2_instance_availability":          tableAwsInstanceAvailability(ctx),
			"aws_ec2_instance_type":                  tableAwsInstanceType(ctx),
			"aws_ec2_key_pair":                       tableAwsEc2KeyPair(ctx),
			"aws_ec2_launch_configuration":           tableAwsEc2LaunchConfiguration(ctx),
			"aws_ec2_load_balancer_listener":         tableAwsEc2ApplicationLoadBalancerListener(ctx),
			"aws_ec2_network_interface":              tableAwsEc2NetworkInterface(ctx),
			"aws_ec2_network_load_balancer":          tableAwsEc2NetworkLoadBalancer(ctx),
			"aws_ec2_target_group":                   tableAwsEc2TargetGroup(ctx),
			"aws_ec2_transit_gateway":                tableAwsEc2TransitGateway(ctx),
			"aws_ec2_transit_gateway_route_table":    tableAwsEc2TransitGatewayRouteTable(ctx),
			"aws_ec2_transit_gateway_vpc_attachment": tableAwsEc2TransitGatewayVpcAttachment(ctx),
			"aws_ecr_repository":                     tableAwsEcrRepository(ctx),
			"aws_ecs_cluster":                        tableAwsEcsCluster(ctx),
			"aws_ecs_task_definition":                tableAwsEcsTaskDefinition(ctx),
			"aws_efs_access_point":                   tableAwsEfsAccessPoint(ctx),
			"aws_efs_file_system":                    tableAwsElasticFileSystem(ctx),
			"aws_eks_cluster":                        tableAwsEksCluster(ctx),
			"aws_elastic_beanstalk_environment":      tableAwsElasticBeanstalkEnvironment(ctx),
			"aws_elasticache_cluster":                tableAwsElastiCacheCluster(ctx),
			"aws_elasticache_parameter_group":        tableAwsElastiCacheParameterGroup(ctx),
			"aws_elasticache_replication_group":      tableAwsElastiCacheReplicationGroup(ctx),
			"aws_elastic_beanstalk_application":      tableAwsElasticBeanstalkApplication(ctx),
			"aws_emr_cluster":                        tableAwsEmrCluster(ctx),
			"aws_eventbridge_rule":                   tableAwsEventBridgeRule(ctx),
			"aws_glacier_vault":                      tableAwsGlacierVault(ctx),
			"aws_guardduty_detector":                 tableAwsGuardDutyDetector(ctx),
			"aws_guardduty_threat_intel_set":         tableAwsGuardDutyThreatIntelSet(ctx),
			"aws_iam_access_advisor":                 tableAwsIamAccessAdvisor(ctx),
			"aws_iam_access_key":                     tableAwsIamAccessKey(ctx),
			"aws_iam_account_password_policy":        tableAwsIamAccountPasswordPolicy(ctx),
			"aws_iam_account_summary":                tableAwsIamAccountSummary(ctx),
			"aws_iam_action":                         tableAwsIamAction(ctx),
			"aws_iam_credential_report":              tableAwsIamCredentialReport(ctx),
			"aws_iam_group":                          tableAwsIamGroup(ctx),
			"aws_iam_policy":                         tableAwsIamPolicy(ctx),
			"aws_iam_policy_simulator":               tableAwsIamPolicySimulator(ctx),
			"aws_iam_role":                           tableAwsIamRole(ctx),
			"aws_iam_server_certificate":             tableAwsIamServerCertificate(ctx),
			"aws_iam_user":                           tableAwsIamUser(ctx),
			"aws_iam_virtual_mfa_device":             tableAwsIamVirtualMfaDevice(ctx),
			"aws_inspector_assessment_target":        tableAwsInspectorAssessmentTarget(ctx),
			"aws_inspector_assessment_template":      tableAwsInspectorAssessmentTemplate(ctx),
			"aws_kinesis_consumer":                   tableAwsKinesisConsumer(ctx),
			"aws_kinesis_firehose_delivery_stream":   tableAwsKinesisFirehoseDeliveryStream(ctx),
			"aws_kinesis_stream":                     tableAwsKinesisStream(ctx),
			"aws_kinesis_video_stream":               tableAwsKinesisVideoStream(ctx),
			"aws_kms_key":                            tableAwsKmsKey(ctx),
			"aws_lambda_alias":                       tableAwsLambdaAlias(ctx),
			"aws_lambda_function":                    tableAwsLambdaFunction(ctx),
			"aws_lambda_version":                     tableAwsLambdaVersion(ctx),
			"aws_rds_db_cluster":                     tableAwsRDSDBCluster(ctx),
			"aws_rds_db_cluster_parameter_group":     tableAwsRDSDBClusterParameterGroup(ctx),
			"aws_rds_db_cluster_snapshot":            tableAwsRDSDBClusterSnapshot(ctx),
			"aws_rds_db_instance":                    tableAwsRDSDBInstance(ctx),
			"aws_rds_db_option_group":                tableAwsRDSDBOptionGroup(ctx),
			"aws_rds_db_parameter_group":             tableAwsRDSDBParameterGroup(ctx),
			"aws_rds_db_snapshot":                    tableAwsRDSDBSnapshot(ctx),
			"aws_rds_db_subnet_group":                tableAwsRDSDBSubnetGroup(ctx),
			"aws_redshift_cluster":                   tableAwsRedshiftCluster(ctx),
			"aws_redshift_event_subscription":        tableAwsRedshiftEventSubscription(ctx),
			"aws_redshift_parameter_group":           tableAwsRedshiftParameterGroup(ctx),
			"aws_redshift_subnet_group":              tableAwsRedshiftSubnetGroup(ctx),
			"aws_region":                             tableAwsRegion(ctx),
			"aws_route53_record":                     tableAwsRoute53Record(ctx),
			"aws_route53_resolver_endpoint":          tableAwsRoute53ResolverEndpoint(ctx),
			"aws_route53_resolver_rule":              tableAwsRoute53ResolverRule(ctx),
			"aws_route53_zone":                       tableAwsRoute53Zone(ctx),
			"aws_s3_account_settings":                tableAwsS3AccountSettings(ctx),
			"aws_s3_bucket":                          tableAwsS3Bucket(ctx),
			"aws_securityhub_hub":                    tableAwsSecurityHub(ctx),
			"aws_securityhub_product":                tableAwsSecurityhubProduct(ctx),
			"aws_sns_topic":                          tableAwsSnsTopic(ctx),
			"aws_sns_topic_subscription":             tableAwsSnsTopicSubscription(ctx),
			"aws_sqs_queue":                          tableAwsSqsQueue(ctx),
			"aws_ssm_association":                    tableAwsSSMAssociation(ctx),
			"aws_ssm_document":                       tableAwsSSMDocument(ctx),
			"aws_ssm_maintenance_window":             tableAwsSSMMaintenanceWindow(ctx),
			"aws_ssm_parameter":                      tableAwsSSMParameter(ctx),
			"aws_ssm_patch_baseline":                 tableAwsSSMPatchBaseline(ctx),
			"aws_vpc":                                tableAwsVpc(ctx),
			"aws_vpc_customer_gateway":               tableAwsVpcCustomerGateway(ctx),
			"aws_vpc_dhcp_options":                   tableAwsVpcDhcpOptions(ctx),
			"aws_vpc_egress_only_internet_gateway":   tableAwsVpcEgressOnlyIGW(ctx),
			"aws_vpc_eip":                            tableAwsVpcEip(ctx),
			"aws_vpc_endpoint":                       tableAwsVpcEndpoint(ctx),
			"aws_vpc_endpoint_service":               tableAwsVpcEndpointService(ctx),
			"aws_vpc_flow_log":                       tableAwsVpcFlowlog(ctx),
			"aws_vpc_internet_gateway":               tableAwsVpcInternetGateway(ctx),
			"aws_vpc_nat_gateway":                    tableAwsVpcNatGateway(ctx),
			"aws_vpc_network_acl":                    tableAwsVpcNetworkACL(ctx),
			"aws_vpc_route":                          tableAwsVpcRoute(ctx),
			"aws_vpc_route_table":                    tableAwsVpcRouteTable(ctx),
			"aws_vpc_security_group":                 tableAwsVpcSecurityGroup(ctx),
			"aws_vpc_security_group_rule":            tableAwsVpcSecurityGroupRule(ctx),
			"aws_vpc_subnet":                         tableAwsVpcSubnet(ctx),
			"aws_vpc_vpn_gateway":                    tableAwsVpcVpnGateway(ctx),
			"aws_wellarchitected_workload":           tableAwsWellArchitectedWorkload(ctx),
		},
	}

	return p
}
