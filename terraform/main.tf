provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "${var.region}"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "beginnerd-firehose-bucket"
  acl = "private"
  force_destroy = true
}

resource "aws_iam_role" "firehose_role" {
   name = "firehose_test_role"
   assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "s3_policy" {
    name = "s3_policy"
    role = "${aws_iam_role.firehose_role.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "beginnerd_stream" {
  name = "beginnerd_firehose_stream"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose_role.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_interval = 60
  }
}