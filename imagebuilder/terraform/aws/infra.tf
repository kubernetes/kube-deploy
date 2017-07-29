variable "access_key" {}
variable "secret_key" {}
variable "region" {
  default = "us-east-1"
}

provider "aws" {
  region = "${var.region}"
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
}

resource "aws_vpc" "imagebuilder" {
  cidr_block = "172.20.0.0/16"

  tags {
    Name = "imagebuilder"
    "k8s.io/role/imagebuilder" = "1"
  }
}

resource "aws_subnet" "imagebuilder" {
  vpc_id     = "${aws_vpc.imagebuilder.id}"
  cidr_block = "172.20.1.0/24"

  tags {
    Name = "imagebuilder"
    "k8s.io/role/imagebuilder" = "1"
  }
}

resource "aws_internet_gateway" "imagebuilder" {
  vpc_id = "${aws_vpc.imagebuilder.id}"

  tags {
    Name = "imagebuilder"
    "k8s.io/role/imagebuilder" = "1"
  }
}

resource "aws_route" "internet_access" {
  route_table_id         = "${aws_vpc.imagebuilder.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.imagebuilder.id}"
}

resource "aws_security_group" "allow_all" {
  name        = "imagebuilder"
  description = "Imagebuilder security group"
  vpc_id = "${aws_vpc.imagebuilder.id}"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "imagebuilder-ssh"
    "k8s.io/role/imagebuilder" = "1"
  }
}
