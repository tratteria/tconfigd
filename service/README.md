# tconfigd Components
This directory holds the source code for tconfigd. Below is an overview of each component and its functionality.

### agentsmanager
It manages agents by enabling agent registrations, handling agent heartbeats, and tracking active agents.

### api
It implements tconfigd's APIs. It holds the following APIs:

- Agent Registration API
- Agent Heartbeat API

### configdispatcher
It dispatches configurations to tratteria and tratteria agents.

### tratcontroller
It implements Kubernetes controller for `tratteria.io` custom resources. It hosts controllers for the following resources:
- `tratteria.io/TraT`
- `tratteria.io/TraTExclusion`

### webhook
It implements Kubernetes Admission Controller Webhooks. It hosts the following webhooks:
- Tratteria Agent Injection Mutating Admission Controller Webhook
- `tratteria.io` custom resources Validating Admission Controller Webhook

## Contributions
Contributions to the project are welcome, including feature enhancements, bug fixes, and documentation improvements.
