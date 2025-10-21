# GitHub Actions Docker Build Troubleshooting

## What's the exact error you're seeing?

Please check which error message you see and follow the corresponding solution:

---

## Error 1: "Password required"

### This means the secret is not being passed to the workflow.

**Solution - Try these in order:**

### A. Verify the secret exists

1. Go to: `https://github.com/scottwalter/axeos-dashboard/settings/secrets/actions`
2. Look for `DOCKER_HUB_TOKEN` in the list
3. If it's NOT there or has a different name, you need to add/fix it

### B. Check the repository

1. Make sure you're in the correct repository: `scottwalter/axeos-dashboard`
2. If this is a fork, secrets from the original repo won't transfer
3. Each repository needs its own secrets

### C. Delete and recreate the secret

1. In GitHub: Settings → Secrets and variables → Actions
2. Find `DOCKER_HUB_TOKEN` and click "Remove"
3. Create new Docker Hub token: https://hub.docker.com/settings/security
4. Add new secret with name: `DOCKER_HUB_TOKEN`

### D. Try the alternative workflow

Instead of the main workflow, try the alternative:

1. Go to Actions tab
2. Select "Build and Push Multi-Arch Docker Image (Alternative)"
3. Click "Run workflow"

This uses a different authentication method that might work better.

---

## Error 2: "unauthorized: authentication required"

### This means the token is invalid or doesn't have permissions.

**Solution:**

1. Generate a NEW Docker Hub token:
   - Go to: https://hub.docker.com/settings/security
   - Create new token
   - Set permissions: **Read, Write, Delete**
   - Copy the token

2. Update the GitHub secret:
   - Delete old `DOCKER_HUB_TOKEN`
   - Add new one with the fresh token

3. Verify Docker Hub username is `scottwalter`

---

## Error 3: "DOCKER_HUB_TOKEN is NOT set or is empty"

### The secret exists but is empty or malformed.

**Solution:**

1. Regenerate Docker Hub token (it might have been revoked)
2. When adding to GitHub, make sure you paste the ENTIRE token
   - Token should start with `dckr_pat_`
   - Should be quite long (50+ characters)
   - No extra spaces before or after

---

## Error 4: Build fails during "Log in to Docker Hub" step

### Authentication is failing at Docker login.

**Try this:**

1. Use the alternative workflow file I created
2. It uses `docker login` command instead of the action
3. Should give more detailed error messages

---

## Still Not Working? Here's the nuclear option:

### Create a completely fresh setup:

```bash
# 1. Delete the GitHub secret
# Go to: Settings → Secrets → Delete DOCKER_HUB_TOKEN

# 2. Revoke old Docker Hub token
# Go to: https://hub.docker.com/settings/security
# Delete the old token

# 3. Create fresh Docker Hub token
# Click "New Access Token"
# Description: "GitHub Actions Fresh"
# Permissions: Read, Write, Delete
# Generate and COPY THE TOKEN

# 4. Add to GitHub
# Settings → Secrets → New repository secret
# Name: DOCKER_HUB_TOKEN
# Value: [paste fresh token]

# 5. Test with alternative workflow
# Actions → "Build and Push (Alternative)" → Run workflow
```

---

## Test Locally First

Before using GitHub Actions, test if your Docker Hub credentials work:

```bash
# Test login locally
echo "YOUR_DOCKER_HUB_TOKEN" | docker login -u scottwalter --password-stdin

# If this works, the token is valid
# If this fails, generate a new token
```

---

## Common Mistakes

❌ **Wrong secret name**: Must be exactly `DOCKER_HUB_TOKEN` (case-sensitive)
❌ **Adding to wrong place**: Must be in Repository secrets, not Organization or Environment
❌ **Incomplete token**: Token must be complete, starts with `dckr_pat_`
❌ **Token expired**: Docker Hub tokens don't expire but can be revoked
❌ **Wrong permissions**: Token needs Read & Write (or Read, Write, Delete)
❌ **Forked repo**: Forks don't inherit secrets, you must add them manually

---

## How to Share Error Details

If still stuck, share these details:

1. **Exact error message** from GitHub Actions log
2. **Which step failed** (name of the step)
3. **Screenshot** of the error (if possible)
4. Confirm: "I can see DOCKER_HUB_TOKEN in my repository secrets list: YES/NO"
5. Confirm: "My Docker Hub username is: scottwalter"

---

## Quick Test

Run this to verify everything:

1. Docker Hub token exists: https://hub.docker.com/settings/security
2. GitHub secret exists: https://github.com/scottwalter/axeos-dashboard/settings/secrets/actions
3. Try alternative workflow: Actions → "Alternative" → Run workflow
4. Check logs carefully for the actual error message
