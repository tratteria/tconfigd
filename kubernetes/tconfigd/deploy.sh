docker build -t tconfigd:latest -f ../../service/Dockerfile ../../service/

kubectl create namespace tratteria

kubectl create configmap config --from-file=config.yaml=config.yaml -n tratteria

cd ../../rules
chmod +x deploy-rules.sh
./deploy-rules.sh example-rules
cd ../kubernetes/tconfigd

kubectl apply -f service-account.yaml
kubectl apply -f role.yaml
kubectl apply -f rolebinding.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f tratteria-agent-injector-mutating-webhook.yaml
