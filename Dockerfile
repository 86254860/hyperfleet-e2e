# [TODO] It requires to add the real file once the e2e test framework is ready
# The image is for Prow CI steps to manage the CLM component testing
FROM golang:1.25-alpine

WORKDIR /workspace

COPY . .
