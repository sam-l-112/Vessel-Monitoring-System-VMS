/**
 * VMS Sidebar Component (Vue 3 CDN 全域註冊形式)
 * 使用方式：每頁 JS 中引入後 app.component('vms-sidebar', VmsSidebar)
 * HTML: <vms-sidebar active-page="alarm"></vms-sidebar>
 * 品牌與登出按鈕已移至全域 .vms-header 頂欄
 */

const VmsSidebar = {
    name: 'VmsSidebar',
    props: {
        activePage: { type: String, default: 'dashboard' }
    },
    data() {
        return {
            isCollapsed: false
        };
    },
    methods: {
        toggleCollapse() {
            this.isCollapsed = !this.isCollapsed;
        }
    },
    template: `
        <aside class="vms-sidebar d-flex flex-column" :class="{ collapsed: isCollapsed }">
            <!-- 收合/展開按鈕 -->
            <div class="sidebar-toggle px-3 pt-3 pb-1 d-flex align-items-center"
                :class="isCollapsed ? 'flex-column gap-2' : 'justify-content-end'">
                <template v-if="isCollapsed">
                    <i class="fas fa-fish text-white-50 fs-6"></i>
                </template>
                <button class="btn btn-sm border-0" @click="toggleCollapse"
                    style="color:rgba(255,255,255,0.5);background:transparent"
                    :title="isCollapsed ? '展開側欄' : '收合側欄'">
                    <i :class="isCollapsed ? 'fas fa-angle-right' : 'fas fa-angle-left'"></i>
                </button>
            </div>

            <hr style="border-color:rgba(255,255,255,0.15)" class="mx-3 my-2">

            <!-- 導覽選單 -->
            <ul class="nav flex-column px-2 flex-grow-1">
                <li class="nav-item" style="padding:0.1rem 0.25rem">
                    <a class="nav-link rounded d-flex align-items-center gap-2" :class="{ active: activePage === 'dashboard' }" href="dashboard.html">
                        <i class="fas fa-chart-pie" style="width:20px;text-align:center;flex-shrink:0"></i>
                        <span v-show="!isCollapsed">儀表板</span>
                    </a>
                </li>
                <li class="nav-item" style="padding:0.1rem 0.25rem">
                    <a class="nav-link rounded d-flex align-items-center gap-2" :class="{ active: activePage === 'alarm' }" href="Alarm_system.html">
                        <i class="fas fa-bell" style="width:20px;text-align:center;flex-shrink:0"></i>
                        <span v-show="!isCollapsed">示警系統</span>
                    </a>
                </li>
                <li class="nav-item" style="padding:0.1rem 0.25rem">
                    <a class="nav-link rounded d-flex align-items-center gap-2" :class="{ active: activePage === 'fish' }" href="fish_data.html">
                        <i class="fas fa-fish" style="width:20px;text-align:center;flex-shrink:0"></i>
                        <span v-show="!isCollapsed">魚類數據</span>
                    </a>
                </li>
                <li class="nav-item" style="padding:0.1rem 0.25rem">
                    <a class="nav-link rounded d-flex align-items-center gap-2" :class="{ active: activePage === 'feed' }" href="Feed_Control.html">
                        <i class="fas fa-utensils" style="width:20px;text-align:center;flex-shrink:0"></i>
                        <span v-show="!isCollapsed">飼料管理</span>
                    </a>
                </li>
                <li class="nav-item" style="padding:0.1rem 0.25rem">
                    <a class="nav-link rounded d-flex align-items-center gap-2" :class="{ active: activePage === 'analytics' }" href="Data_Analytics.html">
                        <i class="fas fa-chart-line" style="width:20px;text-align:center;flex-shrink:0"></i>
                        <span v-show="!isCollapsed">數據分析</span>
                    </a>
                </li>
                <li class="nav-item" style="padding:0.1rem 0.25rem">
                    <a class="nav-link rounded d-flex align-items-center gap-2" :class="{ active: activePage === 'weather' }" href="CWA_WI.html">
                        <i class="fas fa-cloud-sun" style="width:20px;text-align:center;flex-shrink:0"></i>
                        <span v-show="!isCollapsed">天氣資訊</span>
                    </a>
                </li>
            </ul>
        </aside>
    `
};
