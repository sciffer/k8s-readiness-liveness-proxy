# Kubernetes Readiness and Liveness Proxy
This is a naive simplified implementation that aims at exposing readiness and liveness endpoints for kubernetes pods, meant to complement 3rd parties that lack those endpoints as a sidecar.
It allows you to configure the endpoint of the application to pull from as well as a regex that can be compared to the response body and verify ther service is healthy and ready(readiness) or alive(livenss).
