# k8s-readiness-liveness-proxy

[![Build Status](https://via.placeholder.com/150x20/007bff/FFFFFF?text=Build+Status)](https://github.com/sciffer/k8s-readiness-liveness-proxy/actions/workflows/ci.yml/badge.svg?branch=master) <!-- Replace with actual build status badge -->
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

`k8s-readiness-liveness-proxy` is a simple service designed to act as a sidecar in Kubernetes deployments, providing readiness and liveness probes for applications that don't natively expose them. This is particularly useful for situations like Kafka Connect, where jobs/tasks might fail without the pod being restarted due to a lack of proper liveness probes.

This proxy allows you to configure endpoints and regular expressions to check the health of your application and expose these checks as standard Kubernetes readiness and liveness probes.

## Features

*   **Easy Configuration:** Uses a simple YAML configuration file (via Kubernetes ConfigMap).
*   **HTTP Probe Support:**  Checks application health via HTTP requests.
*   **Regex Matching:**  Validates responses using regular expressions.
*   **Negatable Regex:**  Option to invert the result of the regex match.
*   **Kubernetes Integration:** Designed to work seamlessly as a sidecar container.

## Configuration

The proxy is configured via a YAML file, typically mounted as a Kubernetes ConfigMap. The default configuration path is `/app/config/config.yaml`.

Here's an example of the configuration file:

```yaml
liveness:
  type: http
  path: /status  # The path on the target application's endpoint to scrape
  port: 8088     # The port on the target application's service to scrape
  expected_status: 200  # Expected HTTP status code
  target_service_address: "localhost"  # Address of the target service (localhost for sidecar)
  expected_body_regex: "Status: Started"  # Regex to match in the response body
  negate_expected_body_regex: false  # Invert the regex match result?

readiness:
  type: http
  path: /status
  port: 8088
  expected_status: 200
  target_service_address: "localhost"
  expected_body_regex: "Tests: PASS"
  negate_expected_body_regex: false
```

**Configuration Options:**

| Parameter                    | Description                                                                                                                               | Type   | Default |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------- |
| `type`                       | The type of probe. Currently, only `http` is supported.                                                                                   | string | `http`  |
| `path`                       | The path to request on the target application's endpoint.                                                                                 | string |         |
| `port`                       | The port to use for the request.                                                                                                         | number |         |
| `expected_status`            | The expected HTTP status code.                                                                                                           | number | `200`   |
| `target_service_address`     | The address of the target service. Use `"localhost"` when running as a sidecar.                                                           | string |         |
| `expected_body_regex`       | A regular expression to match against the response body. Leave empty to skip body checking.                                               | string | `""`    |
| `negate_expected_body_regex` | If `true`, the probe will be considered successful if the regex *does not* match the response body.                                      | bool   | `false` |

## Usage

### Kubernetes Sidecar (Recommended)

The primary use case for this proxy is as a sidecar container in a Kubernetes deployment. Here's an example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: my-app-image:latest  # Replace with your main application image
          ports:
            - containerPort: 8088  # Replace with your application's port
          livenessProbe:
            httpGet:
              path: /livez  #  Liveness probe path exposed by the proxy
              port: 8080    #  Port the proxy listens on
          readinessProbe:
            httpGet:
              path: /readyz # Readiness probe path exposed by the proxy
              port: 8080    # Port the proxy listens on
          volumeMounts:
            - name: config-volume
              mountPath: /app/config  # Mount the config for your app (optional)
        - name: k8s-readiness-liveness-proxy
          image: scifferous/k8s-readiness-liveness-proxy:latest  # Replace with the correct image tag
          ports:
            - containerPort: 8080 # The port the proxy listens on
          volumeMounts:
            - name: config-volume
              mountPath: /app/config  # Mount the config for the proxy
      volumes:
        - name: config-volume
          configMap:
            name: k8s-rlp-config  # The ConfigMap containing the proxy configuration
```

**Explanation:**

1.  **ConfigMap:** A ConfigMap named `k8s-rlp-config` is used to store the proxy's configuration. You should create this ConfigMap with the configuration details described in the "Configuration" section.
2.  **Deployment:** A deployment named `my-app-deployment` is defined.
3.  **Containers:**
    *   `my-app`: This is your main application container. Replace `my-app-image:latest` and the `containerPort` with your application's details.
    *   `k8s-readiness-liveness-proxy`: This is the sidecar container.  It uses the `scifferous/k8s-readiness-liveness-proxy` image (you should replace the tag with the correct one after building).
4.  **Volume Mounts:** The `config-volume` (containing the ConfigMap) is mounted to `/app/config` in *both* containers. This allows both your application and the proxy to access the configuration.
5. **Probes:** The main application container now uses the `/livez` and `/readyz` endpoints exposed by our proxy for its liveness and readiness probes.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details (you should create a LICENSE file).
