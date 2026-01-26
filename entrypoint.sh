#!/bin/bash
set -e

# Inputs from action.yml
# GitHub Actions maps inputs to INPUT_<NAME_UPPER>
REGISTRY="${INPUT_REGISTRY}"
USERNAME="${INPUT_USERNAME}"
PASSWORD="${INPUT_PASSWORD}"
NAMESPACE="${INPUT_NAMESPACE}"
PATH_VAL="${INPUT_PATH}"
BASE="${INPUT_BASE}"

# Login if credentials are provided
if [ -n "$REGISTRY" ] && [ -n "$USERNAME" ] && [ -n "$PASSWORD" ]; then
    echo "Logging into $REGISTRY..."
    echo "$PASSWORD" | skr registry login "$REGISTRY" -u "$USERNAME" --password-stdin
fi

# Construct command
CMD="skr batch publish $PATH_VAL --registry $REGISTRY --namespace $NAMESPACE"

if [ -n "$BASE" ]; then
    CMD="$CMD --base $BASE"
fi

echo "Running: $CMD"
eval "$CMD"
