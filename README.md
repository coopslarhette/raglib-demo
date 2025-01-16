# RAGLib Demo

A generative search application built to use [`raglib`](https://github.com/coopslarhette/raglib), featuring document retrieval to ground LLM-powered answer generation.

## Features

- Multi-source document retrieval & ranking (web search via SERP API and Exa.ai,)
- Rich answer formatting via Markdown support
- Syntax highlighting
- Proof of work via citations and source references

## Architecture

The application consists of two main components:

### Backend (Go)
- Built using `raglib` for document retrieval and answer generation
- Combines multiple document sources (SERP API and Exa.ai)
- Handles concurrent document retrieval and processing
- Implements a streaming API using Server-Sent Events
- Uses model facade to easily swap between LLM providers, currently using Anthropic's Haiku

### Frontend (Next.js, TypeScript, CSS Modules)
- Real-time result streaming (SSE) and rendering
- Source document display
- Answer attribution via citations
- Code block and snippet highlighting

## Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- Yarn package manager
- API keys for:
    - SERP API
    - Exa API
    - Anthropic, OpenAI, any other model providers etc

## Getting Started

1. Clone the repository:
```bash
git clone https://github.com/coopslarhette/raglib-demo
cd raglib-demo
```

2. Install backend dependencies:
```bash
go mod download
```

3. Install frontend dependencies:
```bash
cd web-client
yarn install
```

4. Set up environment variables:
```bash
# Backend (.env)
ANTHROPIC_API_KEY=your_anthropic_key
SERP_API_KEY=your_serp_key
EXA_API_KEY=your_exa_key

# Frontend (.env.local)
NEXT_PUBLIC_API_URL=http://localhost:5080
```

5. Start the backend server:
```bash
go run main.go
```

6. Start the frontend development server:
```bash
cd web-client
yarn dev
```

The application will be available at `http://localhost:3000`.

## Project Structure

```
raglib-demo/
├── api/              # Backend API handlers and server setup
    ├── search.go     # Main search handler/backend entry point 
├── web-client/       # Frontend Next.js application
    ├── src/
        ├── app/     # Next.js app router components
            ├── search/  # Search-related components and logic
```