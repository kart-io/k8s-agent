#!/bin/bash

# Update imports in pkg directory
echo "Updating imports in pkg directory..."

# Update app package imports
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg/app -name "*.go" -type f -exec sed -i \
  -e 's|"github.com/kart-io/k8s-agent/common/options"|"github.com/kart-io/k8s-agent/common/options"|g' \
  -e 's|"github.com/kart-io/k8s-agent/common/bootstrap"|"github.com/kart-io/k8s-agent/pkg/bootstrap"|g' \
  {} \;

# Update bootstrap package imports
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg/bootstrap -name "*.go" -type f -exec sed -i \
  -e 's|"github.com/kart-io/k8s-agent/common/server"|"github.com/kart-io/k8s-agent/common/server"|g' \
  {} \;

# Update initializers package imports
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg/initializers -name "*.go" -type f -exec sed -i \
  -e 's|"github.com/kart-io/k8s-agent/common/bootstrap"|"github.com/kart-io/k8s-agent/pkg/bootstrap"|g' \
  -e 's|"github.com/kart-io/k8s-agent/common/health"|"github.com/kart-io/k8s-agent/common/health"|g' \
  {} \;

# Update contextx package imports
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg/contextx -name "*.go" -type f -exec sed -i \
  -e 's|"github.com/kart-io/k8s-agent/common/|"github.com/kart-io/k8s-agent/common/|g' \
  {} \;

# Update idempotent package imports
find /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg/idempotent -name "*.go" -type f -exec sed -i \
  -e 's|"github.com/kart-io/k8s-agent/common/cache"|"github.com/kart-io/k8s-agent/common/cache"|g' \
  {} \;

echo "Import updates completed!"