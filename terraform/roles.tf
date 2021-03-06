# ECS task execution role data
data "aws_iam_policy_document" "ecs_task_execution_role" {
  version = "2012-10-17"
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# ECS task execution role
resource "aws_iam_role" "ecs_task_execution_role" {
  name               = var.ecs_task_execution_role_name
  assume_role_policy = data.aws_iam_policy_document.ecs_task_execution_role.json
}

# ECS task execution role policy attachment
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role" "ecs_task_role" {
  name = "delphis-ecsTaskRole"

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Action": "sts:AssumeRole",
     "Principal": {
       "Service": "ecs-tasks.amazonaws.com"
     },
     "Effect": "Allow",
     "Sid": ""
   }
 ]
}
EOF
}

# ECS task execution role policy attachment
//resource "aws_iam_role_policy_attachment" "ecs_task_execution_role-SQS" {
//  role       = aws_iam_role.ecs_task_role.name
//  policy_arn = "arn:aws:iam::aws:policy/AmazonSQSFullAccess"
//}

# ECS task execution role policy attachment
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role-S3" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
}

resource "aws_iam_policy" "dynamodb" {
  name        = "delphis-task-policy-dynamodb"
  description = "Policy that allows access to DynamoDB"

  policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Effect": "Allow",
           "Action": [
               "dynamodb:UpdateTimeToLive",
               "dynamodb:PutItem",
               "dynamodb:BatchWriteItem",
               "dynamodb:DescribeTable",
               "dynamodb:ListTables",
               "dynamodb:DeleteItem",
               "dynamodb:GetItem",
               "dynamodb:BatchGetItem",
               "dynamodb:Scan",
               "dynamodb:Query",
               "dynamodb:UpdateItem"
           ],
           "Resource": "*"
       }
   ]
}
EOF
}

# resource "aws_iam_policy" "rds-read-write" {
#   name = "chatham-task-policy-rds-read-write"
#   description = "Policy that allows access to rds for read and write"

#   assume_role_policy = <<EOF
# {
#   "Version": "2012-10-17",
#   "Statement": [
#     {
#       "Sid": "RDS Read Write",
#       "Effect": "Allow",
#       "Action": [
#         // TODO
#       ],
#       "Resource": {cluster_instances-staging.arn}
#     }
#   ]
# }
# EOF

resource "aws_iam_role_policy_attachment" "ecs-task-role-policy-attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.dynamodb.arn
}

# resource ""

// TODO: This should limit to just the specific secret that we use for delphis.
resource "aws_iam_policy" "secrets-manager" {
  name        = "delphis-task-policy-secrets-manager"
  description = "Policy that allows access to Secrets-manager"

  policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Effect": "Allow",
           "Action": [
              "secretsmanager:Describe*",
              "secretsmanager:Get*",
              "secretsmanager:List*" 
           ],
           "Resource": "*"
       }
   ]
}
EOF
}

resource "aws_iam_policy" "github-actions-deployment" {
  name        = "delphis-github-actions-ecs-deployment"
  description = "Policy that allows Github Actions to deploy ECS"
  policy      = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecs:DeregisterTaskDefinition",
                "ecs:DescribeServices",
                "ecs:DescribeTaskDefinition",
                "ecs:DescribeTasks",
                "ecs:ListTasks",
                "ecs:ListTaskDefinitions",
                "ecs:RegisterTaskDefinition",
                "ecs:StartTask",
                "ecs:StopTask",
                "ecs:UpdateService",
                "iam:PassRole"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs-task-role-policy-attachment-secrets-manager" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.secrets-manager.arn
}

// Zapier user for consuming SQS
resource "aws_iam_user" "zapier_sqs" {
  name = "zapier_sqs"
}

resource "aws_iam_policy_attachment" "zapier_sqs" {
  name       = "sqs_attachment"
  users      = [aws_iam_user.zapier_sqs.name]
  roles      = [aws_iam_role.ecs_task_role.name]
  policy_arn = "arn:aws:iam::aws:policy/AmazonSQSFullAccess"
}

// Zapier user for consuming SQS
resource "aws_iam_user" "github-action-deploy" {
  name = "github-action-deploy"
}

resource "aws_iam_policy_attachment" "github-action-attach" {
  name       = "github-action-attach"
  users      = [aws_iam_user.github-action-deploy.name]
  policy_arn = aws_iam_policy.github-actions-deployment.arn
}