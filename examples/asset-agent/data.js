window.graphData = {
  "graphNodes": [
    {
      "id": "dir-root",
      "label": "sample",
      "group": "Dir",
      "parent": null
    },
    {
      "id": "meta-agent",
      "label": "meta-agent",
      "group": "Agent",
      "parent": "dir-root"
    },
    {
      "id": "dir-roadmap",
      "label": "roadmap",
      "group": "Dir",
      "parent": "dir-root"
    },
    {
      "id": "file-roadmap/未命名.md",
      "label": "未命名.md",
      "group": "File",
      "parent": "dir-roadmap"
    },
    {
      "id": "dir-平台",
      "label": "平台",
      "group": "Dir",
      "parent": "dir-root"
    },
    {
      "id": "platform-agent",
      "label": "platform-agent",
      "group": "Agent",
      "parent": "dir-平台"
    },
    {
      "id": "file-平台/PLAN.md",
      "label": "PLAN.md",
      "group": "File",
      "parent": "dir-平台"
    },
    {
      "id": "file-平台/STATUS.md",
      "label": "STATUS.md",
      "group": "File",
      "parent": "dir-平台"
    },
    {
      "id": "dir-平台/config",
      "label": "config",
      "group": "Dir",
      "parent": "dir-平台"
    },
    {
      "id": "config-agent",
      "label": "config-agent",
      "group": "Agent",
      "parent": "dir-平台/config"
    },
    {
      "id": "dir-平台/docs",
      "label": "docs",
      "group": "Dir",
      "parent": "dir-平台"
    },
    {
      "id": "dir-平台/docs/drd",
      "label": "drd",
      "group": "Dir",
      "parent": "dir-平台/docs"
    },
    {
      "id": "drd-agent",
      "label": "drd-agent",
      "group": "Agent",
      "parent": "dir-平台/docs/drd"
    },
    {
      "id": "file-平台/docs/drd/支付设计.md",
      "label": "支付设计.md",
      "group": "File",
      "parent": "dir-平台/docs/drd"
    },
    {
      "id": "dir-平台/docs/qa",
      "label": "qa",
      "group": "Dir",
      "parent": "dir-平台/docs"
    },
    {
      "id": "qa-agent",
      "label": "qa-agent",
      "group": "Agent",
      "parent": "dir-平台/docs/qa"
    },
    {
      "id": "file-平台/docs/qa/测试策略.md",
      "label": "测试策略.md",
      "group": "File",
      "parent": "dir-平台/docs/qa"
    },
    {
      "id": "dir-平台/src",
      "label": "src",
      "group": "Dir",
      "parent": "dir-平台"
    },
    {
      "id": "code-agent",
      "label": "code-agent",
      "group": "Agent",
      "parent": "dir-平台/src"
    },
    {
      "id": "file-平台/src/payment.ts",
      "label": "payment.ts",
      "group": "File",
      "parent": "dir-平台/src"
    },
    {
      "id": "dir-平台/test",
      "label": "test",
      "group": "Dir",
      "parent": "dir-平台"
    },
    {
      "id": "test-agent",
      "label": "test-agent",
      "group": "Agent",
      "parent": "dir-平台/test"
    },
    {
      "id": "dir-库",
      "label": "库",
      "group": "Dir",
      "parent": "dir-root"
    },
    {
      "id": "lib-agent",
      "label": "lib-agent",
      "group": "Agent",
      "parent": "dir-库"
    },
    {
      "id": "dir-示例",
      "label": "示例",
      "group": "Dir",
      "parent": "dir-root"
    },
    {
      "id": "example-agent",
      "label": "example-agent",
      "group": "Agent",
      "parent": "dir-示例"
    }
  ],
  "graphEdges": [
    {
      "id": "write-meta-agent-file-平台/PLAN.md",
      "from": "meta-agent",
      "to": "file-平台/PLAN.md",
      "type": "write",
      "label": "写入",
      "events": []
    },
    {
      "id": "trigger-meta-agent-platform-agent",
      "from": "meta-agent",
      "to": "platform-agent",
      "type": "trigger",
      "label": "meta-agent→platform-agent",
      "events": [
        {
          "time": "09:42:12",
          "traceId": "trace-001",
          "payload": {
            "plan": "v2"
          }
        },
        {
          "time": "09:43:12",
          "traceId": "trace-002",
          "payload": {
            "plan": "v2.1"
          }
        }
      ]
    },
    {
      "id": "feedback-platform-agent-meta-agent",
      "from": "platform-agent",
      "to": "meta-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-meta-agent-drd-agent",
      "from": "meta-agent",
      "to": "drd-agent",
      "type": "trigger",
      "label": "meta-agent→drd-agent",
      "events": []
    },
    {
      "id": "feedback-drd-agent-meta-agent",
      "from": "drd-agent",
      "to": "meta-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-meta-agent-qa-agent",
      "from": "meta-agent",
      "to": "qa-agent",
      "type": "trigger",
      "label": "meta-agent→qa-agent",
      "events": []
    },
    {
      "id": "feedback-qa-agent-meta-agent",
      "from": "qa-agent",
      "to": "meta-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-meta-agent-code-agent",
      "from": "meta-agent",
      "to": "code-agent",
      "type": "trigger",
      "label": "meta-agent→code-agent",
      "events": []
    },
    {
      "id": "feedback-code-agent-meta-agent",
      "from": "code-agent",
      "to": "meta-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-meta-agent-config-agent",
      "from": "meta-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "meta-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-meta-agent",
      "from": "config-agent",
      "to": "meta-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "write-platform-agent-file-平台/STATUS.md",
      "from": "platform-agent",
      "to": "file-平台/STATUS.md",
      "type": "write",
      "label": "写入",
      "events": []
    },
    {
      "id": "trigger-platform-agent-meta-agent",
      "from": "platform-agent",
      "to": "meta-agent",
      "type": "trigger",
      "label": "platform-agent→meta-agent",
      "events": []
    },
    {
      "id": "feedback-meta-agent-platform-agent",
      "from": "meta-agent",
      "to": "platform-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-platform-agent-drd-agent",
      "from": "platform-agent",
      "to": "drd-agent",
      "type": "trigger",
      "label": "platform-agent→drd-agent",
      "events": [
        {
          "time": "09:44:12",
          "traceId": "trace-001",
          "payload": {
            "task": "支付模块设计"
          }
        },
        {
          "time": "09:45:12",
          "traceId": "trace-002",
          "payload": {
            "task": "退款流程"
          }
        }
      ]
    },
    {
      "id": "feedback-drd-agent-platform-agent",
      "from": "drd-agent",
      "to": "platform-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-platform-agent-qa-agent",
      "from": "platform-agent",
      "to": "qa-agent",
      "type": "trigger",
      "label": "platform-agent→qa-agent",
      "events": []
    },
    {
      "id": "feedback-qa-agent-platform-agent",
      "from": "qa-agent",
      "to": "platform-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-platform-agent-code-agent",
      "from": "platform-agent",
      "to": "code-agent",
      "type": "trigger",
      "label": "platform-agent→code-agent",
      "events": []
    },
    {
      "id": "feedback-code-agent-platform-agent",
      "from": "code-agent",
      "to": "platform-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-platform-agent-config-agent",
      "from": "platform-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "platform-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-platform-agent",
      "from": "config-agent",
      "to": "platform-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-drd-agent-meta-agent",
      "from": "drd-agent",
      "to": "meta-agent",
      "type": "trigger",
      "label": "drd-agent→meta-agent",
      "events": []
    },
    {
      "id": "feedback-meta-agent-drd-agent",
      "from": "meta-agent",
      "to": "drd-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-drd-agent-qa-agent",
      "from": "drd-agent",
      "to": "qa-agent",
      "type": "trigger",
      "label": "drd-agent→qa-agent",
      "events": [
        {
          "time": "09:47:12",
          "traceId": "trace-001",
          "payload": {
            "status": "final"
          }
        }
      ]
    },
    {
      "id": "feedback-qa-agent-drd-agent",
      "from": "qa-agent",
      "to": "drd-agent",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "09:50:12",
          "traceId": "trace-001",
          "payload": {
            "issue": "缺少超时"
          }
        }
      ]
    },
    {
      "id": "trigger-drd-agent-code-agent",
      "from": "drd-agent",
      "to": "code-agent",
      "type": "trigger",
      "label": "drd-agent→code-agent",
      "events": [
        {
          "time": "09:52:12",
          "traceId": "trace-001",
          "payload": {
            "status": "final"
          }
        },
        {
          "time": "09:48:12",
          "traceId": "trace-002",
          "payload": {
            "status": "final"
          }
        }
      ]
    },
    {
      "id": "feedback-code-agent-drd-agent",
      "from": "code-agent",
      "to": "drd-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-drd-agent-config-agent",
      "from": "drd-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "drd-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-drd-agent",
      "from": "config-agent",
      "to": "drd-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-qa-agent-meta-agent",
      "from": "qa-agent",
      "to": "meta-agent",
      "type": "trigger",
      "label": "qa-agent→meta-agent",
      "events": []
    },
    {
      "id": "feedback-meta-agent-qa-agent",
      "from": "meta-agent",
      "to": "qa-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-qa-agent-drd-agent",
      "from": "qa-agent",
      "to": "drd-agent",
      "type": "trigger",
      "label": "qa-agent→drd-agent",
      "events": []
    },
    {
      "id": "feedback-drd-agent-qa-agent",
      "from": "drd-agent",
      "to": "qa-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-qa-agent-code-agent",
      "from": "qa-agent",
      "to": "code-agent",
      "type": "trigger",
      "label": "qa-agent→code-agent",
      "events": []
    },
    {
      "id": "feedback-code-agent-qa-agent",
      "from": "code-agent",
      "to": "qa-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-qa-agent-config-agent",
      "from": "qa-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "qa-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-qa-agent",
      "from": "config-agent",
      "to": "qa-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-code-agent-test-agent",
      "from": "code-agent",
      "to": "test-agent",
      "type": "trigger",
      "label": "code-agent→test-agent",
      "events": [
        {
          "time": "09:57:12",
          "traceId": "trace-001",
          "payload": {
            "status": "implemented"
          }
        },
        {
          "time": "09:51:12",
          "traceId": "trace-002",
          "payload": {
            "status": "implemented"
          }
        }
      ]
    },
    {
      "id": "feedback-test-agent-code-agent",
      "from": "test-agent",
      "to": "code-agent",
      "type": "feedback",
      "label": "反馈",
      "events": [
        {
          "time": "09:54:12",
          "traceId": "trace-002",
          "payload": {
            "bug": "边界崩溃"
          }
        }
      ]
    },
    {
      "id": "trigger-code-agent-config-agent",
      "from": "code-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "code-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-code-agent",
      "from": "config-agent",
      "to": "code-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-test-agent-config-agent",
      "from": "test-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "test-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-test-agent",
      "from": "config-agent",
      "to": "test-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    },
    {
      "id": "trigger-lib-agent-config-agent",
      "from": "lib-agent",
      "to": "config-agent",
      "type": "trigger",
      "label": "lib-agent→config-agent",
      "events": []
    },
    {
      "id": "feedback-config-agent-lib-agent",
      "from": "config-agent",
      "to": "lib-agent",
      "type": "feedback",
      "label": "反馈",
      "events": []
    }
  ]
};
