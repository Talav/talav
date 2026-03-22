#!/bin/bash
# Architecture validation script
# Checks import restrictions and layering rules from .cursor/architecture.yaml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

VIOLATIONS=0

echo "🔍 Validating architecture constraints..."
echo ""

# Function to check if a package imports forbidden patterns
check_imports() {
    local package_path=$1
    local layer=$2
    local forbidden_pattern=$3
    
    # Get all imports for the package
    imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' "$package_path" 2>/dev/null || echo "")
    
    # Check if any import matches the forbidden pattern
    for import in $imports; do
        if [[ $import == $forbidden_pattern ]]; then
            echo -e "${RED}✗ VIOLATION${NC}: $package_path"
            echo "  Layer: $layer"
            echo "  Forbidden import: $import"
            echo ""
            VIOLATIONS=$((VIOLATIONS + 1))
            return 1
        fi
    done
    
    return 0
}

# Check Component Layer (pkg/component/*)
echo "📦 Checking Component Layer..."
for component_dir in pkg/component/*; do
    if [ -d "$component_dir" ] && [ -f "$component_dir/go.mod" ]; then
        package=$(go list "./$component_dir" 2>/dev/null || echo "")
        if [ -n "$package" ]; then
            # Components cannot import pkg/fx/*
            check_imports "$package" "component" "github.com/talav/talav/pkg/fx/*" || true
            
            # Components cannot import pkg/module/*
            check_imports "$package" "component" "github.com/talav/talav/pkg/module/*" || true
        fi
    fi
done

# Check FX Layer (pkg/fx/*)
echo "🔧 Checking FX Layer..."
for fx_dir in pkg/fx/*; do
    if [ -d "$fx_dir" ] && [ -f "$fx_dir/go.mod" ]; then
        package=$(go list "./$fx_dir" 2>/dev/null || echo "")
        if [ -n "$package" ]; then
            # FX modules cannot import pkg/module/*
            check_imports "$package" "fx" "github.com/talav/talav/pkg/module/*" || true
        fi
    fi
done

# Check for circular dependencies
echo "🔄 Checking for circular dependencies..."
for dir in pkg/component/* pkg/fx/* pkg/module/*; do
    if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
        # Use go mod graph to detect cycles
        cd "$dir"
        if go mod graph 2>&1 | grep -q "cycle"; then
            echo -e "${RED}✗ CIRCULAR DEPENDENCY${NC}: $dir"
            VIOLATIONS=$((VIOLATIONS + 1))
        fi
        cd - > /dev/null
    fi
done

# Report results
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $VIOLATIONS -eq 0 ]; then
    echo -e "${GREEN}✓ Architecture validation passed${NC}"
    echo "  No layering violations detected"
    exit 0
else
    echo -e "${RED}✗ Architecture validation failed${NC}"
    echo "  Found $VIOLATIONS violation(s)"
    echo ""
    echo "Fix these violations before committing:"
    echo "  - Components (pkg/component/*) cannot import pkg/fx/* or pkg/module/*"
    echo "  - FX modules (pkg/fx/*) cannot import pkg/module/*"
    echo ""
    echo "See .cursor/architecture.yaml for detailed rules"
    exit 1
fi
