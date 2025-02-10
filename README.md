# Kubernetes Readiness and Liveness Proxy
This implementation aims at exposing readiness and liveness endpoints for kubernetes pods, meant purely at 3rd parties that lack those endpoints.
It allows you to configure the endpoint of the application to pull from as well as a regex that can be compared to the response body and verify ther service is healthy and ready(readiness) or alive(livenss).
