package handler

import (
	"fmt"
	"net/http"
	"strings"
)

// ServiceUnavailableHTML is the template for the "Service Unavailable" page.
const ServiceUnavailableHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Service Unavailable | Teeter</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #0f172a;
            --card-bg: rgba(30, 41, 59, 0.7);
            --text: #f1f5f9;
            --text-muted: #94a3b8;
            --primary: #3b82f6;
            --primary-hover: #2563eb;
            --danger: #ef4444;
            --border: rgba(255, 255, 255, 0.1);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Outfit', sans-serif;
            background-color: var(--bg);
            background-image: 
                radial-gradient(at 0%% 0%%, rgba(59, 130, 246, 0.15) 0px, transparent 50%%),
                radial-gradient(at 100%% 100%%, rgba(147, 51, 234, 0.15) 0px, transparent 50%%);
            color: var(--text);
            height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            overflow: hidden;
        }

        .container {
            max-width: 500px;
            width: 100%%;
            padding: 2rem;
            text-align: center;
            background: var(--card-bg);
            backdrop-filter: blur(12px);
            border: 1px solid var(--border);
            border-radius: 24px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
            animation: fadeIn 0.6s ease-out;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .icon-container {
            width: 80px;
            height: 80px;
            background: rgba(239, 68, 68, 0.1);
            border-radius: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 1.5rem;
            color: var(--danger);
        }

        h1 {
            font-size: 2rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            background: linear-gradient(to right, #fff, #94a3b8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        p {
            color: var(--text-muted);
            line-height: 1.6;
            margin-bottom: 2rem;
        }

        .details {
            background: rgba(15, 23, 42, 0.5);
            border-radius: 12px;
            padding: 1rem;
            text-align: left;
            font-family: monospace;
            font-size: 0.85rem;
            margin-bottom: 2rem;
            border-left: 3px solid var(--primary);
        }

        .details-label {
            color: var(--primary);
            font-weight: bold;
            margin-bottom: 0.25rem;
            display: block;
        }

        .btn {
            display: inline-block;
            background: var(--primary);
            color: white;
            text-decoration: none;
            padding: 0.75rem 2rem;
            border-radius: 12px;
            font-weight: 600;
            transition: all 0.2s ease;
            border: none;
            cursor: pointer;
            width: 100%%;
        }

        .btn:hover {
            background: var(--primary-hover);
            transform: translateY(-2px);
            box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.4);
        }

        .btn:active {
            transform: translateY(0);
        }

        .footer {
            margin-top: 2rem;
            font-size: 0.75rem;
            color: var(--text-muted);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon-container">
            <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
        </div>
        <h1>%s</h1>
        <p>The service you are looking for is currently unavailable or encountered an error.</p>
        
        <div class="details">
            <span class="details-label">Requested Path</span>
            %s
            <br><br>
            <span class="details-label">Error Details</span>
            %s
        </div>

        <button class="btn" onclick="window.location.reload()">Retry Connection</button>

        <div class="footer">
            Teeter Load Balancer &bull; %s
        </div>
    </div>
</body>
</html>`

// ServeErrorPage checks if the request accepts HTML and serves a themed error page.
// Otherwise, it returns a plain text error.
func ServeErrorPage(w http.ResponseWriter, r *http.Request, status int, title string, message string) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(status)
		
		fmt.Fprintf(w, ServiceUnavailableHTML, title, r.URL.Path, message, "System Operational Check Failed")
		return
	}

	// Default to plain text for non-browser requests
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintln(w, message)
}
