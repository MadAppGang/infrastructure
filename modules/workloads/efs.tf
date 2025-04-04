# Add IAM permissions to task role
data "aws_iam_policy_document" "efs_access" {
  dynamic "statement" {
    for_each = length(var.backend_efs_mounts) > 0 ? [1] : []
    content {
      effect = "Allow"
      actions = [
        "elasticfilesystem:ClientMount",
        "elasticfilesystem:ClientWrite",
        "elasticfilesystem:DescribeMountTargets",
        "elasticfilesystem:ClientRootAccess"
      ]
      resources = [
        for mount in var.backend_efs_mounts :
        "arn:aws:elasticfilesystem:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:file-system/${var.available_efs[mount.efs_name].id}"
      ]
    }
  }
}

resource "aws_iam_role_policy" "efs_access" {
  count  = length(var.backend_efs_mounts) > 0 ? 1 : 0
  name   = "efs-access"
  role   = aws_iam_role.backend_task.id
  policy = data.aws_iam_policy_document.efs_access.json
}
