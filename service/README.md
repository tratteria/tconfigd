# tconfigd Components
This directory contains the source code for tconfigd. Below is an overview of tconfigd's components and their functionalities.

### websocketserver
This package contains the WebSocket server that manages connections and communications to Tratteria Agents and Tratteria Service.

### servicemessagehandler
This package provides a high-level interface to propagate rules and messages to Tratteria Agents and Tratteria Service.

### tratcontroller
This package implements Kubernetes controllers for `tratteria.io` custom resources. It hosts controllers for the following resources:

- `tratteria.io/TraT`
- `tratteria.io/TratteriaConfig`
- `tratteria.io/TraTExclusion`

### webhook
This package implements Kubernetes Admission Controller Webhooks. It hosts the following webhooks:

- Tratteria Agent Injection Mutating Admission Controller Webhook

## Contributions
Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
