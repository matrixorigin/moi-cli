# Recipe：审查并变更 IAM

这个流程适用于角色、权限视图、对象授权和 user-role binding。IAM 结论只能来自服务端返回，
Skill 不自行合并角色、owner 或 grants 推断 allow/deny。

1. 固定 workspace ID，读取 `iam current-principal-roles`、`current-principal-capabilities` 和 `permission-schema`。
2. 对目标角色执行 `roles` 与 `role-permission-view --role-id`；对目标资源或对象授权读取对应详情。
3. 需要变更时，先保存当前 role/binding/grant 的真实 ID、版本和生命周期状态，确认是否存在 CAS 要求。
4. 使用目标命令的 `schema` 生成输入文件，并以 `--validate-file` 做本地结构校验；不得拼造版本或权限字段。
5. 明确写入对象、workspace、用户/角色、授权动作和回滚方式后再提交。跨 workspace 变更必须分别固定两侧 ID。
6. 提交后重新读取角色、binding 或 grant 详情，并用权限过滤列表或受控只读请求核验实际可见性。
7. 如果 Core/PDP 不可用、资源无法解析、版本冲突或返回未知状态，按 fail-closed 处理，保留原始 `code`/`reason`/`message`，不加本地 allow fallback。

删除或收回授权前必须先执行产品提供的 preflight/impact 查询；没有影响分析命令时，不把删除当作可安全自动化动作。

## 隔离角色的已验证边界

create-role 使用实时 permission-schema 的 schema_version、唯一 request_id、name、statements 和 data_privilege_mutations。
两类授权均为空会被拒绝；不要把400误报认证失败。无成员的隔离角色可用一条明确目标表的 table.read 应用权限验证管理链路，
但不能据此承诺表SQL可读：table.query/SELECT的数据面授权是另一分支。
按实时Schema把应用权限放入statements、SELECT放入data_privilege_mutations；不能把table.query塞进普通应用语句。

角色创建后回读 role-permission-view、role-members 和 data-privileges，确认没有额外授权或成员。
update-role-lifecycle 的JSON为 `{expected_lifecycle_version, target_state, reason_code, request_id}`；
expected_lifecycle_version必须来自最新服务返回。旧版本应冲突且原状态保持，不能重读后静默覆盖。
使用role-permission-view的整数expected_role_lifecycle_version或角色列表的lifecycle_version；
不要用view中用于诊断的字符串lifecycle_version。data-privileges需独立核对items中的目标表/action/privilege/source/drift_status，
不能用应用权限view代替数据权限回读。
测试生命周期active→disabled→active，不绑定真实成员。删除前 role-delete-impact 检查 binding/default/inheritance/pending 等，
仅在本次角色无引用且范围已授权时删除；最后角色列表应与基线一致。

table + select + grant 的数据权限变更要求非空 `row_column`；不改变现有行列规则时按正式合同显式提供
`"row_column":{"mode":"keep"}`，不能省略，也不能把keep解释为清除限制或任意选择新规则。
授权变更采用实时permission-schema及真实目标，下面仅示意这一条mutation的结构：

```json
{
  "operation":"grant", "action_id":"table.query", "resource_type":"table",
  "resource_id":"实际表ID", "privilege":"select", "with_grant_option":false,
  "row_column":{"mode":"keep"}
}
```

2026-10-30 应用Role生命周期/CAS/impact/删除已通过。首次SELECT请求缺少row_column.mode，
返回ErrIAMRoleWriteRolledBack且无角色残留；按合同补keep的一次重试成功，独立data-privileges回读
确认目标表select、source=matrixone、drift_status=in_sync。测试角色零成员，删除后列表回到基线。
因此该项是补齐输入后写入/回读通过，不再标为未知服务端故障。session_invalidation_status=unknown仍保留；
未切换普通成员身份，不能据角色配置正确声称普通角色SQL的allow/deny或会话失效已通过。

## 当前CLI响应兼容与失败后的实际权限

2026-11-02在正式0.1.8上，`current-principal-capabilities`因未知响应字段
`unsupported_capability_ids`解码失败，`role-permission-view`因`lifecycle_state`失败。
这不是未登录或空策略；重复登录不能补齐响应类型。角色列表、graph、impact及data-privileges
仍可作各自范围的核对，但不能替代完整策略正文。缺完整版本/策略回读时，不根据诊断值猜版本并继续用户权限写入。
历史角色配置通过不代表当前这两个入口仍能解析；先按当前真实返回判断。

零成员隔离Role的独立应用策略写入取得policy3；原policy2重交被409/ErrIAMPolicyConflict拒绝，
impact保持两条授权，但完整策略正文仍无法回读。因此只确认此管理路径的限定写入/冲突证据，
不承诺统一编辑、继承或普通主体allow/deny。禁用角色的策略写入也被拒绝，禁用状态保持。

SELECT grant的`row_column.mode=keep`已有限定创建成功证据，不能照搬到revoke。
本地业务owner提示revoke应clear、set从draft构建；本地代码不等于线上版本。先查正式schema及合法业务例子，
不把结构校验通过当业务有效。本轮canonical-only和补draft的set都返回400，未建立可推荐的行列变更方法；
继承add返回400/ErrResponseInvalid且graph未变，实际原因未确认，不宣称后端没有继承能力。

本轮另建一个获准隔离零成员Role，先确认select/in_sync，再以正式schema合法的revoke+clear和
实际policy/lifecycle/data版本提交一次：返回500/ErrIAMRoleWriteFailed；独立data-privileges仍为data版本1，
但source=moi_operation、drift_status=missing_in_mo、disabled_reason=MISSING_IN_MATRIXONE。
撤权业务验收失败，且失败后数据面状态发生变化；不能凭角色/版本不变、错误名或exit非零就断言全回滚。
此前revoke+keep的422与本次clear的500分别保留，不能混成同一个输入错误。

遇这类响应兼容、写入失败或漂移时，保存原请求/错误，沿原Role读roles、data-privileges、graph/impact，
分别说明哪些可读、哪些缺失；停止自动repair、重交或替换身份。真实用户Role需向操作者明确当前状态与未知原因，
仅对已批准的隔离对象按影响分析精确收尾。本轮自建角色/库表已清理，不表示生产Role可用删除重建恢复。

## 继承图与 `clear` 的已验证路径

继承是否存在以 `role-inheritance-graph` 为准，`role-permission-view` 只列直接授权，不能代替角色图。
2026-11-15 使用双方当前 `lifecycle_version=1`，子角色 `50004` 添加父角色 `50003` 后出现对应边，
移除后边消失、图版本恢复。角色删除前 `role-delete-impact` 也需要显式当前 expected-lifecycle-version。

同一轮先用 `data-privileges` 确认 SELECT 为 `in_sync`、`data_privilege_version=1`，再提交一次 revoke+clear：
返回 ready，数据版本升到2，SELECT项消失，trace为`db203529-3503-4c9c-8ad3-4af9063850b9`。
这证明当前初始授权和真实版本下 `clear` 可完成，不能把此前失败简单归为不支持；它仍不覆盖
`mo_fact_only`、损坏历史授权或普通成员会话失效。旧表已不存在导致的回滚属于失效输入，不是继承或撤权故障。
