// Vessel Monitoring System - Frontend Application
// Vue.js 3 Application with Chart.js integration

const { createApp } = Vue;

createApp({
    data() {
        return {
            currentSection: 'dashboard',
            sidebarVisible: false,
            metrics: {
                totalFish: 1250,
                healthIndex: 87,
                waterTemp: 24.5,
                alerts: 3
            },
            recentActivities: [
                {
                    id: 1,
                    title: '水溫異常警報',
                    description: '池塘A的水溫上升至28°C',
                    time: '5分鐘前'
                },
                {
                    id: 2,
                    title: '飼料補充完成',
                    description: '自動飼料機已補充500kg飼料',
                    time: '1小時前'
                },
                {
                    id: 3,
                    title: '魚類健康檢查',
                    description: 'AI分析顯示魚類健康狀況良好',
                    time: '2小時前'
                }
            ]
        }
    },
    methods: {
        showSection(section) {
            this.currentSection = section;
            this.sidebarVisible = false; // Close sidebar on mobile
        },
        toggleSidebar() {
            this.sidebarVisible = !this.sidebarVisible;
        },
        refreshData() {
            // Simulate data refresh
            this.metrics.healthIndex = Math.floor(Math.random() * 20) + 80;
            this.metrics.waterTemp = (Math.random() * 5 + 20).toFixed(1);
            this.metrics.alerts = Math.floor(Math.random() * 5);
            this.initCharts();
        },
        getSectionTitle(section) {
            const titles = {
                'fish-data': '魚類數據管理',
                'weather': '天氣資訊監控',
                'feed': '飼料管理系統',
                'analytics': '數據分析中心',
                'alerts': '警報通知中心',
                'settings': '系統設定'
            };
            return titles[section] || '未知頁面';
        },
        initCharts() {
            // Growth Chart
            const growthCtx = document.getElementById('growthChart');
            if (growthCtx) {
                // Clear existing chart if it exists
                if (window.growthChart) {
                    window.growthChart.destroy();
                }
                window.growthChart = new Chart(growthCtx, {
                    type: 'line',
                    data: {
                        labels: ['1月', '2月', '3月', '4月', '5月', '6月'],
                        datasets: [{
                            label: '平均體重 (kg)',
                            data: [0.8, 1.2, 1.8, 2.3, 2.8, 3.2],
                            borderColor: 'rgb(37, 99, 235)',
                            backgroundColor: 'rgba(37, 99, 235, 0.1)',
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                display: false
                            }
                        },
                        scales: {
                            y: {
                                beginAtZero: true
                            }
                        }
                    }
                });
            }

            // Water Quality Chart
            const waterCtx = document.getElementById('waterQualityChart');
            if (waterCtx) {
                // Clear existing chart if it exists
                if (window.waterQualityChart) {
                    window.waterQualityChart.destroy();
                }
                window.waterQualityChart = new Chart(waterCtx, {
                    type: 'doughnut',
                    data: {
                        labels: ['優良', '良好', '一般', '需改善'],
                        datasets: [{
                            data: [45, 30, 20, 5],
                            backgroundColor: [
                                'rgb(16, 185, 129)',
                                'rgb(59, 130, 246)',
                                'rgb(245, 158, 11)',
                                'rgb(239, 68, 68)'
                            ]
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                position: 'bottom'
                            }
                        }
                    }
                });
            }
        },
        logout() {
            if (confirm('確定要登出嗎？')) {
                // Implement logout logic
                alert('已登出');
            }
        }
    },
    mounted() {
        this.initCharts();
    }
}).mount('#app');

// Global error handler
window.addEventListener('error', function(e) {
    console.error('Global error:', e.error);
});

// Service Worker registration (for PWA support)
if ('serviceWorker' in navigator) {
    window.addEventListener('load', function() {
        // navigator.serviceWorker.register('/sw.js');
    });
}