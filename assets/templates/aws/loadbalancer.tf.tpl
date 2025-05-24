# Application Load Balancer Configuration
{{if .EnableLoadBalancer}}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "{{.ProjectName}}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false

  tags = {
    Name = "{{.ProjectName}}-alb"
  }
}

# Security group for ALB
resource "aws_security_group" "alb" {
  name_prefix = "{{.ProjectName}}-alb-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "{{.ProjectName}}-alb-sg"
  }
}

# Update application security group to allow traffic from ALB
resource "aws_security_group_rule" "app_from_alb" {
  type                     = "ingress"
  from_port                = 80
  to_port                  = 80
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.alb.id
  security_group_id        = aws_security_group.app.id
  description              = "HTTP from ALB"
}

{{if ne .AppType "static-site"}}
# Target group for application instances
resource "aws_lb_target_group" "app" {
  name     = "{{.ProjectName}}-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/"
    matcher             = "200"
    port                = "traffic-port"
    protocol            = "HTTP"
  }

  tags = {
    Name = "{{.ProjectName}}-tg"
  }
}

{{if eq .AppType "nodejs" "flask"}}
# Attach EC2 instance to target group
resource "aws_lb_target_group_attachment" "app" {
  target_group_arn = aws_lb_target_group.app.arn
  target_id        = aws_instance.app.id
  port             = 80
}
{{end}}

# ALB Listener for HTTP
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

{{if eq .AppType "docker"}}
# Service discovery for ECS
resource "aws_service_discovery_private_dns_namespace" "main" {
  name        = "{{.ProjectName}}.local"
  description = "Service discovery namespace for {{.ProjectName}}"
  vpc         = aws_vpc.main.id
}

resource "aws_service_discovery_service" "app" {
  name = "{{.ProjectName}}-service"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_grace_period_seconds = 30
}

# Update ECS service to use load balancer and service discovery
resource "aws_ecs_service" "app_with_lb" {
  name            = "{{.ProjectName}}-service-lb"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.app.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "{{.ProjectName}}"
    container_port   = 80
  }

  service_registries {
    registry_arn = aws_service_discovery_service.app.arn
  }

  depends_on = [aws_lb_listener.http]
}
{{end}}

{{else}}
# For static sites, redirect to S3 website endpoint
resource "aws_lb_listener" "http_static" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      host        = aws_s3_bucket_website_configuration.main.website_endpoint
      port        = "80"
      protocol    = "HTTP"
      status_code = "HTTP_301"
    }
  }
}
{{end}}

# CloudWatch Log Group for ALB access logs
resource "aws_cloudwatch_log_group" "alb_logs" {
  name              = "/aws/loadbalancer/{{.ProjectName}}"
  retention_in_days = 7
}

# Store ALB DNS name in Parameter Store
resource "aws_ssm_parameter" "alb_dns_name" {
  name  = "/{{.ProjectName}}/loadbalancer/dns_name"
  type  = "String"
  value = aws_lb.main.dns_name

  tags = {
    Name = "{{.ProjectName}}-alb-dns-name"
  }
}

resource "aws_ssm_parameter" "alb_arn" {
  name  = "/{{.ProjectName}}/loadbalancer/arn"
  type  = "String"
  value = aws_lb.main.arn

  tags = {
    Name = "{{.ProjectName}}-alb-arn"
  }
}

resource "aws_ssm_parameter" "alb_zone_id" {
  name  = "/{{.ProjectName}}/loadbalancer/zone_id"
  type  = "String"
  value = aws_lb.main.zone_id

  tags = {
    Name = "{{.ProjectName}}-alb-zone-id"
  }
}

{{end}} 