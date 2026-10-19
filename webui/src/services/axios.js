import axios from 'axios'

// The backend URL comes from __API_URL__ in vite.config.js (the constant used in the
// evaluation). If it still points to localhost but the page was opened from another
// device (e.g. http://192.168.1.10:4173), "localhost" would be that device itself, so
// the host serving the page is used instead, keeping the configured port.
function resolveBackendURL() {
    const url = new URL(__API_URL__)
    const pageHost = window.location.hostname
    const isLoopback = (h) => h === 'localhost' || h === '127.0.0.1' || h === '[::1]' || h.endsWith('.localhost')
    if (isLoopback(url.hostname) && !isLoopback(pageHost)) {
        url.hostname = pageHost
    }
    return url.origin
}

export const apiBaseURL = resolveBackendURL()

export function apiAssetURL(path) {
    return `${apiBaseURL}${path}`
}

const instance = axios.create({
    baseURL: apiBaseURL,
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

// Add auth token to requests
instance.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('wasatext_token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// Handle auth errors
instance.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response && error.response.status === 401) {
            localStorage.removeItem('wasatext_token')
            localStorage.removeItem('wasatext_user_id')
            window.location.href = '/'
        }
        return Promise.reject(error)
    }
)

export default instance
