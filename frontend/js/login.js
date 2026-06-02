// VMS Login Page - Vue.js Application

const { createApp } = Vue;

const app = createApp({
    data() {
        return {
            credentials: {
                username: '',
                password: ''
            },
            rememberMe: false,
            showPassword: false,
            loading: false,
            errorMessage: '',
            successMessage: ''
        }
    },
    methods: {
        async handleLogin() {
            this.errorMessage = '';
            this.successMessage = '';

            if (!this.credentials.username.trim() || !this.credentials.password.trim()) {
                this.errorMessage = '請輸入用戶名稱和密碼';
                return;
            }

            this.loading = true;

            try {
                const response = await this.loginAPI(this.credentials);

                if (response.success) {
                    this.successMessage = '登入成功！正在跳轉...';

                    if (response.token) {
                        localStorage.setItem('vms_token', response.token);
                        if (this.rememberMe) {
                            localStorage.setItem('vms_remember', 'true');
                        }
                    }

                    if (response.user) {
                        sessionStorage.setItem('vms_user', JSON.stringify(response.user));
                    }

                    setTimeout(() => {
                        window.location.href = './dashboard.html';
                    }, 1500);

                } else {
                    this.errorMessage = response.message || '登入失敗，請檢查用戶名稱和密碼';
                }

            } catch (error) {
                console.error('Login error:', error);
                this.errorMessage = this.getErrorMessage(error);
            } finally {
                this.loading = false;
            }
        },

        async loginAPI(credentials) {
            const apiUrl = '/api/auth/login';

            try {
                const response = await axios.post(apiUrl, {
                    username: credentials.username,
                    password: credentials.password
                }, {
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    timeout: 10000
                });

                return {
                    success: true,
                    token: response.data.token,
                    user: response.data.user,
                    message: response.data.message
                };

            } catch (error) {
                if (error.response) {
                    const status = error.response.status;
                    const data = error.response.data;

                    switch (status) {
                        case 401:
                            return { success: false, message: '用戶名稱或密碼錯誤' };
                        case 403:
                            return { success: false, message: '帳號已被停權' };
                        case 429:
                            return { success: false, message: '登入嘗試次數過多，請稍後再試' };
                        case 500:
                            return { success: false, message: '伺服器內部錯誤，請稍後再試' };
                        default:
                            return { success: false, message: data.message || '登入失敗' };
                    }
                } else if (error.request) {
                    return { success: false, message: '網路連接失敗，請檢查網路連線' };
                } else {
                    return { success: false, message: '發生未知錯誤，請稍後再試' };
                }
            }
        },

        getErrorMessage(error) {
            if (error.code === 'NETWORK_ERROR') {
                return '網路連接失敗，請檢查網路連線';
            } else if (error.code === 'TIMEOUT') {
                return '請求超時，請稍後再試';
            } else {
                return '登入失敗，請稍後再試';
            }
        },

        checkExistingSession() {
            const token = localStorage.getItem('vms_token');
            const remember = localStorage.getItem('vms_remember');
            if (token && remember === 'true') {
                window.location.href = './dashboard.html';
            }
        },

        handleKeyPress(event) {
            if (event.key === 'Enter') {
                this.handleLogin();
            }
        }
    },

    mounted() {
        this.checkExistingSession();
        document.addEventListener('keypress', this.handleKeyPress);
        this.$nextTick(() => {
            const usernameField = document.getElementById('username');
            if (usernameField) usernameField.focus();
        });
    },

    beforeUnmount() {
        document.removeEventListener('keypress', this.handleKeyPress);
    }
});

app.mount('#loginApp');
