# 量潮数字资产云服务端


## Deployment

The provider is deployed as an Alibaba Cloud Function Compute custom runtime
code package. GitHub Actions builds `src/provider`, creates
`qtcloud-asset-provider-py311.zip`, and uploads it to:

```text
oss://qtcloud-asset-studio/provider/qtcloud-asset-provider-py311.zip
```

Terraform uses that OSS object to create the `provider` FC function.

The optional provider custom domain is `api.asset.quanttide.com`. It is disabled
by default in Terraform because Alibaba Cloud Function Compute requires the
domain to have an ICP license that belongs to Alibaba Cloud before binding it.
After the ICP requirement is satisfied, set
`enable_provider_custom_domain = true` and apply Terraform.
