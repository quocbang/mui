
# mui UI畫面使用的API

## MES登入

* <font color=yellow>使用者登入</font> <font color=green>[POST]</font>/user/login

## MES登出

* <font color=yellow>使用者登出</font> <font color=green>[POST]</font>/user/logout

## MES修改密碼

* <font color=yellow>修改密碼(限MES帳號)</font> <font color=orange>[PUT]</font>/user/change-password

## 材料更改

* <font color=yellow>取得條碼資訊</font> <font color=blue>[GET]</font>/pda/barcode/{ID}

* <font color=yellow>更新條碼資訊</font> <font color=orange>[PUT]</font>/pda/barcode/{ID}

* <font color=yellow>取得管制/放行可變更狀態清單</font> <font color=blue>[GET]</font>/pda/barcode/update-status-list/ID/{ID}

* <font color=yellow>取得條碼可展延條碼日數</font> <font color=blue>[GET]</font>/pda/barcode/extend-expired-date/ID/{ID}

* <font color=yellow>取得管制地點清單</font> <font color=blue>[GET]</font>/pda/barcode/control-area

* <font color=yellow>取得管制原因清單</font> <font color=blue>[GET]</font>/pda/barcode/reason-list

## 材料新增

* <font color=yellow>新增產品標示卡</font> <font color=green>[POST]</font>/resource/material/stock

## 庫存搬移

* <font color=yellow>取得條碼倉儲位置</font> <font color=blue>[GET]</font>/warehouse/resource/{ID}

* <font color=yellow>物料倉儲搬移</font> <font color=orange>[PUT]</font>/warehouse/resource/{ID}

## 排產派工

* <font color=yellow>取得產品群組規則清單</font> <font color=blue>[GET]</font>/product/groups/department-oid/{departmentOID}/product-type/{productType}

* <font color=yellow>取得部門可生產之產品類別(部門代號必填)</font> <font color=blue>[GET]</font>/product/active-product-types/department-oid/{departmentOID}

* <font color=yellow>查詢排產資料</font> <font color=blue>[GET]</font>/plans/department-oid/{departmentOID}/product-type/{productType}/date/{date}

* <font color=yellow>新增排產資料</font> <font color=green>[POST]</font>/plan

* <font color=yellow>取得產品代號清單</font> <font color=blue>[GET]</font>/product/active-products/product-type/{productType}

* <font color=yellow>取得指定產品之配合表清單</font> <font color=blue>[GET]</font>/product/active-recipes/product-id/{productID}

* <font color=yellow>新增工單</font> <font color=green>[POST]</font>/work-orders

* <font color=yellow>透過檔案建立工單, 現況僅支援excel</font> <font color=green>[POST]</font>/work-orders/upload/department/{department}

## 機台排程

* <font color=yellow>查詢指定部門之機台清單</font> <font color=blue>[GET]</font>/station-list/department-oid/{departmentOID}

* <font color=yellow>取得部門代號清單</font> <font color=blue>[GET]</font>/departments

* <font color=yellow>取得機台指定日期的工單清單</font> <font color=blue>[GET]</font>/schedulings/station/{station}/date/{date}

* <font color=yellow>更新工單排序</font> <font color=orange>[PUT]</font>/work-orders

* <font color=yellow>下載預列印材料標示卡</font> <font color=green>[POST]</font>/print/work-orders/{workOrderID}/pre-material-resource

* <font color=yellow>更新工單資訊</font> <font color=orange>[PUT]</font>/work-orders/{id}

## 配合配方表查詢

* <font color=yellow>取得部門可生產之產品類別(部門代號必填)</font> <font color=blue>[GET]</font>/product/active-product-types/department-oid/{departmentOID}

* <font color=yellow>取得可生產產品類別</font> <font color=blue>[GET]</font>/product/active-product-types

* <font color=yellow>取得產品代號清單</font> <font color=blue>[GET]</font>/product/active-products/product-type/{productType}

* <font color=yellow>途程資訊清單</font> <font color=blue>[GET]</font>/product/recipe-process/recipe-id/{recipeID}

## 材料掛載

* <font color=yellow>取得材料標示卡資訊</font> <font color=blue>[GET]</font>/resource/material/info/resource-id/{ID}

* <font color=yellow>取得指定機台工位之已綁定材料</font> <font color=blue>[GET]</font>/site/material/station/{station}/site-name/{siteName}/site-index/{siteIndex}

* <font color=yellow>綁定/清除/彈出工位材料(可操作多筆材料到一個工位上)</font> <font color=green>[POST]</font>/site/resources/bind/auto

## 載具維護

* <font color=yellow>下載載具標示卡</font> <font color=green>[POST]</font>/print/barcode/code39

* <font color=yellow>查詢載具清單</font>  <font color=blue>[GET]</font>/carrier/department-oid/{departmentOID}

* <font color=yellow>新增載具</font> <font color=green>[POST]</font>/carrier

* <font color=yellow>更新載具資料</font> <font color=orange>[PUT]</font>/carrier/{ID}

* <font color=yellow>刪除載具</font> <font color=red>[DELETE]</font>/carrier/{ID}

## 帳號管理

* <font color=yellow>取得角色清單</font> <font color=blue>[GET]</font>/account/role-list

* <font color=yellow>查詢可新增角色授權帳號清單</font> <font color=blue>[GET]</font>/account/unauthorized/department-oid/{departmentOID}

* <font color=yellow>查詢帳號權限清單</font> <font color=blue>[GET]</font>/account/authorized/department-oid/{departmentOID}

* <font color=yellow>新增帳號權限</font> <font color=green>[POST]</font>/account/authorization

* <font color=yellow>修改帳號角色</font> <font color=orange>[PUT]</font>/account/authorization/{employeeID}

* <font color=yellow>刪除帳號</font> <font color=red>[DELETE]</font>/account/authorization/{employeeID}

## 機台維護

* <font color=yellow>取得工位子類別清單</font> <font color=blue>[GET]</font>/site/sub-type-list

* <font color=yellow>取得工位類別清單</font> <font color=blue>[GET]</font>/site/type-list

* <font color=yellow>取得機台狀態清單</font> <font color=blue>[GET]</font>/station/state

* <font color=yellow>查詢指定部門之機台清單詳細資訊</font> <font color=blue>[GET]</font>/station/maintenance/department-oid/{departmentOID}

* <font color=yellow>新增機台資訊</font> <font color=green>[POST]</font>/station/maintenance

* <font color=yellow>更新機台資訊</font> <font color=#00E3E3>[PATCH]</font>/station/maintenance/{ID}

* <font color=yellow>刪除機台</font> <font color=red>[DELETE]</font>/station/maintenance/{ID}

## 物料查詢

* <font color=yellow>取得材料狀態清單</font> <font color=blue>[GET]</font>/resource/material/status

* <font color=yellow>取得可生產產品類別</font> <font color=blue>[GET]</font>/product/active-product-types

* <font color=yellow>取得產品代號清單</font> <font color=blue>[GET]</font>/product/active-products/product-type/{productType}

* <font color=yellow>取得物料產品資訊</font> <font color=blue>[GET]</font>/resource/material/info/product-type/{productType}

* <font color=yellow>取得材料標示卡資訊</font> <font color=blue>[GET]</font>/resource/material/info/resource-id/{ID}

* <font color=yellow>分批材料標示卡</font> <font color=green>[POST]</font>/resource/material/split

* <font color=yellow>下載材料標示卡</font> <font color=green>[POST]</font>/print/resource/material

## 生產達成率

* <font color=yellow>生產完成率清單</font> <font color=blue>[GET]</font>/work-orders-rate/department/{departmentID}

* <font color=yellow>取得部門代號清單</font> <font color=blue>[GET]</font>/departments

## PDA登入

* <font color=yellow>使用者登入</font> <font color=green>[POST]</font>/user/login

## PDA登出

* <font color=yellow>使用者登出</font> <font color=green>[POST]</font>/user/logout

* <font color=yellow>機台操作人員登出</font> <font color=green>[POST]</font>/stations/sign-out

## 機台設定

* <font color=yellow>取得部門可生產之產品類別(部門代號必填)</font> <font color=blue>[GET]</font>/product/active-product-types/department-oid/{departmentOID}

* <font color=yellow>設置PDA作業畫面欄位設定</font> <font color=green>[POST]</font>/production-flow/config/station/{stationID}

* <font color=yellow>取得PDA作業畫面欄位設定</font> <font color=blue>[GET]</font>/production-flow/config/station/{stationID}

* <font color=yellow>取得機台號清單</font> <font color=blue>[GET]</font>/production-flow/station

## PDA切換機台

* <font color=yellow>更改工單狀態</font> <font color=orange>[PUT]</font>/production-flow/status/work-order/{workOrderID}

* <font color=yellow>取得機台的工單清單</font> <font color=blue>[GET]</font>/production-flow/work-orders/station/{stationID}

* <font color=yellow>取得機台號清單</font> <font color=blue>[GET]</font>/production-flow/station

* <font color=yellow>查詢工單詳細資訊</font> <font color=blue>[GET]</font>/production-flow/work-order/{workOrderID}/information

* <font color=yellow>機台操作人員強制登入</font> <font color=green>[POST]</font>/station/{stationID}/sign-in

* <font color=yellow>取得指定機台工位之作業員資料</font> <font color=green>[POST]</font>/station/{stationID}/operator

* <font color=yellow>取得PDA作業畫面欄位設定</font> <font color=blue>[GET]</font>/production-flow/config/station/{stationID}

## PDA掛載

* <font color=yellow>取得材料標示卡資訊</font> <font color=blue>[GET]</font>/resource/material/info/resource-id/{ID}

* <font color=yellow>取得指定機台工位之已綁定材料</font> <font color=blue>[GET]</font>/site/material/station/{station}/site-name/{siteName}/site-index/{siteIndex}

* <font color=yellow>綁定/清除/彈出工位材料(可操作多筆材料到一個工位上)</font> <font color=green>[POST]</font>/site/resources/bind/auto

* <font color=yellow>工具條碼查詢</font> <font color=blue>[GET]</font>/production-flow/tool-resource/{toolResourceID}

* <font color=yellow>查詢工位資訊</font> <font color=green>[POST]</font>/production-flow/site/information

## PDA作業

* <font color=yellow>更改工單狀態</font> <font color=orange>[PUT]</font>/production-flow/status/work-order/{workOrderID}

* <font color=yellow>MES投料</font> <font color=green>[POST]</font>/mes/feed/station/{stationID}

* <font color=yellow>MES收料</font> <font color=green>[POST]</font>/mes/collect/station/{stationID}

* <font color=yellow>取得PDA作業畫面欄位設定</font> <font color=blue>[GET]</font>/production-flow/config/station/{stationID}

* <font color=yellow>查詢工單詳細資訊</font> <font color=blue>[GET]</font>/production-flow/work-order/{workOrderID}/information

* <font color=yellow>列印材料標示卡</font> <font color=green>[POST]</font>/production-flow/print/material-resource
