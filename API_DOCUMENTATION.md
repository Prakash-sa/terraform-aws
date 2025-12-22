# AI-Assisted Incident Automation API Documentation

## Overview

The AI-Assisted Incident Automation API provides endpoints for creating, managing, and analyzing incidents using AI-powered analysis from OpenAI or Anthropic Claude. This system helps automate incident triage, root cause analysis, and documentation.

## Features

- **Automated Incident Creation**: Create incidents from alerts, logs, or manual input
- **AI-Powered Analysis**: Analyze incidents using GPT-4 or Claude to identify key findings and potential causes
- **Severity Classification**: Automatically classify incident severity based on description and context
- **Log Summarization**: Summarize large log files to extract meaningful insights
- **RCA Generation**: Automatically generate comprehensive Root Cause Analysis documents
- **Recommended Actions**: Get AI-suggested remediation steps and preventive measures

## Authentication

Currently, the API does not require authentication. In production, you should implement proper authentication and authorization.

## Base URL

```
http://localhost:8080
```

## Endpoints

### 1. Create Incident

Creates a new incident in the system.

**Endpoint:** `POST /api/v1/incidents`

**Request Body:**
```json
{
  "title": "High CPU Usage Alert",
  "description": "CPU usage exceeded 90% on production servers",
  "severity": "high",
  "source": "prometheus",
  "alert_data": "cpu_usage{instance=\"prod-1\"} > 90",
  "logs": [
    "2024-12-22 10:15:23 ERROR: CPU at 92%",
    "2024-12-22 10:15:45 WARNING: Memory pressure detected"
  ]
}
```

**Parameters:**
- `title` (string, required): Brief title of the incident
- `description` (string, required): Detailed description of the incident
- `severity` (string, optional): Severity level - "critical", "high", "medium", or "low". If not provided, AI will classify it automatically
- `source` (string, required): Source of the incident (e.g., "prometheus", "cloudwatch", "manual")
- `alert_data` (string, optional): Raw alert data
- `logs` (array of strings, optional): Associated log entries

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "title": "High CPU Usage Alert",
  "description": "CPU usage exceeded 90% on production servers",
  "severity": "high",
  "status": "open",
  "source": "prometheus",
  "alert_data": "cpu_usage{instance=\"prod-1\"} > 90",
  "logs": [...],
  "created_at": "2024-12-22T10:15:23Z",
  "updated_at": "2024-12-22T10:15:23Z"
}
```

### 2. List Incidents

Retrieves all incidents in the system.

**Endpoint:** `GET /api/v1/incidents`

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "title": "High CPU Usage Alert",
    "severity": "high",
    "status": "open",
    "created_at": "2024-12-22T10:15:23Z",
    ...
  }
]
```

### 3. Get Incident

Retrieves details of a specific incident.

**Endpoint:** `GET /api/v1/incidents/{id}`

**Parameters:**
- `id` (string, required): Incident ID

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "title": "High CPU Usage Alert",
  "description": "CPU usage exceeded 90% on production servers",
  "severity": "high",
  "status": "open",
  "source": "prometheus",
  "created_at": "2024-12-22T10:15:23Z",
  "updated_at": "2024-12-22T10:15:23Z"
}
```

### 4. Analyze Incident

Analyzes an incident using AI to provide insights, severity classification, and recommendations.

**Endpoint:** `POST /api/v1/incidents/{id}/analyze`

**Parameters:**
- `id` (string, required): Incident ID

**Response:** `200 OK`
```json
{
  "incident_id": "uuid",
  "summary": "CPU usage spike detected on production server prod-1, reaching 92% utilization...",
  "suggested_severity": "high",
  "key_findings": [
    "CPU usage exceeded 90% threshold",
    "Memory pressure detected concurrently",
    "Issue occurred on production instance prod-1"
  ],
  "potential_causes": [
    "Runaway process or memory leak",
    "Sudden traffic spike",
    "Resource-intensive batch job",
    "Insufficient resource allocation"
  ],
  "recommended_actions": [
    "Identify and investigate top CPU-consuming processes",
    "Check for memory leaks in application",
    "Review recent deployments or configuration changes",
    "Consider scaling up resources if needed",
    "Set up alerts for early warning of similar issues"
  ],
  "generated_at": "2024-12-22T10:16:00Z",
  "provider": "openai"
}
```

### 5. Get Analysis

Retrieves the AI analysis for an incident (if already performed).

**Endpoint:** `GET /api/v1/incidents/{id}/analysis`

**Parameters:**
- `id` (string, required): Incident ID

**Response:** `200 OK` - Same format as Analyze Incident response

**Error Response:** `404 Not Found` if no analysis exists

### 6. Generate RCA Document

Generates a comprehensive Root Cause Analysis document for an incident.

**Endpoint:** `POST /api/v1/incidents/{id}/rca/generate`

**Parameters:**
- `id` (string, required): Incident ID

**Response:** `200 OK`
```json
{
  "incident_id": "uuid",
  "summary": "This RCA documents a high CPU usage incident that occurred on production server prod-1...",
  "timeline": [
    "2024-12-22 10:15:23 - CPU usage alert triggered at 92%",
    "2024-12-22 10:15:45 - Memory pressure detected",
    "2024-12-22 10:20:00 - Investigation started",
    "2024-12-22 10:30:00 - Root cause identified",
    "2024-12-22 10:45:00 - Issue resolved"
  ],
  "root_cause": "A background batch job was consuming excessive CPU resources due to inefficient database queries...",
  "impact_analysis": "The incident affected response times for 15% of users during a 30-minute window...",
  "resolution": "The batch job was terminated and rescheduled to run during off-peak hours...",
  "preventive_measures": [
    "Optimize database queries in the batch job",
    "Implement resource limits for batch processes",
    "Schedule resource-intensive jobs during off-peak hours",
    "Set up proactive monitoring for resource usage"
  ],
  "lessons_learned": [
    "Need better resource management for batch jobs",
    "Importance of query optimization",
    "Value of proactive monitoring"
  ],
  "generated_at": "2024-12-22T11:00:00Z",
  "generated_by": "ai"
}
```

### 7. Get RCA Document

Retrieves the RCA document for an incident (if already generated).

**Endpoint:** `GET /api/v1/incidents/{id}/rca`

**Parameters:**
- `id` (string, required): Incident ID

**Response:** `200 OK` - Same format as Generate RCA response

**Error Response:** `404 Not Found` if no RCA exists

### 8. Summarize Logs

Summarizes a collection of log entries using AI.

**Endpoint:** `POST /api/v1/logs/summarize`

**Request Body:**
```json
{
  "logs": [
    "2024-12-22 10:15:23 ERROR: Connection timeout to database",
    "2024-12-22 10:15:24 WARNING: Retry attempt 1",
    "2024-12-22 10:15:25 WARNING: Retry attempt 2",
    "2024-12-22 10:15:26 ERROR: Max retries exceeded",
    "2024-12-22 10:15:27 CRITICAL: Service unavailable"
  ]
}
```

**Response:** `200 OK`
```json
{
  "summary": "The logs indicate a database connection failure with multiple retry attempts. The connection timed out, three retries were attempted, and ultimately the service became unavailable. This suggests a network or database server issue.",
  "log_count": 5,
  "timestamp": "2024-12-22T10:16:00Z"
}
```

## Error Responses

All endpoints may return the following error responses:

**400 Bad Request**
```json
{
  "message": "Invalid request payload",
  "timestamp": "2024-12-22T10:15:23Z"
}
```

**404 Not Found**
```json
{
  "message": "incident not found: {id}",
  "timestamp": "2024-12-22T10:15:23Z"
}
```

**500 Internal Server Error**
```json
{
  "message": "Failed to analyze incident",
  "timestamp": "2024-12-22T10:15:23Z"
}
```

## Configuration

### Environment Variables

- `AI_PROVIDER`: Set to "openai" or "anthropic" (default: "openai")
- `OPENAI_API_KEY`: Your OpenAI API key (required if using OpenAI)
- `OPENAI_MODEL`: OpenAI model to use (default: "gpt-4")
- `ANTHROPIC_API_KEY`: Your Anthropic API key (required if using Claude)
- `ANTHROPIC_MODEL`: Anthropic model to use (default: "claude-3-5-sonnet-20241022")

### Note on AI Features

If AI API keys are not configured, the service will still accept incident creation and management requests, but AI-powered features (analysis, RCA generation, log summarization) will return errors indicating that the AI client is not configured.

## Example Workflows

### Complete Incident Response Workflow

1. **Create Incident** from alert or manual report
2. **Analyze Incident** with AI to get initial insights
3. **Review Analysis** to understand severity and potential causes
4. **Investigate and Resolve** based on recommended actions
5. **Generate RCA** document for post-mortem
6. **Update Status** to resolved/closed

### Log Analysis Workflow

1. **Collect Logs** from affected services
2. **Summarize Logs** using AI to extract key information
3. **Create Incident** with summarized information
4. **Analyze Incident** for deeper insights
5. **Take Action** based on recommendations

## Best Practices

1. **Provide Context**: Include as much relevant information (logs, alerts, metrics) when creating incidents
2. **Review AI Suggestions**: AI analysis is a tool to assist, not replace human judgment
3. **Update Incidents**: Keep incident status updated as you investigate and resolve issues
4. **Use RCA for Learning**: Generated RCA documents help build organizational knowledge
5. **Monitor Patterns**: Track incidents over time to identify recurring issues

## Limitations

- AI analysis quality depends on the information provided
- API rate limits apply based on your OpenAI/Anthropic subscription
- Large log files should be pre-filtered before summarization
- In-memory storage - incidents are lost on service restart (implement persistent storage for production)
