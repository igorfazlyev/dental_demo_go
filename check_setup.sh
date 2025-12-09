#!/bin/bash

echo "🔍 Checking DentalAI Platform Setup..."
echo ""

# Check Go installation
echo "1. Checking Go installation..."
if command -v go &> /dev/null; then
    echo "   ✅ Go is installed: $(go version)"
else
    echo "   ❌ Go is NOT installed"
    echo "   Please install from: https://golang.org/dl/"
    exit 1
fi

# Check directory structure
echo ""
echo "2. Checking directory structure..."

if [ -f "main.go" ]; then
    echo "   ✅ main.go found"
else
    echo "   ❌ main.go NOT found"
    exit 1
fi

if [ -d "templates" ]; then
    template_count=$(ls templates/*.html 2>/dev/null | wc -l)
    echo "   ✅ templates/ directory found ($template_count HTML files)"
    if [ $template_count -lt 14 ]; then
        echo "   ⚠️  Expected 14 templates, found $template_count"
    fi
else
    echo "   ❌ templates/ directory NOT found"
    exit 1
fi

if [ -d "static/css" ]; then
    echo "   ✅ static/css/ directory found"
    if [ -f "static/css/style.css" ]; then
        echo "   ✅ style.css found"
    else
        echo "   ❌ style.css NOT found"
    fi
else
    echo "   ❌ static/css/ directory NOT found"
    exit 1
fi

echo ""
echo "3. Listing template files..."
ls -1 templates/*.html 2>/dev/null || echo "   ❌ No template files found"

echo ""
echo "✅ Setup looks good!"
echo ""
echo "🚀 To run the application:"
echo "   $ go run main.go"
echo ""
echo "🌐 Then open: http://localhost:8080"
