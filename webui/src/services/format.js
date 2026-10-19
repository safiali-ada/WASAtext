// Formats the date and time of a message for the conversation list.
export function formatDateTime(timestamp) {
    if (!timestamp) return ''
    const date = new Date(timestamp)
    const time = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    const today = new Date()
    if (date.toDateString() === today.toDateString()) {
        return `Today ${time}`
    }
    return `${date.toLocaleDateString()} ${time}`
}

export function formatTime(timestamp) {
    if (!timestamp) return ''
    return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export function snippet(text, max = 60) {
    if (!text) return ''
    return text.length > max ? text.substring(0, max) + '…' : text
}

// Extracts a readable message from an axios error.
export function errorText(err, fallback) {
    const body = err?.response?.data
    if (typeof body === 'string' && body.trim()) {
        return body.trim()
    }
    return fallback
}
