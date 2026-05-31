# openydt 接口索引

> 由 `make index`(scripts/gen-index.sh)从 catalog.json 生成,**勿手改**。统计见 `make counts`。
> 列:cmd | 方向(callable 可主动调 / webhook 平台推送) | 读写 | 是否一等命令 | 说明。
> 一等命令用 `openydt <域> <cmd>` 调;callable 但非一等用 `openydt api <cmd>`;webhook 需自建接收端。

## ad (3)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getAdPvStatistics` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取广告曝光度统计信息 |
| `newAd` | callable | write | · | 第三方接入系统请求智慧停车开放平台新增广告 |
| `startOrFreezeOrDelAd` | callable | write | · | 第三方接入系统请求智慧停车开放平台启用/冻结/删除广告 |

## blacklist (3)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `addBlackListCar` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台添加某个黑名单车辆 |
| `getParkBlackList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询黑名单车辆列表 |
| `removeBlackListCar` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台取消某个黑名单车辆 |

## carSearch (4)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getAuthParkOfMobileInventoryList` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取该账号下的所有授权停车场编码 |
| `getAuthParkOfMobilePatrolList` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取该账号下的所有授权停车场编码 |
| `getInventoryRecord` | callable | read | · | 第三方接入系统请求智慧停车开放平台 获取盘点记录 |
| `inventoryCar` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取城市地图车辆分布数据 |

## cityOperationCoupon (10)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `createCityOperationCouponTemplate` | callable | write | · | 第三方接入系统请求智慧停车开放平台创建城市运营券模板 |
| `deleteCityOperationCouponTemplate` | callable | write | · | 第三方接入系统请求智慧停车开放平台停用城市运营券模板 |
| `getCarValidCityOperationCoupon` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询车牌号可用的城市运营券 |
| `getCityOperationCouponBill` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询发券记录, 券使用记录, 或者指定手机号/车辆号的券记录信息 |
| `grantCityOperationCoupon` | callable | write | · | 第三方接入系统请求智慧停车开放平台给指定车牌发放城市运营券，车辆在停车场可使用优惠券抵扣部分或全部停 |
| `modifyCityOperationCouponTemplate` | callable | write | · | 第三方接入系统请求智慧停车开放平台修改城市运营券模板 |
| `queryCityOperationCouponTemplateInfo` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取获取城市运营券模板信息 |
| `reportCityOperationCouponNotify` | webhook | write | · | 智慧停车开放平台主动上报城市运营券使用券编码变更信息到第三方接入系统 |
| `retrieveCityOperationCoupon` | callable | write | · | 第三方接入系统请求智慧停车开放平台回收城市运营券 |
| `updateReceiveCouponSuccessUrl` | callable | write | · | 第三方接入系统请求智慧停车开放平台更新领券成功后跳转url |

## cloud (1)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `cloudLCDAd` | callable | write | · | 第三方接入系统请求智慧停车开放平台对连接到云端的LCD屏下发屏显广告。 |

## common (10)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getFeeForThi` | webhook | read | · | 智慧停车开放平台主动请求第三方接入系统获取停车费用信息 |
| `payFeeForThi` | webhook | write | · | 智慧停车开放平台主动同步用户缴费信息到第三方车场接入系统，实现车辆行驶至出口后，道闸自动抬杆放行 |
| `reducePrestoreForThi` | callable | write | · | 第三方接入系统主动推送预存款扣费信息到开放平台 |
| `scanChannelQrCodeInForThi` | webhook | write | · | 智慧停车开放平台主动向第三方车场获取进场随机码，第三方车场车辆自助进场 |
| `scanChannelQrCodeOutForThi` | webhook | write | · | 智慧停车开放平台主动向第三方停车场获取离场通道当前车辆的随机码及车牌，车辆自助离场 |
| `selfInOutForThiFlow` | callable | write | · |  |
| `sendTraderCouponForThi` | webhook | write | · | 智慧停车开放平台主动请求第三方接入系统下发商家券到停车场 |
| `setPrestoreForThi` | callable | write | · |  |
| `setPrestoreForThiFlow` | callable | write | · |  |
| `updateTraderCouponForThi` | webhook | write | · | 智慧停车开放平台主动请求第三方接入系统下发商家券变更信息到停车场 |

## community (3)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getAuthCommunities` | callable | read | · | 第三方接入系统请求开放平台获取该账号下的所有授权小区信息 |
| `getCommunityInfo` | callable | read | · | 第三方接入系统请求开放平台查询小区信息 |
| `getCommunityPersonnelInfo` | callable | read | · | 第三方接入系统请求开放平台查询用户的小区通行权限 |

## coupon (31)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `GetTraderInfoByTraderCode` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 获取商家信息 |
| `ValidateTraderAccountAndPassword` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 校验商家用户账户密码 |
| `cancelCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台回收电子券 |
| `checkCouponQrCodeValidStatus` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台检查电子券二维码有效性 |
| `checkCouponWhetherSendAvailable` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台检查电子券是否可发放 |
| `couponFlow` | callable | write | · |  |
| `createCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台创建指定电子优惠券并给指定商户售卖电子券 |
| `createCouponTemplate` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台创建指定电子优惠券并给指定商户售卖电子券 |
| `createFixedCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台创建固定电子优惠券，适合一个二维码多张优惠券的场景。（注意：对同一 |
| `createTrader` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台创建商家 |
| `deleteTrader` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台删除商家(不能恢复，请谨慎) |
| `editTrader` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台编辑商家 |
| `frozenTrader` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台冻结或解冻商家 |
| `getTraderCouponGrantRecordList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询商家列表 |
| `getTraderCouponList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询优惠券列表 |
| `getTraderList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询商家列表 |
| `lockCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台锁定电子券 |
| `printCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台打印电子券，调用此接口后标记券已被发放，调用查询可用的电子券接口不 |
| `queryCarCodeValidCoupon` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询已发放电子券 |
| `queryCoupon` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询电子券信息 |
| `queryCouponAvailableParkByCouponCode` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台根据券编码查询电子券适用车场 |
| `queryCouponPrintRecord` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询电子券打印记录 |
| `queryCouponTemplateByCouponCode` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台根据券模版代码查询券模版信息 |
| `queryCouponTemplateByCouponSn` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台根据券模版代码查询券模版信息 |
| `queryTraderCouponSellRecord` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 查询电子券售卖记录 |
| `queryTraderInfoByCouponCode` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台根据券编码查询商家信息 |
| `queryUsableCoupon` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询可用的电子券 |
| `sellCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台售卖电子券给商家 |
| `sendCoupon` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台给指定车辆发放电子优惠券，车辆在停车场可使用电子优惠券抵扣部分或全 |
| `sendCouponByCouponCode` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台根据券编码发放商家券 |
| `syncScanCouponQrCode` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台同步电子券二维码扫码 |

## data (9)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getBillSummary` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取某停车场缴费账单信息 |
| `getCarTrafficFlowAnalysis` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车流量车牌top分布 |
| `getParkBill` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询某车场账单汇总信息 |
| `getRealTimeParkInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场实时数据 |
| `getRealtimeLeaveAndChargeNum` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车场当天的出车次数、交易次数 |
| `getTrafficFlow` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取单个停车场每分钟车流量情况 |
| `parkingPlaceUsed` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台获取车位使用情况数据 |
| `parkingPlaceUsedForEchart` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台获取车位使用情况数据，用于绘制echart笛卡尔坐标系上的热力图 |
| `parkingTimeAnalyse` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台获取车场停车时长分析数据 |

## dataAnalysis (16)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `cityMapCarDistribution` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取城市地图车辆分布数据 |
| `dealForPark` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车场交易数据 |
| `dealForParkVer2` | callable | write | · | 第三方接入系统请求智慧停车开放平台 车场交易数据-版本2 |
| `dealRealTime` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取实时交易数据 |
| `dealSummarise` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取交易总览数据 |
| `dealSummariseForHKAndMacaoVer2` | callable | write | · | 第三方接入系统请求智慧停车开放平台 交易总览数据(港澳)-版本2 |
| `dealTrendencyByDate` | callable | write | · | 第三方接入系统请求智慧停车开放平台根据日期范围获取交易趋势数据 |
| `dealTrendencyByNum` | callable | write | · | 第三方接入系统请求智慧停车开放平台根据数量（几日内或几周内）获取交易趋势数据 |
| `getCarArrearRecordList` | callable | read | · | 第三方接入系统请求智慧停车开放平台 查询运营商或车牌欠费记录 |
| `parkMessage` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车场数据 |
| `parkOnlineData` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车场上线情况 |
| `parkOnlineDataOfTotalRemain` | callable | write | · | 第三方接入系统请求智慧停车开放平台 车场上线情况数据(取总剩余车位) |
| `parkingPlaceTurnOver` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车位周转数据 |
| `parkingPlaceUsed` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车位使用情况数据 |
| `parkingPlaceUsedForEchart` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车位使用情况数据，用于绘制echart笛卡尔坐标系上的热力图 |
| `parkingTimeAnalyse` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取车场停车时长分析数据 |

## device (19)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `addMidAccount` | callable | write | · | 第三方接入系统请求智慧停车开放平台一体机账号新增接口 |
| `removeLeavePrompt` | callable | write | · | 第三方接入系统请求智慧停车开放平台取消指定车牌出场提示 |
| `setLeavePrompt` | callable | write | · | 第三方接入系统请求智慧停车开放平台指定车牌出场提示 |
| `setShowMsg` | callable | write | · | 第三方接入系统请求智慧停车开放平台设置基本类型显示以及语言播报 |
| `setVipShowMsg` | callable | write | · | 第三方接入系统请求智慧停车开放平台设置VIP基本类型显示以及语言播报 |
| `changeChannelMode` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台 修改一体机闸机模式 |
| `channelSnap` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台手动抓拍 |
| `cloudOpenGate` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对云停车场某通道开/关闸 |
| `opGate` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台开/关闸 |
| `opShowVoice` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台下发信息到显示屏或者语音播报 |
| `cloudScanQrCode` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对连接到云端的扫码机下发扫码请求 |
| `cloudScanUpdateConfig` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对连接到云端的扫码机下发更新配置请求 |
| `cloudScanVoice` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对连接到云端的扫码机下发语音播报。接收语音播报指令后开始播报，间隔 |
| `cloudStopScanCode` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对连接到云端的扫码机下发停止扫码 |
| `getCloudEquipStatus` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取连接到云端的设备状态信息请求 |
| `reportChannelBindForScan` | webhook | write | · | 智慧停车开放平台主动上报扫码机绑定通道信息到第三方接入系统 |
| `reportQrCode` | webhook | write | · | 智慧停车开放平台主动上报扫码机扫码信息到第三方接入系统 |
| `scanMachineFlow` | callable | write | · |  |
| `setDefaultScreen` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对通道绑定的屏显下发默认显示内容。 |

## hidden (1)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `setDefaultScreen` | callable | write | · | 第三方接入系统请求智慧停车开放平台对通道绑定的屏显下发默认显示内容。 |

## invoice (2)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getBillInvoiceDetail` | callable | read | · | 第三方接入系统请求智慧停车开放平台 根据第三方交易号获取账单 |
| `updateInvoiceStatusOfPayBill` | callable | write | · | 第三方接入系统同步用户缴费信息到智慧停车开放平台 |

## misc (9)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `alarmGuide` | callable | write | · |  |
| `appendixData` | callable | write | · |  |
| `demoData` | callable | write | · |  |
| `edition` | callable | write | · |  |
| `guide` | callable | write | · |  |
| `guideForThirdPark` | callable | write | · |  |
| `procedureData` | callable | write | · |  |
| `trade/getParkFee` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取停车费用接口 |
| `sdkData` | callable | write | · |  |

## other (15)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `extend/getCarInOutCount` | callable | write | · | 通过此接口可传递车场列表获取车辆进出数 |
| `extend/getCoopInfoByCoopAdminUser` | callable | write | · | 通过此接口可传递授权商用户名密码获取授权商信息 |
| `extend/getDayorLastIncome` | callable | write | · | 通过此接口 获取当天与昨天的比较 |
| `getParkDataMonth` | callable | read | · | 第三方接入系统请求智慧停车开放平台根据日期范围获取3个月内的交易统计 |
| `extend/getParkSpace` | callable | write | · | 通过此接口可传递车场列表获取路内外车位使用率 |
| `getPompLoginUserInfoByUserFlag` | callable | read | · | 根据POMP用户账号获取POMP资源权限 |
| `extend/getWeekorLastIncome` | callable | write | · | 通过此接口可获取本周与上一周收入数据比对 |
| `automaticLoginPOMP` | callable | write | · | 通过此接口可不输入密码登陆POMP |
| `automaticLoginVEMS` | callable | write | · | 第三方接入系统请求智慧停车开放平台自动登录vems接口 |
| `getPOMPAuthzParks` | callable | read | · | 根据POMP用户账号获取POMP数据权限 |
| `getPOMPAuthzResources` | callable | read | · | 根据POMP用户账号获取POMP资源权限 |
| `aroudPark` | callable | write | · |  |
| `fastPay` | callable | write | · |  |
| `thirdMemberLogin` | callable | write | · | 此接口可实现第三方会员(手机号)无缝登录智慧停车微信端 |
| `mockInOut` | callable | write | · | 第三方接入系统请求智慧停车开放平台模拟车辆进出场（测试用）。该接口需要在vems通道里设置：无牌车放 |

## park (22)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getParkEquipmentInfo` | callable | read | · | 第三方接入系统请求一点停开放平台获取停车场设备信息 |
| `getParkAreaInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 获取车场区域信息 |
| `getALLParkInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场全部信息 |
| `getAreaEpt` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取区域内所有停车场空车位信息 |
| `getAuthParkCodes` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取该账号下的所有授权停车场编码 |
| `getCarCouponRecord` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车辆优惠券信息列表 |
| `getCarFreeParkingInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 获取车辆免费停车信息 |
| `getCarOwnerInfo` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取停车场车主信息 |
| `getChannelInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取云停车场通道信息 |
| `getCloudParkDeviceInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取云停车场设备信息 |
| `getDisplayVoiceByCarCode` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车辆屏显及语音 |
| `getEpt` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取单个停车场空车位信息 |
| `getMonthTicketAppointmentPark` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取月票车场信息 |
| `getParkEquipmentInfo` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取停车场设备信息 |
| `getParkInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取单个停车场基本信息 |
| `getParkList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取授权商所有停车场信息内容 |
| `getParkRemainCarport` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取单个停车场实时车位信息 |
| `getParkRemainCarportAndFreeTime` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场剩余车位和车场免费时长 |
| `getParkSystemInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场系统信息(vems传统停车场/云停车场) |
| `getParkYdtCharge` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 根据车场车场编码获取收费信息 |
| `getParkYdtOtherCarTypeChargeInfo` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台根据车场UUID 测算的收费标准自增ID以及之前返回的查费信息对象 |
| `setParkRemainCarport` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台设置停车场实时车位信息 |

## parking (46)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getHisParkDetail` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取指定车辆指定时间段停车记录信息（车辆驶出停车场后该接口才能查询 |
| `getParkPayBill` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取某停车场指定时间段缴费账单信息 |
| `getParkingPosition` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取停车位置相关信息 |
| `getParkingSpaceInfo` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取车位数据接口 |
| `reportParkinglotChange` | webhook | write | · | 智慧停车开放平台主动上报mpgs车位信息变更情况到第三方接入系统。信息包括某车位由空状态转换为车辆占 |
| `paymentRecordQuery` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取缴费记录接口，查询账单详情（对账用） |
| `paymentRecordQueryByInTime` | callable | write | · | 第三方接入系统请求智慧停车开放平台根据进场时间获取缴费记录接口，查询账单详情（对账用） |
| `selfInOutForCloudPark` | callable | write | · | 第三方接入系统请求智慧停车开放平台，使得车辆自助进出场（车辆可以是无牌车，也可以是有牌车）。推荐：传 |
| `typingRandomCodeInOut` | callable | write | · | 第三方接入系统请求智慧停车开放平台，使得无牌车输入无牌车随机码进出场 |
| `addCarTags` | callable | write | · | 智汇云车场：添加车辆标签 |
| `checkChannelExistCar` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台，检查通道是否有车(可协助开关闸，自助进出场使用) |
| `correctCarNo` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对指定在场停车记录进行车牌校正 |
| `correctCarOnChannel` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对指定车场、通道的待进出记录进行车牌校正 |
| `correctingCarCodeAfterCarInOutConfirmPhone` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对指定车场的进出车确认拍照后进行车牌校正 |
| `createInterceptPolicy` | callable | write | ✓ | 智汇云车场：创建车场拦截策略接口 |
| `delCarTags` | callable | write | · | 智汇云车场：删除车辆标签 |
| `deleteInterceptPolicy` | callable | write | ✓ | 智汇云车场：刪除车场拦截策略接口 |
| `getCarInList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取指定时间段进场记录信息 |
| `getCarLockStatus` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取指定车辆锁定状态 |
| `getCarOutList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取指定时间段出场记录信息 |
| `getChannelPermission` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车辆进出通道权限 |
| `getParkDetail` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场某条停车记录详情 |
| `getParkDetailIgnoreStatus` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场某条停车记录详情 |
| `getParkOnSiteCar` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车场在场车辆信息 |
| `getParkPayBillByCarNosAndPayTime` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台根据车牌号码及支付时间查询车场账单 |
| `getPayBill` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询某个车辆某次停车的缴费记录信息 |
| `getPaymentRecordDetailFile` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车场支付账单文件 |
| `getPaymentRecordDetailList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取支付账单信息 |
| `lockCar` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对指定车辆进行锁定。车辆锁定后离场，将在停车场出入口提示“车辆已锁 |
| `cancellationOfArrears` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台取消欠费 |
| `getAbnormalOpenGateList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询非系统开闸记录 |
| `getAbnormalOutList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询车辆异常离场记录 |
| `getArrearsCount` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询运营商欠费记录条数 |
| `getArrearsDetail` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查看授权商欠费图片详情 |
| `getArrearsListByOperator` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询运营商欠费记录 |
| `getCarArrearageList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询车辆欠费记录信息 |
| `updateWihholdDetailBill` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台根据支付订单号更新代扣流程记录的订单回调时间和支付状态 |
| `roadsideCarCheckIn` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台路边车辆自助登记 |
| `scanChannelCodeInOut` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台，使得车辆扫通道二维码自助进出场（车辆可以是无牌车，也可以是有牌车 |
| `scanChannelCodeInOutFlow` | callable | write | · |  |
| `supplementParkingRecordImage` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台补录进场后图片补录 |
| `supplementParkingRecordIn` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台补录进场记录 |
| `supplyCarIn` | callable | write | · | 第三方接入系统请求智慧停车开放平台车辆进场数据补录接口 |
| `supplyCarOut` | callable | write | · | 第三方接入系统请求智慧停车开放平台车辆离场数据补录接口 |
| `supplyCarPic` | callable | write | · | 第三方接入系统请求智慧停车开放平台车辆及车牌图片补录接口 |
| `unlockCar` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台对指定车辆进行解锁 |

## points (2)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `pointsGet` | webhook | write | · | 智慧停车开放平台调用第三方接口获取手机号/车牌可用的城市运营积分 |
| `pointsUse` | webhook | write | · | 智慧停车使用第三方运营积分后，主动将使用信息通知第三方 |

## preferential (10)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getExchangeRule` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取积分兑换规则 |
| `getUserCredits` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取用户积分 |
| `getUserCreditsByCoop` | webhook | read | · | 智慧停车开放平台主动请求第三方接入系统获取用户积分 |
| `reduceUserCreditsByCoop` | webhook | write | · | 智慧停车开放平台主动请求第三方接入系统扣减用户积分 |
| `thirdCouponFlow` | callable | write | · |  |
| `thirdCouponGetByCarCode` | webhook | write | · | 智慧停车开放平台调用第三方接口获取车牌可用的优惠券 |
| `thirdCouponGetByCouponCode` | webhook | write | · | 智慧停车开放平台调用第三方接口根据优惠券编码获取优惠券详细信息 |
| `thirdCouponRecycle` | callable | write | · | 第三方发券到智慧停车开放平台 |
| `thirdCouponSend` | callable | write | · | 第三方发券到智慧停车开放平台 |
| `thirdCouponUse` | webhook | write | · | 智慧停车使用第三方优惠券后，主动将使用信息通知第三方 |

## redlist (3)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `delRedList` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台添删除白名单规则 |
| `getRedList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台添新增白名单规则 |
| `redListAdd` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台添新增白名单规则 |

## score (3)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `reduceUserCredits` | callable | write | · | 第三方接入系统请求智慧停车开放平台扣减用户积分 |
| `reverseUserCredits` | callable | write | · | 第三方接入系统请求智慧停车开放平台冲正用户积分 |
| `reverseUserCreditsByCoop` | webhook | write | · | 智慧停车开放平台主动请求第三方接入系统冲正用户积分 |

## sdk (1)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `javaSdk` | callable | write | · |  |

## shareBooking (5)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `bookingSpace` | callable | write | · | 第三方接入系统请求智慧停车开放平台进行车位预订 |
| `cancelBookingSpace` | callable | write | · | 第三方接入系统请求智慧停车开放平台取消车位预订 |
| `cancelLeasePot` | callable | write | · | 第三方接入系统请求智慧停车开放平台进行取消租用 |
| `doLeasePot` | callable | write | · | 第三方接入系统请求智慧停车开放平台进行车位租用 |
| `shareOvertimePay` | callable | write | · | 第三方接入系统请求智慧停车开放平台进行共享超时支付 |

## thirdParkForBolian (9)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `downThirdParkBillSync` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，返回线上下发同步账单到车场后，车场响应结果数据 |
| `https://openapi-test.yidianting.com.cn/openydt/comet/ajax` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，获取线上下发车场的同步账单请求 |
| `downThirdParkPrepayOrder` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，返回线上下发预付款到车场后，车场响应结果数据 |
| `https://openapi-test.yidianting.com.cn/openydt/comet/ajax` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，获取线上下发车场的预付账单请求 |
| `downThirdParkQueryFee` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，返回线上查费请求到车场后，车场响应结果数据 |
| `https://openapi-test.yidianting.com.cn/openydt/comet/ajax` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，获取线上下发车场的查费请求 |
| `intrduction` | callable | write | · |  |
| `upThirdParkCarIn` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，上报车辆进场数据 |
| `upThirdParkCarOut` | callable | write | · | 第三方车场缴费系统请求智慧停车开放平台，上报车辆出场数据 |

## ticket (57)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `addMonthTicketAuthorizeVisitor` | callable | write | · | 第三方接入系统请求智慧停车开放平台添加月票访客授权 |
| `addOnlineMonthTicket` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台开通线上月票 |
| `addOnlineMonthTicketType` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台新增线上月票类型 |
| `addSpecialCarType` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台添加特殊车辆类型 |
| `applyMonthTicketFreeze` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台申请月票冻结 |
| `bookOrCancelReservation` | callable | write | · | 第三方接入系统请求智慧停车开放平台预约/取消预约月票 |
| `cancelMonthTicketAuthorizeVisitorByCarNo` | callable | write | · | 第三方接入系统请求智慧停车开放平台根据授权ID取消月票访客授权 |
| `cancelMonthTicketAuthorizeVisitorById` | callable | write | · | 第三方接入系统请求智慧停车开放平台根据授权ID取消月票访客授权 |
| `cancelOnlineMonthTicketByMonthTicketType` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台根据月票类型取消线上月票 |
| `cancelOnlineVipTicket` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台取消线上月票 |
| `checkMonthTicketAppointment` | callable | read | · | 第三方接入系统请求智慧停车开放平台查看月票是否预选购买 |
| `deductMonthTicketConfig` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台月票类型名额扣减 |
| `delMonthTicketCertifiRule` | callable | write | · | 第三方接入系统请求智慧停车开放平台月票管理-删除凭证 |
| `editOnlineVipTicket` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台修改线上月票订单信息 |
| `freezeMonthTicket` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台申请月票冻结 |
| `getCarOwnerAndVipType` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 查询车辆的车主及VIP |
| `getCertificateByInfo` | callable | read | · | 第三方接入系统请求智慧停车开放平台 查询用户月票凭证 |
| `GetMonthTicketAccountTransationRecord` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询月票账号交易记录 |
| `getMonthTicketAccountUseRecord` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询月票账号扣费记录 |
| `getMonthTicketAppointmentByCarCode` | callable | read | · | 第三方接入系统请求智慧停车开放平台通过车牌查看月票预约信息 |
| `getMonthTicketAppointmentDetail` | callable | read | · | 第三方接入系统请求智慧停车开放平台通过车牌查看月票预约信息详情 |
| `getMonthTicketAppointmentLineUp` | callable | read | · | 第三方接入系统请求智慧停车开放平台根据月票id获取月票预约信息 |
| `getMonthTicketAuthorizeVisitor` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询月票访客授权 |
| `getMonthTicketAuthorizeVisitorHis` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询月票访客授权 |
| `getMonthTicketBillDetail` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台通过车牌查看月票预约信息详情 |
| `getMonthTicketCertifiRuleList` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询月票凭证记录列表 |
| `getMonthTicketCertificateInfoList` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询用户月票凭证 |
| `getMonthTicketConfigDetail` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询线上月票类型详情 |
| `getMonthTicketConfigDetailList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询线上月票类型详情列表 |
| `getMonthTicketSellNum` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查看月票已购买数量 |
| `getOnlineMonthTicketByCarCard` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询车牌线上月票 |
| `getOnlineMonthTicketList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询线上月票记录 |
| `getOnlineMonthTicketPayment` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询线上月票支付信息 |
| `getOnlineVipTicket` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台查询线上月票 |
| `getParkAgreement` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取车场协议 |
| `getSpecialCarTypeList` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取特殊车辆类型列表 |
| `getVipByCarNo` | callable | read | ✓ | 获取车辆身份 |
| `getVipByCarNoAndTime` | callable | read | ✓ | 通过车牌和时间获取VIP信息 |
| `getWillExpireMonthTicketBill` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台 查询将要过期的月票 |
| `monthTicketCertifiRuleSave` | callable | write | · | 第三方接入系统请求智慧停车开放平台月票管理-新增/编辑月票凭证 |
| `monthTicketConfigEdit` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台修改线上月票类型 |
| `parkAgreementSave` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台同步电子券二维码扫码 |
| `allowBuyMonthlyTicket` | callable | write | · | 第三方接入系统请求智慧停车开放平台添加月票访客授权 |
| `querySupportMonthTicketParkList` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取支持月票的车场列表 |
| `renewOnlineVipTicket` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台 续费线上月票 |
| `saveOrUpdateMonthTicketCretifi` | callable | write | · | 第三方接入系统请求智慧停车开放平台更新或新增用户月票凭证 |
| `unFreezeMonthTicket` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台申请月票冻结 |
| `addCusVipType` | callable | write | · | 第三方接入系统请求智慧停车开放平台新增卡类型 |
| `addVipTicket` | callable | write | · | 第三方接入系统请求智慧停车开放平台开通月卡 |
| `editCusVipType` | callable | write | · | 第三方接入系统请求智慧停车开放平台修改月票类型信息 |
| `getMonthCard` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取某个车主或某个车辆月卡信息 |
| `getVipType` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取月卡类型信息 |
| `pauseMonthCard` | callable | write | · | 第三方接入系统请求智慧停车开放平台暂停某个月卡 |
| `recoverMonthCard` | callable | write | · | 第三方接入系统请求智慧停车开放平台恢复某个处于暂停状态的月卡 |
| `refundMonthCard` | callable | write | · | 第三方接入系统请求智慧停车开放平台退费某个月卡 |
| `renewMonthCard` | callable | write | · | 第三方接入系统请求智慧停车开放平台续费某个月卡 |
| `setVipChargeRule` | callable | write | · | 第三方接入系统请求智慧停车开放平台对指定月卡类型设置计费规则 |

## trade (16)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getCloudParkChargeInfoMap` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取云停车场收费规则信息 |
| `setChargeRule` | callable | write | · | 第三方接入系统通过智慧停车开放平台接口下发停车场计费规则 |
| `commonGetParkFee` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台按时间获取停车费用接口，一般用于未来未知时间段内停车费用计算 |
| `getChargeRule` | callable | read | · | 第三方接入系统请求智慧停车开放平台查询计费规则接口 |
| `getParkFee` | callable | read | ✓ | 第三方接入系统请求智慧停车开放平台获取停车费用接口（请在查费后10分钟内完成缴费操作） |
| `payParkFee` | callable | write | ✓ | 第三方接入系统同步用户缴费信息到智慧停车开放平台 |
| `paybackBatch` | callable | write | ✓ | 第三方接入系统同步用户缴费信息到智慧停车开放平台 |
| `reducePrestore` | webhook | write | · | 智慧停车开放平台主动请求，通知第三方接入系统对预存款车辆进行扣费，第三方务必在10s内响应，超时我方 |
| `refundParkFee` | webhook | write | · | 智慧停车开放平台主动请求，在订单发生退款时通知第三方接入系统对该订单进行退款，第三方务必在10s内响 |
| `setPoints` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台向某个车辆预置车辆运营积分，用于后续车场端车辆自动抵扣运营积分(平 |
| `setPrestore` | callable | write | · |  |
| `setPrestoreFlow` | callable | write | · |  |
| `setPrestoreForCPark` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台向某个车辆预置预存款，用于后续车辆自动扣费 |
| `setPrestoreForCParkFirstPayBeforeLeave` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台向某个车辆预置预存款，用于后续车辆自动扣费 |
| `setPrestoreForFirstPayThenLeave` | callable | write | · | 第三方接入系统请求智慧停车开放平台下发某个车辆预存款，用于后续车辆自动扣费 |
| `syncAutoCoupon` | callable | write | · | 第三方接入系统请求智慧停车开放平台下发某个车辆的停车费计算系数 |

## upward (43)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `alert` | webhook | write | · | 智慧停车开放平台主动上报告警信息 |
| `asynSuccess` | callable | write | · | 第三方接入系统请求智慧停车开放平台异步响应数据上报结果 |
| `getDataReception` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取授权商数据接收情况 |
| `reportAbnormalCar` | webhook | write | · | 智慧停车开放平台主动上报停车场异常开闸数据到第三方接入系统 |
| `reportAbnormalCarIn` | webhook | write | · | 智慧停车开放平台主动上报停车场异常进场信息到第三方接入系统 |
| `reportAccountLogin` | webhook | write | · | 智慧停车开放平台主动上报账户登录信息 |
| `reportBlacklist` | webhook | write | · | 智慧停车开放平台主动上报黑名单变更信息到第三方接入系统 |
| `reportBlacklistData` | webhook | write | · | 智慧停车开放平台主动上报车辆黑名单信息到第三方接入系统 |
| `reportCar` | webhook | write | · | 智慧停车开放平台主动上报停车场车辆信息到第三方接入系统 |
| `reportCarIn` | webhook | write | · | 智慧停车开放平台主动上报车辆进场信息到第三方接入系统 |
| `reportCarOut` | webhook | write | · | 智慧停车开放平台主动上报车辆离场信息到第三方接入系统 |
| `reportCarOwner` | webhook | write | · | 智慧停车开放平台主动上报停车场车主信息到第三方接入系统 |
| `reportCarport` | webhook | write | · | 智慧停车开放平台主动上报停车场车位信息到第三方接入系统 |
| `reportChannelChange` | webhook | write | · | 智慧停车开放平台主动上报车场通道变更信息到第三方接入系统 |
| `reportDailyParkFee` | webhook | write | · | 智慧停车开放平台主动上报日报表信息 |
| `reportEquipmentStatus` | webhook | write | · | 智慧停车开放平台主动上报停车场设备状态信息到第三方接入系统 |
| `reportModifyCar` | webhook | write | · | 智慧停车开放平台主动上报停车场车牌校正数据到第三方接入系统 |
| `reportMonthCar` | webhook | write | · | 智慧停车开放平台主动上报停车场开通和续费月卡车辆到第三方接入系统 |
| `reportMonthTicketCertificateChanges` | webhook | write | · | 智慧停车开放平台主动上报月票凭证变更信息到第三方接入系统 |
| `reportMonthTicketExpirationNotice` | webhook | write | · | 智慧停车开放平台主动上报月票到期续费提醒到第三方接入系统 |
| `reportMonthTicketReservationChanges` | webhook | write | · | 智慧停车开放平台主动上报月票售罄预约信息到第三方接入系统 |
| `reportMonthTicketSyncMessage` | webhook | write | · | 智慧停车开放平台主动上报月票变更同步信息上报到第三方接入系统 |
| `reportParkChange` | webhook | write | · | 智慧停车开放平台主动上报车场信息变更到第三方接入系统 |
| `reportParkHeartbeat` | webhook | write | · | 智慧停车开放上报车场心跳接口给第三方，用于监控车场系统运行状态，如连续没收到该接口上报，则认为该停车 |
| `reportParkInfo` | webhook | write | · | 智慧停车开放平台主动上报停车场基础信息到第三方接入系统 |
| `reportParkMonthTicket` | webhook | write | · | 智慧停车开放平台主动上报车场月票订单变更信息到第三方接入系统 |
| `reportParkMonthTicketConfig` | webhook | write | · | 智慧停车开放平台主动上报车场月票类型变更信息到第三方接入系统 |
| `reportParkRemainCarport` | webhook | write | · | 智慧停车开放平台主动上报车场实时剩余车位信息到第三方接入系统 |
| `reportParkSpecialCarTypeConfig` | webhook | write | · | 智慧停车开放平台主动上报车场特殊车辆类型信息到第三方接入系统 |
| `reportParkVisitor` | webhook | write | · | 智慧停车开放平台主动上报访客变更信息到第三方接入系统 |
| `reportParkingFee` | webhook | write | · | 智慧停车开放平台主动上报停车场临停交费信息到第三方接入系统 |
| `reportPreCarIn` | webhook | write | · | 智慧停车开放平台主动上报车辆预进场信息到第三方接入系统 |
| `reportPreCarOut` | webhook | write | · | 智慧停车开放平台主动上报车辆预离场信息到第三方接入系统 |
| `reportShift` | webhook | write | · | 智慧停车开放平台主动上报停车场交接班信息到第三方接入系统 |
| `reportTicketUseRecord` | webhook | write | · | 智慧停车开放平台主动上报停车场储值卡扣费流水到第三方接入系统 |
| `reportWhitelist` | webhook | write | · | 智慧停车开放平台主动上报车场白名单变更信息到第三方接入系统 |
| `requestFailData` | callable | write | · | 第三方接入系统请求智慧停车开放平台获取未接收到的上报记录 |
| `requestFailDataResend` | callable | write | · | 智慧停车开放平台主动上报类接口已实现上报数据未成功接收的重传机制。但因网络不通、授权商接收端始终掉线 |
| `upThirdParkCarBill` | callable | write | · | 第三方接入系统主动推送车辆费用账单信息到智慧停车开放平台 |
| `upThirdParkCarIn` | callable | write | · | 第三方接入系统主动推送车辆进场信息到智慧停车开放平台 |
| `upThirdParkCarOut` | callable | write | · | 第三方接入系统主动推送车辆离场信息到智慧停车开放平台 |
| `upThirdParkCarPic` | callable | write | · | 第三方接入系统主动推送车辆图片信息到智慧停车开放平台 |
| `upThirdParkCarPreOut` | callable | write | · | 第三方接入系统主动推送车辆预离场信息到智慧停车开放平台 |

## vip (22)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `addOnlineVipTicket` | callable | write | · | 第三方接入系统请求智慧停车开放平台开通线上vip月票 |
| `addOnlineVipType` | callable | write | · | 第三方接入系统请求智慧停车开放平台新增线上月票VIP类型 |
| `addVipCar` | callable | write | · | 第三方接入系统请求智慧停车开放平台下发某个车辆信息 |
| `addVipCustomer` | callable | write | · | 第三方接入系统请求智慧停车开放平台下发某个车主信息 |
| `addVipTicketSimple` | callable | write | · | 第三方接入系统请求智慧停车开放平台开通月卡（简约版）,只适合vip基础功能，不适合多位多车功能 |
| `addVisitorCar` | callable | write | · | 第三方接入系统请求智慧停车开放平台添加某个访客车辆 |
| `authMonthCard` | callable | write | · | 第三方接入系统请求智慧停车开放平台为某张月卡授权添加车牌，满足一位多车或多位多车功能 |
| `cancelVipTicketSimple` | callable | write | · | 第三方接入系统请求智慧停车开放平台取消月卡（简约版） |
| `cancelVisitorCar` | callable | write | · | 第三方接入系统请求智慧停车开放平台取消访客车辆 |
| `changeVipValue` | callable | write | · | 第三方接入系统请求智慧停车开放平台更改储值或次数VIP的金额/次数 |
| `delAuthMonthCard` | callable | write | · | 第三方接入系统请求智慧停车开放平台为某张月卡绑定的车牌删除授权 |
| `delChargeRuleAll` | callable | write | · | 第三方接入系统请求智慧停车开放平台清除计费规则接口 |
| `delMonthCardByCarNoAndCustomCode` | callable | write | · | 第三方接入系统请求智慧停车开放平台清除车牌月卡接口 |
| `delVipAll` | callable | write | · | 第三方接入系统请求智慧停车开放平台清除所有月卡规则接口 |
| `getMonthCardList` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取停车场所有月票信息接口 |
| `getMonthCardSum` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取车场月卡数详情接口 |
| `monthFeeClearDual` | callable | write | · | 第三方接入系统请求智慧停车开放平台清除月卡接口 |
| `monthFeeNew` | callable | write | · | 第三方接入系统请求智慧停车开放平台新增月卡（二）接口 |
| `monthFeeNewTri` | callable | write | · | 第三方接入系统请求智慧停车开放平台新增月卡（三）接口 |
| `monthFeeRuleNew` | callable | write | · | 第三方接入系统请求智慧停车开放平台月卡规则新增接口 |
| `pauseAllMonthCardByCarNo` | callable | write | · | 第三方接入系统请求智慧停车开放平台暂停车牌所有月卡接口 |
| `recoverAllMonthCardByCarNo` | callable | write | · | 第三方接入系统请求智慧停车开放平台恢复车牌所有月卡接口 |

## visitor (2)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `addVisitorCarNew` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台添加某个访客车辆 |
| `cancelVisitorCarNew` | callable | write | ✓ | 第三方接入系统请求智慧停车开放平台取消某个访客车辆 |

## yard (8)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getCarport` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取车位信息 |
| `getCustomParkcode` | callable | read | · | 第三方接入系统请求智慧停车开放平台根据自定义停车场编号获取单个停车场基本信息 |
| `getMenuData` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取VEMS系统菜单树信息 |
| `getParkAreaPresentCarNum` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取停车场区域在场车位数 |
| `getParkWarning` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取车场离线告警日志记录信息 |
| `getPicture` | callable | read | · | 第三方接入系统请求智慧停车开放平台获取停车信息的图片资源 |
| `getVEMSVersion` | callable | read | · | 第三方接入系统请求智慧停车开放平台根据自定义停车场编号获取vems版本信息 |
| `setMenu` | callable | write | · | 第三方接入系统请求智慧停车开放平台对VEMS系统菜单设置禁用/启用。 |

## ydtPay (5)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `notifyPayInfo` | webhook | write | · | 支付通知接口 |
| `notifyPayInfoWithoutPayOrder` | webhook | write | · | 智慧停车开放平台主动上报账单信息到第三方接入系统 |
| `payOrder` | callable | write | · | 支付下单接口 |
| `queryPayInfo` | callable | read | · | 查询支付信息接口 |
| `refundOrder` | callable | write | · | 退款接口 |

## ydtUser (33)

| cmd | 方向 | 读写 | 一等 | 说明 |
|---|---|---|---|---|
| `getParkMerchantDetail` | callable | read | · | 第三方接入系统请求智慧停车开放平台,授权商需要把有清算需求的停车场进件到清算平台时，可以通过本接口获 |
| `getParkMerchantInfoList` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取支付配置信息列表(仅一点停) |
| `getPayChannelList` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取支付配置信息列表(仅一点停) |
| `authenticateConfirm` | callable | write | · | 第三方接入系统请求智慧停车开放平台,确认认证 |
| `authenticateUser` | callable | write | · | 第三方接入系统请求智慧停车开放平台,认证用户 |
| `getAppIdUsers` | callable | read | · | 第三方接入系统请求智慧停车开放平台,查询公众号用户信息 |
| `getBillRecord` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取用户账单记录 |
| `getCars` | callable | read | · | 第三方接入系统请求智慧停车开放平台,查询车辆信息 |
| `getEnrollUserBindingCar` | callable | read | · | 第三方接入系统请求智慧停车开放平台,查询注册用户绑定车辆 |
| `getInvoiceRecord` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取用户可开发票订单 |
| `getInvoiceStatusDetail` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取账单开票状态详情 |
| `getParkMerchantInfo` | callable | read | · | 第三方接入系统请求智慧停车开放平台,查询车场微信支付到账商户号 |
| `getParkingPayCouponList` | callable | read | · | 第三方接入系统请求智慧停车开放平台,查询指定车场指定车牌支付可用优惠券列表 |
| `getUserBillRecordForInvoice` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取用户账单记录用于开票 |
| `getUserCouponRecord` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取用户优惠券信息列表 |
| `getUserOpenId` | callable | read | · | 根据用户标识表示获取用户基础信息（智慧停车发现模块所带标识） |
| `getUserParkRecord` | callable | read | · | 第三方接入系统请求智慧停车开放平台,获取用户历史停车记录 |
| `getUsers` | callable | read | · | 第三方接入系统请求智慧停车开放平台,查询注册用户信息 |
| `getWeixinAccessToken` | callable | read | · | 第三方系统主动请求开放平台，获取微信公众号access_token |
| `getYdtUrl` | callable | read | · | 第三方系统主动请求开放平台，获取相应的h5页面链接 |
| `getYdtUserBaseInfoByFlag` | callable | read | · | 根据用户标识表示获取用户基础信息（智慧停车发现模块所带标识） |
| `getYdtUserFlagByPhone` | callable | read | · | 根据用户手机号码获取用户标识（智慧停车发现模块所带标识） |
| `openInvoiceByPark` | callable | write | · | 第三方接入系统请求智慧停车开放平台,获取用户可开发票订单 |
| `registerUser` | callable | write | · | 第三方接入系统请求智慧停车开放平台注册用户 |
| `reportAuthUserAuthPic` | webhook | write | · | 智慧停车开放平台主动上报智慧停车用户认证图片到第三方接入系统 |
| `reportAuthUserCarInfo` | webhook | write | · | 智慧停车开放平台主动上报智慧停车用户绑定车牌及变更信息到第三方接入系统 |
| `reportAuthUserInfo` | webhook | write | · | 智慧停车开放平台主动上报智慧停车用户注册信息及身份信息变更到第三方接入系统 |
| `reportCertificationUser` | webhook | write | · | 智慧停车开放平台主动上报智慧停车用户实名认证图片到第三方接入系统 |
| `reportYdtParkFee` | webhook | write | · | 智慧停车开放平台主动上报通过智慧停车公众号缴费的账单信息到第三方接入系统 |
| `sendMsgToYdtUser` | callable | write | · | 通过此接口可发送消息给智慧停车用户(微信消息参考地址:https://mp.weixin.qq.co |
| `syncEnrollUserBindingCar` | callable | write | · | 第三方接入系统请求智慧停车开放平台,变更注册用户绑定车辆 |
| `syncUserCarPermission` | callable | write | · | 同步用户车辆权限 |
| `syncUserInfo` | callable | write | · | 第三方接入系统请求我方开放平台修改微信用户openId等信息 |

