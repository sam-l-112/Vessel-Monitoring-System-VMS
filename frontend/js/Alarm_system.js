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
            expandedAlert: null,
            cwaAlerts: [],
            cwaLoading: false,
            cwaError: null,
            cwaTab: 'typhoon',
            cwaDatasets: {
                typhoon: 'W-C0034-001',
                station: 'C-B0024-001',
                tsunami: 'E-A0014-001',
                hightemp: 'W-C0033-005'
            }
        };
    },
    computed: {
        criticalCount() { return this.alertTypes.filter(a => a.severity === 'critical').length; },
        highCount() { return this.alertTypes.filter(a => a.severity === 'high').length; },
        mediumCount() { return this.alertTypes.filter(a => a.severity === 'medium').length; },
        lowCount() { return this.alertTypes.filter(a => a.severity === 'low').length; },
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
        totalPages() { return Math.max(1, Math.ceil(this.filteredTypes.length / this.pageSize)); },
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
        ncdrTotalPages() { return Math.max(1, Math.ceil(this.filteredMessages.length / this.ncdrPageSize)); },
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
        toggleSidebar() { this.sidebarVisible = !this.sidebarVisible; },
        async loadAlertTypes() {
            this.loading = true; this.error = null;
            try {
                const response = await axios.get(`${this.apiBase}/api/dataset`, {
                    params: { apikey: this.apikey, format: 'json', limit: 100, offset: 0 }, timeout: 10000
                });
                if (response.data && response.data.success) {
                    this.alertTypes = response.data.data.alert_types || [];
                    this.generateRecentActivities();
                    this.$nextTick(() => this.initChart());
                } else {
                    this.error = response.data?.message || '無法載入示警資料';
                }
            } catch (err) { this.error = '無法連線到 API 伺服器'; } finally { this.loading = false; }
        },
        async loadAlertMessages() {
            this.alertMessagesLoading = true; this.alertMessagesError = null;
            try {
                const response = await axios.get(`${this.apiBase}/api/datastore`, {
                    params: { apikey: this.apikey, format: 'json', limit: 50, offset: 0 }, timeout: 15000
                });
                if (response.data && response.data.success) {
                    const result = response.data.result || response.data;
                    const items = result.items || [];
                    this.alertMessages = items.map(item => {
                        const info = item.info && item.info[0] ? item.info[0] : {};
                        return {
                            id: item.capId || item.id || '', capCode: item.capCode || '',
                            msgType: item.msgType || '', sender: item.sender || '', sent: item.sent || '',
                            category: info.category || '', event: info.event || '',
                            urgency: info.urgency || '', severity: info.severity || '',
                            certainty: info.certainty || '', headline: info.headline || '',
                            description: info.description || '', area: info.area || [],
                            web: info.web || '', instruction: info.instruction || ''
                        };
                    });
                } else { this.alertMessages = []; }
            } catch (err) {
                this.alertMessagesError = (err.response && err.response.data && err.response.data.message) || '無法連線到 NCDR 服務';
            } finally { this.alertMessagesLoading = false; }
        },
        getAlertAreas(msg) {
            if (!msg.area) return [];
            return msg.area.filter(a => a.areaDesc).map(a => ({ countyCode: a.geocode ? a.geocode.slice(0,5) : '', areaDesc: a.areaDesc }));
        },
        getSeverityLabel(sev) {
            const m = { Extreme:'極端', Severe:'嚴重', Moderate:'中度', Minor:'輕微', Unknown:'未知' };
            return m[sev] || sev || '-';
        },
        getSeverityClass(sev) {
            const m = { Extreme:'severity-critical', Severe:'severity-high', Moderate:'severity-medium', Minor:'severity-low' };
            return m[sev] || 'severity-medium';
        },
        getUrgencyLabel(u) { const m = { Immediate:'立即', Expected:'預期', Future:'未來', Past:'過去', Unknown:'未知' }; return m[u] || u || '-'; },
        formatSentTime(s) {
            if (!s) return '-';
            try { return new Date(s).toLocaleString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit', hour:'2-digit', minute:'2-digit' }); }
            catch(e) { return s; }
        },
        toggleExpand(msg) { this.expandedAlert = this.expandedAlert === msg.id ? null : msg.id; },
        async loadCWAAlerts() {
            this.cwaLoading = true; this.cwaError = null;
            const dataset = this.cwaDatasets[this.cwaTab];
            try {
                const response = await axios.get(`${this.apiBase}/api/cwa/datastore`, {
                    params: { apikey: this.apikey, dataset, limit: 30 }, timeout: 15000
                });
                if (response.data && (response.data.success === 'true' || response.data.success === true)) {
                    const records = (response.data.result && response.data.result.records) || {};
                    this.cwaAlerts = this.flattenCWARecords(this.cwaTab, records);
                } else {
                    this.cwaAlerts = [];
                }
            } catch (err) {
                this.cwaError = (err.response && err.response.data && err.response.data.message)
                    ? err.response.data.message : '無法取得 CWA 氣象警報';
                this.cwaAlerts = [];
            } finally { this.cwaLoading = false; }
        },
        flattenCWARecords(type, records) {
            let items = [];
            const keys = Object.keys(records);
            for (const key of keys) {
                const val = records[key];
                if (Array.isArray(val)) {
                    val.forEach(item => {
                        const flat = { _type: key };
                        if (item.info) Object.assign(flat, item.info);
                        Object.keys(item).forEach(k => { if (k !== 'info' && typeof item[k] !== 'object') flat[k] = item[k]; });
                        flat._areas = [];
                        if (item.info && item.info.area) {
                            item.info.area.forEach(a => {
                                flat._areas.push(a.CountyName || a.TownName || a.areaDesc || JSON.stringify(a).slice(0,30));
                            });
                        }
                        flat._areaStr = flat._areas.join('、');
                        flat._headline = flat.headline || item.headline || flat.event || '';
                        flat._severity = flat.severity || '';
                        flat._effective = flat.effective || '';
                        items.push(flat);
                    });
                } else if (typeof val === 'object' && val !== null) {
                    const flat = { _type: key };
                    Object.assign(flat, val);
                    items.push(flat);
                }
            }
            return items;
        },
        switchCWATab(tab) { this.cwaTab = tab; this.loadCWAAlerts(); },
        cwaTabLabel(t) {
            const m = { typhoon:'颱風警報', hightemp:'高溫資訊', tsunami:'海嘯資訊', station:'觀測資料' };
            return m[t] || t;
        },
        generateRecentActivities() {
            this.recentActivities = this.alertTypes.slice(0,5).map(a => ({
                id: a.id, name: a.name, severity: a.severity,
                time: a.severity==='critical'?'即時':a.severity==='high'?'5分鐘前':a.severity==='medium'?'30分鐘前':'1小時前'
            }));
        },
        initChart() {
            const ctx = document.getElementById('severityChart');
            if (!ctx || typeof Chart === 'undefined') return;
            if (window._severityChart) { try { window._severityChart.destroy(); } catch(e) {} }
            try {
                window._severityChart = new Chart(ctx, {
                    type:'doughnut',
                    data:{
                        labels:['嚴重','高度','中度','低度'],
                        datasets:[{ data:[this.criticalCount, this.highCount, this.mediumCount, this.lowCount],
                            backgroundColor:['#dc2626','#f59e0b','#3b82f6','#10b981'], borderWidth:2, borderColor:'#fff' }]
                    },
                    options:{ responsive:true, maintainAspectRatio:false,
                        plugins:{ legend:{ position:'bottom', labels:{ font:{ size:12 } } } } }
                });
            } catch(e) {}
        },
        refreshData() { this.loadAlertTypes(); this.loadAlertMessages(); this.loadCWAAlerts(); },
        logout() { if (confirm('確定要登出嗎？')) window.location.href = 'login.html'; }
    },
    mounted() {
        this.apiBase = window.location.origin;
        this.loadAlertTypes();
        this.loadAlertMessages();
        this.loadCWAAlerts();
    }
}).mount('#app');
