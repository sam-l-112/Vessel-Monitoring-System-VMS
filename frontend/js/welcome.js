const btn = document.getElementById("grantBtn");
const status = document.getElementById("status");
const msg = document.getElementById("msg");

btn.addEventListener("click", async () => {
  btn.disabled = true;
  btn.textContent = "請求中...";
  status.textContent = "";
  status.className = "status";

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    stream.getTracks().forEach((track) => track.stop());
    msg.innerHTML = "✅ 授權成功！<br>請關閉此頁面，回到儀表板使用 OpenCLAW 語音輸入。";
    btn.textContent = "已完成";
    btn.className = "btn done";
    status.textContent = "3 秒後自動關閉...";
    status.className = "status success";
    setTimeout(() => window.close(), 3000);
  } catch (err) {
    btn.disabled = false;
    btn.textContent = "允許麥克風權限";
    if (err.name === "NotAllowedError" || err.name === "PermissionDeniedError") {
      status.textContent = "❌ 權限被拒絕。請改用 Chrome Extension 側邊欄開啟儀表板，或透過 localhost 存取。";
    } else if (err.name === "NotFoundError" || err.name === "DevicesNotFoundError") {
      status.textContent = "❌ 未偵測到麥克風裝置，請確認已連接麥克風。";
    } else {
      status.textContent = "❌ 發生錯誤：" + err.message;
    }
    status.className = "status error";
  }
});
