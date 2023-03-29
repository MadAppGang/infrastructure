variable "env" {
  type = string
}

variable "project" {
  type = string
}

variable "vpc_id" {
  type = string
}


variable "subnet_ids" {
  type = list(string)
}

variable "backend_image_port" {
  default = 8080
  type    = number
}

variable "mockoon_image_port" {
  default = 80
  type = number
}

variable "private_dns_name" {
  default = "chubby.private"
  type    = string
}

variable "ecr_repository_policy" {
  type    = string
  default = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "Default ECR policy",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage",
                "ecr:BatchCheckLayerAvailability",
                "ecr:PutImage",
                "ecr:InitiateLayerUpload",
                "ecr:UploadLayerPart",
                "ecr:CompleteLayerUpload",
                "ecr:DescribeRepositories",
                "ecr:GetRepositoryPolicy",
                "ecr:ListImages",
                "ecr:DeleteRepository",
                "ecr:BatchDeleteImage",
                "ecr:SetRepositoryPolicy",
                "ecr:DeleteRepositoryPolicy"
            ]
        }
    ]
}
EOF
}

variable "domain" {
  default = "chubbydog.com"
}

variable "ecr_url" {
  default = ""
}

variable "ecr_lifecycle_policy" {
  type    = string
  default = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Delete untagged images",
            "selection": {
                "tagStatus": "untagged",
                "countType": "imageCountMoreThan",
                "countNumber": 1
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 2,
            "description": "Keep no more than 10 recent images",
            "selection": {
                "tagStatus": "any",
                "countType": "imageCountMoreThan",
                "countNumber": 10
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}
