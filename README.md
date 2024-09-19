# Storagestats
Storagestats is a multifunctional storage status information management tool developed by ND Labs.


# 检索工具逻辑 / Retrieval Tool Logic

## 简介 / Introduction

本项目由 ND Labs 开发，旨在通过一套自动化逻辑来处理大量 deal 订单的数据检索。系统每天处理新增的 deal 订单，并按特定逻辑对其进行抽样和检索，最终通过检索成功率来评估系统性能。

This project, developed by ND Labs, aims to process a large number of deal orders through an automated logic system. The system processes newly added deal orders every day, samples them based on specific rules, and performs retrievals, ultimately evaluating system performance based on the retrieval success rate.

## 检索工具逻辑 / Retrieval Tool Logic

### 1. 数据准备阶段 / Data Preparation Stage

系统将每日新增的 deal 订单插入至自有数据库，目前每日新增约 30 多万条数据。

The system inserts newly added deal orders into its own database every day, with approximately 300,000 new records being added daily.

### 2. 抽样样本的逻辑 / Sampling Logic

系统根据 LDN 及节点号筛选相关的 deal 订单，并按照时间顺序进行排序。系统会取这些 deal 订单中最新的 40% 作为抽样样本。

The system filters relevant deal orders based on LDN and node number, then sorts them by time. The latest 40% of these deal orders are selected as the sampling sample.

### 3. 检索样本的逻辑 / Retrieval Sample Logic

在抽样样本中，系统随机抽取 30% 的 deal 订单作为检索样本。每个节点的检索样本数量不超过 150 条。（具体条数根据程序负载而定）

From the sampled deals, the system randomly selects 30% of the deal orders as the retrieval sample. The number of retrieval samples per node does not exceed 150 orders (the exact number may vary based on system load).

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
