chrome.runtime.onInstalled.addListener(() => {
  chrome.tabs.create({
    url: chrome.runtime.getURL("welcome.html")
  });
});

chrome.sidePanel
  .setPanelBehavior({ openPanelOnActionClick: true })
  .catch((err) => console.error("sidePanel setup error:", err));
