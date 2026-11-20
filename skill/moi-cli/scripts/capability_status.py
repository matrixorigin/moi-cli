#!/usr/bin/env python3
"""Offline status registry: validate, query scoped observations, render derived views."""
from __future__ import annotations

import argparse
from collections import Counter
from datetime import date
import hashlib
import json
from pathlib import Path
import re
import sys

STATUSES = ('passed', 'failed', 'partial', 'blocked', 'cli_missing', 'unverified')
CHANNELS = ('regular', 'candidate')
SHAPES = ('read_only', 'single_domain', 'composed', 'controlled')
BEGIN = '<!-- capability-status:begin -->'
END = '<!-- capability-status:end -->'
VIEWS = ('capability-catalog.md', 'coverage-matrix.md', 'capability-inventory.md', 'verified-capabilities.md')
REQUIRED_CAPABILITIES = ('catalog-discovery', 'sql-query', 'source-staging', 'data-ingestion', 'data-export', 'workflow-lifecycle', 'job-recovery', 'knowledge-preparation', 'dashboard-analytics', 'data-share', 'workspace-context', 'connector-lifecycle', 'compute-lifecycle', 'audit-observation', 'member-lifecycle', 'iam-authorization', 'agent-conversation', 'agent-package-resource', 'agent-skill-tool', 'agent-runtime', 'agent-automation')
REQUIRED_MENUS = ('概览', '连接器', '导入任务', '导出任务', '工作流', 'SQL 编辑器', '数据看板', 'Catalog', '计算资源', '数据分享', '知识库', 'SQL 历史', '作业', '操作日志', '用户管理', '角色权限', '智能体 / 对话', '智能体 / 成果', '智能体 / 智能体', '智能体 / 技能库', '智能体 / 工具库', '智能体 / 知识库', '智能体 / 运行环境', '智能体 / 自动化')


def require(ok, message):
    if not ok:
        raise ValueError(message)


def nonempty(value):
    return isinstance(value, str) and bool(value.strip())


def day(value):
    require(isinstance(value, str), 'date must be an ISO day string')
    require(date.fromisoformat(value).isoformat() == value, f'invalid date: {value}')
    return value


def validate(document, root):
    require(isinstance(document, dict), 'registry must be an object')
    require(document.get('schema_version') == 1 and document.get('registry_version') == 1, 'unsupported registry schema')
    sources = document.get('sources')
    require(isinstance(sources, dict) and sources, 'sources required')
    for sid, source in sources.items():
        require(isinstance(source, dict), f'{sid}: source must be an object')
        require(nonempty(sid) and nonempty(source.get('description')), 'source ID and description required')
        if source.get('kind') == 'packaged_reference':
            path = source.get('path', '')
            require(nonempty(path) and not Path(path).is_absolute() and '..' not in Path(path).parts, f'{sid}: invalid packaged reference')
            require((root / 'references' / path).is_file(), f'{sid}: packaged reference missing: {path}')
        elif source.get('kind') == 'external_evidence':
            require(source.get('availability') == 'external_not_bundled' and nonempty(source.get('origin')) and nonempty(source.get('locator')), f'{sid}: external source semantics required')
            # An evidence locator describes an external origin; never resolve it in a portable package.
        else:
            raise ValueError(f'{sid}: unsupported source kind')
        for field in ('snapshot_sha256', 'historical_body_sha256'):
            if field in source:
                require(isinstance(source[field], str) and bool(re.fullmatch('[a-f0-9]{64}', source[field])), f'{sid}: invalid {field}')
    cards = document.get('cards')
    require(isinstance(cards, list) and all(isinstance(card, dict) for card in cards), 'capability cards must be objects')
    require(tuple(card.get('capability_id') for card in cards) == REQUIRED_CAPABILITIES, 'required capability identities do not match the product registry')
    menus = document.get('menus')
    require(isinstance(menus, list) and all(isinstance(menu, dict) for menu in menus), 'menu mappings must be objects')
    require(tuple(menu.get('menu') for menu in menus) == REQUIRED_MENUS, 'required menu identities do not match the product registry')
    seen = set()
    all_observations = {}
    for card in cards:
        require(isinstance(card, dict), 'capability card must be an object')
        cid = card.get('capability_id', '')
        require(bool(re.fullmatch('[a-z0-9-]+', cid)) and cid not in seen, f'duplicate or invalid capability: {cid}')
        seen.add(cid)
        require(not ({'status', 'current_status', 'readiness', 'verified'} & card.keys()), f'{cid}: capability-wide status is forbidden')
        support = card.get('static_support', {})
        require(isinstance(support, dict), f'{cid}: static support must be an object')
        require(support.get('shape') in SHAPES and support.get('level') in ('L0', 'L1', 'L2', 'L3'), f'{cid}: invalid static support; ready/full are not verification statuses')
        require(not ({'status', 'ready', 'full'} & support.keys()), f'{cid}: static support cannot contain a verification status')
        require(all(nonempty(support.get(field)) for field in ('catalog_goal', 'object_lifecycle', 'minimum_evidence', 'risk_note')), f'{cid}: static description fields required')
        require(nonempty(card.get('risk')) and isinstance(card.get('dependencies'), list), f'{cid}: risk and dependencies required')
        observations = card.get('observations')
        require(isinstance(observations, list) and observations, f'{cid}: observations required (use explicit unverified when absent)')
        for observation in observations:
            require(isinstance(observation, dict), f'{cid}: observation must be an object')
            oid = observation.get('observation_id')
            require(nonempty(oid) and oid not in all_observations, f'duplicate observation: {oid}')
            scenario = observation.get('scenario_id', '')
            require(bool(re.fullmatch('[a-z0-9-]+', scenario)), f'{oid}: invalid scenario ID')
            require(observation.get('status') in STATUSES, f'{oid}: invalid observation status')
            cli = observation.get('cli', {})
            require(isinstance(cli, dict), f'{oid}: CLI metadata must be an object')
            require(cli.get('channel') in CHANNELS and nonempty(cli.get('version')), f'{oid}: CLI channel/version required')
            recorded = day(observation.get('recorded_on'))
            verified = observation.get('validated_on')
            if verified is not None:
                require(day(verified) <= recorded, f'{oid}: verification date after record date')
            require(nonempty(observation.get('date_note')), f'{oid}: date provenance required, including unknown dates')
            conditions = observation.get('conditions')
            require(isinstance(conditions, dict) and conditions and all(nonempty(k) and nonempty(v) for k, v in conditions.items()), f'{oid}: explicit conditions required')
            require(nonempty(observation.get('result')) and isinstance(observation.get('limits'), list) and observation['limits'] and all(nonempty(x) for x in observation['limits']), f'{oid}: result/limits required')
            kind = observation.get('evidence_kind')
            require(kind in ('reported_live', 'cli_contract', 'not_run'), f'{oid}: invalid evidence kind')
            if kind == 'not_run':
                require(verified is None and observation['status'] in ('blocked', 'unverified'), f'{oid}: an unexecuted scenario cannot claim a live verdict')
            if observation['status'] in ('passed', 'failed'):
                require(kind == 'reported_live', f'{oid}: passed/failed requires reported live evidence')
            refs = observation.get('sources')
            require(isinstance(refs, list) and refs and all(isinstance(ref, dict) and ref.get('source_id') in sources and nonempty(ref.get('record')) for ref in refs), f'{oid}: source record required')
            supersedes = observation.get('supersedes', [])
            require(isinstance(supersedes, list) and len(set(supersedes)) == len(supersedes) and all(nonempty(x) for x in supersedes), f'{oid}: invalid supersedes')
            all_observations[oid] = (cid, observation)
    grouped = {}
    replaced = set()
    for oid, (cid, observation) in all_observations.items():
        key = (cid, observation['scenario_id'], observation['cli']['channel'])
        condition = json.dumps(observation['conditions'], sort_keys=True, ensure_ascii=False)
        if key in grouped:
            require(grouped[key] == condition, f'{oid}: scenario/channel reused for different conditions; use a new scenario ID')
        grouped[key] = condition
        for previous_id in observation.get('supersedes', []):
            require(previous_id in all_observations and previous_id != oid, f'{oid}: supersedes unknown/self observation')
            prior_cid, prior = all_observations[previous_id]
            require((prior_cid, prior['scenario_id'], prior['cli']['channel']) == key and prior['conditions'] == observation['conditions'], f'{oid}: cross-capability/channel/condition supersedes forbidden')
            require(observation['recorded_on'] >= prior['recorded_on'], f'{oid}: supersedes predates previous record')
            if prior['validated_on'] and observation['validated_on']:
                require(observation['validated_on'] >= prior['validated_on'], f'{oid}: older verification cannot supersede newer verification')
            if prior['status'] == 'failed' and observation['status'] == 'passed':
                require(observation['evidence_kind'] == 'reported_live' and observation['validated_on'] is not None and any(sources[ref['source_id']]['kind'] == 'external_evidence' for ref in observation['sources']), f'{oid}: replacing a known failure needs dated same-scope external live evidence')
            replaced.add(previous_id)
    def visit(oid, chain):
        require(oid not in chain, f'{oid}: supersedes cycle')
        for previous in all_observations[oid][1].get('supersedes', []):
            visit(previous, chain | {oid})
    active_keys = set()
    def has_failed_ancestor(oid):
        return any(all_observations[prior][1]['status'] == 'failed' or has_failed_ancestor(prior)
                   for prior in all_observations[oid][1].get('supersedes', []))
    for oid, (cid, observation) in all_observations.items():
        visit(oid, set())
        if observation['status'] != 'failed' and has_failed_ancestor(oid):
            require(observation['evidence_kind'] == 'reported_live' and observation['validated_on'] is not None and any(sources[ref['source_id']]['kind'] == 'external_evidence' for ref in observation['sources']), f'{oid}: clearing a known failure needs dated same-scope external live evidence')
        if oid not in replaced:
            key = (cid, observation['scenario_id'], observation['cli']['channel'])
            require(key not in active_keys, f'{oid}: duplicate active scenario observation; explicit same-scope supersedes required')
            active_keys.add(key)
    for card in cards:
        require(all(dep in seen and dep != card['capability_id'] for dep in card['dependencies']), 'invalid dependency')
    for menu in menus:
        require(all(nonempty(menu.get(field)) for field in ('menu', 'goal', 'entry_points', 'limits', 'cli_domains', 'commands', 'coverage_limits', 'priority')) and menu.get('level') in ('L0', 'L1', 'L2', 'L3'), 'menu descriptions and level required')
        require(isinstance(menu.get('capability_ids'), list) and menu['capability_ids'] and len(set(menu['capability_ids'])) == len(menu['capability_ids']) and all(cid in seen for cid in menu['capability_ids']), 'unknown/duplicate menu capability mapping')
        require(not ({'status', 'current_status', 'readiness'} & menu.keys()), 'menu state must be derived, never manually stored')
    return {'capabilities': len(cards), 'observations': len(all_observations), 'active_observations': len(active_keys), 'menus': len(menus), 'external_sources': sum(s['kind'] == 'external_evidence' for s in sources.values())}


def active_observations(document, cards=None, history=False):
    cards = document['cards'] if cards is None else cards
    observations = [(card['capability_id'], observation) for card in cards for observation in card['observations']]
    replaced = {prior for _, observation in observations for prior in observation.get('supersedes', [])}
    return [(cid, o) for cid, o in observations if history or o['observation_id'] not in replaced]


def query(document, capability_id, scenario=None, channel='regular', conditions=None, history=False):
    cards = [c for c in document['cards'] if c['capability_id'] == capability_id]
    require(bool(cards), f'unknown capability: {capability_id}')
    matched = [(cid, o) for cid, o in active_observations(document, cards, history) if (not scenario or o['scenario_id'] == scenario) and (not channel or o['cli']['channel'] == channel) and all(o['conditions'].get(k) == v for k, v in (conditions or {}).items())]
    scoped = bool(scenario or conditions)
    statuses = [s for s in STATUSES if any(o['status'] == s for _, o in matched)]
    return {'capability_id': capability_id, 'static_support': cards[0]['static_support'], 'scope_filter': {'scenario_id': scenario, 'cli_channel': channel, 'conditions': conditions or {}}, 'needs_scope_match': not scoped or not matched, 'status': statuses[0] if scoped and matched and len(statuses) == 1 else ('unverified' if not scoped or not matched else None), 'observed_statuses': statuses, 'all_matched_passed': bool(scoped and matched and all(o['status'] == 'passed' for _, o in matched)), 'note': '仅限返回的场景/条件/渠道，默认regular；没有匹配条件时unverified，能力内其它passed不构成全通过。', 'observations': [o for _, o in matched], 'sources': {ref['source_id']: document['sources'][ref['source_id']] for _, o in matched for ref in o['sources']}}


def cell(value):
    return str(value).replace('|', '\\|').replace('\n', ' ')


def table(headers, rows):
    return '\n'.join(['| ' + ' | '.join(headers) + ' |', '|' + '|'.join('---' for _ in headers) + '|'] + ['| ' + ' | '.join(cell(x) for x in row) + ' |' for row in rows])


def state_summary(document, ids, channel):
    counts = Counter(o['status'] for cid, o in active_observations(document) if cid in ids and o['cli']['channel'] == channel)
    return ', '.join(f'{s}×{counts[s]}' for s in STATUSES if counts[s]) or 'unverified（未登记此渠道）'


def render_blocks(document, digest):
    prefix = '来源：`capability-cards.json`（SHA256 `' + digest + '`）。此段由 `scripts/capability_status.py render` 生成，修改JSON后重生成；手写说明留在标记外。\n\n'
    catalog = table(['capability_id', '业务目标', '对象生命周期', '层级/静态形态', '风险', 'regular场景状态', 'candidate场景状态', '最小证据'], [[c['capability_id'], c['static_support']['catalog_goal'], c['static_support']['object_lifecycle'], c['static_support']['level'] + '/' + c['static_support']['shape'], c['risk'], state_summary(document, [c['capability_id']], 'regular'), state_summary(document, [c['capability_id']], 'candidate'), c['static_support']['minimum_evidence']] for c in document['cards']])
    inventory = table(['产品菜单', '业务闭环', '层级', 'capability IDs', '关联能力regular场景计数', '入口', '限制'], [[m['menu'], m['goal'], m['level'], ', '.join(m['capability_ids']), state_summary(document, m['capability_ids'], 'regular'), m['entry_points'], m['limits']] for m in document['menus']])
    matrix = table(['产品菜单', 'CLI domain', '关键命令/组合', 'capability IDs', '关联能力regular场景计数', '关联能力candidate场景计数', '缺口', '优先级'], [[m['menu'], m['cli_domains'], m['commands'], ', '.join(m['capability_ids']), state_summary(document, m['capability_ids'], 'regular'), state_summary(document, m['capability_ids'], 'candidate'), m['coverage_limits'], m['priority']] for m in document['menus']])
    risks = table(['capability_id', 'scenario_id', '渠道/版本', '验证日期', '状态', '实际结果', '限制/来源'], [[cid, o['scenario_id'], o['cli']['channel'] + '/' + o['cli']['version'], o['validated_on'] or '未逐项确认；记录日' + o['recorded_on'], o['status'], o['result'], '；'.join(o['limits']) + '；来源 ' + ', '.join(ref['source_id'] + ':' + ref['record'] for ref in o['sources'])] for cid, o in active_observations(document) if o['status'] != 'passed'])
    return {name: prefix + value + '\n' for name, value in zip(VIEWS, [catalog, matrix, inventory, risks])}


def render(document, root, digest, check_only=False):
    changed = []
    for name, block in render_blocks(document, digest).items():
        path = root / 'references' / name
        text = path.read_text()
        require(text.count(BEGIN) == text.count(END) == 1, f'{name}: exactly one generated block required')
        before, remainder = text.split(BEGIN, 1)
        _, after = remainder.split(END, 1)
        expected = before + BEGIN + '\n' + block + END + after
        if expected != text:
            changed.append(name)
            if not check_only:
                path.write_text(expected)
    return changed


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest='command', required=True)
    for command in ('validate', 'query', 'render'):
        command_parser = subparsers.add_parser(command)
        command_parser.add_argument('--root', type=Path, default=Path(__file__).resolve().parents[1], help='portable skill directory; no external evidence is read')
        if command == 'query':
            command_parser.add_argument('capability_id')
            command_parser.add_argument('--scenario')
            command_parser.add_argument('--channel', choices=CHANNELS, default='regular', help='regular by default; candidate results require an explicit selection')
            command_parser.add_argument('--condition', action='append', default=[], metavar='KEY=VALUE')
            command_parser.add_argument('--history', action='store_true')
        if command == 'render':
            command_parser.add_argument('--check', action='store_true', help='fail on stale generated content; write nothing')
    args = parser.parse_args(argv)
    try:
        raw = (args.root / 'references/capability-cards.json').read_bytes()
        document = json.loads(raw)
        summary = validate(document, args.root)
        if args.command == 'query':
            conditions = {}
            for condition in args.condition:
                require('=' in condition, '--condition requires KEY=VALUE')
                key, value = condition.split('=', 1)
                require(nonempty(key) and nonempty(value) and key not in conditions, 'empty or duplicate condition key')
                conditions[key] = value
            result = query(document, args.capability_id, args.scenario, args.channel, conditions, args.history)
        elif args.command == 'render':
            changed = render(document, args.root, hashlib.sha256(raw).hexdigest(), args.check)
            result = {'valid': not changed if args.check else True, 'changed_views': changed, 'check_only': args.check}
            print(json.dumps(result, ensure_ascii=False))
            return 1 if args.check and changed else 0
        else:
            result = {'valid': True, **summary, 'external_evidence_policy': 'locator semantics only; external file existence is not required'}
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0
    except (ValueError, KeyError, TypeError, OSError) as error:
        print(json.dumps({'valid': False, 'error': str(error)}, ensure_ascii=False), file=sys.stderr)
        return 2


if __name__ == '__main__':
    sys.exit(main())
