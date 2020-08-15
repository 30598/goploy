# 权限模型角色
一个用户必须先分配一个空间，方可登录系统

用户所有操作均在登录空间下，或有多个空间可以切换

## admin 超级管理员
拥有一切权限，想做什么就做什么

## manager 域管理员
无法添加成员，域内一切权限

## group-manager 域内组管理员
无法添加成员和服务器，域内一切权限

## member 成员
部署发布

## 权限码表
| action                              | member | group-manager | manager | admin |
| ------------------------------------| ------ | ------------- | ------- | ----- |
| 部署发布                             |   ✓    |       ✓       |    ✓    |   ✓   |
| 应用监控                             |        |       ✓       |    ✓    |   ✓   |
| 项目设置                             |        |       ✓       |    ✓    |   ✓   |
| 服务器管理                           |        |               |    ✓    |   ✓   |
| 空间管理-查看、编辑                   |        |               |    ✓    |   ✓   |
| 空间管理-新建、删除                   |        |               |         |   ✓   |
| 成员列表                             |        |               |         |   ✓   |