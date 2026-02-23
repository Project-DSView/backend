# Enhanced Monitoring and Logging System

This directory contains configuration files for the enhanced monitoring and logging system implemented for DSView Backend.

## Overview

The monitoring and logging system provides:

- **Centralized Logging**: All application logs aggregated in Grafana
- **Request Tracing**: Track requests across services with correlation IDs
- **Security Middleware**: Rate limiting, security headers, IP filtering
- **Performance Monitoring**: Compression, caching, circuit breaker
- **Real-time Dashboards**: Beautiful Grafana dashboards for monitoring

## Architecture

```
Client Request
     ↓
Traefik (Load Balancer + Middleware)
     ↓
FastAPI/Go Services (Structured Logging)
     ↓
Promtail (Log Collector)
     ↓
Loki (Log Storage)
     ↓
Grafana (Visualization)
```

## Services

### 1. Traefik
- **Purpose**: Load balancer, reverse proxy, and middleware gateway
- **Port**: 80 (HTTP), 443 (HTTPS), 8080 (Dashboard)
- **Configuration**: `traefik/traefik.yml` (static), `traefik/dynamic/` (dynamic)
- **Features**:
  - Security headers (HSTS, XSS protection, CSP)
  - Rate limiting (100 req/min, burst 200)
  - Compression (gzip/brotli)
  - Circuit breaker (fault tolerance)
  - Request tracing with correlation IDs

### 2. Loki
- **Purpose**: Log aggregation and storage
- **Port**: 3100
- **Configuration**: `loki/loki-config.yml`
- **Features**:
  - Efficient log storage
  - LogQL query language
  - Retention policies
  - Multi-tenant support

### 3. Promtail
- **Purpose**: Log collector and forwarder
- **Configuration**: `promtail/promtail-config.yml`
- **Features**:
  - Scrapes logs from all services
  - Parses JSON logs
  - Adds labels for filtering
  - Forwards to Loki

### 4. Grafana
- **Purpose**: Log visualization and dashboards
- **Port**: 3001
- **Login**: admin/admin
- **Configuration**: `grafana/provisioning/`
- **Features**:
  - Pre-configured dashboards
  - LogQL queries
  - Alerting (optional)
  - User management

## Access Points

| Service | URL | Purpose |
|---------|-----|---------|
| Grafana Dashboard | http://localhost:3001 | Log visualization and monitoring |
| Traefik Dashboard | http://localhost:8080 | Load balancer management |
| Loki API | http://localhost:3100 | Log query API |
| Prometheus Metrics | http://localhost:8080/metrics | Traefik metrics |

## Quick Start

### 1. Start All Services
```bash
# Start the entire stack
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

### 2. Access Grafana
1. Open http://localhost:3001
2. Login with admin/admin
3. Navigate to "Dashboards" to see pre-configured dashboards

### 3. View Logs
- **Traefik Dashboard**: http://localhost:8080
- **Application Logs**: Use Grafana Explore with LogQL queries

## LogQL Queries

### Common Queries

**View all logs:**
```logql
{job=~".+"}
```

**View error logs:**
```logql
{job=~".+"} |= "ERROR"
```

**View FastAPI logs:**
```logql
{service="fastapi"}
```

**View Go service logs:**
```logql
{service="go-app"}
```

**View requests by status code:**
```logql
{job="traefik"} | json | status >= 400
```

**View slow requests:**
```logql
{service=~"fastapi|go-app"} | json | duration > 1.0
```

**View requests by user:**
```logql
{service=~"fastapi|go-app"} | json | user_id != ""
```

**View requests by request ID:**
```logql
{service=~"fastapi|go-app"} | json | request_id="abc123"
```

### Advanced Queries

**Rate of requests per minute:**
```logql
rate({job="traefik"}[1m])
```

**Top error responses:**
```logql
topk(10, count by (status) ({job="traefik"} | json | status >= 400))
```

**Average response time:**
```logql
avg_over_time({service=~"fastapi|go-app"} | json | unwrap duration [5m])
```

## Configuration

### Traefik Configuration

**Static Configuration** (`traefik/traefik.yml`):
- Global settings
- Entry points
- Providers
- Logging format
- Metrics configuration

**Dynamic Configuration** (`traefik/dynamic/`):
- Middleware definitions
- Router rules
- Service configurations

### Environment Variables

Copy `traefik/env.example` to `traefik/.env` and modify:

```bash
# Rate limiting
RATE_LIMIT_AVERAGE=100
RATE_LIMIT_BURST=200

# IP whitelist (comma-separated)
IP_WHITELIST=127.0.0.1/32,10.0.0.0/8

# CORS origins
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://localhost:3000
```

### Log Retention

**Loki Configuration** (`loki/loki-config.yml`):
- Retention period: 7 days (configurable)
- Storage: Local filesystem
- Compression: Enabled

**To modify retention:**
```yaml
limits_config:
  retention_period: 168h  # 7 days
```

## Monitoring Dashboards

### 1. Traefik Dashboard
- Request rate and response codes
- Error logs
- Performance metrics
- Security events

### 2. Application Logs Dashboard
- FastAPI and Go service logs
- Error tracking
- Request tracing
- Business events

## Troubleshooting

### Common Issues

**1. Grafana not accessible**
```bash
# Check if Grafana is running
docker-compose ps grafana

# Check logs
docker-compose logs grafana

# Restart if needed
docker-compose restart grafana
```

**2. No logs in Grafana**
```bash
# Check Promtail status
docker-compose ps promtail

# Check log files exist
ls -la ./traefik/logs/
ls -la ./fastapi/logs/
ls -la ./go/logs/

# Check Promtail configuration
docker-compose logs promtail
```

**3. Traefik not routing requests**
```bash
# Check Traefik configuration
docker-compose logs traefik

# Verify configuration files
docker exec traefik traefik version
```

**4. High memory usage**
```bash
# Check resource usage
docker stats

# Adjust Loki retention if needed
# Edit loki/loki-config.yml
```

### Log File Locations

| Service | Log Directory | Log Files |
|---------|---------------|-----------|
| Traefik | `./traefik/logs/` | `traefik.log`, `access.log` |
| FastAPI | `./fastapi/logs/` | `fastapi_structured.log`, `fastapi_app.log`, `fastapi_error.log` |
| Go | `./go/logs/` | `go_app.log`, `go_error.log`, `go_performance.log` |

### Performance Tuning

**For high-traffic environments:**

1. **Increase Loki limits:**
```yaml
limits_config:
  max_query_parallelism: 64
  max_query_series: 200000
```

2. **Adjust Promtail batch size:**
```yaml
clients:
  - url: http://loki:3100/loki/api/v1/push
    batchwait: 1s
    batchsize: 1024
```

3. **Enable compression:**
```yaml
compression: gzip
```

## Security Considerations

### Access Control
- Grafana: Change default password
- Traefik Dashboard: Consider authentication
- Loki API: Restrict access in production

### Data Privacy
- Logs may contain sensitive information
- Implement log sanitization if needed
- Consider data retention policies

### Network Security
- Use HTTPS in production
- Implement proper firewall rules
- Monitor for suspicious activity

## Backup and Recovery

### Backup Logs
```bash
# Backup Loki data
docker run --rm -v dsview-backend_loki-data:/data -v $(pwd):/backup alpine tar czf /backup/loki-backup.tar.gz -C /data .

# Backup Grafana data
docker run --rm -v dsview-backend_grafana-data:/data -v $(pwd):/backup alpine tar czf /backup/grafana-backup.tar.gz -C /data .
```

### Restore Logs
```bash
# Restore Loki data
docker run --rm -v dsview-backend_loki-data:/data -v $(pwd):/backup alpine tar xzf /backup/loki-backup.tar.gz -C /data

# Restore Grafana data
docker run --rm -v dsview-backend_grafana-data:/data -v $(pwd):/backup alpine tar xzf /backup/grafana-backup.tar.gz -C /data
```

## Upgrades

### Upgrading Components

**Loki/Grafana:**
```bash
# Update image versions in docker-compose.yml
# Then restart services
docker-compose pull
docker-compose up -d
```

**Traefik:**
```bash
# Update Traefik image
docker-compose pull traefik
docker-compose up -d traefik
```

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review service logs: `docker-compose logs <service>`
3. Check Grafana dashboards for insights
4. Consult the official documentation:
   - [Traefik Documentation](https://doc.traefik.io/traefik/)
   - [Loki Documentation](https://grafana.com/docs/loki/)
   - [Grafana Documentation](https://grafana.com/docs/grafana/)
