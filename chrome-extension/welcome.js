const btn = document.getElementById("grantBtn");
const status = document.getElementById("status");

btn.addEventListener("click", async () => {
  btn.disabled = true;
  btn.textContent = "請求中...";
  status.textContent = "";
  status.className = "status";

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    stream.getTracks().forEach((track) => track.stop());
    status.textContent = "✅ 權限已授予！請回到側邊欄使用 OpenCLAW 語音輸入。";
    status.className = "status success";
    btn.textContent = "已完成";
    setTimeout(() => window.close(), 2500);
  } catch (err) {
    btn.disabled = false;
    btn.textContent = "允許麥克風權限";
    if (err.name === "NotAllowedError" || err.name === "PermissionDeniedError") {
      status.textContent = "❌ 權限被拒絕。請在 Chrome 設定 → 隱私權與安全性 → 網站設定 → 麥克風，手動允許此擴充功能。";
    } else if (err.name === "NotFoundError" || err.name === "DevicesNotFoundError") {
      status.textContent = "❌ 未偵測到麥克風裝置，請確認已連接麥克風。";
    } else {
      status.textContent = "❌ 發生錯誤：" + err.message;
    }
    status.className = "status error";
  }
});
