# openydt-cli 端到端测试报告

- 环境: 测试环境 (openapi-test.yidianting.com.cn)
- 接口总数(已尝试): 144
- 结果: ✓PASS=61, BIZFAIL=67, NODEPLOY=11, ERROR=1, SKIP=4
- 成功调通业务层(PASS+BIZFAIL): 128 / 144
- 调用次数: 140

> PASS=业务成功(status=1);BIZFAIL=接口已调通但业务校验未过(status=2,多因测试数据/入参,附 resultCode);NODEPLOY=测试环境未部署该接口(接口不存在,非 CLI 问题);ERROR=传输/签名/鉴权/系统异常;SKIP=前置依赖缺失未尝试(附原因)。

## 需关注(非 PASS)

| 命令 | 域 | 结果 | 说明 |
|---|---|---|---|
| addBlackListCar | blacklist | SKIP | 无特殊车辆类型ID |
| createTrader | coupon | BIZFAIL | resultCode=908 其它错误 \| 手机号不正确 |
| createCouponTemplate | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:traderCode不能为空 |
| sellCoupon | coupon | SKIP | 缺 traderCode / couponTemplateCode |
| sendCoupon | coupon | SKIP | 缺 couponTemplateCode |
| ValidateTraderAccountAndPassword | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:traderUserAccount不能为空 |
| cancelCoupon | coupon | BIZFAIL | resultCode=908 其它错误 \| 商家不存在或未启用 |
| checkCouponQrCodeValidStatus | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| checkCouponWhetherSendAvailable | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| createCoupon | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:parkCodeList中的1ZS7H5PQH9不在授权商车场权限中 |
| createFixedCoupon | coupon | BIZFAIL | resultCode=908 其它错误 \| 系统繁忙 |
| deleteTrader | coupon | BIZFAIL | resultCode=908 其它错误 \| 参数异常 |
| editTrader | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:loginAccount不能为空 |
| frozenTrader | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:status不能为空 |
| lockCoupon | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:url不能为空 |
| printCoupon | coupon | BIZFAIL | resultCode=908 其它错误 \| 商家不存在 |
| queryCouponAvailableParkByCouponCode | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| queryCouponPrintRecord | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| queryCouponTemplateByCouponCode | coupon | BIZFAIL | resultCode=0  \| 券模板不存在 |
| queryTraderCouponSellRecord | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:parkCodeList不能为空 |
| queryUsableCoupon | coupon | BIZFAIL | resultCode=908 其它错误 \| 无权限操作该销售账单 |
| sendCouponByCouponCode | coupon | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| getRealtimeLeaveAndChargeNum | data | BIZFAIL | resultCode=-1 业务码未登记 \| 查询车场交易次数失败 |
| changeChannelMode | device | BIZFAIL | resultCode=908 其它错误 \| 找不到通道 |
| channelSnap | device | BIZFAIL | resultCode=908 其它错误 \| 找不到通道 |
| cloudOpenGate | device | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| cloudScanQrCode | device | BIZFAIL | resultCode=908 其它错误 \| 设备不存在 |
| cloudScanUpdateConfig | device | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:智泊云车场不存在以自定义编码11231231231的通道 |
| cloudScanVoice | device | BIZFAIL | resultCode=908 其它错误 \| 设备不存在 |
| cloudStopScanCode | device | BIZFAIL | resultCode=908 其它错误 \| 设备不存在 |
| opGate | device | BIZFAIL | resultCode=908 其它错误 \| 通道当前车牌与入参车牌不一致 |
| opShowVoice | device | BIZFAIL | resultCode=908 其它错误 \| 找不到设备 |
| setDefaultScreen | hidden | BIZFAIL | resultCode=908 其它错误 \| Cannot invoke "cn.akeparking.bem.common.facade.parkconfig.model.dto.… |
| getCarCouponRecord | park | NODEPLOY | 接口不存在(测试环境未部署) \| status=6 接口不存在:getCarCouponRecord |
| getCarFreeParkingInfo | park | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:parkingCode不能为空 |
| getDisplayVoiceByCarCode | park | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getParkYdtOtherCarTypeChargeInfo | park | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:standardSeq不能为空 |
| supplementParkingRecordIn | parking | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:parkOrArea不能为空 |
| getParkOnSiteCar | parking | BIZFAIL | resultCode=909 请求参数错误 \| 进场时间区间间隔不能大于1个月 |
| cancellationOfArrears | parking | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:recordId不能为空 |
| checkChannelExistCar | parking | BIZFAIL | resultCode=908 其它错误 \| 找不到通道 |
| correctCarNo | parking | BIZFAIL | resultCode=908 其它错误 \| For input string: "20180101101021224433" |
| correctCarOnChannel | parking | BIZFAIL | resultCode=908 其它错误 \| 会话已过期 |
| correctingCarCodeAfterCarInOutConfirmPhone | parking | BIZFAIL | resultCode=-1 业务码未登记 \| 找不到对应的通道编码 |
| getAbnormalOutList | parking | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getArrearsDetail | parking | BIZFAIL | resultCode=908 其它错误 \| 其它错误 |
| getChannelPermission | parking | BIZFAIL | resultCode=-1 业务码未登记 \| 找不到通道 |
| getParkDetailIgnoreStatus | parking | BIZFAIL | resultCode=908 其它错误 \| 没找到对应的停车记录 |
| getParkPayBillByCarNosAndPayTime | parking | BIZFAIL | resultCode=909 请求参数错误 \| carNoList不能为空 |
| roadsideCarCheckIn | parking | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:carNo和positionNo不能为空 |
| scanChannelCodeInOut | parking | BIZFAIL | resultCode=909 请求参数错误 \| 通道信息(id:1)不存在 |
| updateWihholdDetailBill | parking | NODEPLOY | 接口不存在(测试环境未部署) \| status=6 接口不存在:updateWihholdDetailBill |
| delRedList | redlist | BIZFAIL | resultCode=908 其它错误 \| Cannot invoke "cn.akeparking.bem.domain.identity.carclass.model.enti… |
| addSpecialCarType | ticket | BIZFAIL | resultCode=908 其它错误 \| 所选车场范围内已存在同名的特殊车辆类型配置 |
| addOnlineMonthTicket | ticket | BIZFAIL | resultCode=908 其它错误 \| 该月票订单已存在 |
| addSpecialCarType | ticket | BIZFAIL | resultCode=908 其它错误 \| 所选车场范围内已存在同名的特殊车辆类型配置 |
| GetMonthTicketAccountTransationRecord | ticket | NODEPLOY | 接口不存在(测试环境未部署) \| status=6 接口不存在:GetMonthTicketAccountTransationRecord |
| applyMonthTicketFreeze | ticket | BIZFAIL | resultCode=908 其它错误 \| 月票订单id不合法 |
| cancelOnlineMonthTicketByMonthTicketType | ticket | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| cancelOnlineVipTicket | ticket | BIZFAIL | resultCode=908 其它错误 \| 找不到对应的月票 |
| deductMonthTicketConfig | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:monthTicketConfigId不能为空 |
| editOnlineVipTicket | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:userName不能为空 |
| freezeMonthTicket | ticket | BIZFAIL | resultCode=908 其它错误 \| 冻结开始时间不能早于当天 |
| getCarOwnerAndVipType | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:carNo不能为空 |
| getMonthTicketAccountUseRecord | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:monthTicketBillId不能为空 |
| getMonthTicketBillDetail | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:thirdBillCodes不能为空 |
| getMonthTicketConfigDetail | ticket | BIZFAIL | resultCode=908 其它错误 \| 无该月票类型信息 |
| getMonthTicketSellNum | ticket | ERROR | status=3 系统异常 \| 系统异常 |
| getOnlineMonthTicketPayment | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:操作时间范围不能大于3个月 |
| getOnlineVipTicket | ticket | BIZFAIL | resultCode=908 其它错误 \| 找不到对应的月票开通记录 |
| getParkAgreement | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:parkCodeList不能为空 |
| getVipByCarNo | ticket | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getVipByCarNoAndTime | ticket | BIZFAIL | resultCode=908 其它错误 \| 该车辆不在场 |
| getWillExpireMonthTicketBill | ticket | BIZFAIL | resultCode=-1 业务码未登记 \| 网络异常 |
| parkAgreementSave | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:agreementTitle不能为空 |
| renewOnlineVipTicket | ticket | BIZFAIL | resultCode=909 请求参数错误 \| 参数错误:车场1ZS7H5PQH9不存在 |
| unFreezeMonthTicket | ticket | BIZFAIL | resultCode=908 其它错误 \| Cannot invoke "cn.akeparking.bem.domain.identity.monthticket.bill.mo… |
| payParkFee | trade | BIZFAIL | resultCode=908 其它错误 \| 费用超出仍应缴金额 |
| setPoints | trade | NODEPLOY | 接口不存在(测试环境未部署) \| status=6 接口不存在:setPoints |
| setPrestoreForCPark | trade | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| setPrestoreForCParkFirstPayBeforeLeave | trade | NODEPLOY | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| addVisitorCarNew | visitor | SKIP | 无特殊车辆类型ID |
| cancelVisitorCarNew | visitor | BIZFAIL | resultCode=908 其它错误 \| 找不到对应访客记录 |

## 按域明细

### blacklist (3)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| addBlackListCar | write | list | SKIP | - | 无特殊车辆类型ID |
| getParkBlackList | read | list | PASS | 1/0 | 业务成功 |
| removeBlackListCar | write | generic | PASS | 1/0 | 业务成功 |

### coupon (30)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| GetTraderInfoByTraderCode | read | generic | PASS | 1/0 | 业务成功 |
| ValidateTraderAccountAndPassword | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:traderUserAccount不能为空 |
| cancelCoupon | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 商家不存在或未启用 |
| checkCouponQrCodeValidStatus | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| checkCouponWhetherSendAvailable | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| createCoupon | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:parkCodeList中的1ZS7H5PQH9不在授权商车场权限中 |
| createCouponTemplate | write | coupon | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:traderCode不能为空 |
| createFixedCoupon | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 系统繁忙 |
| createTrader | write | coupon | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 手机号不正确 |
| deleteTrader | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 参数异常 |
| editTrader | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:loginAccount不能为空 |
| frozenTrader | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:status不能为空 |
| getTraderCouponGrantRecordList | read | generic | PASS | 1/0 | 业务成功 |
| getTraderCouponList | read | generic | PASS | 1/0 | 业务成功 |
| getTraderList | read | coupon | PASS | 1/0 | 业务成功 |
| lockCoupon | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:url不能为空 |
| printCoupon | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 商家不存在 |
| queryCarCodeValidCoupon | read | generic | PASS | 1/0 | 业务成功 |
| queryCoupon | read | generic | PASS | 1/0 | 业务成功 |
| queryCouponAvailableParkByCouponCode | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| queryCouponPrintRecord | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| queryCouponTemplateByCouponCode | read | generic | BIZFAIL | 2/0 | resultCode=0  \| 券模板不存在 |
| queryCouponTemplateByCouponSn | read | generic | PASS | 1/0 | 业务成功 |
| queryTraderCouponSellRecord | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:parkCodeList不能为空 |
| queryTraderInfoByCouponCode | read | generic | PASS | 1/0 | 业务成功 |
| queryUsableCoupon | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 无权限操作该销售账单 |
| sellCoupon | write | coupon | SKIP | - | 缺 traderCode / couponTemplateCode |
| sendCoupon | write | coupon | SKIP | - | 缺 couponTemplateCode |
| sendCouponByCouponCode | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:couponCode不能为空 |
| syncScanCouponQrCode | write | generic | PASS | 1/0 | 业务成功 |

### data (6)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| getBillSummary | read | generic | PASS | 1/0 | 业务成功 |
| getCarTrafficFlowAnalysis | read | generic | PASS | 1/0 | 业务成功 |
| getParkBill | read | generic | PASS | 1/0 | 业务成功 |
| getRealTimeParkInfo | read | generic | PASS | 1/0 | 业务成功 |
| getRealtimeLeaveAndChargeNum | read | generic | BIZFAIL | 2/-1 | resultCode=-1 业务码未登记 \| 查询车场交易次数失败 |
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
| changeChannelMode | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到通道 |
| channelSnap | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到通道 |
| cloudOpenGate | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| cloudScanQrCode | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 设备不存在 |
| cloudScanUpdateConfig | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:智泊云车场不存在以自定义编码11231231231的通道 |
| cloudScanVoice | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 设备不存在 |
| cloudStopScanCode | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 设备不存在 |
| getCloudEquipStatus | read | generic | PASS | 1/0 | 业务成功 |
| opGate | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 通道当前车牌与入参车牌不一致 |
| opShowVoice | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到设备 |

### hidden (1)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| setDefaultScreen | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| Cannot invoke "cn.akeparking.bem.common.facade.parkconfig.… |

### park (18)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| getALLParkInfo | read | generic | PASS | 1/0 | 业务成功 |
| getAreaEpt | read | generic | PASS | 1/0 | 业务成功 |
| getAuthParkCodes | read | generic | PASS | 1/0 | 业务成功 |
| getCarCouponRecord | read | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:getCarCouponRecord |
| getCarFreeParkingInfo | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:parkingCode不能为空 |
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
| getParkYdtCharge | read | generic | PASS | 1/0 | 业务成功 |
| getParkYdtOtherCarTypeChargeInfo | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:standardSeq不能为空 |
| setParkRemainCarport | write | generic | PASS | 1/0 | 业务成功 |

### parking (31)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| cancellationOfArrears | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:recordId不能为空 |
| checkChannelExistCar | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到通道 |
| correctCarNo | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| For input string: "20180101101021224433" |
| correctCarOnChannel | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 会话已过期 |
| correctingCarCodeAfterCarInOutConfirmPhone | write | generic | BIZFAIL | 2/-1 | resultCode=-1 业务码未登记 \| 找不到对应的通道编码 |
| createInterceptPolicy | write | generic | PASS | 1/0 | 业务成功 |
| deleteInterceptPolicy | write | generic | PASS | 1/0 | 业务成功 |
| getAbnormalOpenGateList | read | generic | PASS | 1/0 | 业务成功 |
| getAbnormalOutList | read | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getArrearsCount | read | generic | PASS | 1/0 | 业务成功 |
| getArrearsDetail | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 其它错误 |
| getArrearsListByOperator | read | generic | PASS | 1/0 | 业务成功 |
| getCarArrearageList | read | generic | PASS | 1/0 | 业务成功 |
| getCarInList | read | generic | PASS | 1/0 | 业务成功 |
| getCarLockStatus | read | generic | PASS | 1/0 | 业务成功 |
| getCarOutList | read | generic | PASS | 1/0 | 业务成功 |
| getChannelPermission | read | generic | BIZFAIL | 2/-1 | resultCode=-1 业务码未登记 \| 找不到通道 |
| getParkDetail | read | billing | PASS | 1/0 | 业务成功 |
| getParkDetailIgnoreStatus | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 没找到对应的停车记录 |
| getParkOnSiteCar | read | billing | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 进场时间区间间隔不能大于1个月 |
| getParkPayBillByCarNosAndPayTime | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| carNoList不能为空 |
| getPayBill | read | billing | PASS | 1/0 | 业务成功 |
| getPaymentRecordDetailFile | read | generic | PASS | 1/0 | 业务成功 |
| getPaymentRecordDetailList | read | billing | PASS | 1/0 | 业务成功 |
| lockCar | write | generic | PASS | 1/0 | 业务成功 |
| roadsideCarCheckIn | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:carNo和positionNo不能为空 |
| scanChannelCodeInOut | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 通道信息(id:1)不存在 |
| supplementParkingRecordImage | write | generic | PASS | 1/0 | 业务成功 |
| supplementParkingRecordIn | write | billing | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:parkOrArea不能为空 |
| unlockCar | write | generic | PASS | 1/0 | 业务成功 |
| updateWihholdDetailBill | write | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:updateWihholdDetailBill |

### redlist (3)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| delRedList | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| Cannot invoke "cn.akeparking.bem.domain.identity.carclass.… |
| getRedList | read | generic | PASS | 1/0 | 业务成功 |
| redListAdd | write | generic | PASS | 1/0 | 业务成功 |

### ticket (30)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| GetMonthTicketAccountTransationRecord | read | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:GetMonthTicketAccountTransationRecord |
| addOnlineMonthTicket | write | monthticket | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 该月票订单已存在 |
| addOnlineMonthTicketType | write | monthticket | PASS | 1/0 | 业务成功 |
| addSpecialCarType | write | monthticket | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 所选车场范围内已存在同名的特殊车辆类型配置 |
| addSpecialCarType | write | list | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 所选车场范围内已存在同名的特殊车辆类型配置 |
| applyMonthTicketFreeze | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 月票订单id不合法 |
| cancelOnlineMonthTicketByMonthTicketType | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| cancelOnlineVipTicket | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到对应的月票 |
| deductMonthTicketConfig | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:monthTicketConfigId不能为空 |
| editOnlineVipTicket | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:userName不能为空 |
| freezeMonthTicket | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 冻结开始时间不能早于当天 |
| getCarOwnerAndVipType | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:carNo不能为空 |
| getMonthTicketAccountUseRecord | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:monthTicketBillId不能为空 |
| getMonthTicketBillDetail | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:thirdBillCodes不能为空 |
| getMonthTicketConfigDetail | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 无该月票类型信息 |
| getMonthTicketConfigDetailList | read | generic | PASS | 1/0 | 业务成功 |
| getMonthTicketSellNum | read | generic | ERROR | 3/-1 | status=3 系统异常 \| 系统异常 |
| getOnlineMonthTicketByCarCard | read | generic | PASS | 1/0 | 业务成功 |
| getOnlineMonthTicketList | read | monthticket | PASS | 1/0 | 业务成功 |
| getOnlineMonthTicketPayment | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:操作时间范围不能大于3个月 |
| getOnlineVipTicket | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到对应的月票开通记录 |
| getParkAgreement | read | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:parkCodeList不能为空 |
| getSpecialCarTypeList | read | list | PASS | 1/0 | 业务成功 |
| getVipByCarNo | read | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| getVipByCarNoAndTime | read | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 该车辆不在场 |
| getWillExpireMonthTicketBill | read | generic | BIZFAIL | 2/-1 | resultCode=-1 业务码未登记 \| 网络异常 |
| monthTicketConfigEdit | write | generic | PASS | 1/0 | 业务成功 |
| parkAgreementSave | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:agreementTitle不能为空 |
| renewOnlineVipTicket | write | generic | BIZFAIL | 2/909 | resultCode=909 请求参数错误 \| 参数错误:车场1ZS7H5PQH9不存在 |
| unFreezeMonthTicket | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| Cannot invoke "cn.akeparking.bem.domain.identity.monthtick… |

### trade (7)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| commonGetParkFee | read | generic | PASS | 1/0 | 业务成功 |
| getParkFee | read | billing | PASS | 1/0 | 业务成功 |
| payParkFee | write | billing | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 费用超出仍应缴金额 |
| paybackBatch | write | generic | PASS | 1/0 | 业务成功 |
| setPoints | write | generic | NODEPLOY | 6/-1 | 接口不存在(测试环境未部署) \| status=6 接口不存在:setPoints |
| setPrestoreForCPark | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |
| setPrestoreForCParkFirstPayBeforeLeave | write | generic | NODEPLOY | 9/-1 | 接口不存在(测试环境未部署) \| status=9 接口不存在 |

### visitor (2)

| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |
|---|---|---|---|---|---|
| addVisitorCarNew | write | list | SKIP | - | 无特殊车辆类型ID |
| cancelVisitorCarNew | write | generic | BIZFAIL | 2/908 | resultCode=908 其它错误 \| 找不到对应访客记录 |
