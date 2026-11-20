# Account / Config 领域

本地配置检查：

```sh
moi-cli -o -no-interactive config show
moi-cli -o -no-interactive account browser-session
moi-cli -o -no-interactive account preferences
```

`config show` 只用于离线诊断并应保持 PAT 脱敏；不要把配置值复制进对话。`config set-*`、
account preferences 更新、浏览器登录/登出和 API key 变更属于身份或持久配置操作，除非用户
明确要求，不自动执行。认证缺失时让用户在本机运行 `astra login`，不读取凭证文件。
