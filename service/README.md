# tconfigd Components
This directory contains the source code for tconfigd. Below is an overview of tconfigd's components and their functionalities.

### websocketserver
This package contains the WebSocket server that manages connections and communications to Tokenetes Agents and Tokenetes Service.

### servicemessagehandler
This package provides a high-level interface to propagate rules and messages to Tokenetes Agents and Tokenetes Service.

### tratcontroller
This package implements Kubernetes controllers for `tokenetes.io` custom resources. It hosts controllers for the following resources:

- `tokenetes.io/TraT`
- `tokenetes.io/TokenetesConfig`
- `tokenetes.io/TraTExclusion`

### webhook
This package implements Kubernetes Admission Controller Webhooks. It hosts the following webhooks:

- Tokenetes Agent Injection Mutating Admission Controller Webhook

## Contributions
Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
