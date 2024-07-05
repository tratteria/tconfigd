#!/bin/bash

info() {
    printf "\e[34m%s\n\e[0m" "$1"
}

error() {
    printf "\e[31m%s\n\e[0m" "$1"
}

success() {
    printf "\e[32m%s\n\e[0m" "$1"
}

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --no-spire    Skips the installation of SPIRE."
    echo "  -h, --help    Displays this help message."
}

apply_k8s_config() {
    kubectl apply -f $1 || { error "Failed to apply configuration for $1"; exit 1; }
}

validate_resource() {
    resource_type=$1
    resource_name=$2
    if ! kubectl get ${resource_type} ${resource_name} -n tratteria-system > /dev/null 2>&1; then
        error "${resource_type} ${resource_name} is not configured properly."
        exit 1
    fi
}

install_spire=true

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --no-spire) install_spire=false ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown option: $1"; usage; exit 1 ;;
    esac
    shift
done

if kubectl get namespace tratteria-system > /dev/null 2>&1; then
    error "tconfigd is already installed. Please uninstall the existing installation before proceeding."
    exit 1
fi

if [ "$install_spire" = true ]; then
    cd resources/spire
    chmod +x install.sh
    ./install.sh
    cd ../../
fi

info "Installing tconfigd..."

apply_k8s_config resources/namespaces.yaml
apply_k8s_config resources/crds
apply_k8s_config resources/service-account.yaml
apply_k8s_config resources/role.yaml
apply_k8s_config resources/rolebinding.yaml
apply_k8s_config resources/deployment.yaml
apply_k8s_config resources/service.yaml
apply_k8s_config resources/tratteria-agent-injector-mutating-webhook.yaml

kubectl create configmap config --from-file=config.yaml=config.yaml -n tratteria-system || {
    error "Failed to create static configuration config map"
    exit 1
}

info "Checking for the readiness of tconfigd..."
attempts=0
max_attempts=5
while ! kubectl get pods -n tratteria-system | grep -q '1/1.*Running'; do
    if [ $attempts -ge $max_attempts ]; then
        error "Failed to verify the readiness of tconfigd."
        exit 1
    fi
    attempts=$((attempts + 1))
    info "Waiting for tconfigd to be ready..."
    sleep 10
done

validate_resource namespace tratteria-system
validate_resource crd trats.tratteria.io
validate_resource crd tratteriaconfigs.tratteria.io
validate_resource serviceaccount tconfigd-service-account
validate_resource clusterrole tconfigd-service-account-role
validate_resource clusterrolebinding tconfigd-service-account-binding
validate_resource deployment tconfigd
validate_resource service tconfigd
validate_resource mutatingwebhookconfiguration tratteria-agent-injector

success "tconfigd installation completed successfully."
