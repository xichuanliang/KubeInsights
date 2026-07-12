# KubeInsights

## 基于 eBPF 的 HTTP/HTTPS 全链路网络路径与性能诊断系统

---

# 1. 项目概述

## 1.1 项目名称

**KubeInsights**

全称：

> KubeInsights: eBPF-based HTTP/HTTPS End-to-End Network Path and Performance Diagnosis System

中文：

> KubeInsights eBPF 的 HTTP/HTTPS 全链路网络路径与性能诊断系统


---

# 2. 项目背景

在生产环境中，经常遇到：

- HTTP 请求响应慢
- HTTPS 请求超时
- 服务调用延迟增加
- 网络抖动
- TCP 重传
- Pod 网络异常
- 服务间调用失败


传统工具：

- tcpdump
- Wireshark
- traceroute
- ping
- ss
- netstat

只能观察局部信息，无法回答：

> 一个 HTTP/HTTPS 请求从产生到结束，经过了哪些网络路径？在哪个阶段耗时？瓶颈在哪里？


因此设计 KubeInsights。


---

# 3. 项目目标

实现一个基于 Linux eBPF 的 HTTP/HTTPS 全链路诊断 Agent。


输入：

```
URL

例如:

https://api.example.com/order?id=100
```


输出：

## 3.1 请求生命周期

示例：

```
HTTP Request

|
|-- DNS Resolve
|
|-- TCP Connect       5ms
|
|-- TLS Handshake    20ms
|
|-- Send Request      1ms
|
|-- Server Queue     50ms
|
|-- Application      20ms
|
|-- MySQL             80ms
|
|-- Response Send     2ms
|
|-- TCP Close

Total:178ms
```


---

## 3.2 网络路径


展示：

```
HTTP Request

client

 |
 eth0
 192.168.1.20

 |
 Linux Routing

 |
 iptables PREROUTING

 |
 cni0

 |
 veth1234

 |
 Pod Namespace

 |
 eth0

 |
 nginx

 |
 application
```


记录：

- 网卡名称
- 网卡 IP
- MAC 地址
- Namespace
- Bridge
- Veth
- VXLAN
- Bond
- 路由
- NAT
- Packet latency


---

## 3.3 自动瓶颈分析


例如：

```
Request latency:

500ms


Breakdown:

TCP:
10ms


TLS:
20ms


Network:
5ms


Application:
50ms


MySQL:
400ms


Conclusion:

Root Cause:
MySQL slow query
```


或者：

```
TCP Retransmission:

50 times


Conclusion:

Network packet loss
```


---

# 4. 总体架构


```
                 Web UI

                    |

              Query API

                    |

        Trace Analyzer Engine

                    |

          +---------+---------+

          |                   |

    Event Collector     Topology Manager


================ eBPF Layer =================


        HTTP Probe

        TLS Probe

        Socket Probe

        TCP Probe

        Network Probe

        Scheduler Probe

        IO Probe


================ Linux Kernel ================


        NIC

        Socket

        TCP/IP Stack

        Netfilter

        Routing

        Bridge

        Container Network

```


---

# 5. 技术选型


## 5.1 Kernel eBPF


语言：

```
C
```


框架：

```
libbpf
```


技术：

```
BPF CO-RE
```


数据传输：

```
BPF Ring Buffer
```


---

## 5.2 User Space


语言：

```
Go
```


负责：

- eBPF 加载
- Event 接收
- Trace 聚合
- 数据分析


---

## 5.3 Storage


开发：

```
SQLite
```


生产：

```
ClickHouse
```


---

## 5.4 Visualization


支持：

- Grafana
- Jaeger
- 自研 Web UI


---

# 6. 项目目录设计


```
netsight-ebpf/


├── bpf/

│
├── socket.bpf.c
├── tcp.bpf.c
├── network.bpf.c
├── http.bpf.c
├── tls.bpf.c
├── syscall.bpf.c
├── sched.bpf.c


├── cmd/

│
└── agent/


├── pkg/


│
├── collector/

├── trace/

├── network/

├── topology/

├── analyzer/

├── storage/


├── api/


├── web/


└── configs/

```


---

# 7. eBPF 功能模块设计


# Module 1: Socket Tracking


## 目标

跟踪 TCP 生命周期。


## Hook


```
inet_csk_accept

tcp_v4_connect

tcp_close
```


## 采集


```
socket cookie

pid

process

src ip

dst ip

src port

dst port
```


核心关联：

```
socket_cookie
```


---

# Module 2: TCP Performance


## 目标

分析网络质量。


## Hook


```
tcp_retransmit_skb

tcp_rcv_established

tcp_sendmsg

tcp_recvmsg
```


## 采集


```
RTT

Retransmission

Window Size

Send Queue

Receive Queue

Packet count
```


---

# Module 3: Network Path Tracking


## 目标

获取数据包经过路径。


---

## TC


Hook：

```
tc ingress

tc egress
```


采集：

```
interface

packet direction

timestamp

skb info
```


---

## XDP


检测：

```
packet drop

redirect

pass
```


---

## Netfilter


Hook：


```
NF_INET_PRE_ROUTING

NF_INET_LOCAL_IN

NF_INET_FORWARD

NF_INET_LOCAL_OUT

NF_INET_POST_ROUTING
```


记录：


```
NAT

packet stage

interface
```


---

# Module 4: Network Topology Discovery


## 目标


发现机器网络结构。


采集：

```
网卡

IP

MAC

MTU

Driver

Namespace

Bridge

Veth

Bond

VXLAN
```


示例：


```
eth0

 |

cni0

 |

veth123

 |

pod
```


实现：

Linux Netlink API


---

# Module 5: HTTP Trace


## 目标


识别 HTTP 请求。


支持：

- HTTP/1.1
- HTTPS


---

## HTTP


Hook：

```
tcp_recvmsg
```


采集：


```
method

url

host

status

request size

response size

latency
```


---

## HTTPS


通过 Uprobe：

OpenSSL：

```
SSL_read

SSL_write
```


Go TLS：

```
crypto/tls.Conn.Read

crypto/tls.Conn.Write
```


---

# Module 6: Application Trace


## 目标

分析应用处理时间。


采集：

```
request enter

handler execution

response send
```


支持：

Go：

```
net/http

gin

grpc
```


Java：

```
Servlet

Spring
```


---

# Module 7: Dependency Trace


监控：

- MySQL
- Redis
- RPC


---

## MySQL


Hook：

```
mysql_real_query

send
```


采集：

```
SQL

latency
```


---

## Redis


解析：

```
RESP
```


采集：

```
GET

SET

DEL
```


---

## RPC


支持：

```
HTTP Client

gRPC

Dubbo
```


---

# Module 8: System Resource


## CPU


Hook：

```
sched_switch
```


采集：

```
CPU wait

running time
```


---

## IO


Hook：

```
block_rq_issue

block_rq_complete
```


采集：

```
IO latency
```


---

# 8. Trace 数据模型


所有事件统一模型：


```go
type Event struct {

 Timestamp uint64

 Type EventType


 TraceID uint64


 SocketCookie uint64


 PID uint32


 TID uint32


 Interface string


 SrcIP string


 DstIP string


 Duration uint64


 Metadata map[string]string

}
```


---

# 9. Trace 关联算法


核心：


```
socket_cookie

+

5 tuple

+

timestamp
```


形成：


```
Trace


HTTP

 |

Socket

 |

TCP

 |

Network Path

 |

Application

 |

Dependency

 |

Response
```


---

# 10. 输出格式


JSON:


```json
{
 "url":
 "https://api.test.com/order",

 "duration":180,


 "network_path":[

 {
 "device":"eth0",
 "ip":"10.0.0.2",
 "latency":3
 },


 {
 "device":"veth123",
 "ip":"10.244.1.10",
 "latency":5
 }

 ],


 "spans":[

 {
 "name":"TCP_CONNECT",
 "duration":5
 },


 {
 "name":"MYSQL",
 "duration":80
 }

 ],


 "rootCause":

 "MYSQL_LATENCY"

}
```


---

# 11. 开发阶段规划


# Phase 1

## TCP / Socket Trace


完成：

- Socket 生命周期
- TCP 建连
- TCP RTT
- TCP 重传


目标：

看到：

```
连接建立

关闭

RTT

Retransmission
```


---

# Phase 2

## Network Path


完成：

- TC
- XDP
- Netfilter
- Network Topology


目标：

看到：

```
eth0

bridge

veth

namespace

pod
```


---

# Phase 3

## HTTP / HTTPS


完成：

- HTTP 解析
- HTTPS TLS Hook


目标：

看到：

```
URL

Method

Latency
```


---

# Phase 4

## Trace Correlation


目标：

形成：

```
HTTP

↓

Network

↓

Application

↓

Dependency

```


---

# Phase 5

## Root Cause Analysis


自动判断：


```
Network problem

Application problem

Database problem

CPU problem
```


---

# Phase 6

## Web UI


展示：

- 请求链路
- 网络拓扑
- 时间轴
- Root Cause


---

# 12. Codex 实现要求


请按照以下规则实现：


## 基础要求


1. 使用 C + libbpf + CO-RE 编写 eBPF 程序。

2. 使用 Go 编写用户态 Agent。

3. 所有事件通过 RingBuffer 传输。

4. 所有事件必须包含：

```
timestamp

pid

socket_cookie

interface

src/dst IP

event type
```


5. 代码模块化设计。

6. 每个 Phase 必须能够独立运行。


---

# 13. 最终目标


实现一个类似：

- Pixie
- Cilium Hubble
- Datadog Network Performance Monitoring


的轻量级 eBPF 网络诊断系统。


最终能力：


输入：

```
https://api.example.com/order
```


输出：

```
完整HTTP生命周期

+

完整网络路径

+

经过网卡信息

+

TCP性能

+

应用耗时

+

依赖调用

+

Root Cause分析
```


---

# 项目最终定位

KubeInsights 是一个：

> 基于 eBPF 的 HTTP/HTTPS 全链路网络路径与性能诊断平台。

它通过 Linux 内核级观测能力，将：

- 网络
- TCP
- TLS
- HTTP
- 应用
- 数据库
- 系统资源

统一关联，最终解决：

> 一个 HTTP 请求为什么慢，以及它到底慢在哪里。







# KubeInsights Kubernetes 部署方案

## 基于 Kubernetes DaemonSet 的 eBPF 全链路网络诊断系统部署方案


---

# 1. 部署目标


KubeInsights 采用 Kubernetes 原生部署方式，通过：

- DaemonSet 部署 Agent
- Deployment 部署 Collector
- Deployment 部署 Web UI
- StatefulSet 部署存储组件


实现 Kubernetes 集群内：

- HTTP/HTTPS 请求链路追踪
- TCP 网络性能分析
- 网络路径发现
- Pod 网络拓扑分析
- 服务调用链分析
- 网络瓶颈定位


最终实现：

```
输入:

https://api.example.com/order


输出:

HTTP请求生命周期

+

网络路径

+

经过网卡

+

Pod链路

+

TCP性能

+

Root Cause分析

```


---

# 2. Kubernetes 总体架构


```
                         User


                          |

                          |

                    NetSight UI


                          |

                          |

                  Trace Query API


                          |

                          |

                 Trace Analyzer


                          |

                          |

                    Collector


                          |

              +-----------+-----------+

              |                       |

        Agent(Node1)            Agent(Node2)

              |                       |

              |                       |

          eBPF Program            eBPF Program

              |                       |

          Linux Kernel           Linux Kernel

              |                       |

       Containers/PODs        Containers/PODs


```


---

# 3. Kubernetes 组件设计


整个系统包含以下组件：


```
KubeInsights

|

├── kubeinsights-agent

|       DaemonSet

|

├── kubeinsights-controller

|       Deployment

|

├── kubeinsights-collector

|       Deployment

|

├── kubeinsights-storage

|       StatefulSet

|

└── kubeinsights-ui

        Deployment

```


---

# 4. kubeinsights-agent


## 4.1 部署方式


使用：

```
DaemonSet
```


原因：

eBPF 必须运行在每个 Kubernetes Node 上。


每个 Node 部署一个 Agent。


例如：

```
Node1

  |

  kubeinsights-agent


Node2

  |

  kubeinsights-agent


Node3

  |

  kubeinsights-agent

```


---

# 4.2 Agent 职责


kubeinsights-agent 负责：


## eBPF 加载


加载：

```
socket.bpf.o

tcp.bpf.o

network.bpf.o

http.bpf.o

tls.bpf.o

sched.bpf.o

io.bpf.o
```


---

## 数据采集


采集：

```
TCP

HTTP

HTTPS

TLS

Socket

Network Path

Network Device

CPU

IO

```


---

## 本地解析


例如：

```
socket_cookie

        |

        |

HTTP Request

        |

        |

Network Path

```


生成 Trace Event。


---

## 数据发送


通过：

```
gRPC

+
protobuf

```


发送到：

```
kubeinsights-collector

```


---

# 5. Agent 权限要求


由于需要访问 Linux Kernel，需要特殊权限。


## 必需权限


```
privileged: true
```


原因：

需要：

```
BPF_PROG_LOAD

BPF_MAP_CREATE

kprobe attach

tracepoint attach

tc attach

xdp attach

```


---

## Linux Capability


推荐：

```
CAP_BPF

CAP_PERFMON

CAP_NET_ADMIN

CAP_SYS_ADMIN

```


---

# 6. Agent Host 挂载


Agent 需要访问宿主机信息。


## /sys/fs/bpf


用途：

BPF Pin Map


挂载：


```
/sys/fs/bpf

```


---

## /sys/kernel/debug


用途：

trace filesystem


挂载：


```
/sys/kernel/debug

```


---

## /lib/modules


用途：

Kernel Module 信息


挂载：

```
/lib/modules

```


---

## Host Network


开启：

```
hostNetwork: true

```


原因：

获取：

```
Node NIC

eth0

ens33

bond0

```


---

## Host PID


开启：

```
hostPID: true

```


原因：

获取：

```
Process

PID

Namespace

```


---

# 7. Agent DaemonSet 示例


```yaml
apiVersion: apps/v1
kind: DaemonSet

metadata:

  name: kubeinsights-agent

  namespace: kubeinsights


spec:


  selector:

    matchLabels:

      app: kubeinsights-agent



  template:


    metadata:


      labels:

        app: kubeinsights-agent



    spec:


      hostNetwork: true


      hostPID: true



      containers:


      - name: agent


        image:

          kubeinsights-agent:v1



        securityContext:


          privileged: true



        volumeMounts:


        - name: bpf

          mountPath: /sys/fs/bpf



        - name: debug

          mountPath: /sys/kernel/debug



        - name: modules

          mountPath: /lib/modules



      volumes:



      - name: bpf


        hostPath:


          path: /sys/fs/bpf



      - name: debug


        hostPath:


          path: /sys/kernel/debug



      - name: modules


        hostPath:


          path: /lib/modules

```


---

# 8. kubeinsights-controller


## 部署方式


```
Deployment
```


---

## 职责


负责：

- Agent 配置管理
- eBPF 参数下发
- 采集规则管理
- Kernel 能力检测


例如：

配置：

```yaml
http:

  enabled: true


tcp:

  retransmission: true


tls:

  enabled: true

```


Controller 下发：

```
Agent

```


---

# 9. kubeinsights-collector


## 部署方式


```
Deployment

```


---

## 职责


负责：

## Event 接收


接收：

```
Agent

    |

    |

gRPC

    |

Collector

```


---

## Trace 聚合


例如：


Agent:

```
TCP_CONNECT

```


Agent:

```
HTTP_REQUEST

```


Agent:

```
MYSQL_QUERY

```


合并：


```
TraceID=1001


HTTP

 |

TCP

 |

Network

 |

MYSQL

```


---

## Root Cause 分析


例如：


输入：

```
Request latency

500ms

```


分析：

```
TCP

10ms


TLS

20ms


Application

30ms


MySQL

440ms

```


输出：

```
Root Cause:

MYSQL_LATENCY

```


---

# 10. kubeinsights-storage


## 推荐


生产环境：

```
ClickHouse

```


存储：


## HTTP Trace


```
http_trace

```


字段：

```
trace_id

url

method

latency

status

```


---

## Network Path


```
network_path

```


字段：

```
trace_id

interface

ip

namespace

latency

```


---

## TCP Metrics


```
tcp_metrics

```


字段：

```
rtt

retransmission

packet_loss

```


---

# 11. kubeinsights-ui


## 部署方式


```
Deployment

```


---

## 功能


## HTTP Trace 页面


展示：


```
GET /order


TCP

 |

TLS

 |

Application

 |

MYSQL

 |

Response

```


---

## Network Path 页面


展示：


```
Client


 |

eth0

10.0.0.10


 |

cni0


 |

veth123


 |

Pod


 |

Service


 |

Backend

```


---

## Root Cause 页面


展示：

```
Request Slow


原因:


MYSQL_LATENCY


影响:

90%

```


---

# 12. Kubernetes 网络关联


kubeinsights 需要建立：


```
Socket


 |

PID


 |

Network Namespace


 |

veth


 |

Pod


 |

Deployment


 |

Service

```


最终显示：


```
payment-service


pod/payment-xxxx


node/node1

```


---

# 13. Service 调用链


支持：


```
Client


 |

Service A


 |

Service B


 |

Database

```


例如：


```
order-service


 |

payment-service


 |

mysql

```


---

# 14. 部署流程


## Step 1

创建 Namespace


```bash
kubectl create namespace netsight
```



---

## Step 2

部署 Storage


```bash
kubectl apply -f storage.yaml

```



---

## Step 3

部署 Collector


```bash
kubectl apply -f collector.yaml

```



---

## Step 4

部署 Controller


```bash
kubectl apply -f controller.yaml

```



---

## Step 5

部署 Agent


```bash
kubectl apply -f agent-daemonset.yaml

```



---

## Step 6

部署 UI


```bash
kubectl apply -f ui.yaml

```


---

# 15. 运行流程


请求：

```
curl https://api.example.com/order

```


流程：


```
Pod


 |

Socket


 |

eBPF


 |

RingBuffer


 |

Agent


 |

gRPC


 |

Collector


 |

ClickHouse


 |

UI


```


---

# 16. 生产环境考虑


## 高可用


Collector：

```
Deployment

+

Replica

```


Storage：

```
ClickHouse Cluster

```


---

## 大规模集群


支持：

```
1000+

Nodes

```


架构：

```
Node

 |

Agent


 |

Kafka


 |

Collector

```


---

## 安全


生产建议：

从：

```
privileged

```


逐步降低到：

```
CAP_BPF

CAP_PERFMON

CAP_NET_ADMIN

```


---

# 17. 最终部署形态


```
                 Kubernetes Cluster


                         |

                    NetSight UI


                         |

                    Collector


                         |

              +----------+----------+

              |                     |

          Node1                 Node2


       Agent                  Agent


          |                     |


       eBPF                  eBPF


          |                     |


       Kernel               Kernel


          |                     |


       Pods                 Pods


```


---

# 18. 最终能力


部署完成后：

输入：

```
https://api.example.com/order

```


可以得到：


```
HTTP生命周期

+

TCP状态

+

TLS耗时

+

数据包路径

+

经过网卡

+

Pod网络路径

+

Service调用链

+

数据库耗时

+

Root Cause分析

```


kubeinsights 最终成为 Kubernetes 环境下的：

> eBPF 驱动的 HTTP/HTTPS 全链路网络性能诊断平台。


