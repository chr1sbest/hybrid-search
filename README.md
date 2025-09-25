# Hybrid Search API

This project is a Go application that demonstrates a complete, end-to-end hybrid search system. It combines keyword-based (lexical) and vector-based (semantic) search to provide highly relevant results, using Reciprocal Rank Fusion (RRF) to merge the result sets.

## Getting Started

### Prerequisites

- Go (1.24+)
- Docker
- A [Pinecone API Key](https://app.pinecone.io/)

### Setup

1.  **Set up Pinecone**:

    -   Create a [Pinecone account](https://app.pinecone.io/) and retrieve your API key.
    -   Create a new index with the following specifications:
        -   **Name**: `semantic-search-api` (or a name of your choice).
        -   **Embedding Model**: Use the integrated `llama-text-embed-v2` model.

2.  **Configure Environment**:

    Create a `.env` file in the project root. Add your Pinecone API key and, if you chose a custom index name, update `PINECONE_INDEX_NAME`.

    ```sh
    # .env
    PINECONE_API_KEY="YOUR_API_KEY_HERE"
    PINECONE_INDEX_NAME="semantic-search-api"
    ```

3.  **Start Services**:

    The entire application stack (Go service and Elasticsearch) is managed with Docker Compose.

    ```sh
    docker compose up --build
    ```

    This command will build the application image, start the services, and stream the logs to your terminal. You can run it in the background with `docker compose up -d`.

## Development

This project uses a `Makefile` to streamline common development tasks.

| Command           | Description                                                              |
| ----------------- | ------------------------------------------------------------------------ |
| `make` / `make all` | Tidy modules, generate code, and vendor dependencies.                    |
| `make docker`       | Build and run the application stack using Docker Compose.                |
| `make test`         | Run all unit tests.                                                      |
| `make generate`     | Generate Go code from the OpenAPI specification (`api/spec.yaml`).       |
| `make mocks`        | Generate mock implementations for all service interfaces.                |

## API Documentation

Once the server is running (`make run`), interactive API documentation is available at:

**[http://localhost:8080/docs](http://localhost:8080/docs)**

This documentation is served directly from the application and provides a UI for exploring and interacting with the API endpoints.

---

## System Design & Architecture

The following sections provide a deeper dive into the core concepts and architectural decisions behind this project.

### Key Concepts

#### 1. Hybrid Search

Hybrid search is an approach that merges the results from two distinct types of search engines:

-   **Lexical Search (Keyword-based)**: This is the traditional search method that looks for exact keyword matches. It's fast, precise, and excellent for finding documents that contain specific terms. Implemented here with `Elasticsearch`.
-   **Semantic Search (Vector-based)**: This method searches based on the *meaning* or *intent* behind a query, not just the keywords. It uses vector embeddings to find documents that are contextually similar. Implemented here with `Pinecone`.

By combining both, you get the precision of lexical search and the contextual understanding of semantic search.

#### 2. Reciprocal Rank Fusion (RRF)

Once you have two different sets of search results, you need a way to combine them into a single, coherent list. RRF is a simple and powerful, score-agnostic algorithm for this. It works by looking at the *rank* of a document in each result list, not its absolute score. The formula for a document's RRF score is:

```
RRF_Score = Σ (1 / (k + rank_i))
```

-   `rank_i` is the document's rank in result set `i`.
-   `k` is a constant (we use `60` in this project) that diminishes the impact of lower-ranked items.

#### 3. Document Chunking

Embedding models have a fixed context window. To handle large documents, we first split them into smaller, semantically coherent pieces called **chunks** using a `RecursiveCharacter` text splitter. This improves search relevance by allowing a user's query to match against a focused chunk of text rather than a diluted vector representing the entire document.

#### 4. Pluggable Architecture

The application is designed with a clean separation of concerns using Go interfaces (`EmbeddingClient`, `VectorStore`, `TextStore`, and `Service`). This makes the system extensible, allowing components like `Pinecone` or `Elasticsearch` to be easily swapped with other implementations.

#### 5. Concurrent Operations

To improve performance, the application queries both the text and vector stores **concurrently** using an `errgroup`. This means the total time for the search phase is determined by the *slower* of the two datastores, not the sum of both.

### Project Structure

```
.
├── api/                      # OpenAPI specification and generated code
├── cmd/app/                  # Application entrypoint
├── pkg/
│   ├── chunker/            # Text chunking logic
│   ├── embeddings/         # Embedding client interface and mocks
│   ├── handlers/           # HTTP handlers and tests
│   ├── ranking/            # RRF implementation
│   ├── search/             # Hybrid search orchestration, service, and mocks
│   └── storage/            # Storage interfaces, clients, and mocks
├── .env
├── .gitignore
├── Makefile                  # Development commands
├── go.mod
├── go.sum
└── README.md
```
