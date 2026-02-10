env = "dev"
port = 8085

auth{
  endpoint= "http://localhost:8084"
}

endpoint_configuration = [
  {
    live_endpoint          = "http://localhost:8086"  #user service
    live_timeout_sec       = 60  #timeout for the service
    gw_endpoint         = "/v1/users"  #which endpoint we'll be asking for in  service
    rate_limit_req_per_sec = 5 #rate limiting for this specific endpoint
  },
  {
    live_endpoint          = "http://localhost:8086"  #user service
    live_timeout_sec       = 60  #timeout for the service
    gw_endpoint         = "/v1/users/??/contact"  #which endpoint we'll be asking for in  service
    rate_limit_req_per_sec = 5 #rate limiting for this specific endpoint
  },
  {
    live_endpoint          = "http://localhost:8087"  #orders service
    live_timeout_sec       = 60  #timeout for the service
    gw_endpoint         = "/v1/orders"  #which endpoint we'll be asking for in  service
    rate_limit_req_per_sec = 5 #rate limiting for this specific endpoint
  },
  {
    live_endpoint          = "http://localhost:8087"  #orders service
    live_timeout_sec       = 60  #timeout for the service
    gw_endpoint         = "/v1/orders/??/details"  #which endpoint we'll be asking for in  service
    rate_limit_req_per_sec = 5 #rate limiting for this specific endpoint
  }
]