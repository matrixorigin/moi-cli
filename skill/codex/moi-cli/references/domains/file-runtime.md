# 沙箱 file 领域

`file` 是 task-scoped Product SDK 表面，凭证和技术 API key 由 sandbox 运行时通过环境提供，
不由 Skill 读取或写入。当前命令：

```sh
moi-cli -o file list --name report --extensions pdf,csv --page-size 100
moi-cli file upload --file ./report.pdf --publish-artifact
moi-cli file download --file-id "$file_id" --output ./downloaded.pdf
moi-cli file analyze-image --file ./prepared.jpg --question "What is shown?"
```

仅在已有活动 runtime task 和短期授权时使用；`file upload` 的 `--publish-artifact` 是命令
契约，不得省略或用 Catalog upload 替代。下载前从 `file list` 或前一步返回取得真实 file ID，
下载后检查本地文件存在性、类型和大小。`analyze-image` 只对准备好的 JPEG 和用户问题执行，
不能将结果解释成 Catalog/Knowledge 写入。

当前 `references/commands/file.md` 只生成了 `list`，其它命令需先运行 `file COMMAND --help`。
因此自动化覆盖等级为 `partial`。

`file upload --publish-artifact` 与 `catalog upload-runtime-artifact` 是同一正式owner的两个入口。
前者不要求桌面端重新实现上传；优先让已授权runtime任务在自身上下文执行。
验收应核对上传返回artifact_id确实存在于同task.artifacts，workspace file对应同task，
再按file_id下载检查字节和hash；这是历史真实任务已经通过的链路。
