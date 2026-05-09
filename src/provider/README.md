# 量潮数字资产云服务端


## Deployment

The provider is deployed as an Alibaba Cloud Function Compute custom runtime
code package. GitHub Actions builds `src/provider`, creates
`qtcloud-asset-provider-py311.zip`, and uploads it to:

```text
oss://qtcloud-asset-studio/provider/qtcloud-asset-provider-py311.zip
```

Terraform uses that OSS object to create the `provider` FC function.
