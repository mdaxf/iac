# Environment Variables Configuration

This document describes the environment variables used by IAC for configuration.

## Overview

IAC uses a combination of configuration files and environment variables. **Environment variables take precedence** over values in `apiconfig.json`, allowing you to keep sensitive credentials out of version control.

## Available Environment Variables

### API Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `IAC_API_KEY` | API key for IAC authentication | (from config file) | `your-api-key-here` |
| `OPENAI_KEY` | OpenAI API key for AI features | (from config file) | `sk-proj-xxxxxxxxxxxxx` |
| `OPENAI_MODEL` | OpenAI model to use | `gpt-4o` | `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo` |

## Configuration Priority

The system loads configuration in the following priority order (highest to lowest):

1. **Environment Variables** - Set in your shell or container environment (highest priority)
2. **Local Configuration File** (`apiconfig.local.json`) - Optional local overrides (git-ignored)
3. **Base Configuration File** (`apiconfig.json`) - Default values (tracked in git)
4. **Hardcoded Defaults** - Built-in fallback values (lowest priority)

### How Configuration Loading Works

1. **Base Config**: Load `apiconfig.json` (tracked in git, contains all controllers and default values)
2. **Local Override** (optional): If `apiconfig.local.json` exists, override only root-level settings:
   - `port` - HTTP server port
   - `timeout` - Request timeout
   - `apikey` - API key for authentication
   - `openaikey` - OpenAI API key
   - `openaimodel` - OpenAI model name
   - `portal` - Portal configuration
   - **Note**: Controllers and endpoints are NOT overridden, they always come from `apiconfig.json`
3. **Environment Variables**: Apply environment variable overrides (takes final precedence)

## Local Configuration Override

You can create an optional `apiconfig.local.json` file to override configuration settings without modifying the tracked `apiconfig.json` file. This file is in `.gitignore` and won't be committed.

### Creating apiconfig.local.json

Copy the example and customize:

```bash
cp apiconfig.local.json.example apiconfig.local.json
# Edit with your local values
```

Example `apiconfig.local.json` (only specify fields you want to override):

```json
{
  "port": 8081,
  "apikey": "my-local-api-key",
  "openaikey": "sk-local-xxxxxxxxxxxxx",
  "openaimodel": "gpt-4-turbo"
}
```

**Important Notes:**
- This file is **optional** - IAC works fine without it
- Only root-level fields are overridden (port, timeout, apikey, openaikey, openaimodel, portal)
- Controllers and endpoints always come from `apiconfig.json` (not overridden)
- Environment variables still take precedence over `apiconfig.local.json`
- The file is git-ignored, safe for local secrets

## Usage Examples

### Local Development

Set environment variables in your shell:

```bash
# Linux/Mac
export IAC_API_KEY="your-api-key"
export OPENAI_KEY="sk-proj-xxxxxxxxxxxxx"
export OPENAI_MODEL="gpt-4o"

# Start the application
./iac
```

```powershell
# Windows PowerShell
$env:IAC_API_KEY="your-api-key"
$env:OPENAI_KEY="sk-proj-xxxxxxxxxxxxx"
$env:OPENAI_MODEL="gpt-4o"

# Start the application
.\iac.exe
```

### Docker

Pass environment variables via docker run:

```bash
docker run -d \
  -e IAC_API_KEY="your-api-key" \
  -e OPENAI_KEY="sk-proj-xxxxxxxxxxxxx" \
  -e OPENAI_MODEL="gpt-4o" \
  -p 8080:8080 \
  iac:latest
```

Or use a `.env` file with docker-compose:

```yaml
# docker-compose.yml
version: '3.8'
services:
  iac:
    image: iac:latest
    ports:
      - "8080:8080"
    env_file:
      - .env
```

```bash
# .env file (add to .gitignore!)
IAC_API_KEY=your-api-key
OPENAI_KEY=sk-proj-xxxxxxxxxxxxx
OPENAI_MODEL=gpt-4o
```

### Kubernetes

Use Kubernetes secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: iac-secrets
type: Opaque
stringData:
  IAC_API_KEY: "your-api-key"
  OPENAI_KEY: "sk-proj-xxxxxxxxxxxxx"
  OPENAI_MODEL: "gpt-4o"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iac
spec:
  template:
    spec:
      containers:
      - name: iac
        image: iac:latest
        envFrom:
        - secretRef:
            name: iac-secrets
```

## When to Use Each Configuration Method

Choose the right configuration method for your use case:

### Use Environment Variables When:
- Deploying to production/staging/cloud environments
- Using Docker or Kubernetes
- Running in CI/CD pipelines
- Need different values per environment
- **Recommended for production deployments**

### Use apiconfig.local.json When:
- Local development with custom settings
- Need to override port or timeout locally
- Want to test with different OpenAI models
- Don't want to set environment variables in your shell
- **Good for local development convenience**

### Use apiconfig.json When:
- Defining default values for all environments
- Adding or modifying API endpoints/controllers
- Setting default port, timeout, model
- **This is the source of truth for API structure**

## Security Best Practices

1. **Never commit secrets to Git**
   - The `apiconfig.json` file now tracks with empty values for secrets
   - Use environment variables or `apiconfig.local.json` for all sensitive data
   - `apiconfig.local.json` is already in `.gitignore`

2. **Use different credentials per environment**
   - Development, staging, and production should use separate API keys
   - Rotate keys regularly

3. **Restrict access to environment variables**
   - Use secret management tools (AWS Secrets Manager, HashiCorp Vault, etc.)
   - Limit who can view production secrets

4. **Audit credential usage**
   - Monitor API key usage through your OpenAI dashboard
   - Set up alerts for unusual activity

## Verification

When IAC starts, it will log the configuration it loaded (with secrets masked):

```
loaded portal and api configuration
  - Port: 8080
  - API Key: your****here
  - OpenAI Key: sk-p****xxxx
  - OpenAI Model: gpt-4o
  - Controllers: 18
```

The `[not set]` indicator means the value is empty and should be configured.

## Troubleshooting

### "OpenAI API key not configured" error

**Solution**: Set the `OPENAI_KEY` environment variable:
```bash
export OPENAI_KEY="sk-proj-xxxxxxxxxxxxx"
```

### API authentication failing

**Solution**: Set the `IAC_API_KEY` environment variable:
```bash
export IAC_API_KEY="your-api-key"
```

### Wrong OpenAI model being used

**Solution**: Override with environment variable:
```bash
export OPENAI_MODEL="gpt-4-turbo"
```

## Migration from Old Configuration

If you previously had secrets in `apiconfig.json`:

1. Copy your secrets to a secure location
2. Pull the latest code (apiconfig.json now has empty values)
3. Set environment variables with your actual secrets
4. Restart the application
5. Verify the configuration is loaded correctly from the logs

## Additional Resources

- [OpenAI API Keys](https://platform.openai.com/api-keys)
- [Docker Environment Variables](https://docs.docker.com/compose/environment-variables/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
