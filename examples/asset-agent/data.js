window.graphData = {
  "graphNodes": [
    {
      "id": "dir-root",
      "label": "sample",
      "group": "Dir",
      "parent": null
    },
    {
      "id": "dir-docs",
      "label": "docs",
      "group": "Dir",
      "parent": "dir-root"
    },
    {
      "id": "dir-docs/drd",
      "label": "drd",
      "group": "Dir",
      "parent": "dir-docs"
    },
    {
      "id": "drd-agent",
      "label": "drd-agent",
      "group": "Agent",
      "parent": "dir-docs/drd"
    },
    {
      "id": "file-docs/drd/支付设计.md",
      "label": "支付设计.md",
      "group": "File",
      "parent": "dir-docs/drd"
    },
    {
      "id": "dir-src",
      "label": "src",
      "group": "Dir",
      "parent": "dir-root"
    },
    {
      "id": "dir-src/backend",
      "label": "backend",
      "group": "Dir",
      "parent": "dir-src"
    },
    {
      "id": "code-backend-agent",
      "label": "code-backend-agent",
      "group": "Agent",
      "parent": "dir-src/backend"
    },
    {
      "id": "file-src/backend/payment.go",
      "label": "payment.go",
      "group": "File",
      "parent": "dir-src/backend"
    },
    {
      "id": "dir-src/frontend",
      "label": "frontend",
      "group": "Dir",
      "parent": "dir-src"
    },
    {
      "id": "code-frontend-agent",
      "label": "code-frontend-agent",
      "group": "Agent",
      "parent": "dir-src/frontend"
    },
    {
      "id": "file-src/frontend/payment.js",
      "label": "payment.js",
      "group": "File",
      "parent": "dir-src/frontend"
    },
    {
      "id": "drd-agent-v1",
      "label": "Drd v1",
      "group": "Version",
      "parent": "dir-docs/drd"
    },
    {
      "id": "code-frontend-agent-v1",
      "label": "Frontend v1",
      "group": "Version",
      "parent": "dir-src/frontend"
    },
    {
      "id": "code-backend-agent-v1",
      "label": "Backend v1",
      "group": "Version",
      "parent": "dir-src/backend"
    },
    {
      "id": "drd-agent-v2",
      "label": "Drd v2",
      "group": "Version",
      "parent": "dir-docs/drd"
    },
    {
      "id": "code-frontend-agent-v2",
      "label": "Frontend v2",
      "group": "Version",
      "parent": "dir-src/frontend"
    },
    {
      "id": "code-backend-agent-v2",
      "label": "Backend v2",
      "group": "Version",
      "parent": "dir-src/backend"
    },
    {
      "id": "drd-agent-v3",
      "label": "Drd v3",
      "group": "Version",
      "parent": "dir-docs/drd"
    },
    {
      "id": "code-frontend-agent-v3",
      "label": "Frontend v3",
      "group": "Version",
      "parent": "dir-src/frontend"
    },
    {
      "id": "code-backend-agent-v3",
      "label": "Backend v3",
      "group": "Version",
      "parent": "dir-src/backend"
    }
  ],
  "graphEdges": [
    {
      "id": "trigger-drd-agent-code-frontend-agent",
      "from": "drd-agent",
      "to": "code-frontend-agent",
      "type": "trigger",
      "label": "drd-agent→code-frontend-agent",
      "events": [
        {
          "time": "11:05:58",
          "traceId": "iter-001",
          "payload": {
            "design": "v1",
            "action": "实现支付页面"
          }
        },
        {
          "time": "11:15:58",
          "traceId": "iter-002",
          "payload": {
            "design": "v2",
            "action": "新增退款页面"
          }
        },
        {
          "time": "11:25:58",
          "traceId": "iter-003",
          "payload": {
            "design": "v3",
            "converged": true
          }
        }
      ]
    },
    {
      "id": "feedback-code-frontend-agent-drd-agent",
      "from": "code-frontend-agent",
      "to": "drd-agent",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "11:10:58",
          "traceId": "iter-001",
          "payload": {
            "issue": "缺少退款入口",
            "suggestion": "新增退款单据字段"
          }
        },
        {
          "time": "11:20:58",
          "traceId": "iter-002",
          "payload": {
            "issue": "退款表单校验不匹配",
            "suggestion": "统一校验规则"
          }
        }
      ]
    },
    {
      "id": "trigger-drd-agent-code-backend-agent",
      "from": "drd-agent",
      "to": "code-backend-agent",
      "type": "trigger",
      "label": "drd-agent→code-backend-agent",
      "events": [
        {
          "time": "11:05:58",
          "traceId": "iter-001",
          "payload": {
            "design": "v1",
            "action": "实现支付接口"
          }
        },
        {
          "time": "11:15:58",
          "traceId": "iter-002",
          "payload": {
            "design": "v2",
            "action": "新增退款接口"
          }
        },
        {
          "time": "11:25:58",
          "traceId": "iter-003",
          "payload": {
            "design": "v3",
            "converged": true
          }
        }
      ]
    },
    {
      "id": "feedback-code-backend-agent-drd-agent",
      "from": "code-backend-agent",
      "to": "drd-agent",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "11:12:58",
          "traceId": "iter-001",
          "payload": {
            "issue": "超时参数未定义",
            "suggestion": "增加 timeout"
          }
        },
        {
          "time": "11:21:58",
          "traceId": "iter-002",
          "payload": {
            "issue": "接口返回格式不一致",
            "suggestion": "统一响应格式"
          }
        }
      ]
    },
    {
      "id": "v-drd-agent-v1-code-frontend-agent-v1",
      "from": "drd-agent-v1",
      "to": "code-frontend-agent-v1",
      "type": "trigger",
      "label": "触发",
      "events": [
        {
          "time": "11:05:58",
          "traceId": "iter-001",
          "payload": {
            "design": "v1",
            "action": "实现支付页面"
          }
        }
      ]
    },
    {
      "id": "v-drd-agent-v1-code-backend-agent-v1",
      "from": "drd-agent-v1",
      "to": "code-backend-agent-v1",
      "type": "trigger",
      "label": "触发",
      "events": [
        {
          "time": "11:05:58",
          "traceId": "iter-001",
          "payload": {
            "design": "v1",
            "action": "实现支付接口"
          }
        }
      ]
    },
    {
      "id": "v-code-frontend-agent-v1-drd-agent-v2",
      "from": "code-frontend-agent-v1",
      "to": "drd-agent-v2",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "11:10:58",
          "traceId": "iter-001",
          "payload": {
            "issue": "缺少退款入口",
            "suggestion": "新增退款单据字段"
          }
        }
      ]
    },
    {
      "id": "v-code-backend-agent-v1-drd-agent-v2",
      "from": "code-backend-agent-v1",
      "to": "drd-agent-v2",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "11:12:58",
          "traceId": "iter-001",
          "payload": {
            "issue": "超时参数未定义",
            "suggestion": "增加 timeout"
          }
        }
      ]
    },
    {
      "id": "v-drd-agent-v2-code-frontend-agent-v2",
      "from": "drd-agent-v2",
      "to": "code-frontend-agent-v2",
      "type": "trigger",
      "label": "触发",
      "events": [
        {
          "time": "11:15:58",
          "traceId": "iter-002",
          "payload": {
            "design": "v2",
            "action": "新增退款页面"
          }
        }
      ]
    },
    {
      "id": "v-drd-agent-v2-code-backend-agent-v2",
      "from": "drd-agent-v2",
      "to": "code-backend-agent-v2",
      "type": "trigger",
      "label": "触发",
      "events": [
        {
          "time": "11:15:58",
          "traceId": "iter-002",
          "payload": {
            "design": "v2",
            "action": "新增退款接口"
          }
        }
      ]
    },
    {
      "id": "v-code-frontend-agent-v2-drd-agent-v3",
      "from": "code-frontend-agent-v2",
      "to": "drd-agent-v3",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "11:20:58",
          "traceId": "iter-002",
          "payload": {
            "issue": "退款表单校验不匹配",
            "suggestion": "统一校验规则"
          }
        }
      ]
    },
    {
      "id": "v-code-backend-agent-v2-drd-agent-v3",
      "from": "code-backend-agent-v2",
      "to": "drd-agent-v3",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "11:21:58",
          "traceId": "iter-002",
          "payload": {
            "issue": "接口返回格式不一致",
            "suggestion": "统一响应格式"
          }
        }
      ]
    },
    {
      "id": "v-drd-agent-v3-code-frontend-agent-v3",
      "from": "drd-agent-v3",
      "to": "code-frontend-agent-v3",
      "type": "trigger",
      "label": "触发",
      "events": [
        {
          "time": "11:25:58",
          "traceId": "iter-003",
          "payload": {
            "design": "v3",
            "converged": true
          }
        }
      ]
    },
    {
      "id": "v-drd-agent-v3-code-backend-agent-v3",
      "from": "drd-agent-v3",
      "to": "code-backend-agent-v3",
      "type": "trigger",
      "label": "触发",
      "events": [
        {
          "time": "11:25:58",
          "traceId": "iter-003",
          "payload": {
            "design": "v3",
            "converged": true
          }
        }
      ]
    }
  ]
};
