variable "resource_name" {
  type        = string
  default     = "turbot-test-20200125-create-update"
  description = "Name of the resource used throughout the test."
}

variable "aws_profile" {
  type        = string
  default     = "default"
  description = "AWS credentials profile used for the test. Default is to use the default profile."
}

variable "aws_region" {
  type        = string
  default     = "us-west-2"
  description = "AWS region used for the test. Does not work with default region in config, so must be defined here."
}

variable "aws_region_alternate" {
  type        = string
  default     = "us-east-2"
  description = "Alternate AWS region used for tests that require two regions (e.g. DynamoDB global tables)."
}

provider "aws" {
  profile = var.aws_profile
  region  = var.aws_region
}

provider "aws" {
  alias   = "alternate"
  profile = var.aws_profile
  region  = var.aws_region_alternate
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}
data "aws_region" "primary" {}
data "aws_region" "alternate" {
  provider = aws.alternate
}

data "null_data_source" "resource" {
  inputs = {
    scope = "arn:${data.aws_partition.current.partition}:::${data.aws_caller_identity.current.account_id}"
  }
}

resource "aws_securityhub_account" "named_test_resource" {}

resource "aws_securityhub_action_target" "named_test_resource" {
  depends_on  = [aws_securityhub_account.named_test_resource]
  name        = "Send notification"
  identifier  = "SendToChat"
  description = "This is custom action sends selected findings to chat"
}

output "name" {
  value = aws_securityhub_action_target.named_test_resource.name
}

output "arn" {
  value = aws_securityhub_action_target.named_test_resource.arn
}

output "aws_region" {
  value = data.aws_region.primary.name
}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
