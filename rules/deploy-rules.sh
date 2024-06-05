#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <trats-rules-directory>"
    exit 1
fi

INPUT_DIRECTORY="$1"
CONFIG_MAP_NAME="trats-rules-config"
NAMESPACE="tratteria"

if [ ! -d "$INPUT_DIRECTORY" ]; then
    echo "Error: Trats rules directory $INPUT_DIRECTORY does not exist."
    exit 1
fi

if [ $(find "$INPUT_DIRECTORY" -type f -name '*.ndjson' | wc -l) -eq 0 ]; then
    echo "Error: No trats rules files found in $INPUT_DIRECTORY."
    exit 1
fi

kubectl create configmap "$CONFIG_MAP_NAME" \
    --from-file=trats-rules.ndjson=<(find "$INPUT_DIRECTORY" -type f -name '*.ndjson' -exec cat {} + | sed 's/^/    /') \
    --namespace="$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -

echo -e "\nTrats rules deployed successfully."
