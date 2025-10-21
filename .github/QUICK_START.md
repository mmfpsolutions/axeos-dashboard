# Quick Start: Setting Up GitHub Actions Docker Build

Follow these steps **in order** to fix the "Password required" error:

## Step 1: Create Docker Hub Access Token (Do This First!)

1. Go to https://hub.docker.com/ and log in
2. Click your **username** (top right) → **Account Settings**
3. Click **Security** in the left menu
4. Click **New Access Token** button
5. Fill in:
   - Description: `GitHub Actions - AxeOS Dashboard`
   - Access permissions: **Read & Write**
6. Click **Generate**
7. **COPY THE TOKEN NOW** - you won't see it again!
   - It looks like: `dckr_pat_1234567890abcdefghijklmnop`

## Step 2: Add Token to GitHub (Do This Second!)

1. Go to: https://github.com/scottwalter/axeos-dashboard/settings/secrets/actions
   - Or navigate: Your repo → Settings → Secrets and variables → Actions
2. Click **New repository secret** (green button)
3. Fill in **EXACTLY**:
   ```
   Name: DOCKER_HUB_TOKEN
   Secret: [paste the token from Step 1]
   ```
4. Click **Add secret**
5. **VERIFY**: You should see `DOCKER_HUB_TOKEN` in the secrets list

## Step 3: Test the Workflow

### Option A: Push a commit
```bash
git add .
git commit -m "Test GitHub Actions build"
git push origin main
```

### Option B: Trigger manually
1. Go to: https://github.com/scottwalter/axeos-dashboard/actions
2. Click **Build and Push Multi-Arch Docker Image**
3. Click **Run workflow** → **Run workflow**

## Step 4: Monitor the Build

1. Go to the Actions tab
2. Click on the running workflow
3. Watch the logs

### What You Should See:

✅ **Check Docker Hub credentials** - Should show:
```
Docker Hub username: scottwalter
Docker Hub token is set: YES
```

✅ **Log in to Docker Hub** - Should show:
```
Login Succeeded
```

❌ **If you see "Password required"** - The secret is not set correctly. Go back to Step 2.

## Common Issues

### Issue: "Password required"
**Solution**: The secret `DOCKER_HUB_TOKEN` is not set or has the wrong name
- Go to: https://github.com/scottwalter/axeos-dashboard/settings/secrets/actions
- Verify the secret name is EXACTLY `DOCKER_HUB_TOKEN` (all caps)
- Delete and recreate it if needed

### Issue: "ERROR: DOCKER_HUB_TOKEN secret is not set!"
**Solution**: Same as above - secret is missing or incorrectly named

### Issue: "unauthorized: authentication required"
**Solution**: The token value is wrong
- Regenerate a new Docker Hub token (Step 1)
- Update the GitHub secret with the new token (Step 2)

### Issue: Can't see "Settings" in GitHub
**Solution**: You need admin access to the repository
- Ask the repository owner to add the secret
- Or fork the repo and add secrets to your fork

## Success!

When the build succeeds, you'll see:
- All steps green in GitHub Actions
- New images on Docker Hub: https://hub.docker.com/r/scottwalter/axeos-dashboard

You can then pull the image:
```bash
docker pull scottwalter/axeos-dashboard:latest
```

## Need More Help?

See the full documentation: [DOCKER_HUB_SETUP.md](DOCKER_HUB_SETUP.md)
