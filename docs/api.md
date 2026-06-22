# 停车场管理系统 API 文档

## 概述

- **Base URL**: `http://localhost:8080`
- **数据格式**: `application/json; charset=utf-8`
- **字符编码**: UTF-8，中文车牌请使用 URL 编码或直接在 JSON body 中传输
- **存储方式**: 内存存储（进程重启数据丢失，仅用于开发/演示）

## 通用响应格式

所有接口统一返回以下 JSON 结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": { }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | number | 业务码，`0` 表示成功，非 0 表示失败 |
| `message` | string | 结果说明 |
| `data` | object / array | 响应数据，部分错误响应不包含该字段 |

常见业务码：

| code | 含义 |
|------|------|
| `0` | 成功 |
| `400` | 请求参数错误 |
| `404` | 资源不存在（如车牌号未在场内 / 未注册月卡）|
| `409` | 资源冲突（如车辆已在场内 / 月卡已注册 / 月卡已过期 / 车位已满）|

---

## 一、车辆出入场

### 1.1 车辆入场

记录车牌号与入场时间，自动判断车型。

- **请求方式**: `POST`
- **路径**: `/api/entry`
- **Content-Type**: `application/json`

**请求体**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `license_plate` | string | ✅ | 车牌号 |
| `car_type` | string | ❌ | 车型：`temp`（临时车）/ `monthly`（月卡车）。**不传时系统会根据月卡自动识别**，有有效月卡自动当月卡车，否则当临时车 |

**请求示例**:

```json
{
  "license_plate": "京A88888"
}
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "入场成功",
  "data": {
    "license_plate": "京A88888",
    "car_type": "monthly",
    "entry_time": "2026-06-23T02:22:51+08:00"
  }
}
```

| data 字段 | 类型 | 说明 |
|-----------|------|------|
| `license_plate` | string | 车牌号 |
| `car_type` | string | 实际登记车型（可能与传入不同，系统会自动识别月卡）|
| `entry_time` | string | ISO8601 格式入场时间 |

**可能的错误**:

| code | message | 含义 |
|------|---------|------|
| `400` | 车牌号不能为空 | 缺少 license_plate |
| `409` | 车辆已在场内 | 该车已入场未出 |
| `409` | 车位已满 | 100 个车位全部占用 |
| `409` | 无效的车辆类型 | car_type 既不是 temp 也不是 monthly |
| `409` | 该车牌未注册月卡 | 传了 monthly 但该车牌没有月卡 |
| `409` | 月卡已过期，请续费 | 传了 monthly 但月卡已过期 |

---

### 1.2 车辆出场

根据停留时长自动算费，写入当日收入流水。

- **请求方式**: `POST`
- **路径**: `/api/exit`
- **Content-Type**: `application/json`

**请求体**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `license_plate` | string | ✅ | 车牌号 |

**请求示例**:

```json
{
  "license_plate": "京A88888"
}
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "出场成功",
  "data": {
    "license_plate": "京A88888",
    "car_type": "temp",
    "entry_time": "2026-06-23T02:22:51+08:00",
    "exit_time": "2026-06-23T05:45:12+08:00",
    "duration_min": 203,
    "fee": 15
  }
}
```

| data 字段 | 类型 | 说明 |
|-----------|------|------|
| `license_plate` | string | 车牌号 |
| `car_type` | string | 出场时实际按哪种车型计费（系统兜底再识别一次月卡）|
| `entry_time` | string | ISO8601 入场时间 |
| `exit_time` | string | ISO8601 出场时间 |
| `duration_min` | number | 停留时长（分钟，向上取整）|
| `fee` | number | 应收费用（元）|

**可能的错误**:

| code | message | 含义 |
|------|---------|------|
| `400` | 车牌号不能为空 | 缺少 license_plate |
| `404` | 该车辆不在场内 | 该车没有入场记录 |

---

### 1.3 查询剩余车位

查询当前车位占用情况。

- **请求方式**: `GET`
- **路径**: `/api/spaces`

**请求示例**:

```
GET /api/spaces
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "total": 100,
    "occupied": 12,
    "available": 88
  }
}
```

| data 字段 | 类型 | 说明 |
|-----------|------|------|
| `total` | number | 总车位数（固定 100）|
| `occupied` | number | 已占用 |
| `available` | number | 剩余可用 |

---

## 二、月卡管理

### 2.1 注册月卡

为指定车牌开通月卡服务。

- **请求方式**: `POST`
- **路径**: `/api/card/register`
- **Content-Type**: `application/json`

**请求体**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `license_plate` | string | ✅ | 车牌号 |
| `owner_name` | string | ✅ | 车主姓名 |
| `months` | number | ✅ | 开通月数（必须 > 0）|

**请求示例**:

```json
{
  "license_plate": "京A88888",
  "owner_name": "张三",
  "months": 3
}
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "注册成功",
  "data": {
    "license_plate": "京A88888",
    "owner_name": "张三",
    "start_date": "2026-06-23T00:00:00+08:00",
    "expire_date": "2026-09-24T00:00:00+08:00",
    "active": true,
    "created_at": "2026-06-23T02:30:00+08:00"
  }
}
```

| data 字段 | 类型 | 说明 |
|-----------|------|------|
| `license_plate` | string | 车牌号 |
| `owner_name` | string | 车主姓名 |
| `start_date` | string | 生效日（注册当日 00:00）|
| `expire_date` | string | 到期日（N 个月后 + 1 天 00:00）|
| `active` | boolean | 是否有效 |
| `created_at` | string | ISO8601 注册时间 |

**可能的错误**:

| code | message | 含义 |
|------|---------|------|
| `400` | 车牌号 / 车主姓名不能为空 | 必填项缺失 |
| `400` | 月数必须大于0 | months <= 0 |
| `409` | 该车牌已注册月卡 | 已存在不可重复注册，用续费接口 |

---

### 2.2 月卡续费

给已有月卡延长有效期。

- **请求方式**: `POST`
- **路径**: `/api/card/renew`
- **Content-Type**: `application/json`

**请求体**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `license_plate` | string | ✅ | 车牌号 |
| `months` | number | ✅ | 续费月数（必须 > 0）|

> 💡 **续费规则**：若月卡仍有效，从当前到期日顺延；若已过期，从当日 00:00 起算。

**请求示例**:

```json
{
  "license_plate": "京A88888",
  "months": 2
}
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "续费成功",
  "data": {
    "license_plate": "京A88888",
    "owner_name": "张三",
    "start_date": "2026-06-23T00:00:00+08:00",
    "expire_date": "2026-11-24T00:00:00+08:00",
    "active": true,
    "created_at": "2026-06-23T02:30:00+08:00"
  }
}
```

**可能的错误**:

| code | message | 含义 |
|------|---------|------|
| `400` | 车牌号不能为空 | 缺失 |
| `400` | 续费月数必须大于0 | months <= 0 |
| `404` | 该车牌未注册月卡 | 该车牌没有月卡记录 |

---

### 2.3 查询月卡状态

查询某个车牌的月卡详情（含有效期判断）。

- **请求方式**: `GET`
- **路径**: `/api/card/status`

**Query 参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `license_plate` | string | ✅ | 车牌号（含中文请 URL 编码）|

**请求示例**:

```
GET /api/card/status?license_plate=%E4%BA%ACA88888
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "license_plate": "京A88888",
    "owner_name": "张三",
    "start_date": "2026-06-23T00:00:00+08:00",
    "expire_date": "2026-09-24T00:00:00+08:00",
    "active": false,
    "created_at": "2026-06-23T02:30:00+08:00"
  }
}
```

> 💡 `active` 字段是**动态判断**的：即使数据库里存的是 `true`，如果当前时间超过 `expire_date`，接口会自动返回 `active: false`。

**可能的错误**:

| code | message | 含义 |
|------|---------|------|
| `400` | license_plate 参数不能为空 | 缺少 query 参数 |
| `404` | 该车牌未注册月卡 | 无月卡记录 |

---

## 三、管理端接口

### 3.1 当日收入流水

查询当天所有出场记录的收入明细。

- **请求方式**: `GET`
- **路径**: `/api/admin/income`

**请求示例**:

```
GET /api/admin/income
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "license_plate": "京A88888",
      "car_type": "temp",
      "entry_time": "2026-06-23T02:22:51+08:00",
      "exit_time": "2026-06-23T05:45:12+08:00",
      "duration_min": 203,
      "fee": 15,
      "created_at": "2026-06-23T05:45:12+08:00"
    },
    {
      "license_plate": "沪B66666",
      "car_type": "monthly",
      "entry_time": "2026-06-23T03:10:00+08:00",
      "exit_time": "2026-06-23T08:20:00+08:00",
      "duration_min": 310,
      "fee": 0,
      "created_at": "2026-06-23T08:20:00+08:00"
    }
  ]
}
```

`data` 为 IncomeRecord 数组，每个元素字段见 1.2 出场响应。

---

### 3.2 在场车辆列表

查询当前场内所有未出场的车辆。

- **请求方式**: `GET`
- **路径**: `/api/admin/vehicles`

**请求示例**:

```
GET /api/admin/vehicles
```

**成功响应** (`200 OK`):

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "license_plate": "京A88888",
      "car_type": "temp",
      "entry_time": "2026-06-23T10:00:00+08:00"
    }
  ]
}
```

---

## 四、计费规则

### 4.1 临时车（`temp`）

| 项目 | 规则 |
|------|------|
| 免费时长 | 前 **30 分钟**免费（含 30 分钟）|
| 计费单位 | 超出 30 分钟后，按 **¥5 / 小时** 计费，不足 1 小时向上取整 1 小时 |
| 计费基数 | `(实际停留分钟 - 30) ÷ 60`，结果向上取整 |
| 每日封顶 | 单日最高 **¥50** |

**计算示例**：

| 实际停留 | 计费说明 | 费用 |
|---------|---------|------|
| 15 分钟 | 免费内 | ¥0 |
| 30 分钟 | 刚好免费 | ¥0 |
| 31 分钟 | 超出 1 分钟，按 1 小时算 | ¥5 |
| 85 分钟 | 超出 55 分钟 = 1 小时 | ¥5 |
| 150 分钟 | 超出 120 分钟 = 2 小时 | ¥10 |
| 12 小时 | 超出 690 分钟 = 12 小时，但封顶 ¥50 | ¥50 |

### 4.2 月卡车（`monthly`）

| 项目 | 规则 |
|------|------|
| 费用 | **免费**（无论停多久，出场 `fee` 均为 0）|
| 准入前提 | 该车牌必须有**已注册且在有效期内**的月卡 |
| 有效期判断 | `active === true` 且 `当前时间 < expire_date` |

### 4.3 车型自动识别（重要！）

为了避免调用方遗漏传 `car_type`，系统提供**双重兜底识别**：

1. **入场时**：若 `car_type` 为空或传了 `temp`，系统先查月卡，有有效月卡 → 自动升级为 `monthly`
2. **出场时**：即使入场登记为 `temp`，出场时若检测到该车牌已有有效月卡 → 按 `monthly` 计费（0 元）

> 💡 **推荐做法**：调用方直接不传 `car_type`，让系统自动识别即可，最省心。

---

## 五、典型调用流程

### 5.1 临时车（简单版）

```
1. POST /api/entry        {"license_plate":"临时XXXX"}     ← 不传 car_type，系统默认 temp
2. GET  /api/spaces                                      ← 确认车位变化
3. POST /api/exit         {"license_plate":"临时XXXX"}     ← 自动算费
```

### 5.2 月卡车（完整流程）

```
1. POST /api/card/register {"license_plate":"京A88888","owner_name":"张三","months":3}  ← 办卡
2. POST /api/entry         {"license_plate":"京A88888"}                                   ← 自动识别为 monthly
3. GET  /api/card/status?license_plate=京A88888                                           ← 随时查月卡状态
4. POST /api/exit          {"license_plate":"京A88888"}                                   ← 出场费 0
5. POST /api/card/renew    {"license_plate":"京A88888","months":1}                        ← 到期后续一个月
```

### 5.3 管理端日常查看

```
GET /api/spaces             ← 看车位余量
GET /api/admin/vehicles     ← 看谁在里面
GET /api/admin/income       ← 看今天收了多少钱
```

---

## 六、curl 调用速查

```bash
# 入场（临时车）
curl -X POST http://localhost:8080/api/entry \
  -H "Content-Type: application/json" \
  -d '{"license_plate":"TEST001"}'

# 出场
curl -X POST http://localhost:8080/api/exit \
  -H "Content-Type: application/json" \
  -d '{"license_plate":"TEST001"}'

# 查车位
curl http://localhost:8080/api/spaces

# 注册月卡
curl -X POST http://localhost:8080/api/card/register \
  -H "Content-Type: application/json" \
  -d '{"license_plate":"MONTH01","owner_name":"Alice","months":12}'

# 续费
curl -X POST http://localhost:8080/api/card/renew \
  -H "Content-Type: application/json" \
  -d '{"license_plate":"MONTH01","months":6}'

# 查月卡状态（中文车牌需 URL 编码）
curl "http://localhost:8080/api/card/status?license_plate=MONTH01"

# 管理端 - 当日收入
curl http://localhost:8080/api/admin/income

# 管理端 - 在场车辆
curl http://localhost:8080/api/admin/vehicles
```
