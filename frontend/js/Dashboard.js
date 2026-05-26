const { createApp } = Vue;

createApp({
    data() {
        return {
            sidebarVisible: false,
            currentSection: 'dashboard',
            aiChatOpen: false,
            openclawOpen: false,
            openclawMessages: [
                {
                    role: 'assistant',
                    content: '您好！我是 OpenCLAW 助手。我可以協助您進行養殖相關操作與查詢。'
                }
            ],
            openclawInput: '',
            openclawLoading: false,
            selectedArea: 'newtaipei',
            areas: [
                { value: 'newtaipei', label: '新北市' },
                { value: 'penghu', label: '澎湖縣' }
            ],
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
            forecastData: [],
            feedData: [],
            showFishModal: false,
            savingFish: false,
            editingFish: null,
            fishForm: {
                fish_type: '',
                quantity: null,
                weight: null,
                health_status: 'good'
            },
            showFeedModal: false,
            savingFeed: false,
            editingFeed: null,
            feedForm: {
                feed_type: '',
                quantity: null,
                unit: 'kg'
            },
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
            apiBase: ''
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
                this.loadForecastData();
            } else if (section === 'fish-data') {
                this.loadFishData();
            } else if (section === 'feed') {
                this.loadFeedData();
            }
        },
        toggleAIChat() {
            this.aiChatOpen = !this.aiChatOpen;
        },
        toggleOpenCLAW() {
            this.openclawOpen = !this.openclawOpen;
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
                // 使用 OpenCLI API (Gemini)
                const response = await axios.post(`${this.apiBase}/api/opencli/gemini/chat`, {
                    message: userMessage
                }, {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: 90000
                });

                if (response.data && response.data.success) {
                    // Handle different response structures
                    let replyText = '';
                    if (response.data.data && response.data.data.response) {
                        replyText = response.data.data.response;
                    } else if (response.data.data && response.data.data && response.data.data.data && response.data.data.data.response) {
                        replyText = response.data.data.data.response;
                    } else if (response.data.reply) {
                        replyText = response.data.reply;
                    }
                    
                    if (replyText) {
                        this.aiMessages.push({
                            role: 'assistant',
                            content: replyText
                        });
                    } else {
                        this.aiMessages.push({
                            role: 'assistant',
                            content: '收到回應，但內容格式有誤'
                        });
                    }
                } else {
                    const errorMsg = response.data?.message || response.data?.error || '未知錯誤';
                    this.aiMessages.push({
                        role: 'assistant',
                        content: '抱歉，發生錯誤：' + errorMsg
                    });
                }
            } catch (error) {
                console.error('AI Query Error:', error);
                let errorMsg = '抱歉，無法連線到 AI 服務。請確認 API 伺服器正在運行。';
                if (error.code === 'ECONNREFUSED') {
                    errorMsg = '無法連線到 API 伺服器。請確認伺服器正在運行。';
                } else if (error.response) {
                    errorMsg = '伺服器錯誤：' + (error.response.data?.message || error.response.status);
                }
                this.aiMessages.push({
                    role: 'assistant',
                    content: errorMsg
                });
            }
            
            this.aiLoading = false;
        },
        async sendOpenCLAWMessage() {
            if (!this.openclawInput.trim() || this.openclawLoading) return;

            const userMessage = this.openclawInput.trim();
            this.openclawMessages.push({
                role: 'user',
                content: userMessage
            });
            this.openclawInput = '';
            this.openclawLoading = true;

            try {
                const response = await axios.post(`${this.apiBase}/api/openclaw/chat`, {
                    message: userMessage
                }, {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: 90000
                });

                if (response.data && response.data.success) {
                    const replyText = response.data.reply || response.data.message || '收到回應';
                    this.openclawMessages.push({
                        role: 'assistant',
                        content: replyText
                    });
                } else {
                    const errorMsg = response.data?.message || response.data?.error || '未知錯誤';
                    this.openclawMessages.push({
                        role: 'assistant',
                        content: '抱歉，發生錯誤：' + errorMsg
                    });
                }
            } catch (error) {
                console.error('OpenCLAW Error:', error);
                let errorMsg = '抱歉，無法連線到 OpenCLAW 服務。';
                if (error.response) {
                    errorMsg = '伺服器錯誤：' + (error.response.data?.message || error.response.status);
                } else if (error.code === 'ECONNREFUSED') {
                    errorMsg = '無法連線到 API 伺服器。請確認伺服器正在運行。';
                }
                this.openclawMessages.push({
                    role: 'assistant',
                    content: errorMsg
                });
            }

            this.openclawLoading = false;
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

                // 根据鱼群健康状态计算健康指数
                if (this.fishData.length > 0) {
                    const healthScores = { 'excellent': 100, 'good': 85, 'fair': 65, 'poor': 40 };
                    let totalScore = 0;
                    let totalWeight = 0;
                    this.fishData.forEach(fish => {
                        const score = healthScores[fish.health_status] || 50;
                        const weight = fish.weight || 1;
                        totalScore += score * weight;
                        totalWeight += weight;
                    });
                    this.metrics.healthIndex = totalWeight > 0 ? Math.floor(totalScore / totalWeight) : 0;
                } else {
                    this.metrics.healthIndex = 0;
                }
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
        showAddFishModal() {
            console.log('showAddFishModal called');
            this.editingFish = null;
            this.fishForm = {
                fish_type: '',
                quantity: null,
                weight: null,
                health_status: 'good'
            };
            this.showFishModal = true;
            console.log('showFishModal set to:', this.showFishModal);
        },
        editFish(fish) {
            this.editingFish = fish;
            this.fishForm = {
                id: fish.id,
                fish_type: fish.fish_type,
                quantity: fish.quantity,
                weight: fish.weight,
                health_status: fish.health_status
            };
            this.showFishModal = true;
        },
        closeFishModal() {
            this.showFishModal = false;
            this.editingFish = null;
            this.fishForm = {
                fish_type: '',
                quantity: null,
                weight: null,
                health_status: 'good'
            };
        },
        async saveFishData() {
            this.savingFish = true;
            try {
                let response;
                if (this.editingFish) {
                    // Update existing fish
                    response = await axios.put(`${this.apiBase}/api/fish/data`, this.fishForm, {
                        headers: { 'Content-Type': 'application/json' }
                    });
                } else {
                    // Create new fish
                    response = await axios.post(`${this.apiBase}/api/fish/data`, this.fishForm, {
                        headers: { 'Content-Type': 'application/json' }
                    });
                }
                if (response.data.success) {
                    this.closeFishModal();
                    this.loadFishData();
                    alert('魚類數據已儲存！');
                } else {
                    alert('儲存失敗：' + response.data.message);
                }
            } catch (error) {
                alert('儲存失敗：' + (error.response?.data?.message || error.message));
            } finally {
                this.savingFish = false;
            }
        },
        showAddFeedModal() {
            this.editingFeed = null;
            this.feedForm = {
                feed_type: '',
                quantity: null,
                unit: 'kg'
            };
            this.showFeedModal = true;
        },
        editFeed(feed) {
            this.editingFeed = feed;
            this.feedForm = {
                id: feed.id,
                feed_type: feed.feed_type,
                quantity: feed.quantity,
                unit: feed.unit || 'kg'
            };
            this.showFeedModal = true;
        },
        closeFeedModal() {
            this.showFeedModal = false;
            this.editingFeed = null;
            this.feedForm = {
                feed_type: '',
                quantity: null,
                unit: 'kg'
            };
        },
        async saveFeedData() {
            this.savingFeed = true;
            try {
                let response;
                if (this.editingFeed) {
                    response = await axios.put(`${this.apiBase}/api/feed/data`, this.feedForm, {
                        headers: { 'Content-Type': 'application/json' }
                    });
                } else {
                    response = await axios.post(`${this.apiBase}/api/feed/data`, this.feedForm, {
                        headers: { 'Content-Type': 'application/json' }
                    });
                }
                if (response.data.success) {
                    this.closeFeedModal();
                    this.loadFeedData();
                    alert('飼料數據已儲存！');
                } else {
                    alert('儲存失敗：' + response.data.message);
                }
            } catch (error) {
                alert('儲存失敗：' + (error.response?.data?.message || error.message));
            } finally {
                this.savingFeed = false;
            }
        },
        async loadWeatherData() {
            this.loading = true;
            try {
                let areaParam = this.selectedArea !== 'newtaipei' ? '?area=' + this.selectedArea : '';
                let response = await this.makeAPIRequest('/api/weather/data/cwa' + areaParam);
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
        async loadForecastData() {
            this.loading = true;
            try {
                let areaParam = this.selectedArea !== 'newtaipei' ? 'area=' + this.selectedArea + '&' : '';
                let response = await this.makeAPIRequest('/api/weather/forecast?' + areaParam + 'type=week');
                if (response.success) {
                    let rawData = response.data;
                    let parsed = typeof rawData === 'string' ? JSON.parse(rawData) : rawData;
                    let result = typeof parsed.result === 'string' ? JSON.parse(parsed.result) : parsed;
                    this.forecastData = this.parseForecastData(result.records);
                }
            } catch (error) {
                console.error('Forecast data load error:', error);
            } finally {
                this.loading = false;
            }
        },
        parseForecastData(records) {
            if (!records || !records.Locations || !records.Locations[0]) return [];
            let loc = records.Locations[0];
            let location = loc.Location[0];
            if (!location) return [];
            
            let forecast = [];
            let tempData = {};
            
            for (let element of location.WeatherElement) {
                let elementName = element.ElementName;
                for (let time of element.Time) {
                    let startTime = time.StartTime;
                    if (!tempData[startTime]) {
                        tempData[startTime] = { time: startTime };
                    }
                    if (elementName === '平均溫度' || elementName === 'Temperature') {
                        tempData[startTime].temp = time.ElementValue ? time.ElementValue[0].Temperature : time.Parameter?.ParameterName;
                    } else if (elementName === '最高溫度' || elementName === 'MaxTemperature') {
                        tempData[startTime].maxTemp = time.ElementValue ? time.ElementValue[0].Temperature : time.Parameter?.ParameterName;
                    } else if (elementName === '最低溫度' || elementName === 'MinTemperature') {
                        tempData[startTime].minTemp = time.ElementValue ? time.ElementValue[0].Temperature : time.Parameter?.ParameterName;
                    } else if (elementName === '天氣現象' || elementName === 'Wx') {
                        tempData[startTime].weather = time.Parameter?.ParameterName;
                    } else if (elementName === '降雨機率' || elementName === 'PoP') {
                        tempData[startTime].rain = time.Parameter?.ParameterName;
                    }
                }
            }
            
            Object.values(tempData).forEach(d => {
                forecast.push({
                    time: this.formatDate(d.time),
                    temp: d.temp || '-',
                    maxTemp: d.maxTemp || '-',
                    minTemp: d.minTemp || '-',
                    weather: d.weather || '-',
                    rain: d.rain || '-'
                });
            });
            
            return forecast;
        },
        changeArea() {
            this.loadWeatherData();
            this.loadForecastData();
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
            if (growthCtx && this.fishData.length > 0 && typeof Chart !== 'undefined') {
                // 使用前綴底線避免與 DOM ID 衝突
                if (window._growthChart && typeof window._growthChart.destroy === 'function') {
                    try {
                        window._growthChart.destroy();
                    } catch(e) {
                        console.warn('Chart destroy failed:', e);
                    }
                }
                // 使用鱼种名称作为标签，数量作为数据
                const fishLabels = this.fishData.map(fish => fish.fish_type || '未知');
                const fishQuantities = this.fishData.map(fish => fish.quantity || 0);
                try {
                    window._growthChart = new Chart(growthCtx, {
                        type: 'bar',
                        data: {
                            labels: fishLabels,
                            datasets: [{
                                label: '數量 (尾)',
                                data: fishQuantities,
                                backgroundColor: [
                                    'rgba(37, 99, 235, 0.7)',
                                    'rgba(16, 185, 129, 0.7)',
                                    'rgba(245, 158, 11, 0.7)',
                                    'rgba(239, 68, 68, 0.7)',
                                    'rgba(139, 92, 246, 0.7)',
                                    'rgba(236, 72, 153, 0.7)'
                                ],
                                borderColor: [
                                    'rgb(37, 99, 235)',
                                    'rgb(16, 185, 129)',
                                    'rgb(245, 158, 11)',
                                    'rgb(239, 68, 68)',
                                    'rgb(139, 92, 246)',
                                    'rgb(236, 72, 153)'
                                ],
                                borderWidth: 1
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: { legend: { display: false } },
                            scales: { y: { beginAtZero: true } }
                        }
                    });
                } catch(e) {
                    console.warn('Growth chart create failed:', e);
                }
            }

            const waterCtx = document.getElementById('waterQualityChart');
            if (waterCtx && this.weatherData.length > 0 && typeof Chart !== 'undefined') {
                // 使用前綴底線避免與 DOM ID 衝突
                if (window._waterQualityChart && typeof window._waterQualityChart.destroy === 'function') {
                    try {
                        window._waterQualityChart.destroy();
                    } catch(e) {
                        console.warn('Chart destroy failed:', e);
                    }
                }
                try {
                    const excellent = this.weatherData.filter(w => w.ph_level >= 7.0 && w.ph_level <= 8.0).length;
                    const good = this.weatherData.filter(w => w.ph_level >= 6.5 && w.ph_level < 7.0).length;
                    const fair = this.weatherData.filter(w => w.ph_level >= 6.0 && w.ph_level < 6.5).length;
                    const poor = this.weatherData.filter(w => w.ph_level < 6.0).length;
                    const total = this.weatherData.length || 1;

                    window._waterQualityChart = new Chart(waterCtx, {
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
                } catch(e) {
                    console.warn('Water chart create failed:', e);
                }
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
        this.apiBase = window.location.origin;
        this.loadDashboardData();
    }
}).mount('#app');