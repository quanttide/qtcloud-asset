#!/usr/bin/env python3
"""生成资产智能体协作图数据 (v3)"""
import os, json, argparse
from datetime import datetime, timedelta, timezone
from fnmatch import fnmatch

AGENTS = [
    {'id': 'drd-agent', 'path': 'docs/drd', 'outputs': ['docs/drd/*.md']},
    {'id': 'code-frontend-agent', 'path': 'src/frontend', 'watch': ['docs/drd/*.md'], 'outputs': ['src/frontend/*.js']},
    {'id': 'code-backend-agent', 'path': 'src/backend', 'watch': ['docs/drd/*.md'], 'outputs': ['src/backend/*.go']},
]

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('root', help='资产仓库根目录')
    parser.add_argument('--output', default='data.json')
    parser.add_argument('--seed-traces', action='store_true')
    args = parser.parse_args()

    agents = AGENTS
    agent_by_id = {a['id']: a for a in agents}

    nodes = []
    root_name = os.path.basename(os.path.normpath(args.root))
    nodes.append({'id': 'dir-root', 'label': root_name, 'group': 'Dir', 'parent': None})

    def scan(path, parent_id):
        for entry in sorted(os.listdir(path)):
            if entry.startswith('.') or entry.startswith('~'):
                continue
            full = os.path.join(path, entry)
            if not os.path.isdir(full):
                file_id = f"file-{full.replace(args.root, '').strip(os.sep).replace(os.sep, '/')}"
                nodes.append({'id': file_id, 'label': entry, 'group': 'File', 'parent': parent_id})
                continue
            dir_id = f"dir-{full.replace(args.root, '').strip(os.sep).replace(os.sep, '/')}"
            nodes.append({'id': dir_id, 'label': entry, 'group': 'Dir', 'parent': parent_id})
            rel = os.path.relpath(full, args.root)
            for agent in agents:
                if agent.get('path') == rel:
                    nodes.append({'id': agent['id'], 'label': agent['id'], 'group': 'Agent', 'parent': dir_id})
            scan(full, dir_id)

    for agent in agents:
        if agent.get('path') == '.':
            nodes.append({'id': agent['id'], 'label': agent['id'], 'group': 'Agent', 'parent': 'dir-root'})
    scan(args.root, 'dir-root')

    edges = []
    file_index = {}
    for n in nodes:
        if n['group'] == 'File':
            file_index.setdefault(n['label'], []).append(n['id'])

    for agent in agents:
        aid = agent['id']
        for out_pat in agent.get('outputs', []):
            fname = os.path.basename(out_pat)
            for fid in file_index.get(fname, []):
                edges.append({'id': f'write-{aid}-{fid}', 'from': aid, 'to': fid, 'type': 'write', 'label': '写入', 'events': []})

        for other in agents:
            if other['id'] == aid: continue
            for out_pat in agent.get('outputs', []):
                fname = os.path.basename(out_pat)
                out_dir = os.path.dirname(out_pat)
                for watch_pat in other.get('watch', []):
                    if not fnmatch(fname, os.path.basename(watch_pat)):
                        continue
                    watch_dir = os.path.dirname(watch_pat)
                    dirs_ok = out_dir == watch_dir or (out_dir and watch_dir and out_dir.startswith(watch_dir.rstrip('/') + '/'))
                    if not dirs_ok:
                        continue
                    b_path = agent_by_id[other['id']].get('path', '')
                    a_path = agent.get('path', '.')
                    if not (out_dir == b_path or out_dir == a_path):
                        continue
                    eid = f'trigger-{aid}-{other["id"]}'
                    if not any(e['id'] == eid for e in edges):
                        edges.append({'id': eid, 'from': aid, 'to': other['id'], 'type': 'trigger', 'label': f'{aid}→{other["id"]}', 'events': []})
                    feedback_id = f'feedback-{other["id"]}-{aid}'
                    if not any(e['id'] == feedback_id for e in edges):
                        edges.append({'id': feedback_id, 'from': other['id'], 'to': aid, 'type': 'feedback', 'label': '反馈', 'events': []})

    if args.seed_traces:
        edge_lookup = {}
        for e in edges:
            if e['type'] in ('trigger', 'feedback'):
                edge_lookup[(e['from'], e['to'], e['type'])] = e['id']

        base = datetime.now(timezone.utc) - timedelta(minutes=30)
        traces = [
            ('iter-001', [
                ('drd-agent', 'code-frontend-agent', 'trigger', 0,   {'design': 'v1', 'action': '实现支付页面'}),
                ('drd-agent', 'code-backend-agent', 'trigger', 0,    {'design': 'v1', 'action': '实现支付接口'}),
                ('code-frontend-agent', 'drd-agent', 'feedback', 5,  {'issue': '缺少退款入口', 'suggestion': '新增退款单据字段'}),
                ('code-backend-agent', 'drd-agent', 'feedback', 7,   {'issue': '超时参数未定义', 'suggestion': '增加 timeout'}),
            ]),
            ('iter-002', [
                ('drd-agent', 'code-frontend-agent', 'trigger', 10,  {'design': 'v2', 'action': '新增退款页面'}),
                ('drd-agent', 'code-backend-agent', 'trigger', 10,   {'design': 'v2', 'action': '新增退款接口'}),
                ('code-frontend-agent', 'drd-agent', 'feedback', 15, {'issue': '退款表单校验不匹配', 'suggestion': '统一校验规则'}),
                ('code-backend-agent', 'drd-agent', 'feedback', 16,  {'issue': '接口返回格式不一致', 'suggestion': '统一响应格式'}),
            ]),
            ('iter-003', [
                ('drd-agent', 'code-frontend-agent', 'trigger', 20,  {'design': 'v3', 'converged': True}),
                ('drd-agent', 'code-backend-agent', 'trigger', 20,   {'design': 'v3', 'converged': True}),
            ]),
        ]

        for trace_id, steps in traces:
            for from_id, to_id, etype, delta, payload in steps:
                eid = edge_lookup.get((from_id, to_id, etype))
                if not eid:
                    eid = edge_lookup.get((from_id, to_id, 'trigger')) or edge_lookup.get((from_id, to_id, 'feedback'))
                if eid:
                    for e in edges:
                        if e['id'] == eid:
                            e.setdefault('events', []).append({
                                'time': (base + timedelta(minutes=delta)).strftime('%H:%M:%S'),
                                'traceId': trace_id, 'payload': payload
                            })
                            break
                else:
                    print(f'警告: 未找到边 {from_id}->{to_id} ({etype})')

        vc = {}
        version_nodes = {}
        version_edges = []

        for trace_id, steps in traces:
            for from_id, to_id, etype, delta, payload in steps:
                for aid in (from_id, to_id):
                    vc.setdefault(aid, 1)
                fv = vc[from_id]
                vc[to_id] = vc[from_id] if (to_id not in vc or etype == 'trigger') else vc[from_id] + 1

                for aid, ver in [(from_id, fv), (to_id, vc[to_id])]:
                    nid = f'{aid}-v{ver}'
                    if nid not in version_nodes:
                        dir_id = None
                        for n in nodes:
                            if n['group'] == 'Agent' and n['id'] == aid:
                                dir_id = n.get('parent')
                                break
                        label = aid.replace('-agent', '').replace('code-', '').replace('-', ' ').title() + f' v{ver}'
                        version_nodes[nid] = {'id': nid, 'label': label, 'group': 'Version', 'parent': dir_id}

                from_vid = f'{from_id}-v{fv}'
                to_vid = f'{to_id}-v{vc[to_id]}'
                veid = f'v-{from_vid}-{to_vid}'
                if not any(e['id'] == veid for e in version_edges):
                    version_edges.append({
                        'id': veid, 'from': from_vid, 'to': to_vid,
                        'type': etype, 'label': '反馈' if etype == 'feedback' else '触发',
                        'events': [{'time': (base + timedelta(minutes=delta)).strftime('%H:%M:%S'), 'traceId': trace_id, 'payload': payload}]
                    })

        nodes.extend(version_nodes.values())
        edges.extend(version_edges)

    ext = os.path.splitext(args.output)[1]
    data = json.dumps({'graphNodes': nodes, 'graphEdges': edges}, ensure_ascii=False, indent=2)
    if ext == '.js':
        data = f'window.graphData = {data};\n'
    with open(args.output, 'w') as f:
        f.write(data)
    print(f'✅ 已生成 {args.output} (节点 {len(nodes)}, 边 {len(edges)})')

if __name__ == '__main__':
    main()
