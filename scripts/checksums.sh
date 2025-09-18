#!/bin/bash
set -e

cd dist
echo "Generating checksums..."

# Generate individual checksums
for file in mvx-*; do
    if [ -f "$file" ]; then
        echo "Generated checksum for $file"
        sha256sum "$file" > "$file.sha256"
    fi
done

# Generate combined checksums
echo "Combined checksums in dist/checksums.txt"
sha256sum mvx-* > checksums.txt
