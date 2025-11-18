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

1. **Environment Variables** - Set in your shell or container environment
2. **Configuration File** (`apiconfig.json`) - Default values
3. **Hardcoded Defaults** - Built-in fallback values

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

## Security Best Practices

1. **Never commit secrets to Git**
   - The `apiconfig.json` file now tracks with empty values
   - Use environment variables for all sensitive data
   - Add `apiconfig.local.json` to `.gitignore` if you need a local override

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
