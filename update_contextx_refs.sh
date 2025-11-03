#!/bin/bash

echo "Updating contextx references to contextutil..."

# Update all Go files that reference contextx
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent -name "*.go" -type f | while read file; do
    # Update import paths
    sed -i 's|"github.com/kart-io/k8s-agent/pkg/contextx"|"github.com/kart-io/k8s-agent/pkg/contextutil"|g' "$file"

    # Update package references
    sed -i 's|contextx\.|contextutil.|g' "$file"
done

echo "References updated successfully!"