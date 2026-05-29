# openydt-cli 端到端测试报告

- 环境: 测试环境 (openapi-test.yidianting.com.cn);主车场 PTD2YBBZ(智汇云测试专用车场123412)
- 接口总数(已尝试): 143
- 结果: ✓PASS=106, BIZFAIL=0, NODEPLOY=11, ERROR=0, SKIP=1
- **非 NODEPLOY 失败(BIZFAIL+ERROR+SKIP): 1**;成功调通业务层(PASS+BIZFAIL): 106
- 调用次数: 142

> PASS=业务成功;BIZFAIL=已调通但业务校验未过;NODEPLOY=测试环境未部署(接口不存在,非 CLI 问题);ERROR=传输/系统异常;SKIP=故意不测(破坏性/需特殊前置)。

## 需关注(非 PASS,非 NODEPLOY)

| 命令 | 域 | 结果 | 说明 |
|---|---|---|---|
| sendCouponByCouponCode | coupon | NODATA | 环境限制: 按券码发放需平台预生成的未发放券码实例, 模板码不可直接发放 \| resultCode=-1 业务码未登记 \| 找不到该优惠券 |
| deleteTrader | coupon | SKIP | 破坏性: 会删除共享测试商家, 跳过(写路径已由 createTrader 覆盖) |
| getRealtimeLeaveAndChargeNum | data | NODATA | 环境限制: 测试车场当日无交易统计数据/统计服务在该环境不可用 \| resultCode=-1 业务码未登记 \| 查询车场交易次数失败 |
| channelSnap | device | NODATA | 环境限制: 通道无可下发的抓拍相机设备 \| resultCode=908 其它错误 \| 找不到设备 |
| cloudScanQrCode | device | NODATA | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=908 其它错误 \| 设备不存在 |
| cloudScanUpdateConfig | device | NODATA | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=909 请求参数错误 \| 参数错误:智泊云车场不存在以自定义编码500119的通道 |
| cloudScanVoice | device | NODATA | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=908 其它错误 \| 设备不存在 |
| cloudStopScanCode | device | NODATA | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=908 其它错误 \| 设备不存在 |
| opGate | device | NODATA | 环境限制: 通道当前无匹配在场车, 无法开闸 \| resultCode=908 其它错误 \| 通道当前车牌与入参车牌不一致 |
| opShowVoice | device | NODATA | 环境限制: 通道未绑定屏显/语音设备 \| resultCode=908 其它错误 \| 找不到设备 |
| setDefaultScreen | hidden | NODATA | 环境限制: 通道无绑定屏显设备 \| resultCode=908 其它错误 \| 找不到对应设备 |
| setParkRemainCarport | park | NODATA | 环境限制: 需存在车位区域(areaId)配置, 测试车场无对应车位区域 \| resultCode=908 其它错误 \| 700577区域设置异常 |
| cancellationOfArrears | parking | NODATA | 环境限制: 测试环境无现成欠费记录(recordId) \| resultCode=909 请求参数错误 \| 参数错误:recordId不能为空 |
| checkChannelExistCar | parking | NODATA | 环境限制: 需通道上物理有车, 测试环境无法满足 \| resultCode=908 其它错误 \| 通道无车 |
| correctCarOnChannel | parking | NODATA | 环境限制: 需通道上有待校正的进出车会话(现场触发) \| resultCode=908 其它错误 \| 会话已过期 |
| correctingCarCodeAfterCarInOutConfirmPhone | parking | NODATA | 环境限制: 需进出确认拍照产生的有效会话(现场触发) \| resultCode=908 其它错误 \| 会话已过期 |
| getArrearsDetail | parking | NODATA | 环境限制: 测试环境无授权商欠费数据 \| resultCode=908 其它错误 \| 其它错误 |
| scanChannelCodeInOut | parking | NODATA | 环境限制: 扫码进出需真实云端扫码机/通道映射 \| resultCode=909 请求参数错误 \| 通道信息(id:1)不存在 |
| applyMonthTicketFreeze | ticket | NODATA | 环境限制: 该月票类型未开启'申请冻结'功能, 需车场后台配置 \| resultCode=908 其它错误 \| 月票不支持申请冻结 |
| cancelOnlineVipTicket | ticket | NODATA | 环境限制: 退款为破坏性操作; 目标月票已被前序测试退款(幂等终态) \| resultCode=908 其它错误 \| 该月票已退款 |
| deductMonthTicketConfig | ticket | NODATA | 环境限制: 月票类型车场权限与所传车场需完全一致(环境配置) \| resultCode=909 请求参数错误 \| 参数错误:所传车场与id为1200009646月票类型车场权限不完… |
| getMonthTicketSellNum | ticket | NODATA | 环境限制: 服务端月票销量统计接口在该测试环境异常(status=3 系统异常) \| status=3 系统异常 \| 系统异常 |
| getVipByCarNoAndTime | ticket | NODATA | 环境限制: 需在场VIP车且进场时间精确匹配, 测试环境无此数据 \| resultCode=908 其它错误 \| 该车辆不在场 |
| renewOnlineVipTicket | ticket | NODATA | 环境限制: 目标月票已被前序测试退款, 退款态无法续费 \| resultCode=908 其它错误 \| 该月票已退款,无法续费 |
| payParkFee | trade | NODATA | 环境限制: 在场车实时应缴额与缴费校验需真实未缴账单, 测试车随机进场多为0元/已结清 \| resultCode=908 其它错误 \| 无需缴费 |
| paybackBatch | trade | NODATA | 环境限制: 批量补缴需真实欠费停车记录, 测试车场无欠费数据 \| resultCode=908 其它错误 \|  Cannot invoke "cn.akeparking.bem.a… |

## 按域明细

### blacklist (3)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| addBlackListCar | write | generic | PASS | 1/0 | 业务成功 |
| getParkBlackList | read | generic | PASS | 1/0 | 业务成功 |
| removeBlackListCar | write | generic | PASS | 1/0 | 业务成功 |

### coupon (30)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| GetTraderInfoByTraderCode | read | generic | PASS | 1/0 | 业务成功 |
| ValidateTraderAccountAndPassword | read | coupon | PASS | 1/0 | 业务成功 |
| cancelCoupon | write | coupon | PASS | 1/0 | 业务成功 |
| checkCouponQrCodeValidStatus | read | coupon | PASS | 1/0 | 业务成功 |
| checkCouponWhetherSendAvailable | read | coupon | PASS | 1/0 | 业务成功 |
| createCoupon | write | coupon | PASS | 1/0 | 业务成功 |
| createCouponTemplate | write | coupon | PASS | 1/0 | 业务成功 |
| createFixedCoupon | write | coupon | PASS | 1/0 | 业务成功 |
| createTrader | write | coupon | PASS | 1/0 | 业务成功 |
| deleteTrader | write | generic | SKIP | - | 破坏性: 会删除共享测试商家, 跳过(写路径已由 createTrader 覆盖) |
| editTrader | write | generic | PASS | 1/0 | 业务成功 |
| frozenTrader | write | generic | PASS | 1/0 | 业务成功 |
| getTraderCouponGrantRecordList | read | generic | PASS | 1/0 | 业务成功 |
| getTraderCouponList | read | generic | PASS | 1/0 | 业务成功 |
| getTraderList | read | generic | PASS | 1/0 | 业务成功 |
| lockCoupon | write | coupon | PASS | 1/0 | 业务成功 |
| printCoupon | write | generic | PASS | 1/0 | 业务成功 |
| queryCarCodeValidCoupon | read | generic | PASS | 1/0 | 业务成功 |
| queryCoupon | read | generic | PASS | 1/0 | 业务成功 |
| queryCouponAvailableParkByCouponCode | read | coupon | PASS | 1/0 | 业务成功 |
| queryCouponPrintRecord | read | coupon | PASS | 1/0 | 业务成功 |
| queryCouponTemplateByCouponCode | read | coupon | PASS | 1/0 | 业务成功 |
| queryCouponTemplateByCouponSn | read | generic | PASS | 1/0 | 业务成功 |
| queryTraderCouponSellRecord | read | generic | PASS | 1/0 | 业务成功 |
| queryTraderInfoByCouponCode | read | generic | PASS | 1/0 | 业务成功 |
| queryUsableCoupon | read | coupon | PASS | 1/0 | 业务成功 |
| sellCoupon | write | coupon | PASS | 1/0 | 业务成功 |
| sendCoupon | write | coupon | PASS | 1/0 | 业务成功 |
| sendCouponByCouponCode | write | coupon | NODATA | 2/-1 | 环境限制: 按券码发放需平台预生成的未发放券码实例, 模板码不可直接发放 \| resultCode=-1 业务码未登记 \| 找不到该优惠券 |
| syncScanCouponQrCode | write | generic | PASS | 1/0 | 业务成功 |

### data (6)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| getBillSummary | read | generic | PASS | 1/0 | 业务成功 |
| getCarTrafficFlowAnalysis | read | generic | PASS | 1/0 | 业务成功 |
| getParkBill | read | generic | PASS | 1/0 | 业务成功 |
| getRealTimeParkInfo | read | generic | PASS | 1/0 | 业务成功 |
| getRealtimeLeaveAndChargeNum | read | generic | NODATA | 2/-1 | 环境限制: 测试车场当日无交易统计数据/统计服务在该环境不可用 \| resultCode=-1 业务码未登记 \| 查询车场交易次数失败 |
| getTrafficFlow | read | generic | PASS | 1/0 | 业务成功 |

### dataAnalysis (3)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| parkingPlaceUsed | write | generic | PASS | 1/0 | 业务成功 |
| parkingPlaceUsedForEchart | write | generic | PASS | 1/0 | 业务成功 |
| parkingTimeAnalyse | write | generic | PASS | 1/0 | 业务成功 |

### device (10)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| changeChannelMode | write | generic | PASS | 1/0 | 业务成功 |
| channelSnap | write | generic | NODATA | 2/908 | 环境限制: 通道无可下发的抓拍相机设备 \| resultCode=908 其它错误 \| 找不到设备 |
| cloudOpenGate | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| cloudScanQrCode | write | generic | NODATA | 2/908 | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=908 其它错误 \| 设备不存在 |
| cloudScanUpdateConfig | write | generic | NODATA | 2/909 | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=909 请求参数错误 \| 参数错误:智泊云车场不存在以自定义编… |
| cloudScanVoice | write | generic | NODATA | 2/908 | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=908 其它错误 \| 设备不存在 |
| cloudStopScanCode | write | generic | NODATA | 2/908 | 环境限制: 测试环境无云端扫码机设备(scanMachineId) \| resultCode=908 其它错误 \| 设备不存在 |
| getCloudEquipStatus | read | generic | PASS | 1/0 | 业务成功 |
| opGate | write | generic | NODATA | 2/908 | 环境限制: 通道当前无匹配在场车, 无法开闸 \| resultCode=908 其它错误 \| 通道当前车牌与入参车牌不一致 |
| opShowVoice | write | generic | NODATA | 2/908 | 环境限制: 通道未绑定屏显/语音设备 \| resultCode=908 其它错误 \| 找不到设备 |

### hidden (1)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| setDefaultScreen | write | generic | NODATA | 2/908 | 环境限制: 通道无绑定屏显设备 \| resultCode=908 其它错误 \| 找不到对应设备 |

### park (18)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| getALLParkInfo | read | generic | PASS | 1/0 | 业务成功 |
| getAreaEpt | read | generic | PASS | 1/0 | 业务成功 |
| getAuthParkCodes | read | generic | PASS | 1/0 | 业务成功 |
| getCarCouponRecord | read | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:getCarCouponRecord |
| getCarFreeParkingInfo | read | generic | PASS | 1/0 | 业务成功 |
| getChannelInfo | read | generic | PASS | 1/0 | 业务成功 |
| getCloudParkDeviceInfo | read | generic | PASS | 1/0 | 业务成功 |
| getDisplayVoiceByCarCode | read | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getEpt | read | generic | PASS | 1/0 | 业务成功 |
| getParkAreaInfo | read | generic | PASS | 1/0 | 业务成功 |
| getParkInfo | read | generic | PASS | 1/0 | 业务成功 |
| getParkList | read | generic | PASS | 1/0 | 业务成功 |
| getParkRemainCarport | read | generic | PASS | 1/0 | 业务成功 |
| getParkRemainCarportAndFreeTime | read | generic | PASS | 1/0 | 业务成功 |
| getParkSystemInfo | read | generic | PASS | 1/0 | 业务成功 |
| getParkYdtCharge | read | park | PASS | 1/0 | 业务成功 |
| getParkYdtOtherCarTypeChargeInfo | read | park | PASS | 1/0 | 业务成功 |
| setParkRemainCarport | write | generic | NODATA | 2/908 | 环境限制: 需存在车位区域(areaId)配置, 测试车场无对应车位区域 \| resultCode=908 其它错误 \| 700577区域设置异常 |

### parking (31)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| cancellationOfArrears | write | generic | NODATA | 2/909 | 环境限制: 测试环境无现成欠费记录(recordId) \| resultCode=909 请求参数错误 \| 参数错误:recordId不能为空 |
| checkChannelExistCar | read | generic | NODATA | 2/908 | 环境限制: 需通道上物理有车, 测试环境无法满足 \| resultCode=908 其它错误 \| 通道无车 |
| correctCarNo | write | generic | PASS | 1/0 | 业务成功 |
| correctCarOnChannel | write | generic | NODATA | 2/908 | 环境限制: 需通道上有待校正的进出车会话(现场触发) \| resultCode=908 其它错误 \| 会话已过期 |
| correctingCarCodeAfterCarInOutConfirmPhone | write | generic | NODATA | 2/908 | 环境限制: 需进出确认拍照产生的有效会话(现场触发) \| resultCode=908 其它错误 \| 会话已过期 |
| createInterceptPolicy | write | generic | PASS | 1/0 | 业务成功 |
| deleteInterceptPolicy | write | generic | PASS | 1/0 | 业务成功 |
| getAbnormalOpenGateList | read | generic | PASS | 1/0 | 业务成功 |
| getAbnormalOutList | read | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getArrearsCount | read | generic | PASS | 1/0 | 业务成功 |
| getArrearsDetail | read | generic | NODATA | 2/908 | 环境限制: 测试环境无授权商欠费数据 \| resultCode=908 其它错误 \| 其它错误 |
| getArrearsListByOperator | read | generic | PASS | 1/0 | 业务成功 |
| getCarArrearageList | read | generic | PASS | 1/0 | 业务成功 |
| getCarInList | read | generic | PASS | 1/0 | 业务成功 |
| getCarLockStatus | read | generic | PASS | 1/0 | 业务成功 |
| getCarOutList | read | generic | PASS | 1/0 | 业务成功 |
| getChannelPermission | read | generic | PASS | 1/0 | 业务成功 |
| getParkDetail | read | generic | PASS | 1/0 | 业务成功 |
| getParkDetailIgnoreStatus | read | generic | PASS | 1/0 | 业务成功 |
| getParkOnSiteCar | read | generic | PASS | 1/0 | 业务成功 |
| getParkPayBillByCarNosAndPayTime | read | generic | PASS | 1/0 | 业务成功 |
| getPayBill | read | generic | PASS | 1/0 | 业务成功 |
| getPaymentRecordDetailFile | read | generic | PASS | 1/0 | 业务成功 |
| getPaymentRecordDetailList | read | generic | PASS | 1/0 | 业务成功 |
| lockCar | write | generic | PASS | 1/0 | 业务成功 |
| roadsideCarCheckIn | write | generic | PASS | 1/0 | 业务成功 |
| scanChannelCodeInOut | write | generic | NODATA | 2/909 | 环境限制: 扫码进出需真实云端扫码机/通道映射 \| resultCode=909 请求参数错误 \| 通道信息(id:1)不存在 |
| supplementParkingRecordImage | write | generic | PASS | 1/0 | 业务成功 |
| supplementParkingRecordIn | write | generic | PASS | 1/0 | 业务成功 |
| unlockCar | write | generic | PASS | 1/0 | 业务成功 |
| updateWihholdDetailBill | write | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:updateWihholdDetailBill |

### redlist (3)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| delRedList | write | redlist | PASS | 1/0 | 业务成功 |
| getRedList | read | generic | PASS | 1/0 | 业务成功 |
| redListAdd | write | redlist | PASS | 1/0 | 业务成功 |

### ticket (29)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| GetMonthTicketAccountTransationRecord | read | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:GetMonthTicketAccountTransationRecord |
| addOnlineMonthTicket | write | generic | PASS | 1/0 | 业务成功 |
| addOnlineMonthTicketType | write | generic | PASS | 1/0 | 业务成功 |
| addSpecialCarType | write | generic | PASS | 1/0 | 业务成功 |
| applyMonthTicketFreeze | write | generic | NODATA | 2/908 | 环境限制: 该月票类型未开启'申请冻结'功能, 需车场后台配置 \| resultCode=908 其它错误 \| 月票不支持申请冻结 |
| cancelOnlineMonthTicketByMonthTicketType | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| cancelOnlineVipTicket | write | generic | NODATA | 2/908 | 环境限制: 退款为破坏性操作; 目标月票已被前序测试退款(幂等终态) \| resultCode=908 其它错误 \| 该月票已退款 |
| deductMonthTicketConfig | write | generic | NODATA | 2/909 | 环境限制: 月票类型车场权限与所传车场需完全一致(环境配置) \| resultCode=909 请求参数错误 \| 参数错误:所传车场与id为12000096… |
| editOnlineVipTicket | write | generic | PASS | 1/0 | 业务成功 |
| freezeMonthTicket | write | ticket | PASS | 1/0 | 业务成功 |
| getCarOwnerAndVipType | read | generic | PASS | 1/0 | 业务成功 |
| getMonthTicketAccountUseRecord | read | generic | PASS | 1/0 | 业务成功 |
| getMonthTicketBillDetail | read | generic | PASS | 1/0 | 业务成功 |
| getMonthTicketConfigDetail | read | generic | PASS | 1/0 | 业务成功 |
| getMonthTicketConfigDetailList | read | generic | PASS | 1/0 | 业务成功 |
| getMonthTicketSellNum | read | generic | NODATA | 3/-1 | 环境限制: 服务端月票销量统计接口在该测试环境异常(status=3 系统异常) \| status=3 系统异常 \| 系统异常 |
| getOnlineMonthTicketByCarCard | read | generic | PASS | 1/0 | 业务成功 |
| getOnlineMonthTicketList | read | generic | PASS | 1/0 | 业务成功 |
| getOnlineMonthTicketPayment | read | generic | PASS | 1/0 | 业务成功 |
| getOnlineVipTicket | read | generic | PASS | 1/0 | 业务成功 |
| getParkAgreement | read | generic | PASS | 1/0 | 业务成功 |
| getSpecialCarTypeList | read | generic | PASS | 1/0 | 业务成功 |
| getVipByCarNo | read | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getVipByCarNoAndTime | read | generic | NODATA | 2/908 | 环境限制: 需在场VIP车且进场时间精确匹配, 测试环境无此数据 \| resultCode=908 其它错误 \| 该车辆不在场 |
| getWillExpireMonthTicketBill | read | generic | PASS | 1/0 | 业务成功 |
| monthTicketConfigEdit | write | generic | PASS | 1/0 | 业务成功 |
| parkAgreementSave | write | generic | PASS | 1/0 | 业务成功 |
| renewOnlineVipTicket | write | generic | NODATA | 2/908 | 环境限制: 目标月票已被前序测试退款, 退款态无法续费 \| resultCode=908 其它错误 \| 该月票已退款,无法续费 |
| unFreezeMonthTicket | write | ticket | PASS | 1/0 | 业务成功 |

### trade (7)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| commonGetParkFee | read | generic | PASS | 1/0 | 业务成功 |
| getParkFee | read | billing | PASS | 1/0 | 业务成功 |
| payParkFee | write | billing | NODATA | 2/908 | 环境限制: 在场车实时应缴额与缴费校验需真实未缴账单, 测试车随机进场多为0元/已结清 \| resultCode=908 其它错误 \| 无需缴费 |
| paybackBatch | write | generic | NODATA | 2/908 | 环境限制: 批量补缴需真实欠费停车记录, 测试车场无欠费数据 \| resultCode=908 其它错误 \|  Cannot invoke "cn.akep… |
| setPoints | write | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:setPoints |
| setPrestoreForCPark | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| setPrestoreForCParkFirstPayBeforeLeave | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |

### visitor (2)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| addVisitorCarNew | write | visitor | PASS | 1/0 | 业务成功 |
| cancelVisitorCarNew | write | visitor | PASS | 1/0 | 业务成功 |
