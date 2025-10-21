# Docker Hub GitHub Actions Setup

This document explains how to configure GitHub Actions to automatically build and push multi-architecture Docker images to Docker Hub.

## Prerequisites

1. A Docker Hub account (username: `scottwalter`)
2. Admin access to this GitHub repository

## Step 1: Create a Docker Hub Access Token

1. Log in to [Docker Hub](https://hub.docker.com/)
2. Click on your username in the top right corner
3. Select **Account Settings**
4. Navigate to **Security** → **Access Tokens** (or **Personal Access Tokens**)
5. Click **New Access Token** or **Generate New Token**
6. Give it a description (e.g., "GitHub Actions - AxeOS Dashboard")
7. Set permissions to **Read, Write, Delete** (or just **Read & Write** if that's the only option)
8. Click **Generate**
9. **IMPORTANT**: Copy the token immediately - you won't be able to see it again!

## Step 2: Add the Token to GitHub Secrets

1. Go to your GitHub repository: `https://github.com/scottwalter/axeos-dashboard`
2. Click on **Settings** (repository settings, not your account)
3. In the left sidebar, expand **Secrets and variables** → click **Actions**
4. Click **New repository secret**
5. Enter the following:
   - **Name**: `DOCKER_HUB_TOKEN`
   - **Secret**: Paste the Docker Hub access token you copied in Step 1
6. Click **Add secret**

## Step 3: Verify the Workflow

The workflow is configured in [`.github/workflows/docker-build.yml`](workflows/docker-build.yml) and will:

- Trigger automatically on every push to the `main` branch
- Can also be triggered manually via GitHub Actions UI
- Build Docker images for both AMD64 and ARM64 architectures
- Push images to Docker Hub as `scottwalter/axeos-dashboard`

### Image Tags

The workflow creates multiple tags:
- `latest` - Always points to the most recent build from main
- `main` - Branch-based tag
- `main-<sha>` - Commit SHA-based tag for version pinning

## Step 4: Test the Workflow

1. Commit and push the workflow file to the `main` branch
2. Go to **Actions** tab in your GitHub repository
3. You should see the workflow running
4. Once complete, verify the images on Docker Hub: https://hub.docker.com/r/scottwalter/axeos-dashboard

## Using the Multi-Arch Images

### Pull the Image

```bash
# Pull latest version (works on both AMD64 and ARM64)
docker pull scottwalter/axeos-dashboard:latest

# Pull specific version
docker pull scottwalter/axeos-dashboard:main-a1b2c3d
```

### Run on Ubuntu Server 25.10 (AMD64)

```bash
docker run -d \
  --name axeos-dashboard \
  -p 3000:3000 \
  -v /path/to/config:/app/config \
  scottwalter/axeos-dashboard:latest
```

### Run on macOS Sequoia (ARM64 - Apple Silicon)

```bash
docker run -d \
  --name axeos-dashboard \
  -p 3000:3000 \
  -v /path/to/config:/app/config \
  scottwalter/axeos-dashboard:latest
```

Docker will automatically pull the correct architecture for your platform.

## Troubleshooting

### Workflow Fails with "unauthorized: authentication required"

- Verify the `DOCKER_HUB_TOKEN` secret is set correctly
- Regenerate the Docker Hub access token and update the secret
- Ensure the token has **Read & Write** permissions

### Workflow Fails During Build

- Check the Actions logs for specific errors
- Verify the Dockerfile builds locally: `docker build -t test .`
- Ensure all dependencies are available

### Images Don't Appear on Docker Hub

- Check that your Docker Hub username is `scottwalter` in the workflow
- Verify the repository `scottwalter/axeos-dashboard` exists or can be auto-created
- Check Docker Hub repository settings (should allow public pushes)

## Manual Workflow Trigger

You can manually trigger a build without pushing code:

1. Go to **Actions** tab in GitHub
2. Select **Build and Push Multi-Arch Docker Image** workflow
3. Click **Run workflow**
4. Select the `main` branch
5. Click **Run workflow**

## Advanced Configuration

### Adding Version Tags

To add semantic versioning, tag your commits:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow will automatically create Docker tags:
- `1.0.0`
- `1.0`
- `latest`

### Customizing Architectures

Edit `.github/workflows/docker-build.yml` line 47 to add/remove platforms:

```yaml
platforms: linux/amd64,linux/arm64,linux/arm/v7
```

## Security Notes

- Never commit Docker Hub tokens to the repository
- Rotate access tokens periodically
- Use the minimum required permissions for tokens
- Review Docker Hub access logs regularly
