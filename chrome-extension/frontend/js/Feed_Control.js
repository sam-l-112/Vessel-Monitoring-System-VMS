const API_BASE = "";

const { createApp } = Vue;

const app = createApp({
    data() {
        return {
            sidebarVisible: false,
            feedData: [],
            loading: false,
            error: null,
            showFeedModal: false,
            editingFeed: null,
            savingFeed: false,
            feedForm: {
                feed_type: '',
                quantity: 0,
                unit: 'kg'
            }
        };
    },
    mounted() {
        this.loadFeedData();
    },
    methods: {
        toggleSidebar() { this.sidebarVisible = !this.sidebarVisible; },
        async makeAPIRequest(url, method = 'GET', data = null) {
            try {
                const config = {
                    method: method,
                    url: `${API_BASE}${url}`,
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };
                if (data) {
                    config.data = data;
                }
                const response = await axios(config);
                return response.data;
            } catch (error) {
                console.error('API Error:', error);
                throw error;
            }
        },
        async loadFeedData() {
            this.loading = true;
            this.error = null;
            try {
                const response = await this.makeAPIRequest('/api/feed/data', 'GET');
                this.feedData = response.data || [];
            } catch (error) {
                this.error = error.response?.data?.message || error.message || '載入飼料數據失敗';
                console.error('Load feed data error:', error);
            } finally {
                this.loading = false;
            }
        },
        showAddFeedModal() {
            this.editingFeed = null;
            this.feedForm = {
                feed_type: '',
                quantity: 0,
                unit: 'kg'
            };
            this.showFeedModal = true;
        },
        editFeed(feed) {
            this.editingFeed = feed;
            this.feedForm = {
                feed_type: feed.feed_type,
                quantity: feed.quantity,
                unit: feed.unit || 'kg'
            };
            this.showFeedModal = true;
        },
        closeFeedModal() {
            this.showFeedModal = false;
            this.editingFeed = null;
        },
        async saveFeedData() {
            this.savingFeed = true;
            try {
                if (this.editingFeed) {
                    await this.makeAPIRequest(`/api/feed/data/${this.editingFeed.id}`, 'PUT', this.feedForm);
                } else {
                    await this.makeAPIRequest('/api/feed/data', 'POST', this.feedForm);
                }
                this.closeFeedModal();
                await this.loadFeedData();
            } catch (error) {
                this.error = error.response?.data?.message || '儲存失敗';
                console.error('Save feed data error:', error);
            } finally {
                this.savingFeed = false;
            }
        },
        async deleteFeed(id) {
            if (!confirm('確定要刪除這筆飼料記錄嗎？')) {
                return;
            }
            try {
                await this.makeAPIRequest(`/api/feed/data/${id}`, 'DELETE');
                await this.loadFeedData();
            } catch (error) {
                this.error = error.response?.data?.message || '刪除失敗';
                console.error('Delete feed data error:', error);
            }
        },
        formatDate(dateStr) {
            if (!dateStr) return 'N/A';
            const date = new Date(dateStr);
            return date.toLocaleString('zh-TW');
        },
        logout() {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = 'login.html';
        }
    }
});

app.component('vms-sidebar', VmsSidebar);
app.mount('#app');