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
            feedData: [],
            dashboardData: {
                totalPonds: 0,
                activeAlerts: 0,
                todayFeed: 0,
                waterQuality: {}
            }
        };
    },
    methods: {
        toggleSidebar() {
            this.sidebarVisible = !this.sidebarVisible;
        },
        showSection(section) {
            this.currentSection = section;
            this.sidebarVisible = false;
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
                const apiBase = 'http://192.168.50.75:8080';
                const response = await axios.post(`${apiBase}/api/ai/query`, {
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
                    content: '抱歉，無法連線到 AI 服務。請確認 opencli 已正確安裝。'
                });
            }
            
            this.aiLoading = false;
        },
        logout() {
            if (confirm('確定要登出嗎？')) {
                window.location.href = 'login.html';
            }
        },
        async loadFeedData() {
            this.loading = true;
            this.error = null;
            setTimeout(() => {
                this.feedData = [
                    {
                        id: 1,
                        feed_type: '顆粒飼料',
                        quantity: 50,
                        unit: 'kg',
                        feed_time: new Date().toISOString(),
                        user: { username: '管理員' }
                    },
                    {
                        id: 2,
                        feed_type: '營養添加劑',
                        quantity: 10,
                        unit: 'kg',
                        feed_time: new Date(Date.now() - 3600000).toISOString(),
                        user: { username: '技術員' }
                    }
                ];
                this.loading = false;
            }, 500);
        },
        formatDate(dateString) {
            const date = new Date(dateString);
            return date.toLocaleString('zh-TW', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
        },
        loadDashboardData() {
            this.dashboardData = {
                totalPonds: 5,
                activeAlerts: 2,
                todayFeed: 150,
                waterQuality: {
                    temperature: 25.5,
                    ph: 7.2,
                    oxygen: 8.5
                }
            };
        }
    },
    mounted() {
        this.loadDashboardData();
        this.loadFeedData();
    }
}).mount('#app');

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
            feedData: [],
            // 其他數據可以根據需要添加
            dashboardData: {
                totalPonds: 0,
                activeAlerts: 0,
                todayFeed: 0,
                waterQuality: {}
            }
        };
    },
    methods: {
        toggleSidebar() {
            this.sidebarVisible = !this.sidebarVisible;
        },
        showSection(section) {
            this.currentSection = section;
            this.sidebarVisible = false; // 在移動設備上切換後隱藏側邊欄
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
                const apiBase = 'http://192.168.50.75:8080';
                const response = await axios.post(`${apiBase}/api/ai/query`, {
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
                // 實際項目中應該清除用戶會話並重定向到登錄頁面
                window.location.href = 'login.html';
            }
        },
        async loadFeedData() {
            this.loading = true;
            this.error = null;
            try {
                // 模擬從 API 加載飼料數據
                // 實際項目中應該調用真實的 API
                setTimeout(() => {
                    this.feedData = [
                        {
                            id: 1,
                            feed_type: '顆粒飼料',
                            quantity: 50,
                            unit: 'kg',
                            feed_time: new Date().toISOString(),
                            user: { username: '管理員' }
                        },
                        {
                            id: 2,
                            feed_type: '營養添加劑',
                            quantity: 10,
                            unit: 'kg',
                            feed_time: new Date(Date.now() - 3600000).toISOString(),
                            user: { username: '技術員' }
                        }
                    ];
                    this.loading = false;
                }, 500);
            } catch (error) {
                this.error = '加載飼料數據失敗';
                this.loading = false;
            }
        },
        formatDate(dateString) {
            const date = new Date(dateString);
            return date.toLocaleString('zh-TW', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
        },
        // 可以添加更多方法來處理其他功能
        loadDashboardData() {
            // 加載儀表板數據
            this.dashboardData = {
                totalPonds: 5,
                activeAlerts: 2,
                todayFeed: 150,
                waterQuality: {
                    temperature: 25.5,
                    ph: 7.2,
                    oxygen: 8.5
                }
            };
        }
    },
    mounted() {
        // 組件掛載時初始化數據
        this.loadDashboardData();
        this.loadFeedData();
    }
}).mount('#app');

