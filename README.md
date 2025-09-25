# Hybrid Search with Reciprocal Rank Fusion (RRF)

This project is a Go application that demonstrates a complete, end-to-end **Hybrid Search** system. It combines the strengths of traditional keyword-based search (lexical search) with modern vector-based search (semantic search) to provide more relevant and context-aware results.

This is a common pattern in modern search systems and a great topic for a system design interview.

## Key System Design Concepts

### 1. Hybrid Search

**What is it?**
Hybrid search is an approach that merges the results from two distinct types of search engines:

- **Lexical Search (Keyword-based)**: This is the traditional search method that looks for exact keyword matches. It's fast, precise, and excellent for finding documents that contain specific terms.
  - *Implementation*: `Elasticsearch`

- **Semantic Search (Vector-based)**: This method searches based on the *meaning* or *intent* behind a query, not just the keywords. It uses vector embeddings to find documents that are contextually similar.
  - *Implementation*: `Pinecone`

**Why use it?**
By combining both, you get the best of both worlds: the precision of lexical search and the contextual understanding of semantic search. A query for "how to fix a flat tire" will match documents with those exact words (thanks to Elasticsearch) as well as documents that talk about "patching a rubber tube" or "inflating a wheel" (thanks to Pinecone).

### 2. Reciprocal Rank Fusion (RRF)

**What is it?**
Once you have two different sets of search results, you need a way to combine them into a single, coherent list. RRF is a simple and powerful algorithm for this.

It works by looking at the *rank* of a document in each result list, not its score. The formula for a document's RRF score is:

```
RRF_Score = Σ (1 / (k + rank_i))
```

- `rank_i` is the document's rank in result set `i`.
- `k` is a constant (we use `60` in this project) that diminishes the impact of lower-ranked items.

**Why use it?**
- **Score-Agnostic**: Different search systems produce scores on different scales. RRF elegantly sidesteps the need to normalize these scores.
- **Rewards Prominence**: It gives significant weight to documents that appear at the top of *any* list, assuming that a top result from any system is likely to be highly relevant.

### 3. Document Chunking for Large Texts

**What is it?**
Embedding models have a fixed context window, meaning they can only process a certain amount of text at once. To handle large documents, we first split them into smaller, semantically coherent pieces called **chunks**.

- **Implementation**: We use the `RecursiveCharacter` text splitter from the `langchaingo` library, a robust implementation of a common and effective chunking algorithm.

**Why use it?**
Chunking improves search relevance. Instead of a single, diluted vector for a large document, we get multiple, focused vectors for each chunk. A specific user query for "how to change a headlight bulb" will have a much stronger match with a specific chunk about headlights than with a general vector for an entire car maintenance manual.

### 4. Pluggable Architecture with Interfaces

The application is designed with a clean separation of concerns using Go interfaces (`EmbeddingClient`, `VectorStore`, and `TextStore`).

- `search_service.go` orchestrates the search, but it doesn't know or care *which* vector database or text engine is being used. It only knows about the interfaces.
- This makes the system **extensible**. You can easily swap `Pinecone` for another vector DB like `Weaviate`, `Elasticsearch` for `OpenSearch`, or the internal embedding model for an external one like `OpenAI` by simply creating a new client that satisfies the appropriate interface.

### 5. Concurrent Operations

To improve performance, the application queries both Elasticsearch and Pinecone **concurrently** using an `errgroup`. This means the total time for the search phase is determined by the *slower* of the two datastores, not the sum of both.

## Project Structure

```
.
├── cmd/app/
│   └── main.go              # Application entrypoint
├── pkg/
│   ├── chunker/
│   │   └── chunker.go       # Text chunking logic
│   ├── embeddings/
│   │   └── embeddings.go    # Embedding client interface
│   ├── handlers/
│   │   └── handlers.go      # HTTP handlers
│   ├── ranking/
│   │   └── ranking.go       # RRF implementation
│   ├── search/
│   │   └── service.go       # Hybrid search orchestration
│   └── storage/
│       ├── elasticsearch.go # Elasticsearch client
│       ├── pinecone.go      # Pinecone client
│       ├── stores.go        # Store interfaces
│       └── models.go        # Core data models
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## How to Run This Project

### Prerequisites

- **Go** (version 1.18+)
- **Docker**
- A **Pinecone API Key**

### 1. Start Elasticsearch

This project uses Docker to run an Elasticsearch instance locally. The `xpack.security.enabled=false` flag is for convenience in a local development environment.

```sh
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "xpack.security.enabled=false" docker.elastic.co/elasticsearch/elasticsearch:8.14.1
```

### 2. Create a .env File

Create a `.env` file in the root of the project and add your Pinecone API key. You can get one from the [Pinecone console](https://app.pinecone.io/).

```
# .env
PINECONE_API_KEY="YOUR_API_KEY_HERE"
```

### 3. Run the Go Application

```sh
go mod tidy
go run ./cmd/app
```

The server will start on `http://localhost:8080`.

### 4. Use the API

**Store a Document:**

```sh
curl -X POST http://localhost:8080/store \
-H "Content-Type: application/json" \
-d '{"text": "Reciprocal Rank Fusion is a powerful way to combine search results."}'
```

**Query for a Document:**

```sh
curl -X GET "http://localhost:8080/query?q=search%20results"
```
