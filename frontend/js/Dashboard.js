const { createApp } = Vue;

createApp({
    data() {
        return {
            sidebarVisible: false,
            currentSection: 'dashboard',
            aiChatOpen: false,
            aiMessages: [
                {
                    role: 'assistant',
                    content: '您好！我是您的養殖監控系統 AI 助手。我可以幫助您分析數據、解答問題，或提供養殖建議。請問有什麼可以幫助您的嗎？'
                }
            ],
            aiInput: '',
            aiLoading: false,
            loading: false,
            error: null,
            fishData: [],
            weatherData: [],
            feedData: [],
            metrics: {
                totalFish: 0,
                healthIndex: 0,
                waterTemp: 0,
                alerts: 0
            },
            recentActivities: [],
            dashboardData: {
                totalPonds: 0,
                activeAlerts: 0,
                todayFeed: 0,
                waterQuality: {}
            },
            apiBase: 'http://192.168.50.75:8080'
        };
    },
    methods: {
        toggleSidebar() {
            this.sidebarVisible = !this.sidebarVisible;
        },
        showSection(section) {
            this.currentSection = section;
            this.sidebarVisible = false;

            if (section === 'weather') {
                this.loadWeatherData();
            } else if (section === 'fish-data') {
                this.loadFishData();
            } else if (section === 'feed') {
                this.loadFeedData();
            }
        },
        toggleAIChat() {
            this.aiChatOpen = !this.aiChatOpen;
        },
        async sendAIMessage() {
            if (!this.aiInput.trim() || this.aiLoading) return;

            const userMessage = this.aiInput.trim();
            this.aiMessages.push({
                role: 'user',
                content: userMessage
            });
            this.aiInput = '';
            this.aiLoading = true;

            try {
                const response = await axios.post(`${this.apiBase}/api/ai/query`, {
                    query: userMessage
                }, {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: 90000
                });
                
                if (response.data.success) {
                    this.aiMessages.push({
                        role: 'assistant',
                        content: response.data.data.response
                    });
                } else {
                    this.aiMessages.push({
                        role: 'assistant',
                        content: '抱歉，發生錯誤：' + response.data.message
                    });
                }
            } catch (error) {
                console.error('AI Query Error:', error);
                this.aiMessages.push({
                    role: 'assistant',
                    content: '抱歉，無法連線到 AI 服務。請確認 API 伺服器正在運行。'
                });
            }
            
            this.aiLoading = false;
        },
        logout() {
            if (confirm('確定要登出嗎？')) {
                window.location.href = 'login.html';
            }
        },
        async makeAPIRequest(endpoint, options = {}) {
            try {
                const response = await axios.get(`${this.apiBase}${endpoint}`, {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: 15000,
                    ...options
                });
                return response.data;
            } catch (error) {
                console.error(`API Error (${endpoint}):`, error);
                throw error;
            }
        },
        async loadDashboardData() {
            this.loading = true;
            this.error = null;

            try {
                const fishResponse = await this.makeAPIRequest('/api/fish/data');
                if (fishResponse.success && fishResponse.data) {
                    this.fishData = fishResponse.data;
                    this.metrics.totalFish = fishResponse.data.reduce((sum, fish) => sum + fish.quantity, 0);
                }

                await this.loadWeatherData();

                this.metrics.healthIndex = Math.floor(Math.random() * 20) + 80;
                this.metrics.alerts = 0;

                this.generateRecentActivities();

                this.$nextTick(() => {
                    this.initCharts();
                });
            } catch (error) {
                console.error('Dashboard data load error:', error);
            } finally {
                this.loading = false;
            }
        },
        async loadFishData() {
            this.loading = true;
            try {
                const response = await this.makeAPIRequest('/api/fish/data');
                if (response.success) {
                    this.fishData = response.data || [];
                }
            } catch (error) {
                this.error = '無法載入魚類數據';
            } finally {
                this.loading = false;
            }
        },
        async loadWeatherData() {
            this.loading = true;
            try {
                let response = await this.makeAPIRequest('/api/weather/data/cwa');
                if (response.success && response.data && response.data.length > 0) {
                    this.weatherData = response.data;
                    if (response.data[0].temperature) {
                        this.metrics.waterTemp = response.data[0].temperature;
                    }
                } else {
                    response = await this.makeAPIRequest('/api/weather/data');
                    if (response.success) {
                        this.weatherData = response.data || [];
                        if (response.data && response.data[0]) {
                            this.metrics.waterTemp = response.data[0].temperature || 0;
                        }
                    }
                }
            } catch (error) {
                console.error('Weather data load error:', error);
            } finally {
                this.loading = false;
            }
        },
        async loadFeedData() {
            this.loading = true;
            try {
                const response = await this.makeAPIRequest('/api/feed/data');
                if (response.success) {
                    this.feedData = response.data || [];
                }
            } catch (error) {
                this.error = '無法載入飼料數據';
            } finally {
                this.loading = false;
            }
        },
        generateRecentActivities() {
            this.recentActivities = [];

            if (this.fishData.length > 0) {
                this.recentActivities.push({
                    id: 1,
                    title: '魚類數據更新',
                    description: `系統記錄了 ${this.fishData.length} 種魚類數據`,
                    time: '剛剛'
                });
            }

            if (this.weatherData.length > 0) {
                const latest = this.weatherData[0];
                this.recentActivities.push({
                    id: 2,
                    title: '天氣監控更新',
                    description: `溫度: ${latest.temperature || 'N/A'}°C, 濕度: ${latest.humidity || 'N/A'}%`,
                    time: '5分鐘前'
                });
            }

            if (this.recentActivities.length === 0) {
                this.recentActivities = [
                    { id: 1, title: '系統初始化', description: 'VMS系統已成功啟動', time: '剛剛' },
                    { id: 2, title: 'API 連接', description: '嘗試連接中央氣象署 API', time: '5分鐘前' }
                ];
            }
        },
        initCharts() {
            const growthCtx = document.getElementById('growthChart');
            if (growthCtx && this.fishData.length > 0) {
                if (window.growthChart) {
                    window.growthChart.destroy();
                }
                const growthData = this.fishData.slice(0, 6).map(fish => fish.weight || 1.0);
                window.growthChart = new Chart(growthCtx, {
                    type: 'line',
                    data: {
                        labels: ['1月', '2月', '3月', '4月', '5月', '6月'],
                        datasets: [{
                            label: '平均體重 (kg)',
                            data: growthData,
                            borderColor: 'rgb(37, 99, 235)',
                            backgroundColor: 'rgba(37, 99, 235, 0.1)',
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } },
                        scales: { y: { beginAtZero: true } }
                    }
                });
            }

            const waterCtx = document.getElementById('waterQualityChart');
            if (waterCtx && this.weatherData.length > 0) {
                if (window.waterQualityChart) {
                    window.waterQualityChart.destroy();
                }
                const excellent = this.weatherData.filter(w => w.ph_level >= 7.0 && w.ph_level <= 8.0).length;
                const good = this.weatherData.filter(w => w.ph_level >= 6.5 && w.ph_level < 7.0).length;
                const fair = this.weatherData.filter(w => w.ph_level >= 6.0 && w.ph_level < 6.5).length;
                const poor = this.weatherData.filter(w => w.ph_level < 6.0).length;
                const total = this.weatherData.length || 1;

                window.waterQualityChart = new Chart(waterCtx, {
                    type: 'doughnut',
                    data: {
                        labels: ['優良', '良好', '一般', '需改善'],
                        datasets: [{
                            data: [
                                Math.round((excellent / total) * 100),
                                Math.round((good / total) * 100),
                                Math.round((fair / total) * 100),
                                Math.round((poor / total) * 100)
                            ],
                            backgroundColor: ['rgb(16, 185, 129)', 'rgb(59, 130, 246)', 'rgb(245, 158, 11)', 'rgb(239, 68, 68)']
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: { legend: { position: 'bottom' } }
                    }
                });
            }
        },
        refreshData() {
            this.loadDashboardData();
        },
        getHealthClass(status) {
            const classes = { 'excellent': 'badge-success', 'good': 'badge-success', 'fair': 'badge-warning', 'poor': 'badge-danger' };
            return `badge ${classes[status] || 'badge-secondary'}`;
        },
        getHealthText(status) {
            const texts = { 'excellent': '優良', 'good': '良好', 'fair': '一般', 'poor': '需改善' };
            return texts[status] || status;
        },
        formatDate(dateString) {
            if (!dateString) return 'N/A';
            const date = new Date(dateString);
            return date.toLocaleString('zh-TW', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
        }
    },
    mounted() {
        this.loadDashboardData();
    }
}).mount('#app');