import axios from 'axios'

export const apiBaseURL = import.meta.env.DEV ? '/api' : 'http://localhost:3000'

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
