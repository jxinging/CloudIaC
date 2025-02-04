# 组织

## 什么是组织

组织是CloudIaC中最高层级的逻辑实体；

多个组织间数据隔离。

每个组织下设有项目、VCS、云模板、变量、环境、Token、通知等管理；

通常情况下一个用户属于一个组织，同时支持一个用户加入多个组织，并在每个组织下可以具有不同的角色和权限。

## 演示组织

CloudIaC默认创建一个『演示组织』，该组织及下的项目权限受限，主要用于环境演示和实验；

所有用户均可访问演示组织，并可以在演示组织中基于演示云模板进行环境部署等操作；

## 选择活动组织

用户首次登录CloudIaC后默认进入演示组织；

顶部导航栏的组织下拉列表中可以列出当前用户所属的组织，通过选择下拉列表中的组织名称可以切换活动组织。

## 加入组织

用户可以通过以下方式加入组织：

1. 通过组织管理员邀请，组织管理员可以设置用户是管理员还是普通用户
2. 平台管理在创建组织时邀请，被邀请的用户将成为新组织的管理员

被邀请加入组织的用户将收到邮件通知。

## 创建组织

只有平台管理员可以创建组织和编辑组织信息；

要创建新组织，请选择右上角下拉菜单的『系统设置』-『组织管理』，然后单击*创建组织*。

## 组织设置
**组织**是 CloudIaC 中配置的最高范围，所有其他实体都继承组织的设置;

只有组织管理员才能更改该组织的设置或邀请新用户加入该组织。

#### SSH 密钥
cloudiac 提供 ssh 密钥管理功能，可以在组织中添加 ssh 密钥，以支持 playbook 的执行。

cloudiac 不生成密钥，只支持添加己生成的 ssh 私钥。私钥可以是您手动生成，或者在云商平台创建后导出。

因为我们使用 ssh 连接执行 playbook，所以若环境配置了 playbook 时，则必须为其配置 ssh 密钥。

为了能进行 ssh 认证，还需要在创建计算资源时绑定对应的公钥。同时，要为计算资源绑定公钥通常需要先在云商创建 ssh 密钥对。以上过程需要您通过云模板进行配置或者手动创建。

可以查看 cloudiac-example 示例模板代码了解如何在模板中配置 ssh 密钥对。

#### 通知
通知功能允许您设置关注的事件，并在事件发生时通过配置的渠道发送通知消息。

**目前支持的事件类型有:**

- 发起部署
- 等待审批
- 部署失败
- 部署成功

**目前支持的通知渠道有:**

- 邮件
- 企业微信
- 钉钉
- Slack
