# 🔐 Automatic Authentication for Bloomberg Employees

CloudDeploy now includes **automatic authentication** specifically designed for Bloomberg's enterprise environment! This eliminates the need to manually run authentication commands.

## 🎯 What's New

Instead of seeing error messages like:
```
ERRO Authentication check failed: not authenticated with Azure. Please run 'az login'
```

CloudDeploy now **automatically initiates the login process** for you with Bloomberg-optimized settings.

## 🏢 Bloomberg Environment Detection

CloudDeploy automatically detects if you're in a Bloomberg environment by checking:
- Hostname containing "bloomberg"
- Username containing "corp" 
- Domain containing "bloomberg"
- `BLOOMBERG_ENV` environment variable

## ☁️ Cloud Provider Support

### Azure (Recommended for Bloomberg)
When you need to authenticate with Azure, CloudDeploy will:

1. **Detect Bloomberg environment** 
2. **Use device code flow** (better for enterprise SSO)
3. **Prompt for B-Unit readiness**
4. **Launch**: `az login --use-device-code --tenant common`
5. **Fallback** to: `az login --allow-no-subscriptions` if needed

**Example flow:**
```
🔐 Not authenticated with Azure. Initiating automatic login...
🚀 Starting Azure authentication flow...
💼 Detected Bloomberg environment. Using enterprise SSO authentication...
🔑 Starting Bloomberg Azure SSO authentication...
📱 Please have your B-Unit or B-Unit phone app ready for 2FA
🖥️  Using device code flow for enterprise compatibility...
```

### AWS (Enterprise SSO)
For AWS authentication, CloudDeploy will:

1. **Try AWS SSO first**: `aws sso login`
2. **Fallback to standard**: `aws configure` if SSO fails

### Google Cloud Platform
For GCP authentication, CloudDeploy will:

1. **Use enterprise-friendly options**: `gcloud auth login --enable-gdrive-access --brief`
2. **Integrate with CORP credentials**

## 📱 What You Need

As a Bloomberg employee, make sure you have:

- **CORP credentials** (same username/password as your PC login)
- **B-Unit device** or **B-Unit phone app** for 2FA authentication
- **Access to the cloud provider** through Bloomberg's enterprise setup

## 🔄 How It Works

### Before (Manual Process)
```bash
clouddeploy init
> ERRO Authentication check failed: not authenticated with Azure. Please run 'az login'
> INFO Please authenticate with your cloud provider and try again.
> FATA Initialization failed: not authenticated with Azure. Please run 'az login'

# You had to manually run:
az login
# Then run clouddeploy again
clouddeploy init
```

### After (Automatic Process)
```bash
clouddeploy init
> 🔐 Not authenticated with Azure. Initiating automatic login...
> 🚀 Starting Azure authentication flow...
> 💼 Detected Bloomberg environment. Using enterprise SSO authentication...
> 🔑 Starting Bloomberg Azure SSO authentication...
> 📱 Please have your B-Unit or B-Unit phone app ready for 2FA
> 🖥️  Using device code flow for enterprise compatibility...
> [Authentication flow opens automatically]
> ✅ Bloomberg Azure authentication successful!
> [Continues with clouddeploy init...]
```

## 🚨 Troubleshooting

### If Authentication Fails

The tool provides helpful error messages:
```
Authentication failed: [error details]. The authentication process was attempted automatically. If you're a Bloomberg employee, make sure your CORP credentials and B-Unit are ready.
```

### Manual Fallback

If automatic authentication doesn't work, you can still authenticate manually:
```bash
# Azure
az login

# AWS  
aws sso login
# or
aws configure

# GCP
gcloud auth login
```

### Environment Variables

You can force Bloomberg environment detection by setting:
```bash
export BLOOMBERG_ENV=true
```

## 🌟 Benefits

1. **Seamless Experience**: No more copying/pasting authentication commands
2. **Bloomberg-Optimized**: Uses the best authentication methods for Bloomberg's SSO setup
3. **Automatic Fallbacks**: If one method fails, it tries alternatives
4. **Better Error Messages**: Clear guidance on what to expect and prepare
5. **2FA Ready**: Prompts you to have your B-Unit ready before starting

## 🔧 Technical Details

The authentication system:
- Detects Bloomberg environment automatically
- Uses device code flow for Azure (better enterprise compatibility)
- Tries AWS SSO before standard configuration
- Handles GCP with enterprise-friendly options
- Provides real-time feedback and guidance
- Verifies authentication success before proceeding

This update makes CloudDeploy much more user-friendly for Bloomberg employees while maintaining compatibility with standard authentication methods for other users. 