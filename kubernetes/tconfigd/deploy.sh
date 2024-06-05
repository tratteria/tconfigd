docker build -t tconfigd:latest -f ../../service/Dockerfile ../../service/

kubectl create namespace tratteria
kubectl apply -f service-account.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f tratteria-agent-injector-mutating-webhook.yaml

cd ../../rules
chmod +x deploy-rules.sh
./deploy-rules.sh example-rules