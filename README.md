# HTMX + Alpine.js + SSE Interactive Components Guide

A minimal, self-contained reference for building interactive web UIs with **HTMX**, **Alpine.js**, and **Server-Sent Events** using **Go** and **Templ**.

```bash
templ generate && go run .
# Open http://localhost:8080
```

---

## What's Inside

| Component | Description |
|-----------|-------------|
| **Modals** | HTML `<dialog>` elements with focus trapping, keyboard navigation, and HTMX form submission |
| **Drawers** | Sliding panels (left, right, bottom) using CSS transforms and Alpine.js transitions |
| **Toasts** | Server-pushed notifications via SSE with auto-dismiss and queue management |
| **Inline Swap** | Classic HTMX pattern with `hx-get` and `hx-swap="innerHTML"` |
| **SSE** | Dead simple Server-Sent Events for real-time updates |

---

## The Stack

```
┌───────────────────────────────────────────────────────┐
│  Browser                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐  │
│  │  HTMX       │  │  Alpine.js  │  │  DaisyUI/     │  │
│  │  ─────────  │  │  ─────────  │  │  Tailwind     │  │
│  │  hx-*       │  │  x-*        │  │  ───────────  │  │
│  │  SSE ext    │  │  State      │  │  UI + Styling │  │
│  └─────────────┘  └─────────────┘  └───────────────┘  │
└───────────────────────────────────────────────────────┘
                          │
                          │ HTTP + SSE
                          ▼
┌───────────────────────────────────────────────────────┐
│  Go Server                                            │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────┐  │
│  │  net/http   │  │  Templ      │  │  SSE          │  │
│  │  ─────────  │  │  ─────────  │  │  ───────────  │  │
│  │  Routing    │  │  Templates  │  │  text/event-  │  │
│  │  Handlers   │  │             │  │  stream       │  │
│  └─────────────┘  └─────────────┘  └───────────────┘  │
└───────────────────────────────────────────────────────┘
```

---

## Components Deep Dive

### Modals

We use the native HTML `<dialog>` element - no JavaScript library needed for the basic open/close behavior.

**Key patterns:**

```html
<!-- Trigger -->
<button onclick="my_modal.showModal()">Open Modal</button>

<!-- Modal -->
<dialog id="my_modal" class="modal">
  <div class="modal-box">
    <h3>Modal Title</h3>
    <p>Modal content here</p>

    <!-- Close button -->
    <form method="dialog">
      <button class="btn">Close</button>
    </form>
  </div>

  <!-- Click outside to close -->
  <form method="dialog" class="modal-backdrop">
    <button>close</button>
  </form>
</dialog>
```

**Alpine.js enhancements with focus restoration:**

When using `x-trap`, the native dialog's focus restoration breaks. Fix by tracking the trigger element:

```html
<!-- Trigger: pass element reference -->
<button onclick="my_modal.showModal()"
  x-data
  x-on:click="$nextTick(() => $dispatch('modal-open', { trigger: $el }))">
  Open Modal
</button>

<!-- Modal: store trigger and restore focus on close -->
<dialog id="my_modal" class="modal"
  x-data="{ open: false, trigger: null }"
  x-on:modal-open.window="open = true; trigger = $event.detail.trigger"
  x-on:close="open = false; $nextTick(() => trigger?.focus())"
  x-on:cancel="open = false; $nextTick(() => trigger?.focus())">

  <div class="modal-box"
    x-show="open"
    x-trap.noscroll="open"
    x-transition:enter="transition ease-out duration-200"
    x-transition:enter-start="opacity-0 scale-95"
    x-transition:enter-end="opacity-100 scale-100">
    <!-- x-trap keeps focus inside modal -->
    <!-- x-trap.noscroll prevents body scroll -->
  </div>
</dialog>
```

**With HTMX form:**

```html
<dialog id="delete_modal" class="modal">
  <div class="modal-box">
    <h3>Confirm Delete</h3>
    <form method="post"
          hx-post="/items/delete"
          hx-swap="none">
      <input type="hidden" name="id" value="123">
      <button type="submit" class="btn btn-error">
        <span class="loading loading-spinner htmx-indicator"></span>
        Delete
      </button>
    </form>
  </div>
</dialog>
```

---

### Drawers

Drawers are modals that slide in from the side. We use `modal-end` (right), `modal-start` (left), or `modal-bottom`.

**Right-sliding drawer:**

```html
<!-- Trigger with focus restoration -->
<button onclick="drawer.showModal()"
  x-data
  x-on:click="$nextTick(() => $dispatch('drawer-open', { trigger: $el }))">
  Open Drawer
</button>

<!-- Drawer -->
<dialog id="drawer" class="modal modal-end"
  x-data="{ open: false, trigger: null }"
  x-on:drawer-open.window="open = true; trigger = $event.detail.trigger"
  x-on:close="open = false; $nextTick(() => trigger?.focus())"
  x-on:cancel="open = false; $nextTick(() => trigger?.focus())">

  <div class="modal-box h-full max-h-full rounded-l-2xl rounded-r-none"
    x-show="open"
    x-trap.noscroll="open"
    x-transition:enter="transition ease-out duration-300"
    x-transition:enter-start="translate-x-full"
    x-transition:enter-end="translate-x-0"
    x-transition:leave="transition ease-in duration-200"
    x-transition:leave-start="translate-x-0"
    x-transition:leave-end="translate-x-full">

    <!-- Drawer content -->
    <h2>Settings</h2>
    <button @click="open = false; $el.closest('dialog').close()">Close</button>
  </div>

  <form method="dialog" class="modal-backdrop">
    <button @click="open = false">close</button>
  </form>
</dialog>
```

**Left navigation drawer:** Use `modal-start` and `-translate-x-full` for the enter animation.

**Bottom sheet:** Use `modal-bottom` and `translate-y-full` for mobile-friendly action sheets.

---

### Toasts

Server-pushed notifications using SSE. The Alpine.js component manages a queue of toasts with auto-dismiss.

**Toast container:**

```html
<div class="toast toast-end toast-bottom"
  x-data="{
    toasts: [],

    addToast(data) {
      const id = Date.now()
      this.toasts.push({ id, ...data, visible: true })
      setTimeout(() => this.removeToast(id), 6000)
    },

    removeToast(id) {
      const toast = this.toasts.find(t => t.id === id)
      if (toast) toast.visible = false
      setTimeout(() => {
        this.toasts = this.toasts.filter(t => t.id !== id)
      }, 300)
    },

    handleSSE(event) {
      if (event.detail.type === 'sse-toast') {
        event.preventDefault()
        this.addToast(JSON.parse(event.detail.data))
      }
    }
  }"
  x-on:htmx:sse-before-message.window="handleSSE($event)">

  <template x-for="toast in toasts" :key="toast.id">
    <div class="alert"
      x-show="toast.visible"
      x-transition:enter="transition ease-out duration-300"
      x-transition:enter-start="opacity-0 translate-y-4"
      x-transition:enter-end="opacity-100 translate-y-0">
      <span x-text="toast.message"></span>
    </div>
  </template>
</div>
```

**Triggering a toast from server:**

```go
func handleSomeAction(w http.ResponseWriter, r *http.Request) {
    // Do something...

    // Broadcast toast to ALL connected clients
    broadcast.Send(`{"type":"success","message":"Action completed!"}`)

    w.WriteHeader(http.StatusOK)
}
```

---

### Server-Sent Events (SSE)

SSE provides a simple way to push updates from server to browser over HTTP.

**Server side (Go):**

```go
var toastChan = make(chan string, 10)

func handleSSE(w http.ResponseWriter, r *http.Request) {
    // Required headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher := w.(http.Flusher)

    for {
        select {
        case <-r.Context().Done():
            // Client disconnected
            return
        case msg := <-toastChan:
            // SSE format: "event: name\ndata: payload\n\n"
            fmt.Fprintf(w, "event: sse-toast\ndata: %s\n\n", msg)
            flusher.Flush()
        }
    }
}
```

**Client side (HTMX SSE extension):**

```html
<!-- Enable SSE extension and connect -->
<body hx-ext="sse" sse-connect="/sse">

  <!-- Option 1: Auto-swap content on SSE event -->
  <div sse-swap="sse-toast"></div>

  <!-- Option 2: Handle manually with Alpine -->
  <div x-on:htmx:sse-before-message.window="handleSSE($event)">
    <!-- Custom handling -->
  </div>

</body>
```

**SSE message format:**

```
event: sse-toast
data: {"type":"success","message":"Hello!"}

```

Note: Two newlines (`\n\n`) mark the end of a message.

---

## HTMX Attributes Reference

| Attribute | Purpose | Example |
|-----------|---------|---------|
| `hx-get` | GET request | `hx-get="/items"` |
| `hx-post` | POST request | `hx-post="/items"` |
| `hx-put` | PUT request | `hx-put="/items/1"` |
| `hx-delete` | DELETE request | `hx-delete="/items/1"` |
| `hx-target` | Where to put response | `hx-target="#content"` |
| `hx-swap` | How to swap content | `hx-swap="innerHTML"` |
| `hx-swap="none"` | No DOM update | For SSE-only feedback |
| `hx-vals` | Extra values to send | `hx-vals='{"key":"val"}'` |
| `hx-indicator` | Loading indicator | `hx-indicator="#spinner"` |
| `hx-ext="sse"` | Enable SSE extension | On parent element |
| `sse-connect` | SSE endpoint URL | `sse-connect="/sse"` |
| `sse-swap` | Swap on SSE event | `sse-swap="event-name"` |

---

## Alpine.js Attributes Reference

| Attribute | Purpose | Example |
|-----------|---------|---------|
| `x-data` | Component state | `x-data="{ open: false }"` |
| `x-show` | Toggle visibility | `x-show="open"` |
| `x-on:click` / `@click` | Event listener | `@click="open = true"` |
| `x-on:keydown.escape` | Keyboard events | `@keydown.escape="close()"` |
| `x-trap` | Focus trap (needs plugin) | `x-trap="open"` |
| `x-trap.noscroll` | + prevent body scroll | `x-trap.noscroll="open"` |
| `x-transition` | CSS transitions | See examples above |
| `x-init` | Run on init | `x-init="fetch()"` |
| `$watch` | Watch for changes | `$watch('open', v => ...)` |
| `$nextTick` | After DOM update | `$nextTick(() => ...)` |
| `$dispatch` | Dispatch custom event | `$dispatch('my-event')` |
| `$el` | Current element | `$el.closest('dialog')` |

---

## File Structure

```
guide-htmx/
├── main.go              # Server entry point, routes, handlers
├── sse.go               # SSE endpoint handler
├── go.mod               # Go module
├── README.md            # This file
└── templates/
    ├── layout.templ     # Base HTML with CDN deps
    ├── index.templ      # Demo page showing all components
    ├── modal.templ      # Modal examples (4 variants)
    ├── drawer.templ     # Drawer examples (left, right, bottom)
    └── toast.templ      # Toast notification component
```

---

## Running the Demo

```bash
# Install templ if you haven't
go install github.com/a-h/templ/cmd/templ@latest

# Generate Go code from .templ files
templ generate

# Run the server
go run .

# Open browser
open http://localhost:8080
```

---

## Dependencies

All frontend deps loaded via CDN - no npm, no build step:

```html
<!-- DaisyUI + Tailwind CSS 4 -->
<link href="https://cdn.jsdelivr.net/npm/daisyui@5.5.14/daisyui.css" rel="stylesheet">
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>

<!-- HTMX core -->
<script src="https://unpkg.com/htmx.org@2.0.8"></script>

<!-- HTMX SSE extension -->
<script src="https://unpkg.com/htmx-ext-sse@2.2.2/sse.js"></script>

<!-- Alpine.js Focus plugin (for x-trap) - must load before Alpine -->
<script defer src="https://unpkg.com/@alpinejs/focus@3.15.3/dist/cdn.min.js"></script>

<!-- Alpine.js -->
<script defer src="https://unpkg.com/alpinejs@3.15.3/dist/cdn.min.js"></script>
```

**Go dependencies:**
- `github.com/a-h/templ` - Type-safe HTML templates

---

## Why This Stack?

| Tool | Why |
|------|-----|
| **HTMX** | HTML-first interactivity, no JavaScript framework needed |
| **Alpine.js** | Lightweight reactivity for local UI state (focus, transitions) |
| **SSE** | Simpler than WebSockets for server→client updates |
| **Templ** | Type-safe Go templates with great DX |
| **DaisyUI** | Beautiful components without writing CSS |

---

## Key Patterns Demonstrated

### 1. Modal with Focus Trap
```
Button click → showModal() → dispatch event → Alpine sets open=true → x-trap activates
```

### 2. Drawer with Slide Animation
```
Button click → showModal() → dispatch event → x-transition slides in → x-trap locks focus
```

### 3. Toast via SSE
```
Form submit → hx-post → Server handler → toastChan <- msg → SSE sends → Alpine receives → Toast appears
```

### 4. Form with SSE Feedback
```
Form submit → hx-post (swap=none) → Server processes → SSE toast → Modal closes
```

### 5. Inline HTMX Swap
```
Button click → hx-get → Server returns HTML → hx-swap="innerHTML" updates target
```

### 6. Focus Restoration with x-trap
```
Button passes $el via event → Modal stores trigger → On close → $nextTick(() => trigger?.focus())
```

---

## Learn More

- [HTMX Documentation](https://htmx.org/docs/)
- [HTMX SSE Extension](https://htmx.org/extensions/sse/)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Alpine.js Focus Plugin](https://alpinejs.dev/plugins/focus)
- [Templ Documentation](https://templ.guide/)
- [DaisyUI Components](https://daisyui.com/components/)
- [MDN: Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [MDN: HTML dialog element](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/dialog)
