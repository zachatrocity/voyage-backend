<div align="center">
  
# Voyage (Backend)

</div>

## Overview

This repository contains the backend for Voyage, a self-hosted travel plan aggregator. The backend is responsible for processing emails using notmuch and mbsync, and providing a REST API for accessing and organizing travel-related information.

## Features

- **Email Integration**: Automatically scans and processes travel-related emails (WIP)
- **Trip Organization**: Provides data and tagging capabilities for organizing related travel items into coherent trips (WIP)
- **Self-Hosted**: Full control over your data with Docker-based deployment

## Architecture

Voyage consists of several key components:

1. **Email Fetching Service**: Uses mbsync to retrieve emails from your accounts
2. **Notmuch Integration**: Indexes and tags emails for efficient searching
3. **REST API**: Simple Go API for searching the notmuch database

## API Usage

Search the notmuch database:
```
GET /api/v1/search?q=airbnb
GET /api/v1/emails/{message_id}
GET /api/v1/search?q=subject:flight` (default, sorts by newest_first)
GET /api/v1/search?q=subject:flight&sort=oldest_first`
GET /api/v1/search?q=tag:my-trip`
```
