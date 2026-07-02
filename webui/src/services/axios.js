import axios from 'axios'

// The production frontend and backend are built as separate images by the
// grader.  Derive the backend origin from the page being viewed so deployments
// reached through a hostname or IP address do not accidentally call the
// visitor's own localhost.
const backendURL = new URL(window.location.href)
backendURL.port = '3000'
backendURL.pathname = ''
backendURL.search = ''
backendURL.hash = ''

export const apiBaseURL = import.meta.env.DEV ? '/api' : backendURL.origin

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
