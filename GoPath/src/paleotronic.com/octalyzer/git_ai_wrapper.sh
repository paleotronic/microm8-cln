#!/bin/bash

# Git AI Commit - A wrapper script for AI-generated commit messages
# Save this as 'git-ai' in your PATH to use as: git ai

set -e

# Check if there are staged changes
if ! git diff --cached --quiet; then
    echo "Generating AI commit message..."
    
    # Use git commit without a message - the commit-msg hook will generate it
    git commit --allow-empty-message
else
    echo "No staged changes found. Stage some files first with 'git add'"
    exit 1
fi
