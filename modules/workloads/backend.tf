

resource "aws_alb_target_group" "backend" {
  name                 = "backend-tg-${var.env}"
  port                 = var.backend_image_port
  protocol             = "HTTP"
  vpc_id               = var.vpc_id
  target_type          = "ip"
  deregistration_delay = 30

  health_check {
    path     = var.backend_health_endpoint
    matcher  = "200-299"
    interval = 30
  }
}

resource "aws_ecs_service" "backend" {
  name                               = "backend_service_${var.env}"
  cluster                            = aws_ecs_cluster.main.id
  task_definition                    = "${aws_ecs_task_definition.backend.family}:${max(aws_ecs_task_definition.backend.revision, data.aws_ecs_task_definition.backend.revision)}"
  desired_count                      = 1
  deployment_minimum_healthy_percent = 50
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"

  network_configuration {
    security_groups  = [aws_security_group.backend.id]
    subnets          = var.subnet_ids
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_alb_target_group.backend.arn
    container_name   = "${var.project}_backend_${var.env}"
    container_port   = var.backend_image_port
  }

  service_registries {
    registry_arn = aws_service_discovery_service.backend.arn
  }

  lifecycle {
    ignore_changes = [task_definition]
  }

  tags = {
    terraform = "true"
    env       = var.env
  }

}

resource "aws_ecs_task_definition" "backend" {
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  family                   = "backend_${var.env}"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.backend_task_execution.arn
  task_role_arn            = aws_iam_role.backend_task.arn

  container_definitions = jsonencode([{
    name        = "${var.project}_backend_${var.env}"
    command     = var.backend_container_command
    cpu         = 256
    memory      = 512
    image       = local.docker_image
    secrets     = local.backend_env_ssm
    environment = concat(local.backend_env, var.backend_env)
    essential   = true

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.backend.name
        awslogs-stream-prefix = "ecs"
        awslogs-region        = data.aws_region.current.name
      }
    }

    portMappings = [{
      protocol      = "tcp"
      containerPort = var.backend_image_port
      hostPort      = var.backend_image_port
    }]
  }])

  tags = {
    terraform = "true"
    env       = var.env
  }
}

data "aws_ecs_task_definition" "backend" {
  task_definition = aws_ecs_task_definition.backend.family
}


resource "aws_security_group" "backend" {
  name   = "${var.project}_backend_${var.env}"
  vpc_id = var.vpc_id

  ingress {
    protocol         = "tcp"
    from_port        = var.backend_image_port
    to_port          = var.backend_image_port
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    protocol         = "-1"
    from_port        = 0
    to_port          = 0
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_cloudwatch_log_group" "backend" {
  name = "${var.project}_backend_${var.env}"

  retention_in_days = 7

  tags = {
    terraform = "true"
    env       = var.env
  }
}

resource "aws_s3_bucket" "images" {
  bucket = "${var.project}-images-${var.env}${var.image_bucket_postfix}"
  tags = {
    terraform = "true"
    env       = var.env
  }
}

resource "aws_s3_bucket_ownership_controls" "images" {
  bucket = aws_s3_bucket.images.id
  rule {
    object_ownership = "ObjectWriter"
  }
}

resource "aws_s3_bucket_public_access_block" "images" {
  bucket = aws_s3_bucket.images.id

  block_public_acls       = !var.image_bucket_public
  block_public_policy     = !var.image_bucket_public
  ignore_public_acls      = !var.image_bucket_public
  restrict_public_buckets = !var.image_bucket_public
}



resource "aws_s3_bucket_acl" "images" {
  bucket = aws_s3_bucket.images.id
  acl    = "private"
  depends_on = [
    aws_s3_bucket_ownership_controls.images,
    aws_s3_bucket_public_access_block.images,
  ]
}

resource "aws_iam_role" "backend_task" {
  name               = "${var.project}_backend_task_${var.env}"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json
}

resource "aws_iam_role" "backend_task_execution" {
  name               = "${var.project}_backend_task_execution_${var.env}"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json
}

resource "aws_iam_policy" "full_access_to_images_bucket" {
  name   = "FullAccessToImagesBucket_${var.project}_${var.env}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Effect": "Allow",
          "Action": [
              "s3:*"        
           ],
          "Resource": [
            "arn:aws:s3:::${aws_s3_bucket.images.id}",
            "arn:aws:s3:::${aws_s3_bucket.images.id}/*"
           ]
      }
  ]
}
EOF

  tags = {
    terraform = "true"
    env       = var.env
  }
}

resource "aws_iam_policy" "send_emails" {
  name   = "SendSESEmails_${var.project}_${var.env}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Effect": "Allow",
          "Action": [
              "ses:SendEmail",
              "ses:SendRawEmail"       
           ],
          "Resource": "*"
      }
  ]
}
EOF

  tags = {
    terraform = "true"
    env       = var.env
  }
}

resource "aws_iam_role_policy_attachment" "backend_task_execution" {
  role       = aws_iam_role.backend_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "backend_task_cloudwatch" {
  role       = aws_iam_role.backend_task.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchFullAccess"
}

resource "aws_iam_role_policy_attachment" "backend_task_images_bucket" {
  role       = aws_iam_role.backend_task.name
  policy_arn = aws_iam_policy.full_access_to_images_bucket.arn
}

resource "aws_iam_role_policy_attachment" "backend_task_ses" {
  role       = aws_iam_role.backend_task.name
  policy_arn = aws_iam_policy.send_emails.arn
}


// SSM IAM access policy
resource "aws_iam_role_policy_attachment" "ssm_parameter_access" {
  role       = aws_iam_role.backend_task_execution.name
  policy_arn = aws_iam_policy.ssm_parameter_access.arn
}

resource "aws_iam_policy" "ssm_parameter_access" {
  name   = "BackendSSMAccessPolicy_${var.project}_${var.env}"
  policy = data.aws_iam_policy_document.ssm_parameter_access.json
}

data "aws_iam_policy_document" "ssm_parameter_access" {
  statement {
    actions   = ["ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath"]
    resources = ["arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter/${var.env}/${var.project}/backend/*"]
  }
}
