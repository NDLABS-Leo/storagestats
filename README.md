# Storagestats
Storagestats is a multifunctional storage status information management tool developed by ND Labs.


# 检索工具逻辑 / Retrieval Tool Logic

## 项目简介 / Project Overview

本项目由 **ND Labs** 开发，旨在通过自动化逻辑系统高效处理大量 `deal` 订单数据。系统每天从外部数据源获取新增订单，并对其进行抽样、检索，最终根据检索成功率来评估系统的性能。该解决方案专为处理大规模订单数据而设计，确保在处理海量数据时系统能够保持高效和稳定。

**ND Labs** developed this project to efficiently process large volumes of `deal` order data through an automated logic system. The system retrieves newly added orders daily from external sources, performs sampling and retrieval tasks, and evaluates system performance based on the retrieval success rate. This solution is designed for large-scale data processing, ensuring the system remains efficient and stable even when handling high data volumes.

---

## 功能模块 / Key Modules

### 1. 数据准备阶段 / Data Preparation Stage

系统每日自动从 [Glif](https://marketdeals.s3.amazonaws.com/StateMarketDeals.json.zst) 的 `deal` 数据库获取最新的订单数据，并将其插入到内部的 **MongoDB** 数据库中进行存储与管理。目前，系统每天大约处理 **10 万条新订单** 数据，确保数据库持续更新且增量更新准确无误。

- **数据源**：Glif `deal` 数据库
- **数据库**：MongoDB
- **每日处理量**：约 100,000 条新订单

The system automatically retrieves the latest `deal` order data daily from [Glif](https://marketdeals.s3.amazonaws.com/StateMarketDeals.json.zst) and inserts it into an internal **MongoDB** database for storage and management. Currently, the system processes approximately **100,000 new orders** each day, ensuring continuous and accurate incremental updates.

### 2. 抽样逻辑 / Sampling Logic

在数据存储后，系统通过 **clientID** 和 **providerID** 字段对订单数据进行分组，然后按 **dealID** 进行降序排列。为确保抽样具有代表性，系统选择排序后的前 **40%** 的订单作为样本，这些样本将用于检索测试。

- **分组依据**：clientID, providerID
- **排序方式**：基于 dealID 降序排列
- **样本比例**：选择前 40% 的订单数据

After data is stored, the system groups the order data by **clientID** and **providerID** fields and then sorts the data in descending order by **dealID**. To ensure representative sampling, the system selects the top **40%** of the sorted orders as the sample, which will be used for retrieval testing.

### 3. 检索测试逻辑 / Retrieval Test Logic

对于已选取的抽样样本，系统会对每个样本执行三种不同的检索任务：**Http**、**Graphsync** 和 **Bitswap**。系统通过节点的 **libp2p** 地址与目标节点进行交互，完成检索任务，确保多种检索协议均能正常运行，以全面测试系统的检索能力。

- **检索任务**：
  1. **Http** 检索
  2. **Graphsync** 检索
  3. **Bitswap** 检索
- **节点交互**：基于 libp2p 地址

For the selected sample, the system performs three different retrieval tasks: **Http**, **Graphsync**, and **Bitswap**. The system interacts with target nodes via their **libp2p** addresses to complete the retrieval tasks, ensuring that all retrieval protocols are functional and the system's overall retrieval capability is fully tested.

### 4. 检索成功数 / Number of Successful Retrievals

在执行检索测试后，系统将记录每个检索任务的结果。系统会统计所有检索任务中成功完成检索的数量，并以此作为评估系统性能的关键指标之一。

- **成功检索次数**：指成功完成的检索任务总数

Once the retrieval tests are executed, the system records the result of each task. It counts the total number of successfully completed retrievals, which serves as one of the key metrics for evaluating system performance.

### 5. 检索成功率 / Retrieval Success Rate

检索成功率是衡量系统检索性能的重要指标，其计算公式如下：

\[
\text{检索成功率} = \frac{\text{检索成功数}}{\text{检索样本总数}}
\]

检索成功率表示系统在一批样本中成功完成检索的比率，该比率越高，表明系统性能越好。

The **Retrieval Success Rate** is a critical metric for evaluating the retrieval performance of the system. It is calculated as follows:

\[
\text{Retrieval Success Rate} = \frac{\text{Number of Successful Retrievals}}{\text{Total Number of Retrieval Samples}}
\]

The success rate reflects the percentage of successful retrievals out of the total sample size. A higher success rate indicates better system performance.

---

## 系统运行 / How to Run

### 访问网站 / Access the Website

系统提供了一个在线平台，用户可以访问该平台来查看相关数据统计与系统性能指标。请访问以下链接：

- **网站链接**：[http://storagestats.ndlabs.io/](http://storagestats.ndlabs.io/)

The system offers an online platform where users can view related data statistics and system performance metrics. Please visit the link below:

- **Website Link**: [http://storagestats.ndlabs.io/](http://storagestats.ndlabs.io/)

---

## 未来发展 / Future Development

未来版本将会进一步优化系统的检索逻辑，包括但不限于：
- 改进检索算法以提升检索成功率。
- 增加更多协议类型以支持更广泛的检索场景。
- 提升系统的扩展性，使其能够处理更大规模的订单数据。

Future versions will further optimize the system's retrieval logic, including but not limited to:
- Improved retrieval algorithms to increase success rates.
- Addition of more protocols to support a wider range of retrieval scenarios.
- Enhanced scalability to handle even larger volumes of order data.

---

## 贡献 / Contributions

我们欢迎社区的贡献！如果你对该项目有任何建议、问题或贡献意向，欢迎通过提交 Issue 或 Pull Request 与我们联系。

We welcome contributions from the community! If you have any suggestions, questions, or would like to contribute, feel free to reach out via submitting an issue or pull request.

---


## 许可证 / License

该项目基于 MIT 许可证进行发布。

This project is licensed under the MIT License.
