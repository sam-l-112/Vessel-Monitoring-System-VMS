const { createApp } = Vue;

const app = createApp({
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
        criticalCount() { return (this.alertTypes||[]).filter(a => a.severity === 'critical').length; },
        highCount() { return (this.alertTypes||[]).filter(a => a.severity === 'high').length; },
        mediumCount() { return (this.alertTypes||[]).filter(a => a.severity === 'medium').length; },
        lowCount() { return (this.alertTypes||[]).filter(a => a.severity === 'low').length; },
        filteredTypes() {
            return (this.alertTypes || []).filter(a => {
                if (this.severityFilter && a.severity !== this.severityFilter) return false;
                if (this.searchQuery) {
                    const q = this.searchQuery.toLowerCase();
                    const name = (a.name || '').toLowerCase();
                    const desc = (a.description || '').toLowerCase();
                    return name.includes(q) || desc.includes(q);
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
                    // CWA records 可能在頂層 (records) 或在 result.records 內
                    const records = response.data.records || (response.data.result && response.data.result.records) || {};
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
            const keys = Object.keys(records || {});
            for (const key of keys) {
                const val = records[key];
                if (!Array.isArray(val)) {
                    if (typeof val === 'object' && val !== null) {
                        items.push({ _type: key, ...val });
                    }
                    continue;
                }
                val.forEach(item => {
                    if (type === 'station' && item.station && item.stationObsTimes) {
                        // 觀測資料: 展開每個 stationObsTime
                        const stn = item.station;
                        const times = (item.stationObsTimes.stationObsTime || []);
                        times.forEach(t => {
                            const we = t.weatherElements || {};
                            items.push({
                                _type: 'stationObs',
                                _headline: stn.StationName || stn.stationName || '-',
                                _stationID: stn.StationID || '',
                                _temperature: we.AirTemperature || '-',
                                _humidity: we.RelativeHumidity || '-',
                                _wind: we.WindSpeed || '-',
                                _pressure: we.AirPressure || '-',
                                _effective: t.obsTime?.DateTime || '',
                                _areaStr: stn.StationName || ''
                            });
                        });
                    } else if (key === 'info' || item.category || item.event) {
                        // 颱風/高溫/海嘯: items 自身就是 info 物件
                        const flat = { _type: key, ...item };
                        flat._areas = [];
                        if (item.area) {
                            item.area.forEach(a => {
                                flat._areas.push(a.CountyName || a.TownName || a.areaDesc || '');
                            });
                        }
                        flat._areaStr = flat._areas.filter(Boolean).join('、');
                        flat._headline = flat.headline || flat.event || '';
                        flat._severity = flat.severity || '';
                        flat._effective = flat.effective || flat.onset || '';
                        // 提取颱風專屬欄位（CWA API 可能用不同 key）
                        flat.typhoonName = flat.typhoonName || flat.TyphoonName || '';
                        flat.cwaTyphoonCategory = flat.cwaTyphoonCategory || flat.CwaTyphoonCategory || flat.TyphoonCategory || '';
                        flat.currentPosition = flat.currentPosition || flat.CurrentPosition || flat.position || '';
                        flat.maxWindSpeed = flat.maxWindSpeed || flat.MaxWindSpeed || flat.CenterMaxWindSpeed || '';
                        flat.gust = flat.gust || flat.Gust || flat.CenterGustWindSpeed || '';
                        flat.radius7 = flat.radius7 || flat.Radius7 || '';
                        flat.radius10 = flat.radius10 || flat.Radius10 || '';
                        flat.direction = flat.direction || flat.Direction || flat.movementDirection || '';
                        flat.speed = flat.speed || flat.Speed || flat.movementSpeed || '';
                        flat.pressure = flat.pressure || flat.Pressure || flat.CenterPressure || '';
                        items.push(flat);
                    } else if (item.info) {
                        // 備用: 部分 API info 包在 item.info 裡
                        const flat = { _type: key, ...item.info };
                        flat._areas = [];
                        if (item.info.area) {
                            item.info.area.forEach(a => {
                                flat._areas.push(a.CountyName || a.TownName || a.areaDesc || '');
                            });
                        }
                        flat._areaStr = flat._areas.filter(Boolean).join('、');
                        flat._headline = flat.headline || flat.event || '';
                        flat._severity = flat.severity || '';
                        flat._effective = flat.effective || flat.onset || '';
                        items.push(flat);
                    } else {
                        // Generic fallback
                        const flat = { _type: key };
                        Object.keys(item).forEach(k => {
                            if (typeof item[k] !== 'object' || k === 'info') flat[k] = item[k];
                        });
                        flat._headline = flat.headline || flat.event || '';
                        flat._severity = flat.severity || '';
                        flat._effective = flat.effective || flat.onset || '';
                        items.push(flat);
                    }
                });
            }
            return items;
        },
        switchCWATab(tab) { this.cwaTab = tab; this.loadCWAAlerts(); },
        cwaTabLabel(t) {
            const m = { typhoon:'颱風警報', hightemp:'高溫資訊', tsunami:'海嘯資訊', station:'觀測資料' };
            return m[t] || t;
        },
        generateRecentActivities() {
            this.recentActivities = (this.alertTypes || []).slice(0,5).map(a => ({
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
        this.apiBase = (typeof chrome !== 'undefined' && chrome.runtime && chrome.runtime.id) ? 'https://192.168.50.75' : window.location.origin;
        this.loadAlertTypes();
        this.loadAlertMessages();
        this.loadCWAAlerts();
    }
});

app.component('vms-sidebar', VmsSidebar);
app.mount('#app');
