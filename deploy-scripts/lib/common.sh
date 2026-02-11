#!/usr/bin/env bash

# common.sh - Common utilities for CLM deployment scripts
#
# This module provides shared functionality used across all deployment scripts:
# - Logging functions
# - Dependency checking
# - Kubernetes context validation

# ============================================================================
# Logging Functions
# ============================================================================

log_info() {
    echo "[INFO] $*"
}

log_success() {
    echo "[SUCCESS] $*"
}

log_warning() {
    echo "[WARNING] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_verbose() {
    if [[ "${VERBOSE}" == "true" ]]; then
        echo "[VERBOSE] $*"
    fi
}

log_section() {
    echo
    echo "==================================================================="
    echo "$*"
    echo "==================================================================="
}

# ============================================================================
# Dependency Checking
# ============================================================================

check_dependencies() {
    log_section "Checking Dependencies"

    local missing_deps=()

    local deps=("kubectl" "helm" "git")
    for dep in "${deps[@]}"; do
        if ! command -v "${dep}" &> /dev/null; then
            missing_deps+=("${dep}")
            log_error "Required dependency '${dep}' not found"
        else
            local version
            case "${dep}" in
                kubectl)
                    version=$(kubectl version --client --short 2>/dev/null | head -n1 || echo "unknown")
                    ;;
                helm)
                    version=$(helm version --short 2>/dev/null || echo "unknown")
                    ;;
                git)
                    version=$(git --version || echo "unknown")
                    ;;
            esac
            log_verbose "Found ${dep}: ${version}"
        fi
    done

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again"
        return 1
    fi

    log_success "All dependencies are available"
    return 0
}

# ============================================================================
# Kubernetes Context Validation
# ============================================================================

validate_kubectl_context() {
    log_section "Validating Kubernetes Context"

    if ! kubectl cluster-info &> /dev/null; then
        log_error "Unable to connect to Kubernetes cluster"
        log_error "Please ensure your kubeconfig is properly configured"
        return 1
    fi

    local context
    context=$(kubectl config current-context)
    log_info "Current kubectl context: ${context}"

    local cluster_info
    cluster_info=$(kubectl cluster-info 2>&1 | head -n1 || echo "unknown")
    log_verbose "Cluster info: ${cluster_info}"

    log_success "Kubectl context validated"
    return 0
}

# ============================================================================
# Pod Health Verification
# ============================================================================

verify_pod_health() {
    local namespace="$1"
    local selector="$2"
    local component_name="${3:-component}"
    local timeout="${4:-60}"
    local interval="${5:-5}"

    log_info "Verifying pod health for ${component_name}..."
    log_verbose "Namespace: ${namespace}, Selector: ${selector}"

    local elapsed=0
    while [[ ${elapsed} -lt ${timeout} ]]; do
        # Get pod status
        local pod_status
        pod_status=$(kubectl get pods -n "${namespace}" -l "${selector}" \
            -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\t"}{range .status.containerStatuses[*]}{.state.waiting.reason}{" "}{.state.terminated.reason}{end}{"\n"}{end}' 2>/dev/null)

        if [[ -z "${pod_status}" ]]; then
            log_warning "No pods found with selector ${selector} in namespace ${namespace}"
            sleep ${interval}
            ((elapsed += interval))
            continue
        fi

        # Check for failure states
        local has_failures=false
        local failure_details=""

        while IFS=$'\t' read -r pod_name phase reasons; do
            log_verbose "Pod ${pod_name}: phase=${phase}, reasons=${reasons}"

            # Check for problematic states
            if [[ "${phase}" == "Failed" ]] || \
               [[ "${reasons}" == *"CrashLoopBackOff"* ]] || \
               [[ "${reasons}" == *"Error"* ]] || \
               [[ "${reasons}" == *"ImagePullBackOff"* ]] || \
               [[ "${reasons}" == *"ErrImagePull"* ]]; then
                has_failures=true
                failure_details="${failure_details}\n  - ${pod_name}: ${phase} (${reasons})"
            fi
        done <<< "${pod_status}"

        if [[ "${has_failures}" == "true" ]]; then
            log_error "Pod health check failed for ${component_name}:"
            echo -e "${failure_details}"
            log_info "Pod details:"
            kubectl get pods -n "${namespace}" -l "${selector}"
            return 1
        fi

        # Check if all pods are running
        local running_count
        running_count=$(kubectl get pods -n "${namespace}" -l "${selector}" \
            -o jsonpath='{range .items[*]}{.status.phase}{"\n"}{end}' 2>/dev/null | grep -c "^Running$" || echo "0")

        local total_count
        total_count=$(kubectl get pods -n "${namespace}" -l "${selector}" --no-headers 2>/dev/null | wc -l | tr -d ' ')

        if [[ ${running_count} -gt 0 ]] && [[ ${running_count} -eq ${total_count} ]]; then
            log_success "All pods for ${component_name} are running (${running_count}/${total_count})"
            return 0
        fi

        log_verbose "Waiting for pods to be ready: ${running_count}/${total_count} running (${elapsed}s/${timeout}s)"
        sleep ${interval}
        ((elapsed += interval))
    done

    log_error "Timeout waiting for ${component_name} pods to become healthy"
    log_info "Current pod status:"
    kubectl get pods -n "${namespace}" -l "${selector}"
    return 1
}

# ============================================================================
# Namespace Management
# ============================================================================

delete_namespace() {
    local namespace="$1"

    log_section "Deleting Namespace"

    # Check if namespace exists
    if ! kubectl get namespace "${namespace}" &> /dev/null; then
        log_warning "Namespace '${namespace}' does not exist"
        return 0
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY-RUN] Would delete namespace: ${namespace}"
        return 0
    fi

    log_info "Deleting namespace: ${namespace}"
    log_warning "This will remove all resources in the namespace"

    if kubectl delete namespace "${namespace}" --wait --timeout=5m; then
        log_success "Namespace '${namespace}' deleted successfully"
        return 0
    else
        log_error "Failed to delete namespace '${namespace}'"
        log_info "You may need to manually remove finalizers or check for stuck resources"
        return 1
    fi
}
