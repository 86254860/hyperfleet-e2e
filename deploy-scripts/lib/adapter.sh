#!/usr/bin/env bash

# adapter.sh - HyperFleet Adapter component deployment functions
#
# This module handles discovery, installation, and uninstallation of adapters
# from the testdata/adapter-configs directory

# ============================================================================
# Adapter Discovery Functions
# ============================================================================

discover_adapters() {
    local adapter_configs_dir="${TESTDATA_DIR}/adapter-configs"

    if [[ ! -d "${adapter_configs_dir}" ]]; then
        log_verbose "Adapter configs directory not found: ${adapter_configs_dir}" >&2
        return 1
    fi

    # Find all directories matching clusters-* or nodepools-* pattern
    local adapter_dirs=()
    while IFS= read -r -d '' dir; do
        local basename=$(basename "$dir")
        if [[ "$basename" =~ ^(clusters|nodepools)- ]]; then
            adapter_dirs+=("$basename")
        fi
    done < <(find "${adapter_configs_dir}" -mindepth 1 -maxdepth 1 -type d -print0)

    if [[ ${#adapter_dirs[@]} -eq 0 ]]; then
        log_verbose "No adapter configurations found (no clusters-* or nodepools-* directories)" >&2
        return 1
    fi

    log_info "Found ${#adapter_dirs[@]} adapter(s) to deploy:" >&2
    for dir in "${adapter_dirs[@]}"; do
        log_info "  - ${dir}" >&2
    done

    # Export for use in other functions
    printf '%s\n' "${adapter_dirs[@]}"
}

get_adapters_by_type() {
    local resource_type="$1"  # "clusters" or "nodepools"
    local adapter_configs_dir="${TESTDATA_DIR}/adapter-configs"

    if [[ ! -d "${adapter_configs_dir}" ]]; then
        return 1
    fi

    # Find all directories matching the resource type pattern
    local adapter_names=()
    while IFS= read -r -d '' dir; do
        local basename=$(basename "$dir")
        if [[ "$basename" =~ ^${resource_type}-(.+)$ ]]; then
            # Extract just the adapter name (everything after "clusters-" or "nodepools-")
            local adapter_name="${BASH_REMATCH[1]}"
            adapter_names+=("${adapter_name}")
        fi
    done < <(find "${adapter_configs_dir}" -mindepth 1 -maxdepth 1 -type d -print0)

    if [[ ${#adapter_names[@]} -eq 0 ]]; then
        return 1
    fi

    # Return comma-separated list
    local IFS=','
    echo "${adapter_names[*]}"
}

parse_adapter_name() {
    local dir_name="$1"

    # Extract resource_type and adapter_name
    # Format: <resource_type>-<adapter_name>
    # Examples: clusters-example1-namespace, nodepools-namespace

    if [[ "$dir_name" =~ ^(clusters|nodepools)-(.+)$ ]]; then
        local resource_type="${BASH_REMATCH[1]}"
        local adapter_name="${BASH_REMATCH[2]}"

        echo "${resource_type}|${adapter_name}"
    else
        log_error "Invalid adapter directory name format: ${dir_name}"
        return 1
    fi
}

# ============================================================================
# Adapter Installation Functions
# ============================================================================

install_adapter_instance() {
    local dir_name="$1"

    log_section "Installing Adapter: ${dir_name}"

    # Parse adapter name
    local parsed
    if ! parsed=$(parse_adapter_name "${dir_name}"); then
        log_error "Failed to parse adapter directory name: ${dir_name}"
        return 1
    fi

    local resource_type="${parsed%%|*}"
    local adapter_name="${parsed##*|}"

    log_info "Resource type: ${resource_type}"
    log_info "Adapter name: ${adapter_name}"

    # Construct release name
    local release_name="${RELEASE_PREFIX}-adapter-${resource_type}-${adapter_name}"

    # Source adapter config directory
    local adapter_configs_dir="${TESTDATA_DIR}/adapter-configs"
    local source_adapter_dir="${adapter_configs_dir}/${dir_name}"

    if [[ ! -d "${source_adapter_dir}" ]]; then
        log_error "Adapter config directory not found: ${source_adapter_dir}"
        return 1
    fi

    # Chart path
    local full_chart_path="${WORK_DIR}/adapter/${ADAPTER_CHART_PATH}"

    # Copy adapter config folder to chart directory
    local dest_adapter_dir="${full_chart_path}/${dir_name}"
    log_info "Copying adapter config from ${source_adapter_dir} to ${dest_adapter_dir}"

    if [[ -d "${dest_adapter_dir}" ]]; then
        log_verbose "Removing existing adapter config directory: ${dest_adapter_dir}"
        rm -rf "${dest_adapter_dir}"
    fi

    cp -r "${source_adapter_dir}" "${dest_adapter_dir}"

    # Values file path (now in the chart directory)
    local values_file="${dest_adapter_dir}/values.yaml"
    if [[ ! -f "${values_file}" ]]; then
        log_error "Values file not found: ${values_file}"
        return 1
    fi

    # Construct subscription ID and topic names
    # Allow override from environment variables, otherwise use auto-generated defaults
    local subscription_id="${ADAPTER_SUBSCRIPTION_ID:-${NAMESPACE}-${resource_type}-${adapter_name}}"
    local topic="${ADAPTER_TOPIC:-${NAMESPACE}-${resource_type}}"
    local dead_letter_topic="${ADAPTER_DEAD_LETTER_TOPIC:-${NAMESPACE}-${resource_type}-dlq}"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY-RUN] Would install adapter with:"
        log_info "  Release name: ${release_name}"
        log_info "  Namespace: ${NAMESPACE}"
        log_info "  Chart path: ${full_chart_path}"
        log_info "  Values file: ${values_file}"
        log_info "  Image: ${IMAGE_REGISTRY}/${ADAPTER_IMAGE_REPO}:${ADAPTER_IMAGE_TAG}"
        log_info "  Subscription ID: ${subscription_id}"
        log_info "  Topic: ${topic}"
        log_info "  Dead Letter Topic: ${dead_letter_topic}"
        return 0
    fi


    # Build helm command
    local helm_cmd=(
        helm upgrade --install
        "${release_name}"
        "${full_chart_path}"
        --namespace "${NAMESPACE}"
        --create-namespace
        --wait
        --timeout 5m
        -f "${values_file}"
        --set "image.registry=${IMAGE_REGISTRY}"
        --set "image.repository=${ADAPTER_IMAGE_REPO}"
        --set "image.tag=${ADAPTER_IMAGE_TAG}"
        --set "broker.googlepubsub.projectId=${GCP_PROJECT_ID}"
        --set "broker.googlepubsub.createTopicIfMissing=${ADAPTER_GOOGLEPUBSUB_CREATE_TOPIC_IF_MISSING}"
        --set "broker.googlepubsub.createSubscriptionIfMissing=${ADAPTER_GOOGLEPUBSUB_CREATE_SUBSCRIPTION_IF_MISSING}"
        --set "broker.googlepubsub.subscriptionId=${subscription_id}"
        --set "broker.googlepubsub.topic=${topic}"
        --set "broker.googlepubsub.deadLetterTopic=${dead_letter_topic}"
    )

    log_info "Executing Helm command:"
    log_info "${helm_cmd[*]}"
    echo

    if "${helm_cmd[@]}"; then
        log_success "Adapter ${adapter_name} for ${resource_type} Helm release created successfully"

        # Verify pod health
        log_info "Verifying pod health..."
        if verify_pod_health "${NAMESPACE}" "app.kubernetes.io/instance=${release_name}" "${adapter_name}" 120 5; then
            log_success "Adapter ${adapter_name} for ${resource_type} is running and healthy"
        else
            log_error "Adapter ${adapter_name} for ${resource_type} deployment failed health check"
            log_info "Checking pod logs for troubleshooting:"
            kubectl logs -n "${NAMESPACE}" -l "app.kubernetes.io/instance=${release_name}" --tail=50 2>/dev/null || true
            return 1
        fi
    else
        log_error "Failed to install adapter ${adapter_name} for ${resource_type}"
        return 1
    fi
}

install_adapters() {
    log_section "Deploying All Adapters"

    # Discover adapters
    local adapters
    if ! adapters=$(discover_adapters); then
        log_warning "No adapters found to deploy"
        return 0
    fi

    # Install each adapter
    local failed=0
    while IFS= read -r adapter_dir; do
        if ! install_adapter_instance "${adapter_dir}"; then
            log_warning "Failed to install adapter: ${adapter_dir}"
            ((failed++))
        fi
    done <<< "${adapters}"

    if [[ ${failed} -gt 0 ]]; then
        log_error "${failed} adapter(s) failed to install"
        return 1
    else
        log_success "All adapters deployed successfully"
    fi
}

# ============================================================================
# Adapter Uninstallation Functions
# ============================================================================

uninstall_adapter_instance() {
    local dir_name="$1"

    log_section "Uninstalling Adapter: ${dir_name}"

    # Parse adapter name
    local parsed
    if ! parsed=$(parse_adapter_name "${dir_name}"); then
        log_error "Failed to parse adapter directory name: ${dir_name}"
        return 1
    fi

    local resource_type="${parsed%%|*}"
    local adapter_name="${parsed##*|}"

    # Construct release name
    local release_name="${RELEASE_PREFIX}-adapter-${resource_type}-${adapter_name}"

    # Check if release exists
    if ! helm list -n "${NAMESPACE}" 2>/dev/null | grep -q "^${release_name}"; then
        log_warning "Release '${release_name}' not found in namespace '${NAMESPACE}'"
        return 0
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY-RUN] Would uninstall adapter (release: ${release_name})"
        return 0
    fi

    log_info "Uninstalling adapter ${adapter_name} for ${resource_type}..."
    log_info "Executing: helm uninstall ${release_name} -n ${NAMESPACE} --wait --timeout 5m"

    if helm uninstall "${release_name}" -n "${NAMESPACE}" --wait --timeout 5m; then
        log_success "Adapter ${adapter_name} for ${resource_type} uninstalled successfully"
    else
        log_error "Failed to uninstall adapter ${adapter_name} for ${resource_type}"
        return 1
    fi
}

uninstall_adapters() {
    log_section "Uninstalling All Adapters"

    # Discover adapters
    local adapters
    if ! adapters=$(discover_adapters); then
        log_warning "No adapters found to uninstall"
        return 0
    fi

    # Uninstall each adapter
    local failed=0
    while IFS= read -r adapter_dir; do
        if ! uninstall_adapter_instance "${adapter_dir}"; then
            log_warning "Failed to uninstall adapter: ${adapter_dir}"
            ((failed++))
        fi
    done <<< "${adapters}"

    if [[ ${failed} -gt 0 ]]; then
        log_warning "${failed} adapter(s) failed to uninstall"
    fi
}
