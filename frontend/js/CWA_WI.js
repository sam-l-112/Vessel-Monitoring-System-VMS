document.addEventListener('DOMContentLoaded', () => {
    const app = Vue.createApp({
        data() {
            return {
                selectedArea: 'penghu',
                areas: [
                    { value: 'newtaipei', label: '新北市' },
                    { value: 'penghu', label: '澎湖縣' }
                ],
                weatherData: [],
                forecastData: [],
                currentWeather: '', // 天氣現象
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
        computed: {
            backgroundColor() {
                const weather = this.currentWeather?.toLowerCase() || '';
                const temp = parseInt(this.weatherData[0]?.temperature) || 25;
                
                console.log('Weather:', weather, 'Temp:', temp);
                
                if (weather.includes('雨') || weather.includes('雷')) {
                    return '#64748b';
                } else if (weather.includes('晴') || weather.includes('陽')) {
                    return temp > 30 ? '#fde047' : '#fef9c3';
                } else if (weather.includes('陰')) {
                    return '#9ca3af';
                } else if (weather.includes('雲')) {
                    return '#f3f4f6';
                } else if (temp > 32) {
                    return '#fde047';
                } else if (temp < 15) {
                    return '#bfdbfe';
                }
                return '#ffffff';
            },
            accentColor() {
                const temp = parseInt(this.weatherData[0]?.temperature) || 25;
                if (temp > 30) return '#dc2626';
                if (temp < 15) return '#2563eb';
                return '#059669';
            }
        },
        methods: {
            changeArea() {
                this.loadData();
            },
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
                this.forecastData = [];
                this.currentWeather = '';

                const areaParam = this.selectedArea;

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
                    const forecastResponse = await axios.get(`/api/weather/forecast?area=${areaParam}&type=week`);
                    if (forecastResponse.data && forecastResponse.data.success) {
                        let rawData = forecastResponse.data.data;
                        let parsed = typeof rawData === 'string' ? JSON.parse(rawData) : rawData;
                        let result = typeof parsed.result === 'string' ? JSON.parse(parsed.result) : parsed;
                        this.forecastData = this.parseForecastData(result.records);
                        
                        // 設定當前天氣現象用於背景顏色
                        if (this.forecastData.length > 0) {
                            this.currentWeather = this.forecastData[0].weather || '';
                        }
                    }
                } catch (err) {
                    console.error('Forecast error:', err);
                }

                this.loading = false;
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
                        // 平均溫度
                        if (elementName === '平均溫度') {
                            tempData[startTime].temp = time.ElementValue ? time.ElementValue[0].Temperature : '-';
                        }
                        // 最高溫度
                        else if (elementName === '最高溫度') {
                            tempData[startTime].maxTemp = time.ElementValue ? time.ElementValue[0].MaxTemperature : '-';
                        }
                        // 最低溫度
                        else if (elementName === '最低溫度') {
                            tempData[startTime].minTemp = time.ElementValue ? time.ElementValue[0].MinTemperature : '-';
                        }
                        // 天氣現象 - 使用 ElementValue[0].Weather 欄位
                        else if (elementName === '天氣現象') {
                            let ev = time.ElementValue ? time.ElementValue[0] : {};
                            let weather = ev.Weather || ev.ParameterName || time.Parameter?.ParameterName || '-';
                            tempData[startTime].weather = weather;
                        }
                        // 降雨機率
                        else if (elementName === '12小時降雨機率') {
                            let val = time.ElementValue ? time.ElementValue[0].ProbabilityOfPrecipitation : null;
                            tempData[startTime].rain = (val !== null && val !== undefined) ? val : '0';
                        }
                        // 風速
                        else if (elementName === '風速') {
                            let ws = time.ElementValue ? time.ElementValue[0].WindSpeed : null;
                            if (ws && ws.startsWith('>=')) {
                                tempData[startTime].wind = ws.replace('>= ', '') + '+';
                            } else {
                                tempData[startTime].wind = ws || '-';
                            }
                        }
                        // 濕度
                        else if (elementName === '平均相對濕度') {
                            tempData[startTime].humidity = time.ElementValue ? time.ElementValue[0].RelativeHumidity : '-';
                        }
                    }
                }
                
                Object.values(tempData).forEach(d => {
                    forecast.push({
                        time: this.formatDateTime(d.time),
                        temp: d.temp || '-',
                        maxTemp: d.maxTemp || '-',
                        minTemp: d.minTemp || '-',
                        weather: d.weather || '-',
                        rain: d.rain || '0',
                        wind: d.wind || '-',
                        humidity: d.humidity || '-'
                    });
                });
                
                return forecast;
            },
            formatDateTime(dateStr) {
                if (!dateStr) return '-';
                const date = new Date(dateStr);
                const month = date.getMonth() + 1;
                const day = date.getDate();
                const hour = date.getHours();
                return `${month}/${day} ${hour}:00`;
            },
            async sendAIMessage() {
                if (!this.aiInput.trim() || this.aiLoading) return;

                const userMessage = this.aiInput.trim();
                this.aiMessages.push({ role: 'user', content: userMessage });
                this.aiInput = '';
                this.aiLoading = true;

                try {
                    const response = await axios.post('/api/ai/query', { 
                        query: userMessage 
                    }, {
                        timeout: 90000
                    });
                    
                    if (response.data && response.data.success) {
                        let replyText = '';
                        if (response.data.data && response.data.data.response) {
                            replyText = response.data.data.response;
                        } else if (response.data.reply) {
                            replyText = response.data.reply;
                        }
                        
                        this.aiMessages.push({
                            role: 'assistant',
                            content: replyText || '無法取得回應'
                        });
                    } else {
                        const errorMsg = response.data?.message || response.data?.error || '未知錯誤';
                        this.aiMessages.push({
                            role: 'assistant',
                            content: '抱歉：' + errorMsg
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
                });
            },
        },
        mounted() {
            this.loadData();
        },
    });

    app.mount('#cwa-app');
});