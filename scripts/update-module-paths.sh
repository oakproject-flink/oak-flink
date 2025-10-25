#!/usr/bin/env bash
# update-module-paths.sh
# Updates all Go module paths when forking the repository
# Usage: ./scripts/update-module-paths.sh github.com/yourorg/yourrepo

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <new-module-path>"
    echo "Example: $0 github.com/mycompany/oak-flink"
    echo "Example: $0 gitlab.com/myorg/flink-platform"
    exit 1
fi

OLD_PATH="github.com/oakproject-flink/oak-flink"
NEW_PATH="$1"

echo "Updating module paths from $OLD_PATH to $NEW_PATH"
echo "=================================================="

# Find all go.mod, .go, and .proto files
FILES=$(find . -type f \( -name "*.go" -o -name "go.mod" -o -name "*.proto" \) | grep -v node_modules | grep -v vendor)

# Count files to update
COUNT=$(echo "$FILES" | wc -l)
echo "Found $COUNT files to update"

# Create backup
BACKUP_DIR=".backup-$(date +%Y%m%d-%H%M%S)"
echo "Creating backup in $BACKUP_DIR"
mkdir -p "$BACKUP_DIR"
echo "$FILES" | while read -r file; do
    if [ -f "$file" ]; then
        mkdir -p "$BACKUP_DIR/$(dirname "$file")"
        cp "$file" "$BACKUP_DIR/$file"
    fi
done

# Update all files
echo "Updating files..."
echo "$FILES" | while read -r file; do
    if [ -f "$file" ]; then
        # Use perl for in-place editing (works on both Linux and macOS)
        perl -pi -e "s|$OLD_PATH|$NEW_PATH|g" "$file"
        echo "  ✓ $file"
    fi
done

echo ""
echo "=================================================="
echo "✅ Module paths updated successfully!"
echo ""
echo "Next steps:"
echo "1. Run: go work sync"
echo "2. Run: go mod tidy in each module directory"
echo "3. Test: go build ./..."
echo ""
echo "Backup created in: $BACKUP_DIR"
echo "If anything goes wrong, restore with: cp -r $BACKUP_DIR/* ."
