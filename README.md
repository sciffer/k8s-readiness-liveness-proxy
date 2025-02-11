# Kubernetes Readiness and Liveness Proxy
While working with solutions like Kafka connect, you quickly run into the reality of no proper liveness probes that will restart the pod in case jobs/tasks are failing although the information is exposed by the system - so quickly I wrote this service to act as a sidecar to expose a useful liveness probe.
This is a naive simplified implementation that aims at exposing readiness and liveness endpoints for kubernetes pods, meant to complement 3rd parties that lack those endpoints as a sidecar.
It allows you to configure the endpoint of the application to pull from as well as a regex that can be compared to the response body and verify the service is healthy and ready(readiness) or alive(livenss).
To control the configuration, please use a config map with the same data structure that can be found in the `config/config.yaml` file and mount it to the container at /app/config.
Like:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-rlp-config
data:
  config.yaml: |
    liveness:
      type: http
      path: /health
      port: 80
      expected_status: 200
      target_service_address: "localhost"
      expected_body_regex: ""
      negate_expected_body_regex: false
    
    readiness:
      type: http
      path: /ready
      port: 80
      expected_status: 200
      target_service_address: "localhost"
      expected_body_regex: ""
      negate_expected_body_regex: false
```
