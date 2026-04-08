// VMS Login Page - Vue.js Application
// Handles user authentication and API communication

const { createApp } = Vue;

createApp({
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
            // Clear previous messages
            this.errorMessage = '';
            this.successMessage = '';

            // Basic validation
            if (!this.credentials.username.trim() || !this.credentials.password.trim()) {
                this.errorMessage = '請輸入用戶名稱和密碼';
                return;
            }

            this.loading = true;

            try {
                // API call to Go backend
                const response = await this.loginAPI(this.credentials);

                if (response.success) {
                    this.successMessage = '登入成功！正在跳轉...';

                    // Store authentication token if provided
                    if (response.token) {
                        localStorage.setItem('vms_token', response.token);
                        if (this.rememberMe) {
                            localStorage.setItem('vms_remember', 'true');
                        }
                    }

                    // Store user info if provided
                    if (response.user) {
                        sessionStorage.setItem('vms_user', JSON.stringify(response.user));
                    }

                    // Redirect to dashboard after successful login
                    setTimeout(() => {
                        window.location.href = '../index.html';
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
            // API endpoint - adjust based on your Go backend
            const apiUrl = 'http://192.168.50.75/api/login'; // Adjust port as needed

            try {
                const response = await axios.post(apiUrl, {
                    username: credentials.username,
                    password: credentials.password
                }, {
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    timeout: 10000 // 10 second timeout
                });

                return {
                    success: true,
                    token: response.data.token,
                    user: response.data.user,
                    message: response.data.message
                };

            } catch (error) {
                if (error.response) {
                    // Server responded with error status
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
                    // Network error
                    return { success: false, message: '網路連接失敗，請檢查網路連線' };
                } else {
                    // Other error
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

        // Check if user is already logged in
        checkExistingSession() {
            const token = localStorage.getItem('vms_token');
            const remember = localStorage.getItem('vms_remember');

            if (token && remember === 'true') {
                // Auto redirect if remember me was checked
                window.location.href = '../index.html';
            }
        },

        // Handle Enter key press
        handleKeyPress(event) {
            if (event.key === 'Enter') {
                this.handleLogin();
            }
        }
    },

    mounted() {
        // Check for existing session on page load
        this.checkExistingSession();

        // Add keyboard event listener
        document.addEventListener('keypress', this.handleKeyPress);

        // Focus on username field
        this.$nextTick(() => {
            const usernameField = document.getElementById('username');
            if (usernameField) {
                usernameField.focus();
            }
        });
    },

    beforeUnmount() {
        // Clean up event listener
        document.removeEventListener('keypress', this.handleKeyPress);
    }
}).mount('#loginApp');

// Global error handler for login page
window.addEventListener('error', function(e) {
    console.error('Login page error:', e.error);
});

// Handle offline/online status
window.addEventListener('online', function() {
    console.log('Network connection restored');
});

window.addEventListener('offline', function() {
    console.log('Network connection lost');
    // Could show a notification to user
});