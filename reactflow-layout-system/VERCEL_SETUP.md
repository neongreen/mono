# Vercel Deployment Setup

This guide explains how to set up automatic deployments to Vercel for the ReactFlow Layout System showcase.

## Prerequisites

- A Vercel account (free tier works fine)
- GitHub repository access (you already have this)
- The repository should already be configured (vercel.json files are in place)

## Setup Steps

### Option 1: Automatic Setup (Recommended)

1. Go to [vercel.com](https://vercel.com) and sign in with your GitHub account

2. Click "Add New Project"

3. Select the `neongreen/mono` repository

4. Vercel will automatically detect the configuration from the root `vercel.json` file

5. Click "Deploy"

That's it! Vercel will:
- Build the project automatically
- Deploy on every push to the main branch
- Create preview deployments for pull requests

### Option 2: Manual Configuration

If automatic detection doesn't work, use these settings:

**Framework Preset**: Other

**Build Command**:
```bash
cd reactflow-layout-system && pnpm install && pnpm build
```

**Output Directory**:
```
reactflow-layout-system/dist
```

**Install Command**:
```bash
pnpm install -g pnpm
```

**Root Directory**: Leave empty (use monorepo root)

## Vercel Configuration Files

Two configuration files are used:

### Root vercel.json
Located at `/vercel.json`, configures the deployment for the entire monorepo:
```json
{
  "buildCommand": "cd reactflow-layout-system && pnpm install && pnpm build",
  "outputDirectory": "reactflow-layout-system/dist",
  "installCommand": "pnpm install -g pnpm",
  "framework": null,
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

### Project vercel.json
Located at `/reactflow-layout-system/vercel.json`, provides project-specific configuration:
```json
{
  "buildCommand": "pnpm build",
  "outputDirectory": "dist",
  "installCommand": "pnpm install",
  "framework": "vite",
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

## How It Works

### Automatic Deployments

- **Main Branch**: Every push to main triggers a production deployment
- **Pull Requests**: Every PR gets a preview deployment with a unique URL
- **Branches**: You can configure additional branch deployments if needed

### Build Process

1. Vercel clones the repository
2. Installs PNPM globally
3. Changes to the `reactflow-layout-system` directory
4. Runs `pnpm install` to get dependencies
5. Runs `pnpm build` to create the production build
6. Deploys the contents of the `dist` directory

### Client-Side Routing

The `rewrites` configuration ensures that:
- All routes (`/`, `/example1`, `/example2`, `/example3`) work correctly
- Direct navigation to any route serves the React app
- The React Router handles the routing client-side

## Preview Deployments for Pull Requests

When you create a pull request:
1. Vercel automatically builds and deploys a preview
2. A unique URL is generated (e.g., `project-name-pr123.vercel.app`)
3. The Vercel bot comments on the PR with the preview link
4. Every push to the PR updates the preview

This allows you to:
- Test changes before merging
- Share live previews with reviewers
- Ensure the build works in production environment

## Environment Variables

Currently, no environment variables are needed. If you add any in the future:

1. Go to your Vercel project settings
2. Navigate to "Environment Variables"
3. Add variables for each environment (Production, Preview, Development)

## Custom Domain (Optional)

To use a custom domain:

1. Go to your Vercel project
2. Navigate to Settings → Domains
3. Add your domain
4. Follow the DNS configuration instructions
5. Vercel automatically handles SSL certificates

## Troubleshooting

### Build Fails

If the build fails, check:
- The build command is correct
- All dependencies are in package.json
- The output directory path is correct
- No TypeScript errors exist

### Routes Don't Work

If routes return 404:
- Ensure the `rewrites` configuration is in place
- Check that `vercel.json` is in the root directory
- Verify the SPA is configured correctly

### Preview Deployments Not Working

If PR previews aren't created:
- Check the Vercel GitHub app has repository access
- Ensure the branch protection rules allow the Vercel bot
- Check the Vercel project settings for PR deployment configuration

## Useful Commands

Test locally before deploying:

```bash
# Install dependencies
pnpm install

# Development server
pnpm dev

# Production build
pnpm build

# Preview production build locally
pnpm preview
```

## Resources

- [Vercel Documentation](https://vercel.com/docs)
- [Vercel CLI](https://vercel.com/docs/cli)
- [Monorepo Guide](https://vercel.com/docs/concepts/deployments/configure-projects)
- [PNPM with Vercel](https://vercel.com/docs/concepts/deployments/build-step#pnpm)

## Support

If you encounter issues:
1. Check the Vercel build logs for detailed error messages
2. Refer to the [Vercel documentation](https://vercel.com/docs)
3. Contact Vercel support through their dashboard
