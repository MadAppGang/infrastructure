resource "aws_ecs_service" "pgadmin" {
  count                              = var.pgadmin_enabled ? 1 : 0
  name                               = "pgadmin_service_${var.env}"
  cluster                            = aws_ecs_cluster.main.id
  task_definition                    = aws_ecs_task_definition.pgadmin[0].arn
  desired_count                      = 1
  deployment_minimum_healthy_percent = 50
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"

  network_configuration {
    security_groups  = [aws_security_group.pgadmin[0].id]
    subnets          = var.subnet_ids
    assign_public_ip = true
  }



  service_connect_configuration {
    enabled   = true
    namespace = aws_service_discovery_private_dns_namespace.local.name
    //TODO: logs
    service {
      port_name      = "pgadmin_service_${var.env}"
      discovery_name = "pgadmin_service_${var.env}"
      client_alias {
        port     = 80
        dns_name = "pgadmin_service_${var.env}"
      }
    }
  }

  tags = {
    Name        = "pgadmin-service-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }

}

resource "aws_ecs_task_definition" "pgadmin" {
  count                    = var.pgadmin_enabled ? 1 : 0
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  family                   = "pgadmin"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.pgadmin_task_execution[0].arn
  task_role_arn            = aws_iam_role.pgadmin_task[0].arn

  container_definitions = jsonencode([{
    name   = "${var.project}_pgadmin_${var.env}"
    cpu    = 256
    memory = 512
    image  = "dpage/pgadmin4:latest"
    environment = [
      { "name" : "PGADMIN_DEFAULT_EMAIL", "value" : tostring(var.pgadmin_email) },
      { "name" : "PGADMIN_DEFAULT_PASSWORD", "value" : tostring(aws_ssm_parameter.pgadmin_password[0].value) },
    ]
    essential = true
    linuxParameters = {
      initProcessEnabled = true
    }
    portMappings = [{
      protocol      = "tcp"
      containerPort = 80
      hostPort      = 80
    }]
  }])

  tags = {
    Name        = "pgadmin-service-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}



resource "aws_security_group" "pgadmin" {
  count  = var.pgadmin_enabled ? 1 : 0
  name   = "${var.project}_mockoon_${var.env}"
  vpc_id = var.vpc_id

  ingress {
    protocol         = "tcp"
    from_port        = 80
    to_port          = 80
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

  tags = {
    Name        = "${var.project}-pgadmin-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}

resource "aws_cloudwatch_log_group" "pgadmin" {
  count = var.pgadmin_enabled ? 1 : 0
  name  = "${var.project}-pgadmin_${var.env}"

  retention_in_days = 1

  tags = {
    Name        = "pgadmin-service-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}


resource "aws_iam_role" "pgadmin_task" {
  count              = var.pgadmin_enabled ? 1 : 0
  name               = "${var.project}_pgadmin_task_${var.env}"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json

  tags = {
    Name        = "${var.project}-pgadmin-task-role-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}

resource "aws_iam_role" "pgadmin_task_execution" {
  count              = var.pgadmin_enabled ? 1 : 0
  name               = "${var.project}_pgadmin_task_execution_${var.env}"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json

  tags = {
    Name        = "${var.project}-pgadmin-task-execution-role-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}


resource "aws_iam_role_policy_attachment" "pgadmin_task_execution" {
  count      = var.pgadmin_enabled ? 1 : 0
  role       = aws_iam_role.pgadmin_task_execution[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "pgadmin_task_cloudwatch" {
  count      = var.pgadmin_enabled ? 1 : 0
  role       = aws_iam_role.pgadmin_task[0].name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchFullAccess"
}

