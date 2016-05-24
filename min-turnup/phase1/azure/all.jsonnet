local cfg = import "../../.config.json";
{
  "azure.tf": (import "lib/azure.jsonnet")(cfg),
}
