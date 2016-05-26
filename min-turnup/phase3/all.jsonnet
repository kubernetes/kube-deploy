function(cfg)
  local enabled = function(addon)
    cfg.phase3[addon];
  if cfg.phase3.run_addons then
    (if enabled("kube_proxy") then
      (import "kube-proxy/kube-proxy.jsonnet")(cfg)) +
    (if enabled("dashboard") then
      (import "dashboard/dashboard.jsonnet")(cfg))
