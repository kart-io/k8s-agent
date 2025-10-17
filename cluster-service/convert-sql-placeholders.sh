#!/bin/bash

# Script to convert PostgreSQL placeholders ($1, $2, etc.) to MySQL placeholders (?)
# in Go SQL queries

echo "Converting PostgreSQL placeholders to MySQL placeholders..."

# Find all .go files with SQL queries
for file in $(find . -name "*.go" -type f); do
    # Check if file contains PostgreSQL placeholders
    if grep -q '\$[0-9]' "$file"; then
        echo "Processing: $file"

        # Create temporary file
        temp_file="${file}.tmp"

        # Process the file
        # Replace $1, $2, ... $99 with ?
        sed -e 's/\$[0-9]\+/?/g' "$file" > "$temp_file"

        # Replace original file
        mv "$temp_file" "$file"
    fi
done

echo "Conversion complete!"
echo ""
echo "⚠️  IMPORTANT: Manual review required!"
echo "The conversion changed all \$N placeholders to ?."
echo "Please verify that:"
echo "1. The number of ? matches the number of parameters"
echo "2. The order of parameters in ExecContext/QueryContext calls is correct"
echo "3. No string literals containing \$ were accidentally changed"
