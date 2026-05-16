# Kubectl Assistant

An AI-powered `kubectl` plugin that generates and applies Kubernetes manifests using OpenAI GPT or Google Gemini.

## Prerequisites

- A valid Kubernetes configuration file located at `~/.kube/config`

## Authentication

The plugin requires one of the following API keys:

- **[OpenAI API key](https://platform.openai.com/overview)** — Set `OPENAI_API_KEY` environment variable
- **[Google Gemini](https://ai.google.dev/gemini-api/docs)** — Free tier available via [Google AI Studio](https://aistudio.google.com/apikey). Set `GEMINI_API_KEY` instead of `OPENAI_API_KEY`
- **[Local AI](https://github.com/go-skynet/LocalAI)** — See [getting started guide](https://localai.io/basics/getting_started/index.html)

### OpenAI Configuration

For OpenAI or LocalAI, set the following environment variables:

```shell
export OPENAI_API_KEY=your_openai_key
export OPENAI_DEPLOYMENT_NAME=gpt-3.5-turbo-1106
export OPENAI_ENDPOINT=your_endpoint  # Optional: e.g., "http://localhost:8080/v1"
```

If `OPENAI_ENDPOINT` is set, the plugin will use that endpoint. Otherwise, it defaults to the OpenAI API.

### Gemini Configuration

Set only the `GEMINI_API_KEY` environment variable. Do not set `OPENAI_API_KEY` when using Gemini.

> **Note:** The `--use-k8s-api` (function-calling) feature is not supported with Gemini. Use the default prompt-based generation instead.

## Quick Start with Gemini

### Step 1: Install Go

Download and install Go 1.20 or later from [go.dev/dl](https://go.dev/dl/).

### Step 2: Build the Plugin

```powershell
cd "D:\GO LAMG\kubernetes-assistant-go-master"
go build -o kubectl-assistant.exe .
```

### Step 3: Configure Your API Key

Obtain a free API key from [Google AI Studio](https://aistudio.google.com/apikey).

**PowerShell (current session):**
```powershell
$env:GEMINI_API_KEY = "YOUR_GEMINI_API_KEY_HERE"
```

**Command Prompt:**
```cmd
set GEMINI_API_KEY=YOUR_GEMINI_API_KEY_HERE
```

### Step 4: Run the Assistant

```powershell
.\kubectl-assistant.exe "create an nginx deployment with 2 replicas"
```

Alternatively, run without building:
```powershell
go run . "create an nginx deployment with 2 replicas"
```

### Optional: Configure Model

The default model is `gemini-2.5-flash`. To use a different model:

```powershell
$env:GEMINI_MODEL = "gemini-2.0-flash"
```

## HTTP API

The assistant can run as an HTTP server, enabling integration with web applications.

### Starting the Server

```powershell
.\kubectl-assistant.exe serve --port 8080
```

The default port is 8080. Use the `-p` flag to specify a different port.

### API Endpoint

**POST** `http://localhost:8080/generate`

**Request Body:**
```json
{
  "prompt": "create an nginx deployment with 2 replicas"
}
```

**Response:**
```json
{
  "manifest": "apiVersion: apps/v1\nkind: Deployment\n..."
}
```

### Client Examples

**cURL:**
```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{"prompt": "create an nginx deployment with 2 replicas"}'
```

**JavaScript/React:**
```javascript
const response = await fetch('http://localhost:8080/generate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ prompt: 'create an nginx deployment with 2 replicas' })
});
const { manifest } = await response.json();
```

CORS is enabled for browser-based clients.

## Configuration Flags

| Flag | Environment Variable | Description | Default |
|------|---------------------|-------------|---------|
| `--require-confirmation` | `REQUIRE_CONFIRMATION` | Prompt for confirmation before applying manifests | `true` |
| `--temperature` | `TEMPERATURE` | Controls randomness (0.0 = deterministic, 1.0 = creative) | `0` |
| `--use-k8s-api` | `USE_K8S_API` | Use Kubernetes OpenAPI Spec for accurate completions including CRDs | `false` |
| `--k8s-openapi-url` | `K8S_OPENAPI_URL` | Custom Kubernetes OpenAPI Spec URL | Kubernetes API Server |

### Notes

- The `--use-k8s-api` flag requires a model that supports [function calling](https://openai.com/blog/function-calling-and-other-api-updates) (GPT-3.5 0613 or later). This option is recommended for accuracy and completeness.
- To generate a custom OpenAPI spec that includes CRDs: `kubectl get --raw /openapi/v2 > swagger.json`

## Usage Examples

### Create a Deployment

```bash
$ go run main.go "create an nginx deployment with 3 replicas"
✨ Attempting to apply the following manifest:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Reprompt/Apply/Don't Apply]:
+   Reprompt
  ▸ Apply
    Don't Apply
```

### Reprompt to Refine Your Prompt

```bash
...
Reprompt: update to 5 replicas and port 8080
✨ Attempting to apply the following manifest:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 5
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 8080
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Reprompt/Apply/Don't Apply]:
+   Reprompt
  ▸ Apply
    Don't Apply
```

### Create Multiple Resources

```bash
$ go run main.go "create a foo namespace then create nginx pod in that namespace"
✨ Attempting to apply the following manifest:
apiVersion: v1
kind: Namespace
metadata:
  name: foo
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: foo
spec:
  containers:
  - name: nginx
    image: nginx:latest
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Reprompt/Apply/Don't Apply]:
+   Reprompt
  ▸ Apply
    Don't Apply
```

### Skip Confirmation

```bash
$ go run main.go "create a service with type LoadBalancer with selector as 'app:nginx'" --require-confirmation=false
✨ Attempting to apply the following manifest:
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

---

**Note:** The plugin generates complete manifests based on your prompt and does not track the current cluster state.
