#!/usr/bin/env python3
"""生成资产智能体协作图数据 (v2.1 精简版)"""
import os, json, argparse, yaml
from datetime import datetime, timedelta
from fnmatch import fnmatch

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('root', help='资产仓库根目录')
    parser.add_argument('--rules', default='agent_rules.yaml')
    parser.add_argument('--output', default='data.json')
    parser.add_argument('--seed-traces', action='store_true')
    args = parser.parse_args()

    # 1. 加载规则
    with open(args.rules) as f:
        agents = yaml.safe_load(f).get('agents', [])
    agent_by_id = {a['id']: a for a in agents}

    # 2. 构建节点
    nodes = []
    root_name = os.path.basename(os.path.normpath(args.root))
    nodes.append({'id': 'dir-root', 'label': root_name, 'group': 'Dir', 'parent': None})

    def scan(path, parent_id):
        for entry in sorted(os.listdir(path)):
            if entry.startswith('.') or entry.startswith('~'):
                continue
            full = os.path.join(path, entry)
            if os.path.isdir(full):
                dir_id = f"dir-{full.replace(args.root, '').strip(os.sep).replace(os.sep, '/')}"
                nodes.append({'id': dir_id, 'label': entry, 'group': 'Dir', 'parent': parent_id})
                # 检查该目录是否有 agent 挂载
                rel = os.path.relpath(full, args.root)
                for agent in agents:
                    if agent.get('path') == rel:
                        nodes.append({'id': agent['id'], 'label': agent['id'], 'group': 'Agent', 'parent': dir_id})
                scan(full, dir_id)
            else:
                file_id = f"file-{full.replace(args.root, '').strip(os.sep).replace(os.sep, '/')}"
                nodes.append({'id': file_id, 'label': entry, 'group': 'File', 'parent': parent_id})

    # 根目录自身的 agent
    for agent in agents:
        if agent.get('path') == '.':
            nodes.append({'id': agent['id'], 'label': agent['id'], 'group': 'Agent', 'parent': 'dir-root'})
    scan(args.root, 'dir-root')

    # 3. 构建边（写入 + 触发 + 反馈）
    edges = []
    # 文件节点索引：filename -> [node_id]
    file_index = {}
    for n in nodes:
        if n['group'] == 'File':
            file_index.setdefault(n['label'], []).append(n['id'])

    for agent in agents:
        aid = agent['id']
        # 写入边
        for out_pat in agent.get('outputs', []):
            fname = os.path.basename(out_pat)
            for fid in file_index.get(fname, []):
                edges.append({'id': f'write-{aid}-{fid}', 'from': aid, 'to': fid, 'type': 'write', 'label': '写入', 'events': []})

        # 触发边：如果 A 的 output 文件名被 B 的 watch 包含
        for other in agents:
            if other['id'] == aid: continue
            for out_pat in agent.get('outputs', []):
                fname = os.path.basename(out_pat)
                for watch_pat in other.get('watch', []):
                    if fnmatch(fname, os.path.basename(watch_pat)):
                        eid = f'trigger-{aid}-{other["id"]}'
                        if not any(e['id'] == eid for e in edges):
                            edges.append({'id': eid, 'from': aid, 'to': other['id'], 'type': 'trigger', 'label': f'{aid}→{other["id"]}', 'events': []})
                            # 同时添加反向反馈边
                            edges.append({'id': f'feedback-{other["id"]}-{aid}', 'from': other['id'], 'to': aid, 'type': 'feedback', 'label': '反馈', 'events': []})

    # 4. 注入模拟 Trace（如果启用）
    if args.seed_traces:
        # 创建 from->to->type 到 edge id 的快速查找
        edge_lookup = {}
        for e in edges:
            if e['type'] in ('trigger', 'feedback'):
                edge_lookup[(e['from'], e['to'], e['type'])] = e['id']

        base = datetime.utcnow() - timedelta(minutes=30)
        traces = [
            ('trace-001', [
                ('meta-agent', 'platform-agent', 'trigger', 0,  {'plan': 'v2'}),
                ('platform-agent', 'drd-agent', 'trigger', 2,  {'task': '支付模块设计'}),
                ('drd-agent', 'qa-agent', 'trigger', 5,         {'status': 'final'}),
                ('qa-agent', 'drd-agent', 'feedback', 8,        {'issue': '缺少超时'}),
                ('drd-agent', 'code-agent', 'trigger', 10,      {'status': 'final'}),
                ('code-agent', 'test-agent', 'trigger', 15,     {'status': 'implemented'}),
            ]),
            ('trace-002', [
                ('meta-agent', 'platform-agent', 'trigger', 1,  {'plan': 'v2.1'}),
                ('platform-agent', 'drd-agent', 'trigger', 3,   {'task': '退款流程'}),
                ('drd-agent', 'code-agent', 'trigger', 6,        {'status': 'final'}),
                ('code-agent', 'test-agent', 'trigger', 9,       {'status': 'implemented'}),
                ('test-agent', 'code-agent', 'feedback', 12,     {'bug': '边界崩溃'}),
            ]),
            ('trace-003', [
                ('meta-agent', 'lib-agent', 'trigger', 0,        {'task': '公共库重构'}),
                ('lib-agent', 'drd-agent', 'trigger', 4,         {'design': '接口变更'}),
            ])
        ]

        for trace_id, steps in traces:
            for from_id, to_id, etype, delta, payload in steps:
                eid = edge_lookup.get((from_id, to_id, etype))
                if not eid:  # 宽松匹配
                    eid = edge_lookup.get((from_id, to_id, 'trigger')) or edge_lookup.get((from_id, to_id, 'feedback'))
                if eid:
                    for e in edges:
                        if e['id'] == eid:
                            e.setdefault('events', []).append({
                                'time': (base + timedelta(minutes=delta)).strftime('%H:%M:%S'),
                                'traceId': trace_id,
                                'payload': payload
                            })
                            break
                else:
                    print(f'警告: 未找到边 {from_id}->{to_id} ({etype})')

    # 5. 输出
    ext = os.path.splitext(args.output)[1]
    data = json.dumps({'graphNodes': nodes, 'graphEdges': edges}, ensure_ascii=False, indent=2)
    if ext == '.js':
        data = f'window.graphData = {data};\n'
    with open(args.output, 'w') as f:
        f.write(data)
    print(f'✅ 已生成 {args.output} (节点 {len(nodes)}, 边 {len(edges)})')

if __name__ == '__main__':
    main()