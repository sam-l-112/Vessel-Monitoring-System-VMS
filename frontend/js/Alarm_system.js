const { createApp } = Vue;

createApp({
    data() {
        return {
            sidebarVisible: false,
            loading: false,
            error: null,
            alertTypes: [],
            severityFilter: '',
            searchQuery: '',
            currentPage: 1,
            pageSize: 10,
            recentActivities: [],
            apiBase: '',
            apikey: 'openclaw_vms_secret_key_2026',
            alertMessages: [],
            alertMessagesLoading: false,
            alertMessagesError: null,
            ncdrCountyCode: '',
            ncdrCapCode: '',
            ncdrCurrentPage: 1,
            ncdrPageSize: 10,
            expandedAlert: null
        };
    },
    computed: {
        criticalCount() {
            return this.alertTypes.filter(a => a.severity === 'critical').length;
        },
        highCount() {
            return this.alertTypes.filter(a => a.severity === 'high').length;
        },
        mediumCount() {
            return this.alertTypes.filter(a => a.severity === 'medium').length;
        },
        lowCount() {
            return this.alertTypes.filter(a => a.severity === 'low').length;
        },
        filteredTypes() {
            return this.alertTypes.filter(a => {
                if (this.severityFilter && a.severity !== this.severityFilter) return false;
                if (this.searchQuery) {
                    const q = this.searchQuery.toLowerCase();
                    return a.name.toLowerCase().includes(q) || a.description.toLowerCase().includes(q);
                }
                return true;
            });
        },
        totalPages() {
            return Math.max(1, Math.ceil(this.filteredTypes.length / this.pageSize));
        },
        paginatedTypes() {
            const start = (this.currentPage - 1) * this.pageSize;
            return this.filteredTypes.slice(start, start + this.pageSize);
        },
        filteredMessages() {
            return this.alertMessages.filter(m => {
                if (this.ncdrCountyCode) {
                    const areas = this.getAlertAreas(m);
                    if (!areas.some(a => a.countyCode && a.countyCode.includes(this.ncdrCountyCode))) return false;
                }
                if (this.ncdrCapCode && m.capCode && !m.capCode.includes(this.ncdrCapCode)) return false;
                return true;
            });
        },
        ncdrTotalPages() {
            return Math.max(1, Math.ceil(this.filteredMessages.length / this.ncdrPageSize));
        },
        paginatedMessages() {
            const start = (this.ncdrCurrentPage - 1) * this.ncdrPageSize;
            return this.filteredMessages.slice(start, start + this.ncdrPageSize);
        }
    },
    watch: {
        severityFilter() { this.currentPage = 1; },
        searchQuery() { this.currentPage = 1; },
        pageSize() { this.currentPage = 1; },
        filteredTypes() { this.$nextTick(() => this.initChart()); },
        ncdrCountyCode() { this.ncdrCurrentPage = 1; },
        ncdrCapCode() { this.ncdrCurrentPage = 1; },
        ncdrPageSize() { this.ncdrCurrentPage = 1; }
    },
    methods: {
        toggleSidebar() {
            this.sidebarVisible = !this.sidebarVisible;
        },
        async loadAlertTypes() {
            this.loading = true;
            this.error = null;
            try {
                const response = await axios.get(`${this.apiBase}/api/dataset`, {
                    params: { apikey: this.apikey, format: 'json', limit: 100, offset: 0 },
                    timeout: 10000
                });
                if (response.data && response.data.success) {
                    this.alertTypes = response.data.data.alert_types || [];
                    this.generateRecentActivities();
                    this.$nextTick(() => this.initChart());
                } else {
                    this.error = response.data?.message || '無法載入示警資料';
                }
            } catch (err) {
                console.error('Load alert types error:', err);
                this.error = '無法連線到 API 伺服器，請確認伺服器正在運行';
            } finally {
                this.loading = false;
            }
        },
        async loadAlertMessages() {
            this.alertMessagesLoading = true;
            this.alertMessagesError = null;
            try {
                const response = await axios.get(`${this.apiBase}/api/datastore`, {
                    params: { apikey: this.apikey, format: 'json', limit: 50, offset: 0 },
                    timeout: 15000
                });
                if (response.data && response.data.success) {
                    const result = response.data.result || response.data;
                    const items = result.items || result.data || [];
                    this.alertMessages = items.map(item => {
                        const info = item.info && item.info[0] ? item.info[0] : {};
                        return {
                            id: item.capId || item.id || '',
                            capCode: item.capCode || '',
                            msgType: item.msgType || '',
                            sender: item.sender || '',
                            sent: item.sent || '',
                            category: info.category || '',
                            event: info.event || '',
                            urgency: info.urgency || '',
                            severity: info.severity || '',
                            certainty: info.certainty || '',
                            headline: info.headline || '',
                            description: info.description || '',
                            area: info.area || [],
                            web: info.web || '',
                            instruction: info.instruction || ''
                        };
                    });
                } else {
                    this.alertMessages = [];
                }
            } catch (err) {
                console.error('Load NCDR alert messages error:', err);
                this.alertMessagesError = '無法取得 NCDR 示警訊息';
                if (err.response && err.response.status === 401) {
                    this.alertMessagesError = 'NCDR API 金鑰未設定';
                }
            } finally {
                this.alertMessagesLoading = false;
            }
        },
        getAlertAreas(msg) {
            if (!msg.area) return [];
            let areas = [];
            msg.area.forEach(a => {
                if (a.areaDesc) {
                    areas.push({
                        countyCode: a.geocode ? a.geocode.slice(0, 5) : '',
                        areaDesc: a.areaDesc
                    });
                }
            });
            return areas;
        },
        getSeverityLabel(sev) {
            const labels = { Extreme: '極端', Severe: '嚴重', Moderate: '中度', Minor: '輕微', Unknown: '未知' };
            return labels[sev] || sev || '-';
        },
        getSeverityClass(sev) {
            const map = { Extreme: 'severity-critical', Severe: 'severity-high', Moderate: 'severity-medium', Minor: 'severity-low' };
            return map[sev] || 'severity-medium';
        },
        getUrgencyLabel(urg) {
            const labels = { Immediate: '立即', Expected: '預期', Future: '未來', Past: '過去', Unknown: '未知' };
            return labels[urg] || urg || '-';
        },
        formatSentTime(sent) {
            if (!sent) return '-';
            try {
                const d = new Date(sent);
                return d.toLocaleString('zh-TW', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
            } catch(e) { return sent; }
        },
        toggleExpand(msg) {
            this.expandedAlert = this.expandedAlert === msg.id ? null : msg.id;
        },
        generateRecentActivities() {
            this.recentActivities = this.alertTypes.map(a => ({
                id: a.id,
                name: a.name,
                severity: a.severity,
                time: `${a.severity === 'critical' ? '即時' : a.severity === 'high' ? '5分鐘前' : a.severity === 'medium' ? '30分鐘前' : '1小時前'}`
            })).slice(0, 5);
        },
        initChart() {
            const ctx = document.getElementById('severityChart');
            if (!ctx || typeof Chart === 'undefined') return;
            if (window._severityChart) {
                try { window._severityChart.destroy(); } catch(e) {}
            }
            try {
                window._severityChart = new Chart(ctx, {
                    type: 'doughnut',
                    data: {
                        labels: ['嚴重 (critical)', '高度 (high)', '中度 (medium)', '低度 (low)'],
                        datasets: [{
                            data: [this.criticalCount, this.highCount, this.mediumCount, this.lowCount],
                            backgroundColor: ['#dc2626', '#f59e0b', '#3b82f6', '#10b981'],
                            borderWidth: 2,
                            borderColor: '#fff'
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: { legend: { position: 'bottom', labels: { font: { size: 12 } } } }
                    }
                });
            } catch(e) { console.warn('Chart init error:', e); }
        },
        refreshData() {
            this.loadAlertTypes();
            this.loadAlertMessages();
        },
        logout() {
            if (confirm('確定要登出嗎？')) {
                window.location.href = 'login.html';
            }
        }
    },
    mounted() {
        this.apiBase = window.location.origin;
        this.loadAlertTypes();
        this.loadAlertMessages();
    }
}).mount('#app');
