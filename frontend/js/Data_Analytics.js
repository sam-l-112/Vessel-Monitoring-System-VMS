/**
 * VMS — 數據分析頁面 (開發中)
 * 使用 VmsSidebar 元件
 */

const { createApp } = Vue;

const app = createApp({
    data() {
        return { sidebarVisible: false };
    },
    methods: {
        toggleSidebar() { this.sidebarVisible = !this.sidebarVisible; },
        logout() { if (confirm('確定要登出嗎？')) window.location.href = 'login.html'; }
    }
});

app.component('vms-sidebar', VmsSidebar);
app.mount('#app');
