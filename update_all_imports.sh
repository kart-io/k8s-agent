#!/bin/bash

# Script to update imports throughout the project

echo "Updating imports in common directory for moved packages..."

# Update references to moved packages
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent -name "*.go" -type f ! -path "*/vendor/*" ! -path "*/.git/*" -exec grep -l '"github.com/kart-io/k8s-agent/common/app"' {} \; | while read file; do
    echo "Updating $file"
    sed -i 's|"github.com/kart-io/k8s-agent/common/app"|"github.com/kart-io/k8s-agent/pkg/app"|g' "$file"
done

find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent -name "*.go" -type f ! -path "*/vendor/*" ! -path "*/.git/*" -exec grep -l '"github.com/kart-io/k8s-agent/common/bootstrap"' {} \; | while read file; do
    echo "Updating $file"
    sed -i 's|"github.com/kart-io/k8s-agent/common/bootstrap"|"github.com/kart-io/k8s-agent/pkg/bootstrap"|g' "$file"
done

find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent -name "*.go" -type f ! -path "*/vendor/*" ! -path "*/.git/*" -exec grep -l '"github.com/kart-io/k8s-agent/common/contextx"' {} \; | while read file; do
    echo "Updating $file"
    sed -i 's|"github.com/kart-io/k8s-agent/common/contextx"|"github.com/kart-io/k8s-agent/pkg/contextx"|g' "$file"
done

find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent -name "*.go" -type f ! -path "*/vendor/*" ! -path "*/.git/*" -exec grep -l '"github.com/kart-io/k8s-agent/common/idempotent"' {} \; | while read file; do
    echo "Updating $file"
    sed -i 's|"github.com/kart-io/k8s-agent/common/idempotent"|"github.com/kart-io/k8s-agent/pkg/idempotent"|g' "$file"
done

find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent -name "*.go" -type f ! -path "*/vendor/*" ! -path "*/.git/*" -exec grep -l '"github.com/kart-io/k8s-agent/common/initializers"' {} \; | while read file; do
    echo "Updating $file"
    sed -i 's|"github.com/kart-io/k8s-agent/common/initializers"|"github.com/kart-io/k8s-agent/pkg/initializers"|g' "$file"
done

echo "Import updates completed!"