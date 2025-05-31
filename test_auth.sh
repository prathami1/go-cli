#!/bin/bash

echo "=== Testing Bloomberg Environment Detection ==="
echo ""

echo "Current environment:"
echo "- Hostname: $(hostname)"
echo "- User: $USER"
echo "- Domain: $USERDOMAIN"
echo "- Bloomberg ENV: $BLOOMBERG_ENV"
echo ""

echo "Testing with Bloomberg environment variable..."
export BLOOMBERG_ENV=true
echo "✅ Set BLOOMBERG_ENV=true"
echo ""

echo "You can now test authentication by running:"
echo "  ./clouddeploy init"
echo ""
echo "The tool should detect Bloomberg environment and use enterprise authentication."
echo ""

echo "To test non-Bloomberg mode, run:"
echo "  unset BLOOMBERG_ENV"
echo "  ./clouddeploy init" 