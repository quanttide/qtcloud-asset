//! OSS 子命令：作为 Provider 的 HTTP 客户端，复用其只读接口。
//!
//! 不直接访问阿里云 OSS，也不持有 AK/SK；凭证只在 Provider 一侧。
//! 对应 Provider 的三个只读端点：
//! - `GET /buckets`                 → 列出所有桶
//! - `GET /buckets/{name}/objects`  → 列出桶内对象
//! - `GET /buckets/{name}/object-url?key=…&expires=…` → 生成访问链接

use anyhow::{bail, Context, Result};
use serde::Deserialize;

/// Provider 服务地址（与 Studio 一致）。
const DEFAULT_PROVIDER_URL: &str = "http://127.0.0.1:9000";

/// OSS 桶。
#[derive(Debug, Deserialize)]
pub struct Bucket {
    pub name: String,
    #[serde(default)]
    pub region: String,
    #[serde(default)]
    pub storage_class: String,
    #[serde(default)]
    pub created_at: String,
}

/// OSS 对象（文件或目录）。
#[derive(Debug, Deserialize)]
pub struct OssObject {
    pub key: String,
    #[serde(default)]
    pub size: i64,
    #[serde(default)]
    #[serde(rename = "type")]
    pub object_type: String,
    #[serde(default)]
    pub storage_class: String,
    #[serde(default)]
    pub last_modified: String,
}

/// Provider 客户端，封装 OSS 只读接口。
pub struct Client {
    base_url: String,
}

impl Client {
    pub fn new(base_url: &str) -> Self {
        let base = base_url.trim_end_matches('/');
        Client {
            base_url: base.to_string(),
        }
    }

    /// 列出所有桶。sort: name/created；order: asc/desc（空 = 不排序）。
    pub fn list_buckets(&self, sort: &str, order: &str) -> Result<Vec<Bucket>> {
        #[derive(Deserialize)]
        struct Resp {
            buckets: Vec<Bucket>,
        }

        let mut path = "/buckets".to_string();
        let mut has_query = false;
        if !sort.is_empty() {
            path.push_str(&format!("?sort={}", url_encode(sort)));
            has_query = true;
        }
        if !order.is_empty() {
            path.push_str(&format!(
                "{}order={}",
                if has_query { "&" } else { "?" },
                order
            ));
        }

        let resp: Resp = self.get::<Resp>(&path)?;
        Ok(resp.buckets)
    }

    /// 列出桶内对象。返回对象列表及分页信息（next_marker / truncated）。
    pub fn list_objects(
        &self,
        bucket: &str,
        prefix: &str,
        sort: &str,
        order: &str,
        limit: i64,
        marker: &str,
    ) -> Result<(Vec<OssObject>, Option<String>, bool)> {
        #[derive(Deserialize)]
        struct Resp {
            objects: Vec<OssObject>,
            #[serde(default)]
            next_marker: Option<String>,
            #[serde(default)]
            truncated: bool,
        }

        let path = build_object_path(bucket, prefix, sort, order, limit, marker);
        let resp: Resp = self.get::<Resp>(&path)?;
        Ok((resp.objects, resp.next_marker, resp.truncated))
    }

    /// 生成对象访问链接。私密桶 expires 最大 604800 秒，公开桶忽略此参数。
    pub fn object_url(&self, bucket: &str, key: &str, expires: i64) -> Result<String> {
        #[derive(Deserialize)]
        struct Resp {
            url: String,
        }

        let path = format!(
            "/buckets/{}/object-url?key={}&expires={}",
            url_encode(bucket),
            url_encode(key),
            expires
        );
        let resp: Resp = self.get::<Resp>(&path)?;
        Ok(resp.url)
    }

    /// 发起 GET 请求并反序列化为 JSON。
    fn get<T: serde::de::DeserializeOwned>(&self, path: &str) -> Result<T> {
        let url = format!("{}{}", self.base_url, path);
        let resp = ureq::get(&url)
            .call()
            .with_context(|| format!("请求失败: {url}"))?;

        let status = resp.status();
        let body = resp
            .into_string()
            .with_context(|| format!("读取响应失败: {url}"))?;

        if status != 200 {
            bail!("Provider 返回 HTTP {status}: {body}");
        }

        serde_json::from_str(&body).with_context(|| format!("JSON 解析失败: {body}"))
    }
}

/// 构造列出对象的请求路径（含查询参数），纯函数便于测试。
fn build_object_path(
    bucket: &str,
    prefix: &str,
    sort: &str,
    order: &str,
    limit: i64,
    marker: &str,
) -> String {
    let mut path = format!("/buckets/{}/objects", url_encode(bucket));
    let mut params: Vec<String> = Vec::new();
    if !prefix.is_empty() {
        params.push(format!("prefix={}", url_encode(prefix)));
    }
    if !sort.is_empty() {
        params.push(format!("sort={}", url_encode(sort)));
    }
    if !order.is_empty() {
        params.push(format!("order={}", order));
    }
    if limit > 0 {
        params.push(format!("limit={}", limit));
    }
    if !marker.is_empty() {
        params.push(format!("marker={}", url_encode(marker)));
    }
    if !params.is_empty() {
        path.push('?');
        path.push_str(&params.join("&"));
    }
    path
}

/// 对路径段做 URL 编码（桶名与对象 key 可能含特殊字符）。
fn url_encode(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    for b in s.bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                out.push(b as char)
            }
            _ => out.push_str(&format!("%{:02X}", b)),
        }
    }
    out
}

/// 默认的 Provider 服务地址。
pub fn default_provider_url() -> &'static str {
    DEFAULT_PROVIDER_URL
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_url_encode_plain() {
        assert_eq!(url_encode("qtcloud-studio"), "qtcloud-studio");
    }

    #[test]
    fn test_url_encode_slash() {
        // 对象 key 中的 '/' 会被编码为 %2F，避免被当成路径分隔
        assert_eq!(url_encode("a/b"), "a%2Fb");
    }

    #[test]
    fn test_url_encode_space() {
        assert_eq!(url_encode("a b"), "a%20b");
    }

    #[test]
    fn test_client_new_trims_trailing_slash() {
        let c = Client::new("http://127.0.0.1:9000/");
        assert_eq!(c.base_url, "http://127.0.0.1:9000");
    }

    #[test]
    fn test_build_object_path_no_params() {
        assert_eq!(
            build_object_path("my-bucket", "", "", "", 0, ""),
            "/buckets/my-bucket/objects"
        );
    }

    #[test]
    fn test_build_object_path_prefix_only() {
        assert_eq!(
            build_object_path("my-bucket", "assets/", "", "", 0, ""),
            "/buckets/my-bucket/objects?prefix=assets%2F"
        );
    }

    #[test]
    fn test_build_object_path_sort_order() {
        assert_eq!(
            build_object_path("my-bucket", "", "size", "desc", 0, ""),
            "/buckets/my-bucket/objects?sort=size&order=desc"
        );
    }

    #[test]
    fn test_build_object_path_limit_marker() {
        assert_eq!(
            build_object_path("my-bucket", "", "", "", 100, "next-token"),
            "/buckets/my-bucket/objects?limit=100&marker=next-token"
        );
    }

    #[test]
    fn test_build_object_path_all_params() {
        assert_eq!(
            build_object_path("my-bucket", "docs/", "date", "asc", 50, "m1"),
            "/buckets/my-bucket/objects?prefix=docs%2F&sort=date&order=asc&limit=50&marker=m1"
        );
    }
}
