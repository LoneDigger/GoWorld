# GoWorld

---

![](https://i.imgur.com/C1eRdtF.png)

綠色表示玩家已訂閱區塊
黑色表示玩家未訂閱區塊
藍點表示玩家
每個區塊都有自己的獨立編號

---

![](https://i.imgur.com/DYtmhto.gif)

將地圖劃分數個區域
當有玩家在綠色區塊內移動，與區塊重疊玩家，會推送移動封包
當玩家離開綠色區塊時，伺服器會取消訂閱區塊並發送移除封包

---

###### Example:
玩家會自動訂閱附近區塊(本身站一塊，向外一塊)
當玩家移動，地區跨越由 83 走到 63 範圍內
會取消 102、103、104 區塊訂閱
並訂閱新區塊 42、43、44

[測試用客戶端](https://github.com/LoneDigger/GoPlayer)
[測試影片](https://imgur.com/3cRvlu3)
