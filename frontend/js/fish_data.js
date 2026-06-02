const API_BASE = "";

const { createApp } = Vue;

const app = createApp({
    data() {
        return {
            sidebarVisible: false,
            fishData: [],
            loading: false,
            error: null,
            showFishModal: false,
            editingFish: null,
            savingFish: false,
            fishForm: {
                fish_type: '',
                quantity: 0,
                weight: 0,
                health_status: 'good'
            }
        };
    },
    mounted() {
        this.loadFishData();
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
        async loadFishData() {
            this.loading = true;
            this.error = null;
            try {
                const response = await this.makeAPIRequest('/api/fish/data', 'GET');
                this.fishData = response.data || [];
            } catch (error) {
                this.error = error.response?.data?.message || error.message || '載入魚類數據失敗';
                console.error('Load fish data error:', error);
            } finally {
                this.loading = false;
            }
        },
        showAddFishModal() {
            this.editingFish = null;
            this.fishForm = {
                fish_type: '',
                quantity: 0,
                weight: 0,
                health_status: 'good'
            };
            this.showFishModal = true;
        },
        editFish(fish) {
            this.editingFish = fish;
            this.fishForm = {
                fish_type: fish.fish_type,
                quantity: fish.quantity,
                weight: fish.weight || 0,
                health_status: fish.health_status || 'good'
            };
            this.showFishModal = true;
        },
        closeFishModal() {
            this.showFishModal = false;
            this.editingFish = null;
        },
        async saveFishData() {
            this.savingFish = true;
            try {
                if (this.editingFish) {
                    await this.makeAPIRequest(`/api/fish/data/${this.editingFish.id}`, 'PUT', this.fishForm);
                } else {
                    await this.makeAPIRequest('/api/fish/data', 'POST', this.fishForm);
                }
                this.closeFishModal();
                await this.loadFishData();
            } catch (error) {
                this.error = error.response?.data?.message || '儲存失敗';
                console.error('Save fish data error:', error);
            } finally {
                this.savingFish = false;
            }
        },
        async deleteFish(id) {
            if (!confirm('確定要刪除這筆魚類數據嗎？')) {
                return;
            }
            try {
                await this.makeAPIRequest(`/api/fish/data/${id}`, 'DELETE');
                await this.loadFishData();
            } catch (error) {
                this.error = error.response?.data?.message || '刪除失敗';
                console.error('Delete fish data error:', error);
            }
        },
        getHealthClass(status) {
            const classes = {
                'excellent': 'text-success',
                'good': 'text-primary',
                'fair': 'text-warning',
                'poor': 'text-danger'
            };
            return classes[status] || 'text-muted';
        },
        getHealthText(status) {
            const texts = {
                'excellent': '优良',
                'good': '良好',
                'fair': '一般',
                'poor': '需改善'
            };
            return texts[status] || status;
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