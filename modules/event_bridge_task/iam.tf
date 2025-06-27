
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "task" {
  name               = "${var.project}_${var.task}_task_${var.env}"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json

  tags = {
    Name        = "${var.project}-${var.task}-task-role-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}

resource "aws_iam_role_policy_attachment" "task_cloudwatch" {
  role       = aws_iam_role.task.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchFullAccess"
}

resource "aws_iam_role" "task_execution" {
  name               = "${var.project}_scheduler_${var.task}_task_execution_${var.env}"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json

  tags = {
    Name        = "${var.project}-scheduler-${var.task}-execution-role-${var.env}"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}


resource "aws_iam_role_policy_attachment" "task_execution" {
  role       = aws_iam_role.task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}


data "aws_iam_policy_document" "ecs_tasks_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }


  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type = "Service"
      identifiers = ["events.amazonaws.com"]
    }
  }
}

// SSM IAM access policy
resource "aws_iam_role_policy_attachment" "ssm_parameter_access" {
  role       = aws_iam_role.task.name
  policy_arn = aws_iam_policy.ssm_parameter_access.arn
}

resource "aws_iam_role_policy_attachment" "sqs_access" {
  count      = var.sqs_enable == true ? 1 : 0
  role       = aws_iam_role.task.name
  policy_arn = var.sqs_policy_arn
}
resource "aws_iam_policy" "ssm_parameter_access" {
  name   = "Task${var.task}SSMAccessPolicy"
  policy = data.aws_iam_policy_document.ssm_parameter_access.json

  tags = {
    Name        = "Task-${var.task}-SSM-Access-Policy"
    Environment = var.env
    Project     = var.project
    ManagedBy   = "meroku"
    Application = "${var.project}-${var.env}"
  }
}

data "aws_iam_policy_document" "ssm_parameter_access" {
  statement {
    actions   = ["ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath"]
    resources = ["arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:/${var.env}/${var.project}/task/${var.task}/*"]
  }
}
