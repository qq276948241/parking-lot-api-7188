# 停车场管理系统 - 架构说明

> 给新来同事的快速上手指南，用人话讲清楚这个项目是怎么组织的。

---

## 一、项目长啥样？
------------------

```
project3/
├── main.go                    # 程序入口，启动服务、注册路由
├── go.mod                    # Go 模块依赖
└── internal/
    ├── model/
    │   └── model.go       # 数据结构定义（停车记录、支付记录等）
    ├── store/
    │   └── store.go         # 内存存储，增删改查都在这
    ├── service/
    │   └── billing.go       # 计费逻辑（临时车/月卡车怎么算钱）
    └── handler/
        └── handler.go       # HTTP 接口层，处理请求和响应
```

就 4 层，每层各司其职，不越级。

---

## 二、调用关系是怎样的？
----------------------

```
请求来了 → main.go 路由分发 → handler 处理请求 → service 算费用 → store 存数据
                    ↓
            handler 封装成 JSON 返回给前端
```

具体到代码：

- **main.go** 只干两件事：
  1. `store.NewMemoryStore(100)` 初始化内存存储
  2. 把 URL 路径和 handler 函数对应起来，比如：
     ```go
     mux.HandleFunc("/api/entry", handler.WithMethod(http.MethodPost, h.Entry))
     ```

- **handler.go** 是"前台接待"：
  - 检查请求参数对不对
  - 调用 `store` 或者 `service` 干活
  - 把结果包装成 JSON 给前端

- **service/billing.go** 是"会计"：
  - 只负责算停车费，不管数据存哪
  - 临时车怎么打折、月卡车怎么免费，都在这

- **store.go** 是"仓库管理员"：
  - 所有数据都存在内存的 map 里
  - 提供 `Entry()` `Exit()` `GetActiveRecordByPlate()` 等方法
  - handler 和 service 都不直接操作内存，全靠 store 来管

**最重要的调用链：**

1. 车辆入场：`handler.Entry` → `store.Entry`
2. 车辆出场：`handler.Exit` → `store.GetActiveRecordByPlate` → `billing.CalculateFee` → `store.Exit`
3. 查询车位：`handler.ParkingLotStatus` → `store.GetParkingLot`
4. 按车牌查：`handler.QueryVehicle` → `store.GetActiveRecordByPlate`

---

## 三、每层的核心文件说明
------------------

### 1. model/model.go — 大家都要用的"零件"
----------------------------------

定义了项目里的"东西长什么样"，比如：

```go
type ParkingRecord struct {
    ID          string      // 记录ID
    PlateNumber string      // 车牌号
    VehicleType VehicleType // temp 或 monthly
    EntryTime   time.Time   // 入场时间
    ExitTime    *time.Time // 出场时间（还在里面就是 nil）
    Fee         float64   // 费用
    IsPaid      bool      // 有没有付钱
}
```

还有 `ParkingLot`（车位状态）、`PaymentRecord`（流水记录）等。

**所有层都 import 这个包，因为大家都要传这些结构体。**

### 2. store/store.go — 内存数据库
------------------------------

用 Go 的 `map` 和 `slice` 存所有数据，加了读写锁 `sync.RWMutex` 并发安全。

核心方法：

| 方法 | 干啥的 |
|--------|---------|
| `Entry(record)` | 车辆入场，记录进去 |
| `GetActiveRecordByPlate(plate)` | 根据车牌找**在场**的记录 |
| `Exit(recordID, fee, exitTime)` | 车辆出场，更新记录 |
| `GetParkingLot()` | 看车位还剩多少 |
| `GetActiveVehicles()` | 所有在场车辆列表 |
| `GetTodayPayments()` | 今天的所有流水 |

### 3. service/billing.go — 算账的
--------------------------

**只干一件事：算停车费。**

计费规则写死在代码里（不用数据库）：

- **临时车（temp）**：15 分钟内免费，超过后 5 元/小时，每日封顶 60 元
- **月卡车（monthly）**：随便停，0 元

核心方法就一个：
```go
CalculateFee(record, exitTime) → 费用金额
```

### 4. handler/handler.go — 接客的
------------------------

每个接口对应一个函数。

还提供两个公共工具：

| 工具函数 | 作用 |
|-----------|------|
| `WithMethod(method, next)` | 中间件，检查 HTTP 方法对不对，不对直接返回 405 |
| `decodeJSON(r, v)` | 把请求体的 JSON 解析到结构体，顺便关 body |
| `writeJSON(w, status, data)` | 写 JSON 响应 |
| `writeError(w, status, msg)` | 写错误响应，格式 `{"error": "xxx"}` |

---

## 四、API 接口手册
-------------

所有接口响应都是 JSON，错误响应统一 `Content-Type: application/json`。

---

### 1. 车辆入场
**POST** `/api/entry`

**请求体：**
```json
{
  "plate_number": "京A12345",
  "vehicle_type": "temp"     // 可选，不填就自动判断是不是月卡
}
```

**成功响应 200：**
```json
{
  "id": "uuid-xxx",
  "plate_number": "京A12345",
  "vehicle_type": "temp",
  "entry_time": "2026-06-22T22:06:42+08:00",
  "is_paid": false
}
```

**错误 400：**
```json
{ "error": "plate_number is required" }
{ "error": "vehicle already in parking lot" }
{ "error": "parking lot is full" }
```

---

### 2. 车辆出场
**POST** `/api/exit`

**请求体：**
```json
{
  "plate_number": "京A12345"
}
```

**成功响应 200：**
```json
{
  "id": "uuid-xxx",
  "plate_number": "京A12345",
  "vehicle_type": "temp",
  "entry_time": "2026-06-22T22:06:42+08:00",
  "exit_time": "2026-06-22T22:07:13+08:00",
  "duration_min": 31.5,
  "fee": 0
}
```

**错误 404：**
```json
{ "error": "vehicle not found in parking lot" }
```

---

### 3. 车位状态查询
**GET** `/api/parking/status`

**成功响应 200：**
```json
{
  "total_spots": 100,
  "occupied_spots": 3,
  "available_spots": 97
}
```

---

### 4. 按车牌号查车辆状态
**GET** `/api/vehicle/query?plate_number=京A12345`

**在场时响应 200：**
```json
{
  "plate_number": "京A12345",
  "is_in_parking": true,
  "vehicle_type": "temp",
  "entry_time": "2026-06-22T22:06:42+08:00",
  "duration_min": 125.3,
  "duration_human": "2小时5分12秒"
}
```

**不在场时响应 200（注意是 200 不是 404，方便前端判断）：**
```json
{
  "plate_number": "京A12345",
  "is_in_parking": false
}
```

**错误 400：**
```json
{ "error": "plate_number is required" }
```

---

### 5. 在场车辆列表（管理端）
**GET** `/api/admin/active-vehicles`

**成功响应 200：**
```json
[
  {
    "id": "uuid-xxx",
    "plate_number": "京A12345",
    "vehicle_type": "temp",
    "entry_time": "2026-06-22T22:06:42+08:00",
    "is_paid": false
  }
]
```

---

### 6. 当日收入流水（管理端）
**GET** `/api/admin/today-income`

**成功响应 200：**
```json
{
  "date": "2026-06-22",
  "total": 125.0,
  "count": 8,
  "payments": [
    {
      "id": "PAY-uuid-xxx",
      "plate_number": "京A12345",
      "amount": 15.0,
      "pay_time": "2026-06-22T22:07:13+08:00",
      "vehicle_type": "temp"
    }
  ]
}
```

---

## 五、常见问题 & 调试技巧
------------------

### 怎么启动服务？
```bash
cd project3
go build -o parking-api.exe .
.\parking-api.exe
```
默认端口 `:8080`

### 怎么测试接口？
```bash
# 入场
curl -X POST http://localhost:8080/api/entry \
  -H "Content-Type: application/json" \
  -d '{"plate_number":"TEST001"}'

# 查状态
curl "http://localhost:8080/api/parking/status"
```

### 月卡车怎么加？
在 `main.go` 里调用 `s.AddMonthlyPlate("京A12345")`，程序启动时会加进去。

### 数据会丢吗？
会！因为存在内存里，程序重启就没了。需求说不用数据库，就这样。

### 并发安全吗？
安全，`store.go` 里所有读写都加了 `sync.RWMutex` 锁了。

### 计费规则要改怎么办？
去 `service/billing.go` 里改 `PricingConfig` 就行，不用动其他地方。

---

## 六、改代码的正确姿势
------------------

记住：**该哪层改哪层，别瞎串**。

- 要加字段？→ `model/model.go`
- 要加存数据的方法？→ `store/store.go`
- 要改计费规则？→ `service/billing.go`
- 要加接口？→ `handler/handler.go` 加函数，`main.go` 注册路由
- 要改接口路径？→ `main.go` 里改路由就行

比如要加个"历史记录查询"接口：
1. store 里加 `GetHistoryByPlate(plate)` 方法
2. handler 里加 `HistoryVehicle` 函数
3. main.go 里注册路由 `mux.HandleFunc("/api/vehicle/history", ...)`

就这么简单。
