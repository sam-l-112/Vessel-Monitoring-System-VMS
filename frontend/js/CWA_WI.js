document.addEventListener('DOMContentLoaded', () => {
    const app = Vue.createApp({
        data() {
            return {
                selectedArea: 'penghu',
                weatherData: [],
                forecastRaw: '',
                loading: false,
                error: null,
                sidebarVisible: false,
                aiChatOpen: false,
                aiMessages: [
                    { role: 'assistant', content: '您好！我是 VMS AI 助手，有什麼關於天氣資訊或系統操作的問題嗎？' }
                ],
                aiInput: '',
                aiLoading: false
            };
        },
        methods: {
            switchArea(area) {
                if (this.selectedArea !== area) {
                    this.selectedArea = area;
                    this.loadData();
                }
            },
            toggleSidebar() {
                this.sidebarVisible = !this.sidebarVisible;
            },
            toggleAIChat() {
                this.aiChatOpen = !this.aiChatOpen;
            },
            async loadData() {
                this.loading = true;
                this.error = null;
                this.weatherData = [];
                this.forecastRaw = '';

                const areaParam = this.selectedArea === 'penghu' ? 'penghu' : 'newtaipei';

                try {
                    const weatherResponse = await axios.get(`/api/weather/data/cwa?area=${areaParam}`);
                    if (weatherResponse.data && weatherResponse.data.success) {
                        this.weatherData = weatherResponse.data.data || [];
                    } else {
                        this.error = weatherResponse.data?.message || '無法取得天氣資料。';
                    }
                } catch (err) {
                    this.error = err.response?.data?.message || err.message || '取得即時觀測資料失敗。';
                }

                try {
                    const forecastResponse = await axios.get(`/api/weather/forecast?area=${areaParam}`);
                    if (forecastResponse.data && forecastResponse.data.success) {
                        this.forecastRaw = JSON.stringify(forecastResponse.data.data, null, 2);
                    } else {
                        if (!this.error) {
                            this.error = forecastResponse.data?.message || '無法取得預報資料。';
                        }
                    }
                } catch (err) {
                    if (!this.error) {
                        this.error = err.response?.data?.message || err.message || '取得天氣預報失敗。';
                    }
                }

                this.loading = false;
            },
            async sendAIMessage() {
                if (!this.aiInput.trim() || this.aiLoading) return;

                const userMessage = this.aiInput.trim();
                this.aiMessages.push({ role: 'user', content: userMessage });
                this.aiInput = '';
                this.aiLoading = true;

                try {
                    // 模擬 AI 回應 - 實際應用中應該調用真實的 AI API
                    setTimeout(() => {
                        let response = '';
                        if (userMessage.includes('天氣') || userMessage.includes('氣象')) {
                            response = '關於天氣資訊，您可以查看即時觀測站資料和天氣預報。系統會自動從中央氣象局獲取最新資料。';
                        } else if (userMessage.includes('澎湖') || userMessage.includes('新北')) {
                            response = '您可以通過側邊欄切換不同地區的天氣資訊。目前支援澎湖地區和新北市。';
                        } else {
                            response = '我可以協助您了解天氣資訊、系統操作等問題。請問您需要什麼幫助？';
                        }
                        this.aiMessages.push({ role: 'assistant', content: response });
                        this.aiLoading = false;
                    }, 1000);
                } catch (error) {
                    this.aiMessages.push({ role: 'assistant', content: '抱歉，我現在無法回應您的問題。請稍後再試。' });
                    this.aiLoading = false;
                }
            },
            logout() {
                if (confirm('確定要登出嗎？')) {
                    window.location.href = 'login.html';
                }
            },
            formatDate(value) {
                if (!value) return '未知';
                const date = new Date(value);
                return date.toLocaleString('zh-TW', {
                    year: 'numeric',
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit',
                });
            },
        },
        mounted() {
            this.loadData();
        },
    });

    app.mount('#cwa-app');
});
