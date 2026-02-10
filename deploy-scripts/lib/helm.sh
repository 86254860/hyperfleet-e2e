#!/usr/bin/env bash

# helm.sh - Helm chart management functions
#
# This module provides functions for cloning and managing Helm charts

# ============================================================================
# Helm Chart Management
# ============================================================================

clone_helm_chart() {
    local component="$1"
    local repo_url="$2"
    local ref="$3"
    local chart_path="$4"

    log_info "Cloning ${component} Helm chart from ${repo_url}@${ref} (sparse: ${chart_path})"

    local component_dir="${WORK_DIR}/${component}"

    if [[ -z "${WORK_DIR}" || "${WORK_DIR}" == "/" ]]; then
       log_error "WORK_DIR must be set to a non-root directory"
       return 1
    fi
    if [[ -z "${component}" ]]; then
       log_error "Component name is required"
       return 1
    fi

    # Clean up any existing directory to ensure fresh clone
    if [[ -d "${component_dir}" ]]; then
        log_verbose "Removing existing directory: ${component_dir}"
        rm -rf "${component_dir}"
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY-RUN] Would clone (sparse): git clone --depth 1 --filter=blob:none --sparse --branch ${ref} ${repo_url}"
        log_info "[DRY-RUN] Would checkout: ${chart_path}"
        return 0
    fi

    # Clone with sparse checkout - only download the chart directory
    log_verbose "Executing sparse checkout: git clone --depth 1 --filter=blob:none --sparse --no-checkout --branch ${ref} ${repo_url} ${component_dir}"
    if ! git clone --depth 1 --filter=blob:none --sparse --no-checkout --branch "${ref}" "${repo_url}" "${component_dir}" >/dev/null 2>&1; then
        log_error "Failed to clone ${component} Helm chart"
        return 1
    fi

    # Configure sparse checkout to only include the chart path (no cone mode to avoid root files)
    log_verbose "Configuring sparse checkout for: ${chart_path}"
    if ! (cd "${component_dir}" && \
          git sparse-checkout init --no-cone >/dev/null 2>&1 && \
          git sparse-checkout set "${chart_path}" >/dev/null 2>&1 && \
          git checkout "${ref}" >/dev/null 2>&1); then
        log_error "Failed to checkout chart path: ${chart_path}"
        return 1
    fi

    # Verify chart path exists
    local full_chart_path="${component_dir}/${chart_path}"
    if [[ ! -f "${full_chart_path}/Chart.yaml" ]]; then
        log_error "Chart.yaml not found at ${full_chart_path}"
        log_error "Please verify the chart path is correct"
        return 1
    fi

    log_success "Cloned ${component} Helm chart"
    log_verbose "Chart location: ${full_chart_path}"
}
