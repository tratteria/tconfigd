kubectl delete namespace tratteria
kubectl delete clusterrole tconfigd-service-account-role
kubectl delete clusterrolebinding tconfigd-service-account-binding
kubectl delete mutatingwebhookconfiguration tratteria-agent-injector