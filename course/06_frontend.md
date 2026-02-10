# Module 6: Frontend Architecture (Vue.js)

In this module, we move to `webui/`. We built a **Single Page Application (SPA)** using **Vue.js**.

## 6.1 Core Concepts
### Reactive State
Vue components have a `data()` function. When `this.messages` changes, the UI updates automatically.
```js
data() {
    return {
        messages: [],
        newMessage: ""
    }
}
```

### Component Lifecycle
- **`mounted()`**: This hook runs when the component is inserted into the DOM. We use it to trigger initial data fetching (`loadConversation`).

## 6.2 Key Components

### `App.vue` (The Root)
- Contains `<router-view>`.
- This is a placeholder. Depending on the URL (`/login` or `/conversations`), different components are rendered here.

### `LoginView.vue`
- simple form.
- On success: `localStorage.setItem('token', response.identifier)`.
- This token is used for all subsequent requests.

### `ChatView.vue` (The Heavy Lifter)
This component handles the entire messaging experience.
- **Short Polling**:
    - We use `setInterval(this.refresh, 5000)` to fetch new messages every 5 seconds.
    - *Why?* WebSockets are complex. Polling is simple and effective for this project.
- **Message Sending**:
    - Text messages are sent as JSON.
    - Photo messages use `FormData` object (Multipart upload).
- **Forwarding**:
    - Select a message -> Call `/forward` endpoint.

## 6.3 API Integration (`services/axios.js`)
We use the **Axios** library for HTTP requests.
### The Interceptor Pattern
Instead of manually adding the token to every request, we intercept all outgoing requests.
```js
instance.interceptors.request.use(config => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
});
```
- This ensures creating a new API call is as simple as `axios.post('/url')`. Authentication is handled globally.

## 6.4 CSS & Styling
We use standard CSS in `assets/main.css`.
- Why no Tailwind/Bootstrap? To keep dependencies minimal and demonstrate understanding of CSS fundamentals (Flexbox for layout).
