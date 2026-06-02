    /* ================================================================
       MOCK DATA — 測試用 JSON 物件 (正常狀態 vs 極端氣候風險狀態)
       ================================================================ */
    const MOCK_DATA = {
        normal: {
            healthIndex: 87,
            dissolvedOxygen: 7.2,
            phLevel: 7.6,
            waterTemp: 25.8,
            weather: { icon: 'fa-sun', iconLabel: '晴朗', rain: 15, wind: 12, pressure: 1013 }
        },
        alert: {
            healthIndex: 38,
            dissolvedOxygen: 3.1,
            phLevel: 4.5,
            waterTemp: 33.2,
            weather: { icon: 'fa-poo-storm', iconLabel: '暴風雨', rain: 95, wind: 52, pressure: 982 }
        }
    };

    /* ================================================================
       THRESHOLD CONFIGURATION
       ================================================================ */
    const THRESHOLDS = {
        health:   { min: 50,  label: '健康指數', unit: '%' },
        do:       { min: 5.0, label: '溶解氧',   unit: 'mg/L' },
        ph:       { min: 6.0, max: 9.0, label: 'pH值', unit: '' },
        temp:     { min: 15,  max: 32, label: '水溫',  unit: '°C' }
    };

    /* ================================================================
       Vue 3 Application
       ================================================================ */
    const { createApp } = Vue;

    const app = createApp({
        data() {
            return {
                /* --- original state --- */
                sidebarVisible: false,
                currentSection: 'dashboard',
                aiChatOpen: false,
                openclawOpen: false,
                openclawMessages: [{ role: 'assistant', content: '您好！我是 OpenCLAW 助手。我可以協助您進行養殖相關操作與查詢。' }],
                openclawInput: '',
                openclawLoading: false,
                voiceListening: false,
                voicePermissionDenied: false,
                showVoiceModal: false,
                selectedArea: 'newtaipei',
                areas: [
                    { value: 'newtaipei', label: '新北市' },
                    { value: 'penghu',   label: '澎湖縣' }
                ],
                aiMessages: [{ role: 'assistant', content: '您好！我是您的養殖監控系統 AI 助手。我可以幫助您分析數據、解答問題，或提供養殖建議。請問有什麼可以幫助您的嗎？' }],
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
                fishForm: { fish_type: '', quantity: null, weight: null, health_status: 'good' },
                showFeedModal: false,
                savingFeed: false,
                editingFeed: null,
                feedForm: { feed_type: '', quantity: null, unit: 'kg' },
                metrics: { totalFish: 0, healthIndex: 0, waterTemp: 0, alerts: 0 },
                recentActivities: [],
                dashboardData: { totalPonds: 0, activeAlerts: 0, todayFeed: 0, waterQuality: {} },
                apiBase: '',

                /* --- new state for glassmorphism dashboard --- */
                testMode: false,
                testType: 'normal',
                weatherSkeleton: true,
                alertCards: { health: false, do: false, ph: false, temp: false },
                dashboardMetrics: { healthIndex: 0, dissolvedOxygen: 0, phLevel: 0, waterTemp: 0 },
                weatherDisplay: { icon: 'fa-sun', iconLabel: '晴朗', rain: 15, wind: 12, pressure: 1013 },

                /* --- Chart.js instance references --- */
                _chartHealth: null,
                _chartDO: null,
                _chartPH: null,
                _chartTemp: null
            };
        },

        methods: {
            /* ==========================================================
               ORIGINAL API LOGIC (PRESERVED)
               ========================================================== */
            toggleSidebar() { this.sidebarVisible = !this.sidebarVisible; },

            showSection(section) {
                this.currentSection = section;
                this.sidebarVisible = false;
                if (section === 'weather')    { this.loadWeatherData(); this.loadForecastData(); }
                if (section === 'fish-data')  { this.loadFishData(); }
                if (section === 'feed')       { this.loadFeedData(); }
            },

            toggleAIChat()   { console.log('[toggleAIChat] open:', !this.aiChatOpen); this.aiChatOpen = !this.aiChatOpen; },
            toggleOpenCLAW() { console.log('[toggleOpenCLAW] open:', !this.openclawOpen, '| sidebar element:', !!document.querySelector('.openclaw-chat-sidebar')); this.openclawOpen = !this.openclawOpen; },

            toggleVoiceInput() {
                if (this.voiceListening) { this.stopVoiceInput(); return; }
                const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
                if (!SpeechRecognition) {
                    this.voicePermissionDenied = true;
                    this._voiceError = '您的瀏覽器不支援語音辨識，請使用 Chrome。';
                    return;
                }
                this._voiceError = '';
                this.voicePermissionDenied = false;
                this.voiceListening = true;
                const recognition = new SpeechRecognition();
                recognition.lang = 'zh-TW';
                recognition.interimResults = false;
                recognition.maxAlternatives = 1;
                this._voiceRecognition = recognition;
                recognition.onresult = (event) => {
                    const transcript = event.results[0][0].transcript;
                    if (transcript) {
                        this.openclawInput = transcript.replace(/[。，、！？\.\,\!]/g, '').trim();
                    }
                };
                recognition.onerror = (event) => {
                    this.voiceListening = false;
                    this._voiceRecognition = null;
                    if (event.error === 'not-allowed' || event.error === 'permission-denied') {
                        this.showVoiceModal = true;
                    } else if (event.error === 'no-speech') {
                        this.voicePermissionDenied = true;
                        this._voiceError = '未偵測到語音，請確認麥克風可正常運作。';
                    } else if (event.error !== 'aborted') {
                        this.voicePermissionDenied = true;
                        this._voiceError = '語音辨識錯誤：' + event.error;
                    }
                };
                recognition.onend = () => {
                    this.voiceListening = false;
                    this._voiceRecognition = null;
                };
                recognition.start();
            },

            dismissVoiceHint() {
                this.voicePermissionDenied = false;
                this._voiceError = '';
            },

            openVoiceModal() {
                this.voiceListening = false;
                this.showVoiceModal = true;
            },

            closeVoiceModal() {
                this.showVoiceModal = false;
            },

            openVoicePermissionPage() {
                this.closeVoiceModal();
                window.open('./welcome.html', 'vms_voice', 'width=500,height=400');
            },

            stopVoiceInput() {
                if (this._voiceRecognition) {
                    this._voiceRecognition.abort();
                }
                this.voiceListening = false;
                this._voiceRecognition = null;
            },

            async sendAIMessage() {
                if (!this.aiInput.trim() || this.aiLoading) return;
                const userMessage = this.aiInput.trim();
                this.aiMessages.push({ role: 'user', content: userMessage });
                this.aiInput = '';
                this.aiLoading = true;
                try {
                    const response = await axios.post(`${this.apiBase}/api/opencli/gemini/chat`, { message: userMessage }, {
                        headers: { 'Content-Type': 'application/json' }, timeout: 90000
                    });
                    if (response.data && response.data.success) {
                        let replyText = '';
                        if (response.data.data && response.data.data.response) replyText = response.data.data.response;
                        else if (response.data.data && response.data.data.data && response.data.data.data.response) replyText = response.data.data.data.response;
                        else if (response.data.reply) replyText = response.data.reply;
                        if (replyText) {
                            this.aiMessages.push({ role: 'assistant', content: replyText });
                        } else {
                            this.aiMessages.push({ role: 'assistant', content: '收到回應，但內容格式有誤' });
                        }
                    } else {
                        const errorMsg = response.data?.message || response.data?.error || '未知錯誤';
                        this.aiMessages.push({ role: 'assistant', content: '抱歉，發生錯誤：' + errorMsg });
                    }
                } catch (error) {
                    console.error('AI Query Error:', error);
                    let errorMsg = '抱歉，無法連線到 AI 服務。請確認 API 伺服器正在運行。';
                    if (error.code === 'ECONNREFUSED') errorMsg = '無法連線到 API 伺服器。請確認伺服器正在運行。';
                    else if (error.response) errorMsg = '伺服器錯誤：' + (error.response.data?.message || error.response.status);
                    this.aiMessages.push({ role: 'assistant', content: errorMsg });
                }
                this.aiLoading = false;
            },

            async sendOpenCLAWMessage() {
                if (!this.openclawInput.trim() || this.openclawLoading) return;
                const userMessage = this.openclawInput.trim();
                this.openclawMessages.push({ role: 'user', content: userMessage });
                this.openclawInput = '';
                this.openclawLoading = true;
                try {
                    const response = await axios.post(`${this.apiBase}/api/openclaw/chat`, { message: userMessage }, {
                        headers: { 'Content-Type': 'application/json' }, timeout: 90000
                    });
                    if (response.data && response.data.success) {
                        const replyText = response.data.reply || response.data.message || '收到回應';
                        this.openclawMessages.push({ role: 'assistant', content: replyText });
                    } else {
                        const errorMsg = response.data?.message || response.data?.error || '未知錯誤';
                        this.openclawMessages.push({ role: 'assistant', content: '抱歉，發生錯誤：' + errorMsg });
                    }
                } catch (error) {
                    console.error('OpenCLAW Error:', error);
                    let errorMsg = '抱歉，無法連線到 OpenCLAW 服務。';
                    if (error.response) errorMsg = '伺服器錯誤：' + (error.response.data?.message || error.response.status);
                    else if (error.code === 'ECONNREFUSED') errorMsg = '無法連線到 API 伺服器。請確認伺服器正在運行。';
                    this.openclawMessages.push({ role: 'assistant', content: errorMsg });
                }
                this.openclawLoading = false;
            },

            logout() {
                if (confirm('確定要登出嗎？')) { window.location.href = 'login.html'; }
            },

            async makeAPIRequest(endpoint, options = {}) {
                try {
                    const response = await axios.get(`${this.apiBase}${endpoint}`, {
                        headers: { 'Content-Type': 'application/json' }, timeout: 15000, ...options
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
                    if (this.fishData.length > 0) {
                        const healthScores = { excellent: 100, good: 85, fair: 65, poor: 40 };
                        let totalScore = 0, totalWeight = 0;
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

                    /* --- Update glass dashboard metrics from real data --- */
                    this.dashboardMetrics.healthIndex = this.metrics.healthIndex;
                    this.dashboardMetrics.waterTemp = this.metrics.waterTemp;
                    if (this.weatherData.length > 0) {
                        const latest = this.weatherData[0];
                        this.dashboardMetrics.dissolvedOxygen = parseFloat(latest.dissolved_oxygen) || 0;
                        this.dashboardMetrics.phLevel = parseFloat(latest.ph_level) || 0;
                    }

                    /* --- Run threshold checks --- */
                    this.checkThreshold(this.dashboardMetrics.healthIndex, 'health');
                    this.checkThreshold(this.dashboardMetrics.dissolvedOxygen, 'do');
                    this.checkThreshold(this.dashboardMetrics.phLevel, 'ph');
                    this.checkThreshold(this.dashboardMetrics.waterTemp, 'temp');

                    /* --- Update weather placeholder --- */
                    this.weatherSkeleton = false;
                    if (this.weatherData.length > 0) {
                        const w = this.weatherData[0];
                        this.updateWeatherPlaceholder({
                            temperature: w.temperature,
                            humidity: w.humidity,
                            rain: w.humidity > 70 ? 65 : 15,
                            wind: 12,
                            pressure: 1013
                        });
                    }

                    this.$nextTick(() => { this.initKpiCharts(); });
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
                    if (response.success) { this.fishData = response.data || []; }
                } catch (error) {
                    this.error = '無法載入魚類數據';
                } finally { this.loading = false; }
            },

            showAddFishModal() {
                this.editingFish = null;
                this.fishForm = { fish_type: '', quantity: null, weight: null, health_status: 'good' };
                this.showFishModal = true;
            },
            editFish(fish) {
                this.editingFish = fish;
                this.fishForm = { id: fish.id, fish_type: fish.fish_type, quantity: fish.quantity, weight: fish.weight, health_status: fish.health_status };
                this.showFishModal = true;
            },
            closeFishModal() {
                this.showFishModal = false; this.editingFish = null;
                this.fishForm = { fish_type: '', quantity: null, weight: null, health_status: 'good' };
            },
            async saveFishData() {
                this.savingFish = true;
                try {
                    let response;
                    if (this.editingFish) {
                        response = await axios.put(`${this.apiBase}/api/fish/data`, this.fishForm, { headers: { 'Content-Type': 'application/json' } });
                    } else {
                        response = await axios.post(`${this.apiBase}/api/fish/data`, this.fishForm, { headers: { 'Content-Type': 'application/json' } });
                    }
                    if (response.data.success) { this.closeFishModal(); this.loadFishData(); alert('魚類數據已儲存！'); }
                    else { alert('儲存失敗：' + response.data.message); }
                } catch (error) {
                    alert('儲存失敗：' + (error.response?.data?.message || error.message));
                } finally { this.savingFish = false; }
            },

            showAddFeedModal() {
                this.editingFeed = null; this.feedForm = { feed_type: '', quantity: null, unit: 'kg' }; this.showFeedModal = true;
            },
            editFeed(feed) {
                this.editingFeed = feed;
                this.feedForm = { id: feed.id, feed_type: feed.feed_type, quantity: feed.quantity, unit: feed.unit || 'kg' };
                this.showFeedModal = true;
            },
            closeFeedModal() {
                this.showFeedModal = false; this.editingFeed = null;
                this.feedForm = { feed_type: '', quantity: null, unit: 'kg' };
            },
            async saveFeedData() {
                this.savingFeed = true;
                try {
                    let response;
                    if (this.editingFeed) {
                        response = await axios.put(`${this.apiBase}/api/feed/data`, this.feedForm, { headers: { 'Content-Type': 'application/json' } });
                    } else {
                        response = await axios.post(`${this.apiBase}/api/feed/data`, this.feedForm, { headers: { 'Content-Type': 'application/json' } });
                    }
                    if (response.data.success) { this.closeFeedModal(); this.loadFeedData(); alert('飼料數據已儲存！'); }
                    else { alert('儲存失敗：' + response.data.message); }
                } catch (error) {
                    alert('儲存失敗：' + (error.response?.data?.message || error.message));
                } finally { this.savingFeed = false; }
            },

            async loadWeatherData() {
                this.loading = true;
                try {
                    let areaParam = this.selectedArea !== 'newtaipei' ? '?area=' + this.selectedArea : '';
                    let response = await this.makeAPIRequest('/api/weather/data/cwa' + areaParam);
                    if (response.success && response.data && response.data.length > 0) {
                        this.weatherData = response.data;
                        if (response.data[0].temperature) { this.metrics.waterTemp = response.data[0].temperature; }
                    } else {
                        response = await this.makeAPIRequest('/api/weather/data');
                        if (response.success) {
                            this.weatherData = response.data || [];
                            if (response.data && response.data[0]) { this.metrics.waterTemp = response.data[0].temperature || 0; }
                        }
                    }
                } catch (error) {
                    console.error('Weather data load error:', error);
                } finally { this.loading = false; }
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
                } finally { this.loading = false; }
            },

            parseForecastData(records) {
                if (!records || !records.Locations || !records.Locations[0]) return [];
                let loc = records.Locations[0];
                let location = loc.Location[0];
                if (!location) return [];
                let forecast = [], tempData = {};
                for (let element of location.WeatherElement) {
                    let elementName = element.ElementName;
                    for (let time of element.Time) {
                        let startTime = time.StartTime;
                        if (!tempData[startTime]) { tempData[startTime] = { time: startTime }; }
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
                        time: this.formatDate(d.time), temp: d.temp || '-',
                        maxTemp: d.maxTemp || '-', minTemp: d.minTemp || '-',
                        weather: d.weather || '-', rain: d.rain || '-'
                    });
                });
                return forecast;
            },

            changeArea() { this.loadWeatherData(); this.loadForecastData(); },

            async loadFeedData() {
                this.loading = true;
                try {
                    const response = await this.makeAPIRequest('/api/feed/data');
                    if (response.success) { this.feedData = response.data || []; }
                } catch (error) { this.error = '無法載入飼料數據'; }
                finally { this.loading = false; }
            },

            generateRecentActivities() {
                this.recentActivities = [];
                if (this.fishData.length > 0) {
                    this.recentActivities.push({ id: 1, title: '魚類數據更新', description: `系統記錄了 ${this.fishData.length} 種魚類數據`, time: '剛剛' });
                }
                if (this.weatherData.length > 0) {
                    const latest = this.weatherData[0];
                    this.recentActivities.push({ id: 2, title: '天氣監控更新', description: `溫度: ${latest.temperature || 'N/A'}°C, 濕度: ${latest.humidity || 'N/A'}%`, time: '5分鐘前' });
                }
                if (this.recentActivities.length === 0) {
                    this.recentActivities = [
                        { id: 1, title: '系統初始化', description: 'VMS系統已成功啟動', time: '剛剛' },
                        { id: 2, title: 'API 連接', description: '嘗試連接中央氣象署 API', time: '5分鐘前' }
                    ];
                }
            },

            refreshData() {
                if (this.testMode) {
                    const mock = MOCK_DATA[this.testType];
                    this.updateDashboard(mock);
                } else {
                    this.loadDashboardData();
                }
            },

            getHealthClass(status) {
                const classes = { excellent: 'badge-success', good: 'badge-success', fair: 'badge-warning', poor: 'badge-danger' };
                return `badge ${classes[status] || 'badge-secondary'}`;
            },
            getHealthText(status) {
                const texts = { excellent: '優良', good: '良好', fair: '一般', poor: '需改善' };
                return texts[status] || status;
            },
            formatDate(dateString) {
                if (!dateString) return 'N/A';
                const date = new Date(dateString);
                return date.toLocaleString('zh-TW', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
            },

            /* ==========================================================
               NEW METHODS — 玻璃擬態儀表板專用
               ========================================================== */

            /**
             * updateDashboard(data)
             * 統一的數據更新接口，接收 JSON 物件直接更新所有監控卡片與圖表。
             * @param {Object} data - { healthIndex, dissolvedOxygen, phLevel, waterTemp, weather? }
             */
            updateDashboard(data) {
                if (!data) return;
                console.log('[updateDashboard] 接收數據:', JSON.stringify(data, null, 2));

                /* Update reactive metrics */
                this.dashboardMetrics.healthIndex    = data.healthIndex    ?? this.dashboardMetrics.healthIndex;
                this.dashboardMetrics.dissolvedOxygen = data.dissolvedOxygen ?? this.dashboardMetrics.dissolvedOxygen;
                this.dashboardMetrics.phLevel        = data.phLevel        ?? this.dashboardMetrics.phLevel;
                this.dashboardMetrics.waterTemp       = data.waterTemp      ?? this.dashboardMetrics.waterTemp;

                /* Threshold checks */
                this.checkThreshold(this.dashboardMetrics.healthIndex, 'health');
                this.checkThreshold(this.dashboardMetrics.dissolvedOxygen, 'do');
                this.checkThreshold(this.dashboardMetrics.phLevel, 'ph');
                this.checkThreshold(this.dashboardMetrics.waterTemp, 'temp');

                /* Weather placeholder update */
                if (data.weather) {
                    this.weatherSkeleton = false;
                    this.updateWeatherPlaceholder(data.weather);
                }

                /* Chart updates */
                this.$nextTick(() => { this.initKpiCharts(); });
            },

            /**
             * checkThreshold(value, type)
             * 檢查數值是否超過安全閾值，自動為對應卡片新增/移除 alert-pulse CSS class。
             * @param {number} value - 當前數值
             * @param {string} type  - 'health' | 'do' | 'ph' | 'temp'
             * @returns {boolean} 是否觸發警報
             */
            checkThreshold(value, type) {
                if (value == null || isNaN(value)) return false;
                const t = THRESHOLDS[type];
                if (!t) { console.warn(`[checkThreshold] 未知類型: ${type}`); return false; }

                let isAlert = false;
                if (t.max !== undefined) {
                    isAlert = value < t.min || value > t.max;
                } else {
                    isAlert = value < t.min;
                }

                const cardKey = type === 'do' ? 'do' : type;
                this.alertCards[cardKey] = isAlert;

                if (isAlert) {
                    console.warn(`[ALERT] ${t.label} 異常：當前值 = ${value}${t.unit}，閾值 = ${t.min}${t.unit}${t.max ? ' ~ ' + t.max + t.unit : ''}`);
                } else {
                    console.log(`[OK] ${t.label} 正常：當前值 = ${value}${t.unit}`);
                }

                return isAlert;
            },

            /**
             * toggleTestMode()
             * 切換儀表板「正常狀態」與「極端氣候風險狀態」的視覺效果。
             */
            toggleTestMode() {
                this.testMode = !this.testMode;
                if (this.testMode) {
                    this.testType = 'alert';
                    const alertData = MOCK_DATA.alert;
                    console.log('[toggleTestMode] 切換至極端氣候風險狀態');
                    this.updateDashboard(alertData);
                } else {
                    this.testType = 'normal';
                    console.log('[toggleTestMode] 恢復正常狀態，重新載入 API 數據');
                    this.loadDashboardData();
                }
            },

            /**
             * testDashboardUpdate()
             * 模擬警報數據 — 在瀏覽器 Console 執行此函數，即可驗證卡片顏色、動畫與警告圖示。
             * 用法：在 Console 輸入 testDashboardUpdate()
             */
            testDashboardUpdate() {
                console.log('[testDashboardUpdate] 注入模擬警報數據...');
                const alertData = {
                    healthIndex: 32,
                    dissolvedOxygen: 2.8,
                    phLevel: 4.2,
                    waterTemp: 34.5,
                    weather: { icon: 'fa-poo-storm', iconLabel: '暴風雨', rain: 98, wind: 58, pressure: 975 }
                };
                this.updateDashboard(alertData);
                this.testMode = true;
                this.testType = 'alert';
                console.log('[testDashboardUpdate] 模擬警報已啟用。卡片應出現紅色邊框 + animate-pulse 效果。');
            },

            /* --- Weather placeholder update --- */
            updateWeatherPlaceholder(data) {
                if (!data) return;
                this.weatherSkeleton = false;
                let icon = 'fa-cloud-sun';
                if (data.rain > 70) icon = 'fa-cloud-rain';
                else if (data.rain > 40) icon = 'fa-cloud';
                else if (data.temperature > 32) icon = 'fa-sun';

                this.weatherDisplay = {
                    icon: data.icon || icon,
                    iconLabel: data.iconLabel || this.getWeatherLabel(data),
                    rain: data.rain ?? 0,
                    wind: data.wind ?? 0,
                    pressure: data.pressure ?? 1013
                };
            },

            getWeatherLabel(data) {
                if (data.rain > 70) return '暴雨';
                if (data.rain > 40) return '陰雨';
                if (data.temperature > 30) return '炎熱';
                return '晴朗';
            },

            /* --- KPI Doughnut Charts --- */
            initKpiCharts() {
                if (typeof Chart === 'undefined') return;
                this.createDoughnut('chartHealth', '健康指數', this.dashboardMetrics.healthIndex, 100, ['#10b981', '#e2e8f0']);
                this.createDoughnut('chartDO', '溶解氧', this.dashboardMetrics.dissolvedOxygen, 14, ['#0ea5e9', '#e2e8f0']);
                this.createDoughnut('chartPH', 'pH值', this.dashboardMetrics.phLevel, 14, ['#8b5cf6', '#e2e8f0']);
                this.createDoughnut('chartTemp', '水溫', this.dashboardMetrics.waterTemp, 40, ['#f59e0b', '#e2e8f0']);
            },

            createDoughnut(canvasId, label, value, max, colors) {
                const ctx = document.getElementById(canvasId);
                if (!ctx) return;
                const instanceKey = '_chart' + canvasId.replace('chart', '');

                /* Destroy existing chart if any */
                if (this[instanceKey] && typeof this[instanceKey].destroy === 'function') {
                    try { this[instanceKey].destroy(); } catch(e) { console.warn('Chart destroy failed:', e); }
                }

                const currentVal = Math.min(Math.max(parseFloat(value) || 0, 0), max);
                const remaining = Math.max(max - currentVal, 0);

                try {
                    this[instanceKey] = new Chart(ctx, {
                        type: 'doughnut',
                        data: {
                            labels: [label, '剩餘'],
                            datasets: [{
                                data: [currentVal, remaining],
                                backgroundColor: [colors[0], colors[1]],
                                borderColor: 'rgba(255,255,255,0.8)',
                                borderWidth: 3,
                                borderRadius: 6,
                                hoverBorderWidth: 4
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            cutout: '70%',
                            plugins: {
                                legend: { display: false },
                                tooltip: {
                                    callbacks: {
                                        label: function(ctx) {
                                            return ctx.label + ': ' + ctx.raw.toFixed(1);
                                        }
                                    }
                                }
                            }
                        }
                    });
                } catch(e) {
                    console.warn(`[Chart] ${canvasId} create failed:`, e);
                }
            }
        },

        mounted() {
            this.apiBase = window.location.origin;
            this.loadDashboardData();

            /* Expose test functions to global scope for Console access */
            window.updateDashboard      = (data) => this.updateDashboard(data);
            window.checkThreshold       = (value, type) => this.checkThreshold(value, type);
            window.toggleTestMode       = () => this.toggleTestMode();
            window.testDashboardUpdate  = () => this.testDashboardUpdate();
            window.MOCK_DATA            = MOCK_DATA;

            console.log('%c[VMS Dashboard] 初始化完成 %c| %c可用測試指令: %ctestDashboardUpdate() %c| %ctoggleTestMode()',
                'color:#10b981;font-weight:bold;', '',
                'color:#f59e0b;', 'color:#3b82f6;font-weight:bold;', '',
                'color:#f59e0b;', 'color:#3b82f6;font-weight:bold;');
            console.log('%c[Mock Data] %cMOCK_DATA.normal %c& %cMOCK_DATA.alert %c已就緒',
                'color:#10b981;', 'color:#94a3b8;', '', 'color:#ef4444;', '');
        }
    });

app.component('vms-sidebar', VmsSidebar);
app.mount('#app');
