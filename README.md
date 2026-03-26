# VIDI (Validate, Integrate, Data, Input)

A lightweight, standard-library-only Go middleware for processing HTML forms. VIDI converts form submissions into structured data (JSON, XML, or SQL) and persists them to various backends (Filesystem, In-Memory, or PostgreSQL-ready interfaces). It also supports GET requests as dynamic queries against stored data.

## Features

- 🛠️ **Zero External Dependencies**: Built entirely with the Go standard library.
- 💾 **Pluggable Storage**: Supports Filesystem (JSON files) and In-Memory backends out of the box.
- 🔄 **Multi-Format Output**: Automatically converts data to JSON, XML, or SQL `INSERT` statements.
- 🔍 **Query Engine**: GET requests with query parameters act as filters against stored data.
- 🧩 **Modular Design**: Can be layered with VENI (frontend components) and VICI (auth/context).
- 🚀 **Drop-in Middleware**: Wraps `http.FileServer` seamlessly.

## Quick Start

### Installation

```bash
git clone https://github.com/Emperor42/vidi.git
cd vidi
make run