# GitHub Secret Setup Checklist

Use this checklist to ensure the Docker Hub secret is configured correctly.

## Pre-Flight Checklist

- [ ] I am logged into Docker Hub as `scottwalter`
- [ ] I have admin access to the GitHub repository
- [ ] I am looking at the **repository** settings (not my user settings)

## Docker Hub Token Creation

- [ ] Went to https://hub.docker.com/
- [ ] Account Settings → Security → Access Tokens
- [ ] Clicked "New Access Token"
- [ ] Set description to: `GitHub Actions - AxeOS Dashboard`
- [ ] Set permissions to: **Read & Write**
- [ ] Clicked "Generate"
- [ ] **COPIED THE TOKEN** (it starts with `dckr_pat_`)
- [ ] Token is saved somewhere safe (you can't see it again!)

## GitHub Secret Configuration

- [ ] Went to https://github.com/scottwalter/axeos-dashboard/settings/secrets/actions
- [ ] Clicked "New repository secret" button
- [ ] Entered name as: `DOCKER_HUB_TOKEN` (exact spelling, all caps)
- [ ] Pasted the Docker Hub token into the "Secret" field
- [ ] Clicked "Add secret"
- [ ] See `DOCKER_HUB_TOKEN` in the secrets list

## Verification

- [ ] Secret name is EXACTLY: `DOCKER_HUB_TOKEN`
- [ ] Secret is under "Repository secrets" (not Environment or Organization)
- [ ] No typos in the secret name
- [ ] Token was copied completely (no extra spaces)

## Test the Workflow

Option 1: Manual trigger
- [ ] Go to Actions tab
- [ ] Click "Build and Push Multi-Arch Docker Image"
- [ ] Click "Run workflow" → "Run workflow"
- [ ] Watch the logs

Option 2: Push to main branch
- [ ] Make a small change (edit README)
- [ ] `git commit -am "Test build"`
- [ ] `git push origin main`
- [ ] Go to Actions tab and watch

## Expected Results

When the workflow runs, you should see:

✅ Step "Check Docker Hub credentials" shows:
```
Docker Hub username: scottwalter
Docker Hub token is set: YES
```

✅ Step "Log in to Docker Hub" shows:
```
Login Succeeded
```

✅ Step "Build and push Docker image" completes successfully

✅ Images appear on Docker Hub: https://hub.docker.com/r/scottwalter/axeos-dashboard

## If Something Goes Wrong

### See "Password required" error?
→ Secret is not set. Go back to "GitHub Secret Configuration" section above.

### See "unauthorized" error?
→ Token is invalid. Generate a new Docker Hub token and update the secret.

### See "DOCKER_HUB_TOKEN secret is not set!" error?
→ Secret name is wrong or missing. Check spelling: `DOCKER_HUB_TOKEN`

### Still stuck?
→ Delete the secret and start over from "Docker Hub Token Creation"
