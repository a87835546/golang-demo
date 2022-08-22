
## 6.git使用的规范
1. 分支管理 -- 同一分支 强制使用 git pull 使用 rebase
    * master 分支 -- 主分支，对应当前线上版本
    * develop 分支 -- 开发分支
    * feature 分支 -- 功能分支，开发新功能的分支，并且由develop分支切出来, feature/new_task,开发完成后需要删除。
    * release 分支 -- 发布分支，新功能合并到 develop 分支，准备发布新版本时使用的分支
    * hotfix 分支 -- 紧急修复线上 bug 分支
2. 提交信息规范 --  不要随意添加commit，必须说明此commit的功能 example: (fix: 用户登录参数校验)
    * feat/add: 新功能
    * fix: 修复 bug
    * update: 更新内容
    * docs: 文档变动
    * style: 格式调整，对代码实际运行没有改动，例如添加空行、格式化等
    * refactor: bug 修复和添加新功能之外的代码改动
    * perf: 提升性能的改动
    * test: 添加或修正测试代码
    * chore: 构建过程或辅助工具和库（如文档生成）的更改