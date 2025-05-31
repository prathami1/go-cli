# Changelog

## [v2.0.0] - 2024-12-XX - Bloomberg Enterprise Authentication

### 🎉 Major New Features

#### **Automatic Authentication for Bloomberg Employees**
- **No more manual authentication steps!** CloudDeploy now automatically handles authentication for Bloomberg employees
- **Smart environment detection** - Automatically detects Bloomberg corporate environment
- **Enterprise SSO optimization** - Uses the best authentication methods for Bloomberg's SSO setup

### ✨ New Features

#### **Bloomberg Environment Detection**
- Automatically detects Bloomberg environment by checking:
  - Hostname containing "bloomberg"
  - Username containing "corp"
  - Domain containing "bloomberg" 
  - `BLOOMBERG_ENV` environment variable

#### **Azure Authentication (Optimized for Bloomberg)**
- **Device code flow** for better enterprise SSO compatibility
- **Automatic tenant handling** with fallback options
- **B-Unit integration prompts** to prepare users for 2FA
- Commands used: `az login --use-device-code --tenant common`

#### **AWS Authentication (Enterprise SSO)**
- **AWS SSO first** - Tries `aws sso login` before standard methods
- **Automatic fallback** to `aws configure` if SSO fails
- **Bloomberg enterprise setup aware**

#### **GCP Authentication (Enterprise-friendly)**
- **Enhanced login options** with `--enable-gdrive-access --brief`
- **CORP credential integration**
- **Enterprise authentication flow**

### 🔧 Improvements

#### **Better Error Messages**
- Instead of: `"Please run 'az login'"`
- Now shows: `"The authentication process was attempted automatically. If you're a Bloomberg employee, make sure your CORP credentials and B-Unit are ready"`

#### **User Experience**
- **Proactive 2FA prompts** - Warns users to have B-Unit ready
- **Real-time feedback** during authentication process
- **Clear Bloomberg-specific guidance**
- **Automatic retry mechanisms** with fallback options

#### **Non-Bloomberg Compatibility**
- **Standard authentication** still works for non-Bloomberg users
- **Automatic detection** - No configuration needed
- **Manual override** available with environment variables

### 📁 New Files
- `AUTHENTICATION.md` - Comprehensive authentication documentation
- `test_auth.sh` - Bloomberg environment detection test script

### 📝 Updated Files
- `internal/providers/providers.go` - Complete authentication overhaul
- `cmd/init.go` - Updated error messaging
- `cmd/deploy.go` - Updated error messaging  
- `README.md` - Enhanced troubleshooting section

### 🧪 Testing
- Added Bloomberg environment detection test script
- Verified build compatibility
- Maintained backward compatibility for non-Bloomberg users

### 💡 Usage Examples

#### Before (Manual)
```bash
clouddeploy init
# ERROR: not authenticated with Azure. Please run 'az login'
az login
clouddeploy init
```

#### After (Automatic)
```bash
clouddeploy init
# 🔐 Not authenticated with Azure. Initiating automatic login...
# 💼 Detected Bloomberg environment...
# [Authentication happens automatically]
# ✅ Bloomberg Azure authentication successful!
```

### 🔗 Related
- Based on Bloomberg NEXI documentation for enterprise SSO
- Integrates with Bloomberg's CORP credential system
- Supports B-Unit and B-Unit phone app for 2FA 