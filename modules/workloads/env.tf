resource "aws_ssm_parameter" "backend_env" {
  name  = "/${var.env}/${var.project}/backend/env"
  type  = "SecureString"
  value = " "

  // if we manually change the value, don't rewrite it
  lifecycle {
    ignore_changes = [
      value,
    ]
  }

  tags = {
    Name        = "/${var.env}/${var.project}/backend/env"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}



data "aws_ssm_parameters_by_path" "backend" {
  path      = "/${var.env}/${var.project}/backend"
  recursive = true
  depends_on = [
    aws_ssm_parameter.backend_env
  ]
}



locals {
  backend_env_ssm = [
    for i in range(length(data.aws_ssm_parameters_by_path.backend.names)) : {
      name      = upper(reverse(split("/", data.aws_ssm_parameters_by_path.backend.names[i]))[0])
      valueFrom = data.aws_ssm_parameters_by_path.backend.names[i]
    }
  ]


}

locals {
  backend_env = [
    { "name" : "PG_DATABASE_HOST", "value" : var.db_endpoint },
    { "name" : "PG_DATABASE_USERNAME", "value" : var.db_user },
    { "name" : "PORT", "value" : tostring(var.backend_image_port) },
    { "name" : "PG_DATABASE_NAME", "value" : var.db_name },
    { "name" : "AWS_S3_BUCKET", "value" : "${aws_s3_bucket.backend.bucket}" },
    { "name" : "AWS_REGION", "value" : data.aws_region.current.name },
    { "name" : "URL", "value" : var.api_domain },
    { "name" : "SQS_QUEUE_URL", "value" : var.sqs_queue_url },
    { "name" : "AWS_QUEUE_URL", "value" : var.sqs_queue_url },
  ]
}
