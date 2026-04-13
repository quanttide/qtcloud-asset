# Q004: 阿里云基础设施

**[ 需求来源 ]**：`ADD` 规划部署到阿里云函数计算（FC）+ OSS。
**[ 核心契约 ]**：`manifests/iac/`

---

## 1. 现实需求：一键部署函数计算
* **验收点**：`terraform apply` 后，FC 函数正常运行。
* **验证资产**：`manifests/iac/main.tf`
* **状态**：⏸️ *Pending*

## 2. 现实需求：OSS 桶创建
* **验收点**：部署后 OSS 桶存在且可访问。
* **验证资产**：`manifests/iac/modules/oss/main.tf`
* **状态**：⏸️ *Pending*

## 3. 现实需求：VPC 网络隔离
* **验收点**：函数运行在 VPC 内，外部不可直接访问。
* **验证资产**：`manifests/iac/modules/vpc/main.tf`
* **状态**：⏸️ *Pending*

## 4. 现实需求：触发器绑定
* **验收点**：OSS 上传事件触发 FC 函数执行。
* **验证资产**：`manifests/iac/modules/trigger/main.tf`
* **状态**：⏸️ *Pending*
