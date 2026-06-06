# LLM Inference Server

A lightweight Go-based reverse proxy that exposes an OpenAI-compatible API, forwarding requests to a locally running llama.cpp server.

---

## Architecture

```
Client (curl / OpenAI SDK / any HTTP client)
        │
        │  POST /v1/chat/completions
        ▼
┌──────────────────────────┐
│  Go Inference Gateway    │  :9090
│  (Your Reverse Proxy)    │
└──────────────────────────┘
        │
        │  Forwards to
        ▼
┌──────────────────────────┐
│  llama-server (llama.cpp)│  :8080
│  Running Llama 3.2 8B    │
└──────────────────────────┘
```

---

## Prerequisites

- Linux / macOS / Windows
- Go 1.20+
- Git
- At least 8GB RAM (16GB recommended for 8B model)
- 6–8 GB free disk space for the model

---

## Part 1 — Setting Up llama.cpp

### Step 1: Install build dependencies

```bash
sudo apt update
sudo apt install -y \
  build-essential \
  cmake \
  git \
  curl \
  wget \
  libcurl4-openssl-dev
```

### Step 2: Clone llama.cpp

```bash
git clone https://github.com/ggerganov/llama.cpp.git
cd llama.cpp
```

### Step 3: Build llama.cpp

```bash
cmake -B build
cmake --build build --config Release -j$(nproc)
```

After a successful build, binaries will be in `build/bin/`:

```bash
ls build/bin/
# llama-cli  llama-server  llama-quantize  ...
```

### Step 4: Download the Llama 3.2 8B GGUF model

You need a GGUF-format model. The recommended quantization for CPU-only setups is **Q4_K_M** (good balance of quality vs memory).

**Option A — Using Hugging Face CLI:**

```bash
pip install huggingface-hub

huggingface-cli download \
  bartowski/Meta-Llama-3.2-8B-Instruct-GGUF \
  Meta-Llama-3.2-8B-Instruct-Q4_K_M.gguf \
  --local-dir ./models
```

**Option B — Direct wget (if you have the URL):**

```bash
mkdir -p models
wget -O models/llama-3.2-8b-q4.gguf <YOUR_GGUF_URL>
```

> **Note:** You may need to accept Meta's license on Hugging Face before downloading.

### Step 5: Verify the model loads correctly

Run a quick CLI test before starting the server:

```bash
./build/bin/llama-cli \
  --model models/Meta-Llama-3.2-8B-Instruct-Q4_K_M.gguf \
  --prompt "Hello, who are you?" \
  --n-predict 64 \
  --log-disable
```

You should see a response generated token by token. If this works, the model is good.

### Step 6: Start the llama-server

```bash
./build/bin/llama-server \
  --model models/Meta-Llama-3.2-8B-Instruct-Q4_K_M.gguf \
  --host 0.0.0.0 \
  --port 8081 \
  --ctx-size 4096 \
  --n-predict 512 \
  --parallel 4 \
  --log-disable
```

**Flag reference:**

| Flag | Description |
|------|-------------|
| `--host 0.0.0.0` | Accept connections from all interfaces |
| `--port 8081` | Port to listen on |
| `--ctx-size 4096` | Context window size (tokens) |
| `--n-predict 512` | Max tokens to generate per request |
| `--parallel 4` | Max concurrent generation slots |
| `--log-disable` | Suppress verbose logs |

### Step 7: Verify llama-server is working

```bash
curl http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3.2-8b",
    "messages": [
      {"role": "user", "content": "Say hello in one sentence."}
    ],
    "stream": false,
    "max_tokens": 64
  }'
```

Expected: a JSON response with `choices[0].message.content` populated.

---

## Part 2 — Running the Go Inference Gateway

### Step 1: Clone this repository

```bash
git clone https://github.com/amithathreya/inference_layer.git
cd inference_layer
```

### Step 2: Configure the backend URL

Edit the `main.go` file to set the llama-server URL if it's not running on the default `localhost:8080`:

```go
llamaBackend := backend.NewLlamaBackend("http://localhost:8080")
```

### Step 3: Build and run

```bash
go build -o inference-gateway
./inference-gateway
```

The gateway starts on port **9090** by default.

### Step 4: Test through the gateway

```bash
curl http://localhost:9090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3.2-8b",
    "messages": [
      {"role": "user", "content": "What is the capital of France?"}
    ],
    "stream": false
  }'
```

### Step 5: Test with the OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:9090/v1",
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="llama-3.2-8b",
    messages=[{"role": "user", "content": "Explain Newton's first law simply."}]
)

print(response.choices[0].message.content)
```

---

## Troubleshooting

| Problem | Likely cause | Fix |
|--------|-------------|-----|
| `cmake` fails | Missing build tools | Re-run `apt install build-essential cmake` |
| Model download fails | Not logged in to HuggingFace | Run `huggingface-cli login` |
| llama-server OOM crashes | Not enough RAM | Use a smaller quantization: `Q2_K` or `Q3_K_M` |
| Gateway returns 502 | llama-server not running | Start llama-server first on port 8081 |
| Slow generation | CPU-only, no GPU | Expected; 8B Q4 does ~5–10 tok/sec on modern CPU |

---

## Project Structure

```
inference_layer/
├── main.go                    # Entry point, starts the gateway server
├── go.mod                     # Go module definition
├── api/
│   ├── handler.go             # HTTP request handler for /v1/chat/completions
│   └── types.go               # Request/response data structures
├── backend/
│   └── llamacpp.go            # Backend logic to forward to llama.cpp server
└── README.md
```

