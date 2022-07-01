SELECT
  db_snapshot_identifier,
  arn,
  TYPE,
  allocated_storage,
  db_instance_identifier,
  encrypted,
  engine,
  iam_database_authentication_enabled,
  license_model,
  master_user_name,
  port,
  storage_type,
  vpc_id,
  tags_src
FROM
  aws.aws_rds_db_snapshot
WHERE
  db_snapshot_identifier = '{{ resourceName }}'
