# Storagestats
Storagestats is a multifunctional storage status information management tool developed by ND Labs.


# 检索工具逻辑 / Retrieval Tool Logic

## 简介 / Introduction

本项目由 ND Labs 开发，旨在通过一套自动化逻辑来处理大量 deal 订单的数据检索。系统每天处理新增的 deal 订单，并按特定逻辑对其进行抽样和检索，最终通过检索成功率来评估系统性能。

This project, developed by ND Labs, aims to process a large number of deal orders through an automated logic system. The system processes newly added deal orders every day, samples them based on specific rules, and performs retrievals, ultimately evaluating system performance based on the retrieval success rate.

## 检索工具逻辑 / Retrieval Tool Logic

### 1. 数据准备阶段 / Data Preparation Stage


系统每日通过glif的deal数据库就行数据增量的更新，订单插入至自有Mongodb数据库。目前每日新增月10万条订单数据。
https://marketdeals.s3.amazonaws.com/StateMarketDeals.json.zst

### 2. 抽样样本的逻辑 / Sampling Logic

系统根据clentID和providerID两个字段进行分组，对完成分组的数据进行排序，排序方式为以dealID降序排序。并取前40%的订单作为抽样样本数据。


### 3. 检索测试 / Retrieval Sample Logic

1、将取值完成的前40%的抽样样本数据进行Sample。
2、取节点的libp2p地址
3、跟节点做交互，一笔deal会进行三种检索任务（Http、Graphsync、Bitswap）

### 4. 检索成功数 / Number of Successful Retrievals

系统对检索样本进行检索测试，记录检索成功的总数量。

The system performs retrieval tests on the sample and records the total number of successful retrievals.

### 5. 检索率 / Retrieval Success Rate

检索率 = 检索成功数 / 检索样本。

Retrieval Success Rate = Number of Successful Retrievals / Number of Retrieval Samples.

## 运行方式 / How to Run

访问网站：http://storagestats.ndlabs.io/

Access the website: http://storagestats.ndlabs.io/

## 贡献 / Contributing

欢迎提交 pull requests 或提出 issues 来帮助改进此项目。

We welcome pull requests or issues to help improve this project.

## 许可证 / License

该项目基于 MIT 许可证进行发布。

This project is licensed under the MIT License.
