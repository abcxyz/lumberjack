output "server_url" {
  value = module.server_service.audit_log_server_url
}

output "grpc_client_urls" {
  value = [
    for _, value in var.grpc_client_images :
    lookup(google_cloud_run_service.grpc_client_services, value).status[0].url
  ]
}

output "http_client_urls" {
  value = [
    for _, value in var.http_client_images :
    lookup(google_cloud_run_service.http_client_services, value).status[0].url
  ]
}
