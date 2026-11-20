# Recipe：解析节点参数与结果核验

用于文档、邮件、音视频、清洗和 XML 节点。先用
`workflow workitem-catalog-item --node-id` 读取当前版本的真实输入，再构造独立 DSL；
不要把功能清单中的名称直接当成可传参数。每个节点都要核对本次 execution 的
`get-run-result`，必要时用 `get-run-node` 读取完整 `runtime_output` 和文档元数据。

## 通用文档解析

`moi:parse` 的 native/standard/enhanced、页码和增强选项分别验收。标准档用
`options.page_selector` 选择一页，要求返回块全部属于该页且不含其它页标记。增强档用
多页合成样例，分别核对：

- `cross_page_table_merge=true` 时只有一张目标表，跨页各行完整且
  `cross_page_merged=true`；
- 图片 OCR/描述在正文或图片描述中返回预期标记，保存图片输出还需检查真实产物；
- `reading_order=xy_cut` 的块顺序与页面布局一致，多文件输入保留各自 file ID；
- native/OpenXML 的既有结果不能用 standard/enhanced 的新结果替代，三档分别留存。

2026-11-17 的 3 页 PDF 中，standard 只返回第 2 页；enhanced 只处理第 1-2 页，
A-01 至 A-04 合并为一张表，IMAGE-742 与 OCR/描述均返回。该结果证明当前样例和配置，
不推广到所有扫描件、复杂排版或模型组合。

## 邮件附件和 CSV 输出

`moi:email.archive.etl` 的 `output_mode=catalog_files` 要同时核对返回的
`output_file_ids`、文件名、目标 Volume，并下载 CSV 检查正文、主题、收发件人和来源
file ID。只在 case_result 中出现 file ID 不算文件交付成功。

含 DOCX/PDF 附件的邮件还需在输出中寻找附件独有标记。2026-11-17 生成了 5 个真实
CSV，但内容只有正文和元数据，未提取附件中的 WORD-7421 或 PDF 文本；因此 CSV 输出
通过，附件提取仍保持部分通过。JSONL 没有当前正式 `output_mode`，不能猜参数。

## 音频与视频

`moi:parser.convert.audio.rich` 当前正式选项为 auto/zh/en、auto/off/ffmpeg_afftdn
和切片参数。转写文本之外，读取 `runtime_output.documents[].metadata` 的
`segment_index/start_time/end_time/language/speaker_id`；顶层 `segments` 为空不能证明
没有切片。中文 12.57 秒样例已返回 3 个连续分段，三种 denoise 均完成。

说话人分离必须用至少两个不同说话人的音频。双人样例虽正确切成 2 段，但两段
`speaker_id=null`，所以不能宣称分离通过。视频节点当前同样只有语言、去噪和切片输入；
音轨转写成功不能证明关键帧、独立字幕或场景检测，后者没有正式配置/输出时保留缺口。

## 清洗与 XML

`moi:parser.clean.text` 分别验证 n-gram、特殊字符阈值和正常文本保留。输入全部被过滤时，
当前节点返回明确失败 `clean text removed all content under ...`，不会继续后续节点；
需要覆盖其它分支时拆分 Workflow，不能让首个预期失败掩盖后续结果。schema 没有通用
翻译输入，不能用开放字段猜一个。

XML 除正文和属性边界外，至少用 DocBook、DITA、图片引用和数据型 XML 核对结构。
2026-11-17 两类文档标题/正文正确，两个图片路径成为 `type=image` 且保留
`external_url/image_url`，数据型 XML 返回明确文本；标签属性仍未进入正文。
这只覆盖常见代表方言，不宣称支持任意出版或技术 schema。
