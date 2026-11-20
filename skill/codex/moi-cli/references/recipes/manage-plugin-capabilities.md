# Recipe：发现、使用并核验插件能力

用于MOI产品内插件，不是Codex宿主的插件管理。先固定目标workspace、插件和允许的动作；插件有读写和外部消息能力，
安装、授权、事件触发按本次已获准范围操作，不能从“测试插件”扩展为发送消息或修改外部仓库。

## 正式引用和归属

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent search-plugins --q "$plugin_name"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent describe-plugin-capability --ref "$capability_ref"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-plugin-release --namespace "$release_namespace" --release-id "$release_id"
```

从search结果保存capability_ref、release.id、release.workspace_id及installation版本；不要自行构造引用。
`plugin://github/gh`这类连接地址不能代替capability_ref。本次使用它会400，而使用搜索返回的正式引用后描述成功。
系统发布的namespace是发现到的`system`，不因当前workspace不同而替换；缺namespace曾400，补正式值后成功。

Skill类型的描述会返回包内完整 `skill_path`；`read-plugin-skill --path` 需要的却是 Skill 根目录下的
相对路径。主文件传 `SKILL.md`，嵌套引用传 `references/...`，不能照抄 `skills/github/SKILL.md`，
否则服务端会再次拼接 Skill 根目录并形成重复目录。此前完整路径404属于调用输入错误，不是产品缺正文。
2026-11-15 改用 `--path SKILL.md` 后读取成功，正文3574字节，SHA256与发布源逐字节一致。
其它插件和其它路径仍需分别核验，不能用release可读代替正文可读验收。

把插件capability ID传给托管Skill的`get-skill-file`会404，`list-resource-meta`导出另返回401
`RUNTIME_GRANT_REJECTED`；这些都不是已验证的正文替代入口，不能拿元数据或其它资源的下载命令拼接出插件正文。

## 安装前置和实际调用

先查list-plugin-installations与已有绑定。只修改本次授权对象；既有账号连接或OAuth凭据不得输出。
系统存在的插件可能尚未安装，Agent搜索中的installation_required、prepare返回setup_required=true和pending
只证明需要安装准备。不能将外层任务failed直接认作外部MCP服务不通，也不能把pending说成安装成功。

确需安装且已授权时，先读取 `schema agent install-plugin`。使用实际release_namespace/release_id；
若来自当前任务的安装准备，携带真实setup_operation_id。connections.requirements为空的Context7本次无需外部账号，
正式安装返回installation ID/version后，再发新A2A任务查询公开文档。不要把此结果推广到需要OAuth的插件。

按A2A合同保存task ID并独立get-task。本次成功链为plugins_search发现installed/ready → 激活返回的真实工具schema →
query-docs实际调用 → 工具结果含来源正文 → 同task completed。MOI运行时返回的选择键是运行时合同，
不能把CLI的capability_ref、连接地址或别的会话中的临时选择键混用。工具结果摘要被截断时保留字节数/hash及已见内容，
不能将模型最终回答中的每句话都算已独立验证。

## 测试收尾

只在测试明确要求撤销本次安装时，按返回ID和当前version执行：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent delete-plugin-installation --id "$installation_id" --expected-version "$version"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent list-plugin-installations
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-plugin-installation --id "$installation_id"
```

本次Context7测试安装已删除，列表恢复原两个安装，单独详情404。原GitHub/企业微信安装对象保持一致；
未发送外部消息、未做OAuth或真实渠道事件。插件调用成功不证明这些分支也通过。

## 重复安装、安装版本和卸载后的新任务

安装请求不预设幂等。本轮沿同一正文路径顺序再提交返回503/PLUGIN_STORE_UNAVAILABLE，随后列表仅一份原安装、仍v1；保留真实失败，按原release/installation只读查数量和版本，不重提或推断外部MCP故障。若要证明重传字节相同，应在每次发送前绑定正文摘要；本轮只有独立正文/SHA和原argv路径，未严格绑定逐次发送字节，不能称不可变重传证明。

发布version与installation.version分别记录。本轮同release、空bindings以expected_version1更新安装为v2，旧1更新/删除均409/PLUGIN_STATE_CONFLICT，独立get保持v2。版本冲突先重读差异，不能自动按新版本覆盖他人改动；这不证明不同发布升级或有账号连接迁移可用。运行时未返回installation ID/version/digest时，不声称精确安装版本冻结已验证。

若任务要求验证撤销，按当前version卸载后另发新context任务，明确不prepare、不重装、不授权。installed搜索空、system为installation_required且无可激活capabilities时应停止；外层task completed只证明正确停止，不是插件调用成功。正常发现路径回到未安装态，不证明旧在途令牌、缓存工具或直接调用均已撤销。最终核安装列表/原绑定回基线及原completed结果保持；旧工具结果保留不等于安装仍可用。
