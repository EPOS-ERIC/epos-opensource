#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="epos"

# Apply prerequisites sequentially (these need to be in place first)
kubectl apply -f namespace.yaml
kubectl apply -f configmap-epos-env.yaml
kubectl apply -f secret-epos-secret.yaml
kubectl apply -f pvc-psqldata.yaml
kubectl apply -f pvc-converter-plugins.yaml

# Deploy infrastructure components in parallel
infra=(rabbitmq metadata-database)
echo "Deploying infrastructure components..."
for comp in "${infra[@]}"; do
  kubectl apply -f deployment-${comp}.yaml -f service-${comp}.yaml &
done
wait  # Wait for all kubectl apply commands to complete

kubectl apply -f service-rabbitmq-management.yaml

# Wait for infrastructure to be ready before proceeding
echo "Waiting for infrastructure to be ready..."
kubectl rollout status deployment/rabbitmq -n $NAMESPACE
kubectl rollout status deployment/metadata-database -n $NAMESPACE

# Deploy all services in parallel
services=(resources-service ingestor-service external-access-service converter-service converter-routine backoffice-service)
echo "Deploying services..."
for svc in "${services[@]}"; do
  kubectl apply -f deployment-${svc}.yaml -f service-${svc}.yaml &
done
wait  # Wait for all kubectl apply commands to complete

# Wait for all services to be ready in parallel
echo "Waiting for services to be ready..."
pids=()
for svc in "${services[@]}"; do
  kubectl rollout status deployment/${svc} -n $NAMESPACE &
  pids+=($!)
done

# Wait for all rollout status checks to complete
for pid in "${pids[@]}"; do
  wait $pid
done

# Deploy final components
echo "Deploying gateway and dataportal..."
kubectl apply -f deployment-gateway.yaml -f service-gateway.yaml &
kubectl apply -f deployment-dataportal.yaml -f service-dataportal.yaml &
wait

kubectl rollout status deployment/gateway -n $NAMESPACE &
kubectl rollout status deployment/dataportal -n $NAMESPACE &
wait

echo "EPOS platform deployed."
